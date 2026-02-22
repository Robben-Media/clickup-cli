package clickup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/builtbyrobben/clickup-cli/internal/api"
)

var (
	errIDRequired          = errors.New("id is required")
	errNameRequired        = errors.New("name is required")
	errTextRequired        = errors.New("comment text is required")
	errWorkspaceIDRequired = errors.New("workspace ID required for v3 API; set CLICKUP_WORKSPACE_ID or use --workspace flag")
	errSourceTasksRequired = errors.New("at least one source task ID is required")
)

const defaultBaseURL = "https://api.clickup.com/api"

// Client wraps the API client with ClickUp-specific methods.
type Client struct {
	*api.Client
	workspaceID string
}

// ClientOption is a functional option for configuring the Client.
type ClientOption func(*Client)

// WithWorkspaceID sets the workspace ID for v3 API calls.
func WithWorkspaceID(workspaceID string) ClientOption {
	return func(c *Client) {
		c.workspaceID = workspaceID
	}
}

// NewClient creates a new ClickUp API client.
func NewClient(apiKey string, opts ...ClientOption) *Client {
	c := &Client{
		Client: api.NewClient(apiKey,
			api.WithBaseURL(defaultBaseURL),
			api.WithUserAgent("clickup-cli/1.0"),
		),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// v3Path builds a v3 API path with the workspace ID prefix.
// Returns an error if workspace ID is not configured.
func (c *Client) v3Path(path string) (string, error) {
	if c.workspaceID == "" {
		return "", errWorkspaceIDRequired
	}

	return fmt.Sprintf("/v3/workspaces/%s%s", c.workspaceID, path), nil
}

// Tasks provides methods for the Tasks API.
func (c *Client) Tasks() *TasksService {
	return &TasksService{client: c}
}

// Spaces provides methods for the Spaces API.
func (c *Client) Spaces() *SpacesService {
	return &SpacesService{client: c}
}

// Lists provides methods for the Lists API.
func (c *Client) Lists() *ListsService {
	return &ListsService{client: c}
}

// Members provides methods for the Members API.
func (c *Client) Members() *MembersService {
	return &MembersService{client: c}
}

// Comments provides methods for the Comments API.
func (c *Client) Comments() *CommentsService {
	return &CommentsService{client: c}
}

// Time provides methods for the Time Tracking API.
func (c *Client) Time() *TimeService {
	return &TimeService{client: c}
}

// Folders provides methods for the Folders API.
func (c *Client) Folders() *FoldersService {
	return &FoldersService{client: c}
}

// --- TasksService ---

// TasksService handles task operations.
type TasksService struct {
	client *Client
}

// List returns tasks for a given list.
func (s *TasksService) List(ctx context.Context, listID string, status, assignee string) (*TasksListResponse, error) {
	if listID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/list/%s/task?include_closed=true", listID)

	if status != "" {
		path += fmt.Sprintf("&statuses[]=%s", url.QueryEscape(status))
	}

	if assignee != "" {
		path += fmt.Sprintf("&assignees[]=%s", url.QueryEscape(assignee))
	}

	var result TasksListResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}

	return &result, nil
}

// Get returns a task by ID.
func (s *TasksService) Get(ctx context.Context, taskID string) (*Task, error) {
	if taskID == "" {
		return nil, errIDRequired
	}

	var result Task

	path := fmt.Sprintf("/v2/task/%s", taskID)
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}

	return &result, nil
}

// Create creates a new task in a list.
func (s *TasksService) Create(ctx context.Context, listID string, req CreateTaskRequest) (*Task, error) {
	if listID == "" {
		return nil, errIDRequired
	}

	if req.Name == "" {
		return nil, errNameRequired
	}

	var result Task

	path := fmt.Sprintf("/v2/list/%s/task", listID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}

	return &result, nil
}

// Update updates a task.
func (s *TasksService) Update(ctx context.Context, taskID string, req UpdateTaskRequest) (*Task, error) {
	if taskID == "" {
		return nil, errIDRequired
	}

	var result Task

	path := fmt.Sprintf("/v2/task/%s", taskID)
	if err := s.client.Put(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}

	return &result, nil
}

