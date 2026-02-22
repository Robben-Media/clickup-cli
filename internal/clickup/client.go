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
	errDependsOnRequired   = errors.New("either depends_on or dependency_of is required")
	errLinkTaskIDRequired  = errors.New("linked task ID is required")
	errFieldIDRequired     = errors.New("field ID is required")
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

// Tags provides methods for the Tags API.
func (c *Client) Tags() *TagsService {
	return &TagsService{client: c}
}

// Checklists provides methods for the Checklists API.
func (c *Client) Checklists() *ChecklistsService {
	return &ChecklistsService{client: c}
}

// Relationships provides methods for the Relationships API.
func (c *Client) Relationships() *RelationshipsService {
	return &RelationshipsService{client: c}
}

// CustomFields provides methods for the Custom Fields API.
func (c *Client) CustomFields() *CustomFieldsService {
	return &CustomFieldsService{client: c}
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

// --- TagsService ---

// TagsService handles space tag and task tag operations.
type TagsService struct {
	client *Client
}

// List returns all tags in a space.
func (s *TagsService) List(ctx context.Context, spaceID string) (*SpaceTagsResponse, error) {
	if spaceID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/space/%s/tag", spaceID)

	var result SpaceTagsResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}

	return &result, nil
}

// Create creates a new tag in a space.
func (s *TagsService) Create(ctx context.Context, spaceID string, req CreateSpaceTagRequest) error {
	if spaceID == "" {
		return errIDRequired
	}

	if req.Tag.Name == "" {
		return errNameRequired
	}

	path := fmt.Sprintf("/v2/space/%s/tag", spaceID)
	if err := s.client.Post(ctx, path, req, nil); err != nil {
		return fmt.Errorf("create tag: %w", err)
	}

	return nil
}

// Update updates a tag in a space.
func (s *TagsService) Update(ctx context.Context, spaceID, tagName string, req EditSpaceTagRequest) error {
	if spaceID == "" {
		return errIDRequired
	}

	if tagName == "" {
		return errNameRequired
	}

	path := fmt.Sprintf("/v2/space/%s/tag/%s", spaceID, url.PathEscape(tagName))
	if err := s.client.Put(ctx, path, req, nil); err != nil {
		return fmt.Errorf("update tag: %w", err)
	}

	return nil
}

// Delete deletes a tag from a space.
func (s *TagsService) Delete(ctx context.Context, spaceID, tagName string) error {
	if spaceID == "" {
		return errIDRequired
	}

	if tagName == "" {
		return errNameRequired
	}

	path := fmt.Sprintf("/v2/space/%s/tag/%s", spaceID, url.PathEscape(tagName))
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("delete tag: %w", err)
	}

	return nil
}

// AddToTask adds a tag to a task.
func (s *TagsService) AddToTask(ctx context.Context, taskID, tagName string) error {
	if taskID == "" {
		return errIDRequired
	}

	if tagName == "" {
		return errNameRequired
	}

	path := fmt.Sprintf("/v2/task/%s/tag/%s", taskID, url.PathEscape(tagName))
	if err := s.client.Post(ctx, path, nil, nil); err != nil {
		return fmt.Errorf("add tag to task: %w", err)
	}

	return nil
}

// RemoveFromTask removes a tag from a task.
func (s *TagsService) RemoveFromTask(ctx context.Context, taskID, tagName string) error {
	if taskID == "" {
		return errIDRequired
	}

	if tagName == "" {
		return errNameRequired
	}

	path := fmt.Sprintf("/v2/task/%s/tag/%s", taskID, url.PathEscape(tagName))
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("remove tag from task: %w", err)
	}

	return nil
}

// --- ChecklistsService ---

// ChecklistsService handles checklist operations.
type ChecklistsService struct {
	client *Client
}

// Create creates a new checklist on a task.
func (s *ChecklistsService) Create(ctx context.Context, taskID string, req CreateChecklistRequest) (*Checklist, error) {
	if taskID == "" {
		return nil, errIDRequired
	}

	if req.Name == "" {
		return nil, errNameRequired
	}

	// ClickUp wraps the response in {"checklist": {...}}
	var result struct {
		Checklist Checklist `json:"checklist"`
	}

	path := fmt.Sprintf("/v2/task/%s/checklist", taskID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create checklist: %w", err)
	}

	return &result.Checklist, nil
}

// Update updates a checklist.
func (s *ChecklistsService) Update(ctx context.Context, checklistID string, req EditChecklistRequest) (*Checklist, error) {
	if checklistID == "" {
		return nil, errIDRequired
	}

	// ClickUp wraps the response in {"checklist": {...}}
	var result struct {
		Checklist Checklist `json:"checklist"`
	}

	path := fmt.Sprintf("/v2/checklist/%s", checklistID)
	if err := s.client.Put(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("update checklist: %w", err)
	}

	return &result.Checklist, nil
}

// Delete deletes a checklist.
func (s *ChecklistsService) Delete(ctx context.Context, checklistID string) error {
	if checklistID == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/checklist/%s", checklistID)
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("delete checklist: %w", err)
	}

	return nil
}

// AddItem creates a new checklist item.
func (s *ChecklistsService) AddItem(ctx context.Context, checklistID string, req CreateChecklistItemRequest) (*Checklist, error) {
	if checklistID == "" {
		return nil, errIDRequired
	}

	if req.Name == "" {
		return nil, errNameRequired
	}

	// ClickUp returns the full checklist with items
	var result struct {
		Checklist Checklist `json:"checklist"`
	}

	path := fmt.Sprintf("/v2/checklist/%s/checklist_item", checklistID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("add checklist item: %w", err)
	}

	return &result.Checklist, nil
}

