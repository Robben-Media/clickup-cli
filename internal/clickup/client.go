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
	errEmailRequired       = errors.New("email is required")
	errOAuthFieldsRequired = errors.New("client_id, client_secret, and code are required")
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

// UserGroups provides methods for the User Groups API.
func (c *Client) UserGroups() *UserGroupsService {
	return &UserGroupsService{client: c}
}

// Roles provides methods for the Roles API.
func (c *Client) Roles() *RolesService {
	return &RolesService{client: c}
}

// Guests provides methods for the Guests API.
func (c *Client) Guests() *GuestsService {
	return &GuestsService{client: c}
}

// SharedHierarchy provides methods for the Shared Hierarchy API.
func (c *Client) SharedHierarchy() *SharedHierarchyService {
	return &SharedHierarchyService{client: c}
}

// Templates provides methods for the Templates API.
func (c *Client) Templates() *TemplatesService {
	return &TemplatesService{client: c}
}

// CustomTaskTypes provides methods for the Custom Task Types API.
func (c *Client) CustomTaskTypes() *CustomTaskTypesService {
	return &CustomTaskTypesService{client: c}
}

// LegacyTime provides methods for the Legacy Time Tracking API.
func (c *Client) LegacyTime() *LegacyTimeService {
	return &LegacyTimeService{client: c}
}

// AuditLogs provides methods for the Audit Logs API.
func (c *Client) AuditLogs() *AuditLogsService {
	return &AuditLogsService{client: c}
}

// ACLs provides methods for the ACLs API.
func (c *Client) ACLs() *ACLsService {
	return &ACLsService{client: c}
}

// Workspaces provides methods for the Workspaces API.
func (c *Client) Workspaces() *WorkspacesService {
	return &WorkspacesService{client: c}
}

// Auth provides methods for the Authorization API.
func (c *Client) Auth() *AuthService {
	return &AuthService{client: c}
}

// --- AuthService ---

// AuthService handles authorization operations.
type AuthService struct {
	client *Client
}

// Whoami returns the authorized user.
func (s *AuthService) Whoami(ctx context.Context) (*AuthorizedUserResponse, error) {
	path := "/v2/user"

	var result AuthorizedUserResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get authorized user: %w", err)
	}

	return &result, nil
}

// Token exchanges an OAuth authorization code for an access token.
// Note: This endpoint doesn't require the Authorization header.
func (s *AuthService) Token(ctx context.Context, req OAuthTokenRequest) (*OAuthTokenResponse, error) {
	if req.ClientID == "" || req.ClientSecret == "" || req.Code == "" {
		return nil, errOAuthFieldsRequired
	}

	// Use the API client's base URL but make a public request (no auth)
	path := "/v2/oauth/token"

	var result OAuthTokenResponse
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("exchange oauth token: %w", err)
	}

	return &result, nil
}

// --- WorkspacesService ---

// WorkspacesService handles workspace operations.
type WorkspacesService struct {
	client *Client
}

// List returns all workspaces the authorized user has access to.
func (s *WorkspacesService) List(ctx context.Context) (*WorkspacesResponse, error) {
	path := "/v2/team"

	var result WorkspacesResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list workspaces: %w", err)
	}

	return &result, nil
}

// Plan returns the workspace plan.
func (s *WorkspacesService) Plan(ctx context.Context, teamID string) (*WorkspacePlanResponse, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/team/%s/plan", teamID)

	var result WorkspacePlanResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get workspace plan: %w", err)
	}

	return &result, nil
}

// Seats returns the workspace seat usage.
func (s *WorkspacesService) Seats(ctx context.Context, teamID string) (*WorkspaceSeatsResponse, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/team/%s/seats", teamID)

	var result WorkspaceSeatsResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get workspace seats: %w", err)
	}

	return &result, nil
}

// --- SharedHierarchyService ---

// SharedHierarchyService handles shared hierarchy operations.
type SharedHierarchyService struct {
	client *Client
}

// List returns all resources shared with the authenticated user.
func (s *SharedHierarchyService) List(ctx context.Context, teamID string) (*SharedHierarchyResponse, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/team/%s/shared", teamID)

	var result SharedHierarchyResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list shared hierarchy: %w", err)
	}

	return &result, nil
}

// --- TemplatesService ---

// TemplatesService handles task template operations.
type TemplatesService struct {
	client *Client
}

// List returns all task templates for a team.
func (s *TemplatesService) List(ctx context.Context, teamID string, page int) (*TaskTemplatesResponse, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/team/%s/taskTemplate?page=%d", teamID, page)

	var result TaskTemplatesResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}

	return &result, nil
}