// Delete deletes a task.
func (s *TasksService) Delete(ctx context.Context, taskID string) error {
	if taskID == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/task/%s", taskID)
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	return nil
}

// Search returns filtered tasks across a workspace/team.
func (s *TasksService) Search(ctx context.Context, teamID string, params FilteredTeamTasksParams) (*FilteredTeamTasksResponse, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/team/%s/task", teamID)

	// Build query string manually for array parameters
	values := url.Values{}

	if params.Page > 0 {
		values.Set("page", fmt.Sprintf("%d", params.Page))
	}

	if params.OrderBy != "" {
		values.Set("order_by", params.OrderBy)
	}

	if params.Reverse {
		values.Set("reverse", "true")
	}

	if params.Subtasks {
		values.Set("subtasks", "true")
	}

	if params.IncludeClosed {
		values.Set("include_closed", "true")
	}

	for _, status := range params.Statuses {
		values.Add("statuses[]", status)
	}

	for _, assignee := range params.Assignees {
		values.Add("assignees[]", fmt.Sprintf("%d", assignee))
	}

	for _, tag := range params.Tags {
		values.Add("tags[]", tag)
	}

	if params.DueDateGt > 0 {
		values.Set("due_date_gt", fmt.Sprintf("%d", params.DueDateGt))
	}

	if params.DueDateLt > 0 {
		values.Set("due_date_lt", fmt.Sprintf("%d", params.DueDateLt))
	}

	if params.DateCreatedGt > 0 {
		values.Set("date_created_gt", fmt.Sprintf("%d", params.DateCreatedGt))
	}

	if params.DateCreatedLt > 0 {
		values.Set("date_created_lt", fmt.Sprintf("%d", params.DateCreatedLt))
	}

	if params.DateUpdatedGt > 0 {
		values.Set("date_updated_gt", fmt.Sprintf("%d", params.DateUpdatedGt))
	}

	if params.DateUpdatedLt > 0 {
		values.Set("date_updated_lt", fmt.Sprintf("%d", params.DateUpdatedLt))
	}

	if len(values) > 0 {
		path += "?" + values.Encode()
	}

	var result FilteredTeamTasksResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("search tasks: %w", err)
	}

	return &result, nil
}

// TimeInStatus returns the time-in-status data for a single task.
func (s *TasksService) TimeInStatus(ctx context.Context, taskID string) (*TimeInStatusResponse, error) {
	if taskID == "" {
		return nil, errIDRequired
	}

	var result TimeInStatusResponse

	path := fmt.Sprintf("/v2/task/%s/time_in_status", taskID)
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get time in status: %w", err)
	}

	return &result, nil
}

// BulkTimeInStatus returns time-in-status data for multiple tasks.
func (s *TasksService) BulkTimeInStatus(ctx context.Context, taskIDs []string) (BulkTimeInStatusResponse, error) {
	if len(taskIDs) == 0 {
		return nil, errIDRequired
	}

	// Build query string with task_ids
	values := url.Values{}
	for _, id := range taskIDs {
		values.Add("task_ids", id)
	}

	path := "/v2/task/bulk_time_in_status/task_ids?" + values.Encode()

	var result BulkTimeInStatusResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get bulk time in status: %w", err)
	}

	return result, nil
}

// Merge merges source tasks into the target task.
func (s *TasksService) Merge(ctx context.Context, targetTaskID string, sourceTaskIDs []string) (*Task, error) {
	if targetTaskID == "" {
		return nil, errIDRequired
	}

	if len(sourceTaskIDs) == 0 {
		return nil, errSourceTasksRequired
	}

	req := MergeTasksRequest{MergedTaskIDs: sourceTaskIDs}

	var result Task

	path := fmt.Sprintf("/v2/task/%s/merge", targetTaskID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("merge tasks: %w", err)
	}

	return &result, nil
}