// UpdateItem updates a checklist item.
func (s *ChecklistsService) UpdateItem(ctx context.Context, checklistID, itemID string, req EditChecklistItemRequest) (*Checklist, error) {
	if checklistID == "" {
		return nil, errIDRequired
	}

	if itemID == "" {
		return nil, errIDRequired
	}

	// ClickUp returns the full checklist with items
	var result struct {
		Checklist Checklist `json:"checklist"`
	}

	path := fmt.Sprintf("/v2/checklist/%s/checklist_item/%s", checklistID, itemID)
	if err := s.client.Put(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("update checklist item: %w", err)
	}

	return &result.Checklist, nil
}

// DeleteItem deletes a checklist item.
func (s *ChecklistsService) DeleteItem(ctx context.Context, checklistID, itemID string) error {
	if checklistID == "" {
		return errIDRequired
	}

	if itemID == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/checklist/%s/checklist_item/%s", checklistID, itemID)
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("delete checklist item: %w", err)
	}

	return nil
}

// --- RelationshipsService ---

// RelationshipsService handles task dependency and link operations.
type RelationshipsService struct {
	client *Client
}

// AddDependency adds a dependency to a task.
// Use DependsOn to wait for another task, or DependencyOf to block another task.
func (s *RelationshipsService) AddDependency(ctx context.Context, taskID string, req AddDependencyRequest) error {
	if taskID == "" {
		return errIDRequired
	}

	if req.DependsOn == "" && req.DependencyOf == "" {
		return errDependsOnRequired
	}

	path := fmt.Sprintf("/v2/task/%s/dependency", taskID)
	if err := s.client.Post(ctx, path, req, nil); err != nil {
		return fmt.Errorf("add dependency: %w", err)
	}

	return nil
}

// RemoveDependency removes a dependency from a task.
// Use DependsOn to remove "waiting on", or DependencyOf to remove "blocking".
func (s *RelationshipsService) RemoveDependency(ctx context.Context, taskID string, req AddDependencyRequest) error {
	if taskID == "" {
		return errIDRequired
	}

	if req.DependsOn == "" && req.DependencyOf == "" {
		return errDependsOnRequired
	}

	// Build query params for DELETE request
	params := url.Values{}
	if req.DependsOn != "" {
		params.Set("depends_on", req.DependsOn)
	}

	if req.DependencyOf != "" {
		params.Set("dependency_of", req.DependencyOf)
	}

	path := fmt.Sprintf("/v2/task/%s/dependency?%s", taskID, params.Encode())
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("remove dependency: %w", err)
	}

	return nil
}

// Link adds a link between two tasks.
func (s *RelationshipsService) Link(ctx context.Context, taskID, linkTaskID string) error {
	if taskID == "" {
		return errIDRequired
	}

	if linkTaskID == "" {
		return errLinkTaskIDRequired
	}

	path := fmt.Sprintf("/v2/task/%s/link/%s", taskID, linkTaskID)
	if err := s.client.Post(ctx, path, nil, nil); err != nil {
		return fmt.Errorf("add link: %w", err)
	}

	return nil
}

// Unlink removes a link between two tasks.
func (s *RelationshipsService) Unlink(ctx context.Context, taskID, linkTaskID string) error {
	if taskID == "" {
		return errIDRequired
	}

	if linkTaskID == "" {
		return errLinkTaskIDRequired
	}

	path := fmt.Sprintf("/v2/task/%s/link/%s", taskID, linkTaskID)
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("remove link: %w", err)
	}

	return nil
}

// --- CustomFieldsService ---

// CustomFieldsService handles custom field operations.
type CustomFieldsService struct {
	client *Client
}

// ListByList returns custom fields accessible from a list.
func (s *CustomFieldsService) ListByList(ctx context.Context, listID string) (*CustomFieldsResponse, error) {
	if listID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/list/%s/field", listID)

	var result CustomFieldsResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list custom fields by list: %w", err)
	}

	return &result, nil
}

// ListByFolder returns custom fields accessible from a folder.
func (s *CustomFieldsService) ListByFolder(ctx context.Context, folderID string) (*CustomFieldsResponse, error) {
	if folderID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/folder/%s/field", folderID)

	var result CustomFieldsResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list custom fields by folder: %w", err)
	}

	return &result, nil
}

// ListBySpace returns custom fields accessible from a space.
func (s *CustomFieldsService) ListBySpace(ctx context.Context, spaceID string) (*CustomFieldsResponse, error) {
	if spaceID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/space/%s/field", spaceID)

	var result CustomFieldsResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list custom fields by space: %w", err)
	}

	return &result, nil
}

// ListByTeam returns custom fields accessible from a workspace.
func (s *CustomFieldsService) ListByTeam(ctx context.Context, teamID string) (*CustomFieldsResponse, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/team/%s/field", teamID)

	var result CustomFieldsResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list custom fields by team: %w", err)
	}

	return &result, nil
}

// Set sets a custom field value on a task.
func (s *CustomFieldsService) Set(ctx context.Context, taskID, fieldID string, value interface{}) error {
	if taskID == "" {
		return errIDRequired
	}

	if fieldID == "" {
		return errFieldIDRequired
	}

	req := SetCustomFieldRequest{Value: value}

	path := fmt.Sprintf("/v2/task/%s/field/%s", taskID, fieldID)
	if err := s.client.Post(ctx, path, req, nil); err != nil {
		return fmt.Errorf("set custom field: %w", err)
	}

	return nil
}

// Remove removes a custom field value from a task.
func (s *CustomFieldsService) Remove(ctx context.Context, taskID, fieldID string) error {
	if taskID == "" {
		return errIDRequired
	}

	if fieldID == "" {
		return errFieldIDRequired
	}

	path := fmt.Sprintf("/v2/task/%s/field/%s", taskID, fieldID)
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("remove custom field: %w", err)
	}

	return nil
}