// --- CustomTaskTypesService ---

// CustomTaskTypesService handles custom task type operations.
type CustomTaskTypesService struct {
	client *Client
}

// List returns all custom task types for a team.
func (s *CustomTaskTypesService) List(ctx context.Context, teamID string) (*CustomTaskTypesResponse, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/team/%s/custom_item", teamID)

	var result CustomTaskTypesResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list custom task types: %w", err)
	}

	return &result, nil
}

// --- LegacyTimeService ---

// LegacyTimeService handles legacy time tracking operations at the task level.
type LegacyTimeService struct {
	client *Client
}

// List returns all time intervals for a task.
func (s *LegacyTimeService) List(ctx context.Context, taskID string, customTaskIDs bool, teamID string) (*LegacyTimeResponse, error) {
	if taskID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/task/%s/time", taskID)

	params := url.Values{}
	if customTaskIDs {
		params.Set("custom_task_ids", "true")
	}

	if teamID != "" {
		params.Set("team_id", teamID)
	}

	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var result LegacyTimeResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list legacy time: %w", err)
	}

	return &result, nil
}

// Track creates a time interval on a task.
func (s *LegacyTimeService) Track(ctx context.Context, taskID string, req TrackTimeRequest) (*TrackTimeResponse, error) {
	if taskID == "" {
		return nil, errIDRequired
	}

	var result TrackTimeResponse

	path := fmt.Sprintf("/v2/task/%s/time", taskID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("track legacy time: %w", err)
	}

	return &result, nil
}

// Edit updates a time interval on a task.
func (s *LegacyTimeService) Edit(ctx context.Context, taskID, intervalID string, req EditTimeRequest) error {
	if taskID == "" || intervalID == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/task/%s/time/%s", taskID, intervalID)
	if err := s.client.Put(ctx, path, req, nil); err != nil {
		return fmt.Errorf("edit legacy time: %w", err)
	}

	return nil
}

// Delete removes a time interval from a task.
func (s *LegacyTimeService) Delete(ctx context.Context, taskID, intervalID string) error {
	if taskID == "" || intervalID == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/task/%s/time/%s", taskID, intervalID)
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("delete legacy time: %w", err)
	}

	return nil
}

// --- AuditLogsService ---

// AuditLogsService handles audit log operations.
type AuditLogsService struct {
	client *Client
}

// Query returns audit logs for a workspace.
func (s *AuditLogsService) Query(ctx context.Context, req AuditLogQuery) (*AuditLogsResponse, error) {
	path, err := s.client.v3Path("/auditlogs")
	if err != nil {
		return nil, err
	}

	var result AuditLogsResponse
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("query audit logs: %w", err)
	}

	return &result, nil
}

// --- ACLsService ---

// ACLsService handles ACL operations.
type ACLsService struct {
	client *Client
}

// Update updates access control settings for an object.
func (s *ACLsService) Update(ctx context.Context, objectType, objectID string, req UpdateACLRequest) error {
	if objectType == "" || objectID == "" {
		return errIDRequired
	}

	path, err := s.client.v3Path(fmt.Sprintf("/%s/%s/acls", objectType, objectID))
	if err != nil {
		return err
	}

	if err := s.client.Patch(ctx, path, req, nil); err != nil {
		return fmt.Errorf("update ACLs: %w", err)
	}

	return nil
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

// --- UserGroupsService ---

// UserGroupsService handles user group operations.
type UserGroupsService struct {
	client *Client
}

// List returns all user groups.
func (s *UserGroupsService) List(ctx context.Context) (*UserGroupsResponse, error) {
	path := "/v2/group"

	var result UserGroupsResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list user groups: %w", err)
	}

	return &result, nil
}

// Create creates a new user group.
func (s *UserGroupsService) Create(ctx context.Context, teamID string, req CreateUserGroupRequest) (*UserGroup, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	if req.Name == "" {
		return nil, errNameRequired
	}

	var result UserGroup

	path := fmt.Sprintf("/v2/team/%s/group", teamID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create user group: %w", err)
	}

	return &result, nil
}

// Update updates a user group.
func (s *UserGroupsService) Update(ctx context.Context, groupID string, req UpdateUserGroupRequest) (*UserGroup, error) {
	if groupID == "" {
		return nil, errIDRequired
	}

	var result UserGroup

	path := fmt.Sprintf("/v2/group/%s", groupID)
	if err := s.client.Put(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("update user group: %w", err)
	}

	return &result, nil
}