// Move moves a task to a new list (v3 API).
func (s *TasksService) Move(ctx context.Context, taskID, listID string) error {
	if taskID == "" || listID == "" {
		return errIDRequired
	}

	path, err := s.client.v3Path(fmt.Sprintf("/tasks/%s/home_list/%s", taskID, listID))
	if err != nil {
		return fmt.Errorf("move task: %w", err)
	}

	if err := s.client.Put(ctx, path, nil, nil); err != nil {
		return fmt.Errorf("move task: %w", err)
	}

	return nil
}

// FromTemplate creates a task from a template.
func (s *TasksService) FromTemplate(ctx context.Context, listID, templateID string, req CreateTaskFromTemplateRequest) (*Task, error) {
	if listID == "" {
		return nil, errIDRequired
	}

	if templateID == "" {
		return nil, errIDRequired
	}

	var result Task

	path := fmt.Sprintf("/v2/list/%s/taskTemplate/%s", listID, templateID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create task from template: %w", err)
	}

	return &result, nil
}

// --- SpacesService ---

// SpacesService handles space operations.
type SpacesService struct {
	client *Client
}

// List returns all spaces for a team.
func (s *SpacesService) List(ctx context.Context, teamID string) (*SpacesListResponse, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/team/%s/space", teamID)

	var result SpacesListResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list spaces: %w", err)
	}

	return &result, nil
}

// Get returns a space by ID with full details including statuses and features.
func (s *SpacesService) Get(ctx context.Context, spaceID string) (*SpaceDetail, error) {
	if spaceID == "" {
		return nil, errIDRequired
	}

	var result SpaceDetail

	path := fmt.Sprintf("/v2/space/%s", spaceID)
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get space: %w", err)
	}

	return &result, nil
}

// Create creates a new space in a team.
func (s *SpacesService) Create(ctx context.Context, teamID string, req CreateSpaceRequest) (*SpaceDetail, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	if req.Name == "" {
		return nil, errNameRequired
	}

	var result SpaceDetail

	path := fmt.Sprintf("/v2/team/%s/space", teamID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create space: %w", err)
	}

	return &result, nil
}

// Update updates a space.
func (s *SpacesService) Update(ctx context.Context, spaceID string, req UpdateSpaceRequest) (*SpaceDetail, error) {
	if spaceID == "" {
		return nil, errIDRequired
	}

	var result SpaceDetail

	path := fmt.Sprintf("/v2/space/%s", spaceID)
	if err := s.client.Put(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("update space: %w", err)
	}

	return &result, nil
}

// Delete deletes a space.
func (s *SpacesService) Delete(ctx context.Context, spaceID string) error {
	if spaceID == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/space/%s", spaceID)
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("delete space: %w", err)
	}

	return nil
}

// --- ListsService ---

// ListsService handles list operations.
type ListsService struct {
	client *Client
}

// ListByFolder returns lists in a folder.
func (s *ListsService) ListByFolder(ctx context.Context, folderID string) (*ListsListResponse, error) {
	if folderID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/folder/%s/list", folderID)

	var result ListsListResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list lists by folder: %w", err)
	}

	return &result, nil
}

// ListFolderless returns folderless lists in a space.
func (s *ListsService) ListFolderless(ctx context.Context, spaceID string) (*FolderlessListsResponse, error) {
	if spaceID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/space/%s/list", spaceID)

	var result FolderlessListsResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list folderless lists: %w", err)
	}

	return &result, nil
}

// ListFolders returns folders in a space (used to discover lists).
func (s *ListsService) ListFolders(ctx context.Context, spaceID string) (*FoldersListResponse, error) {
	if spaceID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/space/%s/folder", spaceID)

	var result FoldersListResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list folders: %w", err)
	}

	return &result, nil
}

// Get returns a list by ID with full details.
func (s *ListsService) Get(ctx context.Context, listID string) (*ListDetail, error) {
	if listID == "" {
		return nil, errIDRequired
	}

	var result ListDetail

	path := fmt.Sprintf("/v2/list/%s", listID)
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get list: %w", err)
	}

	return &result, nil
}

