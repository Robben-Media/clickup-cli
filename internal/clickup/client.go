package clickup

import (
	"context"
	"errors"
	"fmt"

	"github.com/builtbyrobben/clickup-cli/internal/api"
)

var (
	errIDRequired   = errors.New("id is required")
	errNameRequired = errors.New("name is required")
	errTextRequired = errors.New("comment text is required")
)

const defaultBaseURL = "https://api.clickup.com/api/v2"

// Client wraps the API client with ClickUp-specific methods.
type Client struct {
	*api.Client
}

// NewClient creates a new ClickUp API client.
func NewClient(apiKey string) *Client {
	return &Client{
		Client: api.NewClient(apiKey,
			api.WithBaseURL(defaultBaseURL),
			api.WithUserAgent("clickup-cli/1.0"),
		),
	}
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

	path := fmt.Sprintf("/list/%s/task?include_closed=true", listID)

	if status != "" {
		path += fmt.Sprintf("&statuses[]=%s", status)
	}

	if assignee != "" {
		path += fmt.Sprintf("&assignees[]=%s", assignee)
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

	path := fmt.Sprintf("/task/%s", taskID)
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

	path := fmt.Sprintf("/list/%s/task", listID)
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

	path := fmt.Sprintf("/task/%s", taskID)
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

	path := fmt.Sprintf("/task/%s", taskID)
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

	path := fmt.Sprintf("/team/%s/space", teamID)

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

	path := fmt.Sprintf("/folder/%s/list", folderID)

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

	path := fmt.Sprintf("/space/%s/list", spaceID)

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

	path := fmt.Sprintf("/space/%s/folder", spaceID)

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
	path := fmt.Sprintf("/team/%s", teamID)

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

	path := fmt.Sprintf("/task/%s/comment", taskID)

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

	// ClickUp returns the comment ID in a wrapper
	var result struct {
		ID string `json:"id"`
	}

	path := fmt.Sprintf("/task/%s/comment", taskID)
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

	path := fmt.Sprintf("/team/%s/time_entries?task_id=%s", teamID, taskID)

	var result TimeEntriesListResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list time entries: %w", err)
	}

	return &result, nil
}

// Log creates a time entry for a task.
func (s *TimeService) Log(ctx context.Context, teamID, taskID string, durationMs int64) (*TimeEntry, error) {
	if teamID == "" || taskID == "" {
		return nil, errIDRequired
	}

	req := map[string]any{
		"duration": durationMs,
		"tid":      taskID,
	}

	var result struct {
		Data TimeEntry `json:"data"`
	}

	path := fmt.Sprintf("/team/%s/time_entries", teamID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("log time: %w", err)
	}

	return &result.Data, nil
}