// Delete deletes a user group.
func (s *UserGroupsService) Delete(ctx context.Context, groupID string) error {
	if groupID == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/group/%s", groupID)
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("delete user group: %w", err)
	}

	return nil
}

// --- RolesService ---

// RolesService handles custom role operations.
type RolesService struct {
	client *Client
}

// List returns all custom roles for a team.
func (s *RolesService) List(ctx context.Context, teamID string) (*CustomRolesResponse, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/team/%s/customroles", teamID)

	var result CustomRolesResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list custom roles: %w", err)
	}

	return &result, nil
}

// --- GuestsService ---

// GuestsService handles guest operations.
type GuestsService struct {
	client *Client
}

// Get returns a guest by ID.
func (s *GuestsService) Get(ctx context.Context, teamID string, guestID int) (*Guest, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	var result GuestResponse

	path := fmt.Sprintf("/v2/team/%s/guest/%d", teamID, guestID)
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get guest: %w", err)
	}

	return &result.Guest, nil
}

// Invite invites a new guest to the workspace.
func (s *GuestsService) Invite(ctx context.Context, teamID string, req InviteGuestRequest) (*Guest, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	if req.Email == "" {
		return nil, errEmailRequired
	}

	var result GuestResponse

	path := fmt.Sprintf("/v2/team/%s/guest", teamID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("invite guest: %w", err)
	}

	return &result.Guest, nil
}

// Update updates a guest's permissions.
func (s *GuestsService) Update(ctx context.Context, teamID string, guestID int, req EditGuestRequest) (*Guest, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	var result GuestResponse

	path := fmt.Sprintf("/v2/team/%s/guest/%d", teamID, guestID)
	if err := s.client.Put(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("update guest: %w", err)
	}

	return &result.Guest, nil
}

// Remove removes a guest from the workspace.
func (s *GuestsService) Remove(ctx context.Context, teamID string, guestID int) error {
	if teamID == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/team/%s/guest/%d", teamID, guestID)
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("remove guest: %w", err)
	}

	return nil
}

// AddToTask adds a guest to a task with the specified permission level.
func (s *GuestsService) AddToTask(ctx context.Context, taskID string, guestID int, permissionLevel string) (*Guest, error) {
	if taskID == "" {
		return nil, errIDRequired
	}

	req := AddGuestToResourceRequest{PermissionLevel: permissionLevel}

	var result GuestResponse

	path := fmt.Sprintf("/v2/task/%s/guest/%d", taskID, guestID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("add guest to task: %w", err)
	}

	return &result.Guest, nil
}

// RemoveFromTask removes a guest from a task.
func (s *GuestsService) RemoveFromTask(ctx context.Context, taskID string, guestID int) error {
	if taskID == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/task/%s/guest/%d", taskID, guestID)
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("remove guest from task: %w", err)
	}

	return nil
}

// AddToList adds a guest to a list with the specified permission level.
func (s *GuestsService) AddToList(ctx context.Context, listID string, guestID int, permissionLevel string) (*Guest, error) {
	if listID == "" {
		return nil, errIDRequired
	}

	req := AddGuestToResourceRequest{PermissionLevel: permissionLevel}

	var result GuestResponse

	path := fmt.Sprintf("/v2/list/%s/guest/%d", listID, guestID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("add guest to list: %w", err)
	}

	return &result.Guest, nil
}

// RemoveFromList removes a guest from a list.
func (s *GuestsService) RemoveFromList(ctx context.Context, listID string, guestID int) error {
	if listID == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/list/%s/guest/%d", listID, guestID)
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("remove guest from list: %w", err)
	}

	return nil
}

// AddToFolder adds a guest to a folder with the specified permission level.
func (s *GuestsService) AddToFolder(ctx context.Context, folderID string, guestID int, permissionLevel string) (*Guest, error) {
	if folderID == "" {
		return nil, errIDRequired
	}

	req := AddGuestToResourceRequest{PermissionLevel: permissionLevel}

	var result GuestResponse

	path := fmt.Sprintf("/v2/folder/%s/guest/%d", folderID, guestID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("add guest to folder: %w", err)
	}

	return &result.Guest, nil
}

// RemoveFromFolder removes a guest from a folder.
func (s *GuestsService) RemoveFromFolder(ctx context.Context, folderID string, guestID int) error {
	if folderID == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/folder/%s/guest/%d", folderID, guestID)
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("remove guest from folder: %w", err)
	}

	return nil
}