// CreateInFolder creates a new list in a folder.
func (s *ListsService) CreateInFolder(ctx context.Context, folderID string, req CreateListRequest) (*ListDetail, error) {
	if folderID == "" {
		return nil, errIDRequired
	}

	if req.Name == "" {
		return nil, errNameRequired
	}

	var result ListDetail

	path := fmt.Sprintf("/v2/folder/%s/list", folderID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create list: %w", err)
	}

	return &result, nil
}

// CreateFolderless creates a new folderless list in a space.
func (s *ListsService) CreateFolderless(ctx context.Context, spaceID string, req CreateListRequest) (*ListDetail, error) {
	if spaceID == "" {
		return nil, errIDRequired
	}

	if req.Name == "" {
		return nil, errNameRequired
	}

	var result ListDetail

	path := fmt.Sprintf("/v2/space/%s/list", spaceID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create folderless list: %w", err)
	}

	return &result, nil
}

// Update updates a list.
func (s *ListsService) Update(ctx context.Context, listID string, req UpdateListRequest) (*ListDetail, error) {
	if listID == "" {
		return nil, errIDRequired
	}

	var result ListDetail

	path := fmt.Sprintf("/v2/list/%s", listID)
	if err := s.client.Put(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("update list: %w", err)
	}

	return &result, nil
}

// Delete deletes a list.
func (s *ListsService) Delete(ctx context.Context, listID string) error {
	if listID == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/list/%s", listID)
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("delete list: %w", err)
	}

	return nil
}

// AddTask adds a task to a list (enables "Tasks in Multiple Lists" feature).
func (s *ListsService) AddTask(ctx context.Context, listID, taskID string) error {
	if listID == "" {
		return errIDRequired
	}

	if taskID == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/list/%s/task/%s", listID, taskID)
	if err := s.client.Post(ctx, path, nil, nil); err != nil {
		return fmt.Errorf("add task to list: %w", err)
	}

	return nil
}

// RemoveTask removes a task from a list.
func (s *ListsService) RemoveTask(ctx context.Context, listID, taskID string) error {
	if listID == "" {
		return errIDRequired
	}

	if taskID == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/list/%s/task/%s", listID, taskID)
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("remove task from list: %w", err)
	}

	return nil
}

// FromTemplateInFolder creates a list from a template in a folder.
func (s *ListsService) FromTemplateInFolder(ctx context.Context, folderID, templateID string, req CreateListFromTemplateRequest) (*ListDetail, error) {
	if folderID == "" {
		return nil, errIDRequired
	}

	if templateID == "" {
		return nil, errIDRequired
	}

	var result ListDetail

	path := fmt.Sprintf("/v2/folder/%s/list_template/%s", folderID, templateID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create list from template: %w", err)
	}

	return &result, nil
}

// FromTemplateInSpace creates a folderless list from a template in a space.
func (s *ListsService) FromTemplateInSpace(ctx context.Context, spaceID, templateID string, req CreateListFromTemplateRequest) (*ListDetail, error) {
	if spaceID == "" {
		return nil, errIDRequired
	}

	if templateID == "" {
		return nil, errIDRequired
	}

	var result ListDetail

	path := fmt.Sprintf("/v2/space/%s/list_template/%s", spaceID, templateID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create list from template: %w", err)
	}

	return &result, nil
}

// --- MembersService ---

// MembersService handles team member operations.
type MembersService struct {
	client *Client
}

// List returns all members of a team (workspace).
func (s *MembersService) List(ctx context.Context, teamID string) (*MembersListResponse, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	// The ClickUp v2 API returns members as part of the team endpoint
	path := fmt.Sprintf("/v2/team/%s", teamID)

	// The /team/{team_id} endpoint returns a team object with members
	var result struct {
		Team struct {
			Members []Member `json:"members"`
		} `json:"team"`
	}

	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}

	return &MembersListResponse{Members: result.Team.Members}, nil
}

// --- CommentsService ---

// CommentsService handles comment operations.
type CommentsService struct {
	client *Client
}

// List returns all comments for a task.
func (s *CommentsService) List(ctx context.Context, taskID string) (*CommentsListResponse, error) {
	if taskID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/task/%s/comment", taskID)

	var result CommentsListResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}

	return &result, nil
}

// Add creates a new comment on a task.
func (s *CommentsService) Add(ctx context.Context, taskID string, text string) (*Comment, error) {
	if taskID == "" {
		return nil, errIDRequired
	}

	if text == "" {
		return nil, errTextRequired
	}

	req := CreateCommentRequest{CommentText: text}

	// ClickUp returns the comment ID as a number in a wrapper
	var result struct {
		ID json.Number `json:"id"`
	}

	path := fmt.Sprintf("/v2/task/%s/comment", taskID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("add comment: %w", err)
	}

	return &Comment{ID: result.ID, Text: text}, nil
}

// Delete removes a comment.
func (s *CommentsService) Delete(ctx context.Context, commentID string) error {
	if commentID == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/comment/%s", commentID)
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}

	return nil
}

// Update modifies an existing comment.
func (s *CommentsService) Update(ctx context.Context, commentID string, req UpdateCommentRequest) error {
	if commentID == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/comment/%s", commentID)
	if err := s.client.Put(ctx, path, req, nil); err != nil {
		return fmt.Errorf("update comment: %w", err)
	}

	return nil
}

// Replies returns all threaded replies to a comment.
func (s *CommentsService) Replies(ctx context.Context, commentID string) (*CommentsListResponse, error) {
	if commentID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/comment/%s/reply", commentID)

	var result CommentsListResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get comment replies: %w", err)
	}

	return &result, nil
}

// Reply creates a threaded reply to a comment.
func (s *CommentsService) Reply(ctx context.Context, commentID string, text string, assignee int) (*Comment, error) {
	if commentID == "" {
		return nil, errIDRequired
	}

	if text == "" {
		return nil, errTextRequired
	}

	req := struct {
		CommentText string `json:"comment_text"`
		Assignee    int    `json:"assignee,omitempty"`
	}{
		CommentText: text,
		Assignee:    assignee,
	}

	var result struct {
		ID json.Number `json:"id"`
	}

	path := fmt.Sprintf("/v2/comment/%s/reply", commentID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create comment reply: %w", err)
	}

	return &Comment{ID: result.ID, Text: text}, nil
}

// ListComments returns all comments for a list.
func (s *CommentsService) ListComments(ctx context.Context, listID string) (*CommentsListResponse, error) {
	if listID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/list/%s/comment", listID)

	var result CommentsListResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list comments on list: %w", err)
	}

	return &result, nil
}

// AddList creates a new comment on a list.
func (s *CommentsService) AddList(ctx context.Context, listID string, req CreateListCommentRequest) (*Comment, error) {
	if listID == "" {
		return nil, errIDRequired
	}

	if req.CommentText == "" {
		return nil, errTextRequired
	}

	var result struct {
		ID json.Number `json:"id"`
	}

	path := fmt.Sprintf("/v2/list/%s/comment", listID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("add list comment: %w", err)
	}

	return &Comment{ID: result.ID, Text: req.CommentText}, nil
}

// ViewComments returns all comments for a view with optional pagination.
func (s *CommentsService) ViewComments(ctx context.Context, viewID string, params ViewCommentsParams) (*CommentsListResponse, error) {
	if viewID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/view/%s/comment", viewID)

	if params.Start > 0 || params.StartID != "" {
		values := url.Values{}
		if params.Start > 0 {
			values.Set("start", fmt.Sprintf("%d", params.Start))
		}

		if params.StartID != "" {
			values.Set("start_id", params.StartID)
		}

		path = path + "?" + values.Encode()
	}

	var result CommentsListResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list view comments: %w", err)
	}

	return &result, nil
}

// AddView creates a new comment on a view.
func (s *CommentsService) AddView(ctx context.Context, viewID string, req CreateViewCommentRequest) (*Comment, error) {
	if viewID == "" {
		return nil, errIDRequired
	}

	if req.CommentText == "" {
		return nil, errTextRequired
	}

	var result struct {
		ID json.Number `json:"id"`
	}

	path := fmt.Sprintf("/v2/view/%s/comment", viewID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("add view comment: %w", err)
	}

	return &Comment{ID: result.ID, Text: req.CommentText}, nil
}

// Subtypes returns available post subtypes for a comment type (v3 API).
func (s *CommentsService) Subtypes(ctx context.Context, typeID string) (*PostSubtypesResponse, error) {
	if typeID == "" {
		return nil, errIDRequired
	}

	path, err := s.client.v3Path(fmt.Sprintf("/comments/types/%s/subtypes", typeID))
	if err != nil {
		return nil, err
	}

	var result PostSubtypesResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get post subtypes: %w", err)
	}

	return &result, nil
}

// --- TimeService ---

// TimeService handles time tracking operations.
type TimeService struct {
	client *Client
}

// List returns time entries for a task within a team.
func (s *TimeService) List(ctx context.Context, teamID, taskID string) (*TimeEntriesListResponse, error) {
	if teamID == "" || taskID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/team/%s/time_entries?task_id=%s", teamID, url.QueryEscape(taskID))

	var result TimeEntriesListResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list time entries: %w", err)
	}

	return &result, nil
}

// Log creates a time entry for a task.
func (s *TimeService) Log(ctx context.Context, teamID, taskID string, durationMs, startMs int64) (*TimeEntry, error) {
	if teamID == "" || taskID == "" {
		return nil, errIDRequired
	}

	req := map[string]any{
		"duration": durationMs,
		"tid":      taskID,
		"start":    startMs,
	}

	var result struct {
		Data TimeEntry `json:"data"`
	}

	path := fmt.Sprintf("/v2/team/%s/time_entries", teamID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("log time: %w", err)
	}

	return &result.Data, nil
}

// --- FoldersService ---

// FoldersService handles folder operations.
type FoldersService struct {
	client *Client
}

// Get returns a folder by ID with full details.
func (s *FoldersService) Get(ctx context.Context, folderID string) (*FolderDetail, error) {
	if folderID == "" {
		return nil, errIDRequired
	}

	var result FolderDetail

	path := fmt.Sprintf("/v2/folder/%s", folderID)
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get folder: %w", err)
	}

	return &result, nil
}

// Create creates a new folder in a space.
func (s *FoldersService) Create(ctx context.Context, spaceID string, req CreateFolderRequest) (*FolderDetail, error) {
	if spaceID == "" {
		return nil, errIDRequired
	}

	if req.Name == "" {
		return nil, errNameRequired
	}

	var result FolderDetail

	path := fmt.Sprintf("/v2/space/%s/folder", spaceID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create folder: %w", err)
	}

	return &result, nil
}

// Update updates a folder.
func (s *FoldersService) Update(ctx context.Context, folderID string, req UpdateFolderRequest) (*FolderDetail, error) {
	if folderID == "" {
		return nil, errIDRequired
	}

	var result FolderDetail

	path := fmt.Sprintf("/v2/folder/%s", folderID)
	if err := s.client.Put(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("update folder: %w", err)
	}

	return &result, nil
}

// Delete deletes a folder.
func (s *FoldersService) Delete(ctx context.Context, folderID string) error {
	if folderID == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/folder/%s", folderID)
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("delete folder: %w", err)
	}

	return nil
}

// FromTemplate creates a folder from a template.
func (s *FoldersService) FromTemplate(ctx context.Context, spaceID, templateID string, req CreateFolderFromTemplateRequest) (*FolderDetail, error) {
	if spaceID == "" {
		return nil, errIDRequired
	}

	if templateID == "" {
		return nil, errIDRequired
	}

	var result FolderDetail

	path := fmt.Sprintf("/v2/space/%s/folder_template/%s", spaceID, templateID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create folder from template: %w", err)
	}

	return &result, nil
}
