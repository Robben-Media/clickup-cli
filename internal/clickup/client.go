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
	errIDRequired            = errors.New("id is required")
	errNameRequired          = errors.New("name is required")
	errTextRequired          = errors.New("comment text is required")
	errWorkspaceIDRequired   = errors.New("workspace ID required for v3 API; set CLICKUP_WORKSPACE_ID or use --workspace flag")
	errEmailRequired         = errors.New("email is required")
	errOAuthFieldsRequired   = errors.New("client_id, client_secret, and code are required")
	errTaskIDRequired        = errors.New("at least one task ID is required")
	errSourceTaskIDRequired  = errors.New("at least one source task ID is required")
	errEndpointRequired      = errors.New("endpoint URL is required")
	errEventsRequired        = errors.New("at least one event is required")
	errKeyResultTypeRequired = errors.New("key result type is required")
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

// Tags provides methods for the Tags API.
func (c *Client) Tags() *TagsService {
	return &TagsService{client: c}
}

// Checklists provides methods for the Checklists API.
func (c *Client) Checklists() *ChecklistsService {
	return &ChecklistsService{client: c}
}

// Relationships provides methods for the Task Relationships API.
func (c *Client) Relationships() *RelationshipsService {
	return &RelationshipsService{client: c}
}

// CustomFields provides methods for the Custom Fields API.
func (c *Client) CustomFields() *CustomFieldsService {
	return &CustomFieldsService{client: c}
}

// Views provides methods for the Views API.
func (c *Client) Views() *ViewsService {
	return &ViewsService{client: c}
}

// Webhooks provides methods for the Webhooks API.
func (c *Client) Webhooks() *WebhooksService {
	return &WebhooksService{client: c}
}

// Goals provides methods for the Goals API.
func (c *Client) Goals() *GoalsService {
	return &GoalsService{client: c}
}

// Users provides methods for the Users API.
func (c *Client) Users() *UsersService {
	return &UsersService{client: c}
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

// Search returns tasks filtered across a workspace (GetFilteredTeamTasks).
func (s *TasksService) Search(ctx context.Context, teamID string, params FilteredTeamTasksParams) (*FilteredTeamTasksResponse, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/team/%s/task", teamID)

	// Build query string manually for array parameters
	query := url.Values{}
	if params.Page > 0 {
		query.Set("page", fmt.Sprintf("%d", params.Page))
	}

	if params.OrderBy != "" {
		query.Set("order_by", params.OrderBy)
	}

	if params.Reverse {
		query.Set("reverse", "true")
	}

	if params.Subtasks {
		query.Set("subtasks", "true")
	}

	for _, status := range params.Statuses {
		query.Add("statuses[]", status)
	}

	if params.IncludeClosed {
		query.Set("include_closed", "true")
	}

	for _, assignee := range params.Assignees {
		query.Add("assignees[]", fmt.Sprintf("%d", assignee))
	}

	for _, tag := range params.Tags {
		query.Add("tags[]", tag)
	}

	if params.DueDateGt > 0 {
		query.Set("due_date_gt", fmt.Sprintf("%d", params.DueDateGt))
	}

	if params.DueDateLt > 0 {
		query.Set("due_date_lt", fmt.Sprintf("%d", params.DueDateLt))
	}

	if params.DateCreatedGt > 0 {
		query.Set("date_created_gt", fmt.Sprintf("%d", params.DateCreatedGt))
	}

	if params.DateCreatedLt > 0 {
		query.Set("date_created_lt", fmt.Sprintf("%d", params.DateCreatedLt))
	}

	if params.DateUpdatedGt > 0 {
		query.Set("date_updated_gt", fmt.Sprintf("%d", params.DateUpdatedGt))
	}

	if params.DateUpdatedLt > 0 {
		query.Set("date_updated_lt", fmt.Sprintf("%d", params.DateUpdatedLt))
	}

	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	var result FilteredTeamTasksResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("search tasks: %w", err)
	}

	return &result, nil
}

// TimeInStatus returns time-in-status data for a single task.
func (s *TasksService) TimeInStatus(ctx context.Context, taskID string) (*TimeInStatusResponse, error) {
	if taskID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/task/%s/time_in_status", taskID)

	var result TimeInStatusResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get time in status: %w", err)
	}

	return &result, nil
}

// BulkTimeInStatus returns time-in-status data for multiple tasks.
func (s *TasksService) BulkTimeInStatus(ctx context.Context, taskIDs []string) (BulkTimeInStatusResponse, error) {
	if len(taskIDs) == 0 {
		return nil, errTaskIDRequired
	}

	// Build query string with task_ids
	query := url.Values{}
	for _, id := range taskIDs {
		query.Add("task_ids", id)
	}

	path := "/v2/task/bulk_time_in_status/task_ids?" + query.Encode()

	var result BulkTimeInStatusResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get bulk time in status: %w", err)
	}

	return result, nil
}

// Merge merges source tasks into a target task.
func (s *TasksService) Merge(ctx context.Context, targetTaskID string, sourceTaskIDs []string) (*MergeTasksResponse, error) {
	if targetTaskID == "" {
		return nil, errIDRequired
	}

	if len(sourceTaskIDs) == 0 {
		return nil, errSourceTaskIDRequired
	}

	req := MergeTasksRequest{MergedTaskIDs: sourceTaskIDs}

	var result MergeTasksResponse

	path := fmt.Sprintf("/v2/task/%s/merge", targetTaskID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("merge tasks: %w", err)
	}

	return &result, nil
}

// Move moves a task to a different list (v3 API).
func (s *TasksService) Move(ctx context.Context, taskID, listID string) (*MoveTaskResponse, error) {
	if taskID == "" || listID == "" {
		return nil, errIDRequired
	}

	path, err := s.client.v3Path(fmt.Sprintf("/tasks/%s/home_list/%s", taskID, listID))
	if err != nil {
		return nil, err
	}

	var result MoveTaskResponse
	if err := s.client.Put(ctx, path, nil, &result); err != nil {
		return nil, fmt.Errorf("move task: %w", err)
	}

	return &result, nil
}

// CreateFromTemplate creates a new task from a template.
func (s *TasksService) CreateFromTemplate(ctx context.Context, listID, templateID string, req CreateTaskFromTemplateRequest) (*Task, error) {
	if listID == "" || templateID == "" {
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

// Get returns a space by ID with full details.
func (s *SpacesService) Get(ctx context.Context, spaceID string) (*SpaceDetail, error) {
	if spaceID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/space/%s", spaceID)

	var result SpaceDetail
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

	path := fmt.Sprintf("/v2/team/%s/space", teamID)

	var result SpaceDetail
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

	path := fmt.Sprintf("/v2/space/%s", spaceID)

	var result SpaceDetail
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

// Folders provides methods for the Folders API.
func (c *Client) Folders() *FoldersService {
	return &FoldersService{client: c}
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

	path := fmt.Sprintf("/v2/folder/%s", folderID)

	var result FolderDetail
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

	path := fmt.Sprintf("/v2/space/%s/folder", spaceID)

	var result FolderDetail
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

	path := fmt.Sprintf("/v2/folder/%s", folderID)

	var result FolderDetail
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

// CreateFromTemplate creates a folder from a template in a space.
func (s *FoldersService) CreateFromTemplate(ctx context.Context, spaceID, templateID string, req CreateFolderFromTemplateRequest) (*FolderDetail, error) {
	if spaceID == "" || templateID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/space/%s/folder_template/%s", spaceID, templateID)

	var result FolderDetail
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create folder from template: %w", err)
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

// Get returns a list by ID with full details.
func (s *ListsService) Get(ctx context.Context, listID string) (*ListDetail, error) {
	if listID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/list/%s", listID)

	var result ListDetail
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

	path := fmt.Sprintf("/v2/folder/%s/list", folderID)

	var result ListDetail
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create list in folder: %w", err)
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

	path := fmt.Sprintf("/v2/space/%s/list", spaceID)

	var result ListDetail
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

	path := fmt.Sprintf("/v2/list/%s", listID)

	var result ListDetail
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

// CreateFromTemplateInFolder creates a list from a template in a folder.
func (s *ListsService) CreateFromTemplateInFolder(ctx context.Context, folderID, templateID string, req CreateListFromTemplateRequest) (*ListDetail, error) {
	if folderID == "" || templateID == "" {
		return nil, errIDRequired
	}

	if req.Name == "" {
		return nil, errNameRequired
	}

	path := fmt.Sprintf("/v2/folder/%s/list_template/%s", folderID, templateID)

	var result ListDetail
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create list from template in folder: %w", err)
	}

	return &result, nil
}

// CreateFromTemplateInSpace creates a folderless list from a template in a space.
func (s *ListsService) CreateFromTemplateInSpace(ctx context.Context, spaceID, templateID string, req CreateListFromTemplateRequest) (*ListDetail, error) {
	if spaceID == "" || templateID == "" {
		return nil, errIDRequired
	}

	if req.Name == "" {
		return nil, errNameRequired
	}

	path := fmt.Sprintf("/v2/space/%s/list_template/%s", spaceID, templateID)

	var result ListDetail
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create list from template in space: %w", err)
	}

	return &result, nil
}

// AddTask adds a task to a list.
func (s *ListsService) AddTask(ctx context.Context, listID, taskID string) error {
	if listID == "" || taskID == "" {
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
	if listID == "" || taskID == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/list/%s/task/%s", listID, taskID)
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("remove task from list: %w", err)
	}

	return nil
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

// Update updates a comment.
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

// Replies returns threaded replies to a comment.
func (s *CommentsService) Replies(ctx context.Context, commentID string) (*ThreadedCommentsResponse, error) {
	if commentID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/comment/%s/reply", commentID)

	var result ThreadedCommentsResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get comment replies: %w", err)
	}

	return &result, nil
}

// Reply creates a threaded reply to a comment.
func (s *CommentsService) Reply(ctx context.Context, commentID string, text string) (*Comment, error) {
	if commentID == "" {
		return nil, errIDRequired
	}

	if text == "" {
		return nil, errTextRequired
	}

	req := CreateCommentRequest{CommentText: text}

	var result Comment

	path := fmt.Sprintf("/v2/comment/%s/reply", commentID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create comment reply: %w", err)
	}

	return &result, nil
}

// ListComments returns list-level comments.
func (s *CommentsService) ListComments(ctx context.Context, listID string) (*CommentsListResponse, error) {
	if listID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/list/%s/comment", listID)

	var result CommentsListResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get list comments: %w", err)
	}

	return &result, nil
}

// AddList creates a comment on a list.
func (s *CommentsService) AddList(ctx context.Context, listID string, req CreateListCommentRequest) (*Comment, error) {
	if listID == "" {
		return nil, errIDRequired
	}

	if req.CommentText == "" {
		return nil, errTextRequired
	}

	// ClickUp returns the comment ID as a number in a wrapper
	var result struct {
		ID json.Number `json:"id"`
	}

	path := fmt.Sprintf("/v2/list/%s/comment", listID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("add list comment: %w", err)
	}

	return &Comment{ID: result.ID, Text: req.CommentText}, nil
}

// ViewComments returns view-level comments with pagination.
func (s *CommentsService) ViewComments(ctx context.Context, viewID string, start int, startID string) (*CommentsListResponse, error) {
	if viewID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/view/%s/comment", viewID)

	params := url.Values{}
	if start > 0 {
		params.Set("start", fmt.Sprintf("%d", start))
	}

	if startID != "" {
		params.Set("start_id", startID)
	}

	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var result CommentsListResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list view comments: %w", err)
	}

	return &result, nil
}

// AddView creates a comment on a view.
func (s *CommentsService) AddView(ctx context.Context, viewID string, req CreateViewCommentRequest) (*Comment, error) {
	if viewID == "" {
		return nil, errIDRequired
	}

	if req.CommentText == "" {
		return nil, errTextRequired
	}

	// ClickUp returns the comment ID as a number in a wrapper
	var result struct {
		ID json.Number `json:"id"`
	}

	path := fmt.Sprintf("/v2/view/%s/comment", viewID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("add view comment: %w", err)
	}

	return &Comment{ID: result.ID, Text: req.CommentText}, nil
}

// Subtypes returns post subtype IDs for a type (v3 API).
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

// Get returns a single time entry by ID.
func (s *TimeService) Get(ctx context.Context, teamID, entryID string) (*TimeEntryDetail, error) {
	if teamID == "" || entryID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/team/%s/time_entries/%s", teamID, entryID)

	var result TimeEntryDetailResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get time entry: %w", err)
	}

	return &result.Data, nil
}

// Current returns the currently running time entry.
func (s *TimeService) Current(ctx context.Context, teamID string) (*TimeEntryDetail, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/team/%s/time_entries/current", teamID)

	var result TimeEntryDetailResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get current timer: %w", err)
	}

	return &result.Data, nil
}

// Start starts a new timer.
func (s *TimeService) Start(ctx context.Context, teamID string, req StartTimeEntryRequest) (*TimeEntryDetail, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	var result TimeEntryDetailResponse

	path := fmt.Sprintf("/v2/team/%s/time_entries/start", teamID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("start timer: %w", err)
	}

	return &result.Data, nil
}

// Stop stops the currently running timer.
func (s *TimeService) Stop(ctx context.Context, teamID string) (*TimeEntryDetail, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	var result TimeEntryDetailResponse

	path := fmt.Sprintf("/v2/team/%s/time_entries/stop", teamID)
	if err := s.client.Post(ctx, path, nil, &result); err != nil {
		return nil, fmt.Errorf("stop timer: %w", err)
	}

	return &result.Data, nil
}

// Update updates a time entry.
func (s *TimeService) Update(ctx context.Context, teamID, entryID string, req UpdateTimeEntryRequest) (*TimeEntryDetail, error) {
	if teamID == "" || entryID == "" {
		return nil, errIDRequired
	}

	var result TimeEntryDetailResponse

	path := fmt.Sprintf("/v2/team/%s/time_entries/%s", teamID, entryID)
	if err := s.client.Put(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("update time entry: %w", err)
	}

	return &result.Data, nil
}

// Delete removes a time entry.
func (s *TimeService) Delete(ctx context.Context, teamID, entryID string) error {
	if teamID == "" || entryID == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/team/%s/time_entries/%s", teamID, entryID)
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("delete time entry: %w", err)
	}

	return nil
}

// History returns the change history for a time entry.
func (s *TimeService) History(ctx context.Context, teamID, entryID string) (*TimeEntryHistoryResponse, error) {
	if teamID == "" || entryID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/team/%s/time_entries/%s/history", teamID, entryID)

	var result TimeEntryHistoryResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get time entry history: %w", err)
	}

	return &result, nil
}

// ListTags returns all tags used in time entries.
func (s *TimeService) ListTags(ctx context.Context, teamID string) (*TimeEntryTagsResponse, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/team/%s/time_entries/tags", teamID)

	var result TimeEntryTagsResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list time entry tags: %w", err)
	}

	return &result, nil
}

// AddTags adds tags to time entries.
func (s *TimeService) AddTags(ctx context.Context, teamID string, req TimeEntryTagsRequest) error {
	if teamID == "" {
		return errIDRequired
	}

	if len(req.TimeEntryIDs) == 0 || len(req.Tags) == 0 {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/team/%s/time_entries/tags", teamID)
	if err := s.client.Post(ctx, path, req, nil); err != nil {
		return fmt.Errorf("add time entry tags: %w", err)
	}

	return nil
}

// RemoveTags removes tags from time entries.
func (s *TimeService) RemoveTags(ctx context.Context, teamID string, req TimeEntryTagsRequest) error {
	if teamID == "" {
		return errIDRequired
	}

	if len(req.TimeEntryIDs) == 0 || len(req.Tags) == 0 {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/team/%s/time_entries/tags", teamID)
	if err := s.client.DeleteWithBody(ctx, path, req); err != nil {
		return fmt.Errorf("remove time entry tags: %w", err)
	}

	return nil
}

// RenameTag renames a tag across all time entries.
func (s *TimeService) RenameTag(ctx context.Context, teamID string, req RenameTimeEntryTagRequest) error {
	if teamID == "" {
		return errIDRequired
	}

	if req.Name == "" || req.NewName == "" {
		return errNameRequired
	}

	path := fmt.Sprintf("/v2/team/%s/time_entries/tags", teamID)
	if err := s.client.Put(ctx, path, req, nil); err != nil {
		return fmt.Errorf("rename time entry tag: %w", err)
	}

	return nil
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

// --- TagsService ---

// TagsService handles space tag operations.
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
		return nil, fmt.Errorf("list space tags: %w", err)
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
		return fmt.Errorf("create space tag: %w", err)
	}

	return nil
}

// Update updates a tag in a space.
func (s *TagsService) Update(ctx context.Context, spaceID, tagName string, req EditSpaceTagRequest) error {
	if spaceID == "" || tagName == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/space/%s/tag/%s", spaceID, url.QueryEscape(tagName))
	if err := s.client.Put(ctx, path, req, nil); err != nil {
		return fmt.Errorf("update space tag: %w", err)
	}

	return nil
}

// Delete deletes a tag from a space.
func (s *TagsService) Delete(ctx context.Context, spaceID, tagName string) error {
	if spaceID == "" || tagName == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/space/%s/tag/%s", spaceID, url.QueryEscape(tagName))
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("delete space tag: %w", err)
	}

	return nil
}

// AddToTask adds a tag to a task.
func (s *TagsService) AddToTask(ctx context.Context, taskID, tagName string) error {
	if taskID == "" || tagName == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/task/%s/tag/%s", taskID, url.QueryEscape(tagName))
	if err := s.client.Post(ctx, path, nil, nil); err != nil {
		return fmt.Errorf("add tag to task: %w", err)
	}

	return nil
}

// RemoveFromTask removes a tag from a task.
func (s *TagsService) RemoveFromTask(ctx context.Context, taskID, tagName string) error {
	if taskID == "" || tagName == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/task/%s/tag/%s", taskID, url.QueryEscape(tagName))
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

	var result ChecklistResponse

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

	var result ChecklistResponse

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

// AddItem creates a new item in a checklist.
func (s *ChecklistsService) AddItem(ctx context.Context, checklistID string, req CreateChecklistItemRequest) (*Checklist, error) {
	if checklistID == "" {
		return nil, errIDRequired
	}

	if req.Name == "" {
		return nil, errNameRequired
	}

	var result ChecklistResponse

	path := fmt.Sprintf("/v2/checklist/%s/checklist_item", checklistID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("add checklist item: %w", err)
	}

	return &result.Checklist, nil
}

// UpdateItem updates a checklist item.
func (s *ChecklistsService) UpdateItem(ctx context.Context, checklistID, itemID string, req EditChecklistItemRequest) (*Checklist, error) {
	if checklistID == "" || itemID == "" {
		return nil, errIDRequired
	}

	var result ChecklistResponse

	path := fmt.Sprintf("/v2/checklist/%s/checklist_item/%s", checklistID, itemID)
	if err := s.client.Put(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("update checklist item: %w", err)
	}

	return &result.Checklist, nil
}

// DeleteItem deletes a checklist item.
func (s *ChecklistsService) DeleteItem(ctx context.Context, checklistID, itemID string) error {
	if checklistID == "" || itemID == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/checklist/%s/checklist_item/%s", checklistID, itemID)
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("delete checklist item: %w", err)
	}

	return nil
}

// --- RelationshipsService ---

// RelationshipsService handles task relationship operations.
type RelationshipsService struct {
	client *Client
}

// AddDependency adds a dependency to a task.
func (s *RelationshipsService) AddDependency(ctx context.Context, taskID string, req AddDependencyRequest) error {
	if taskID == "" {
		return errIDRequired
	}

	if req.DependsOn == "" && req.DependencyOf == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/task/%s/dependency", taskID)
	if err := s.client.Post(ctx, path, req, nil); err != nil {
		return fmt.Errorf("add dependency: %w", err)
	}

	return nil
}

// DeleteDependency removes a dependency from a task.
func (s *RelationshipsService) DeleteDependency(ctx context.Context, taskID string, req AddDependencyRequest) error {
	if taskID == "" {
		return errIDRequired
	}

	if req.DependsOn == "" && req.DependencyOf == "" {
		return errIDRequired
	}

	// Build query params
	path := fmt.Sprintf("/v2/task/%s/dependency?", taskID)
	if req.DependsOn != "" {
		path += fmt.Sprintf("depends_on=%s", url.QueryEscape(req.DependsOn))
	}

	if req.DependencyOf != "" {
		if req.DependsOn != "" {
			path += "&"
		}

		path += fmt.Sprintf("dependency_of=%s", url.QueryEscape(req.DependencyOf))
	}

	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("delete dependency: %w", err)
	}

	return nil
}

// AddLink adds a link between two tasks.
func (s *RelationshipsService) AddLink(ctx context.Context, taskID, linkedTaskID string) error {
	if taskID == "" || linkedTaskID == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/task/%s/link/%s", taskID, linkedTaskID)
	if err := s.client.Post(ctx, path, nil, nil); err != nil {
		return fmt.Errorf("add task link: %w", err)
	}

	return nil
}

// DeleteLink removes a link between two tasks.
func (s *RelationshipsService) DeleteLink(ctx context.Context, taskID, linkedTaskID string) error {
	if taskID == "" || linkedTaskID == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/task/%s/link/%s", taskID, linkedTaskID)
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("delete task link: %w", err)
	}

	return nil
}

// --- CustomFieldsService ---

// CustomFieldsService handles custom field operations.
type CustomFieldsService struct {
	client *Client
}

// ListByList returns custom fields for a list.
func (s *CustomFieldsService) ListByList(ctx context.Context, listID string) (*CustomFieldsResponse, error) {
	if listID == "" {
		return nil, errIDRequired
	}

	var result CustomFieldsResponse

	path := fmt.Sprintf("/v2/list/%s/field", listID)
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list custom fields: %w", err)
	}

	return &result, nil
}

// ListByFolder returns custom fields for a folder.
func (s *CustomFieldsService) ListByFolder(ctx context.Context, folderID string) (*CustomFieldsResponse, error) {
	if folderID == "" {
		return nil, errIDRequired
	}

	var result CustomFieldsResponse

	path := fmt.Sprintf("/v2/folder/%s/field", folderID)
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list custom fields: %w", err)
	}

	return &result, nil
}

// ListBySpace returns custom fields for a space.
func (s *CustomFieldsService) ListBySpace(ctx context.Context, spaceID string) (*CustomFieldsResponse, error) {
	if spaceID == "" {
		return nil, errIDRequired
	}

	var result CustomFieldsResponse

	path := fmt.Sprintf("/v2/space/%s/field", spaceID)
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list custom fields: %w", err)
	}

	return &result, nil
}

// ListByTeam returns custom fields for a workspace/team.
func (s *CustomFieldsService) ListByTeam(ctx context.Context, teamID string) (*CustomFieldsResponse, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	var result CustomFieldsResponse

	path := fmt.Sprintf("/v2/team/%s/field", teamID)
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list custom fields: %w", err)
	}

	return &result, nil
}

// Set sets a custom field value on a task.
func (s *CustomFieldsService) Set(ctx context.Context, taskID, fieldID string, value interface{}) error {
	if taskID == "" || fieldID == "" {
		return errIDRequired
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
	if taskID == "" || fieldID == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/task/%s/field/%s", taskID, fieldID)
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("remove custom field: %w", err)
	}

	return nil
}

// --- ViewsService ---

// ViewsService handles view operations.
type ViewsService struct {
	client *Client
}

// ListByTeam returns all views in a workspace/team.
func (s *ViewsService) ListByTeam(ctx context.Context, teamID string) (*ViewsResponse, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/team/%s/view", teamID)

	var result ViewsResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list team views: %w", err)
	}

	return &result, nil
}

// ListBySpace returns all views in a space.
func (s *ViewsService) ListBySpace(ctx context.Context, spaceID string) (*ViewsResponse, error) {
	if spaceID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/space/%s/view", spaceID)

	var result ViewsResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list space views: %w", err)
	}

	return &result, nil
}

// ListByFolder returns all views in a folder.
func (s *ViewsService) ListByFolder(ctx context.Context, folderID string) (*ViewsResponse, error) {
	if folderID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/folder/%s/view", folderID)

	var result ViewsResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list folder views: %w", err)
	}

	return &result, nil
}

// ListByList returns all views in a list.
func (s *ViewsService) ListByList(ctx context.Context, listID string) (*ViewsResponse, error) {
	if listID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/list/%s/view", listID)

	var result ViewsResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list views for list: %w", err)
	}

	return &result, nil
}

// Get returns a view by ID.
func (s *ViewsService) Get(ctx context.Context, viewID string) (*View, error) {
	if viewID == "" {
		return nil, errIDRequired
	}

	var result ViewResponse

	path := fmt.Sprintf("/v2/view/%s", viewID)
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get view: %w", err)
	}

	return &result.View, nil
}

// Tasks returns tasks matching the view's filters.
func (s *ViewsService) Tasks(ctx context.Context, viewID string, page int) (*ViewTasksResponse, error) {
	if viewID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/view/%s/task", viewID)

	if page > 0 {
		path += fmt.Sprintf("?page=%d", page)
	}

	var result ViewTasksResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get view tasks: %w", err)
	}

	return &result, nil
}

// CreateInTeam creates a view in a workspace/team.
func (s *ViewsService) CreateInTeam(ctx context.Context, teamID string, req CreateViewRequest) (*View, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	if req.Name == "" || req.Type == "" {
		return nil, errNameRequired
	}

	var result ViewResponse

	path := fmt.Sprintf("/v2/team/%s/view", teamID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create team view: %w", err)
	}

	return &result.View, nil
}

// CreateInSpace creates a view in a space.
func (s *ViewsService) CreateInSpace(ctx context.Context, spaceID string, req CreateViewRequest) (*View, error) {
	if spaceID == "" {
		return nil, errIDRequired
	}

	if req.Name == "" || req.Type == "" {
		return nil, errNameRequired
	}

	var result ViewResponse

	path := fmt.Sprintf("/v2/space/%s/view", spaceID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create space view: %w", err)
	}

	return &result.View, nil
}

// CreateInFolder creates a view in a folder.
func (s *ViewsService) CreateInFolder(ctx context.Context, folderID string, req CreateViewRequest) (*View, error) {
	if folderID == "" {
		return nil, errIDRequired
	}

	if req.Name == "" || req.Type == "" {
		return nil, errNameRequired
	}

	var result ViewResponse

	path := fmt.Sprintf("/v2/folder/%s/view", folderID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create folder view: %w", err)
	}

	return &result.View, nil
}

// CreateInList creates a view in a list.
func (s *ViewsService) CreateInList(ctx context.Context, listID string, req CreateViewRequest) (*View, error) {
	if listID == "" {
		return nil, errIDRequired
	}

	if req.Name == "" || req.Type == "" {
		return nil, errNameRequired
	}

	var result ViewResponse

	path := fmt.Sprintf("/v2/list/%s/view", listID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create list view: %w", err)
	}

	return &result.View, nil
}

// Update updates a view.
func (s *ViewsService) Update(ctx context.Context, viewID string, req UpdateViewRequest) (*View, error) {
	if viewID == "" {
		return nil, errIDRequired
	}

	var result ViewResponse

	path := fmt.Sprintf("/v2/view/%s", viewID)
	if err := s.client.Put(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("update view: %w", err)
	}

	return &result.View, nil
}

// Delete deletes a view.
func (s *ViewsService) Delete(ctx context.Context, viewID string) error {
	if viewID == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/view/%s", viewID)
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("delete view: %w", err)
	}

	return nil
}

// --- WebhooksService ---

// WebhooksService handles webhook operations.
type WebhooksService struct {
	client *Client
}

// List returns all webhooks for a team.
func (s *WebhooksService) List(ctx context.Context, teamID string) (*WebhooksResponse, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/team/%s/webhook", teamID)

	var result WebhooksResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list webhooks: %w", err)
	}

	return &result, nil
}

// Create creates a new webhook for a team.
func (s *WebhooksService) Create(ctx context.Context, teamID string, req CreateWebhookRequest) (*Webhook, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	if req.Endpoint == "" {
		return nil, errEndpointRequired
	}

	if len(req.Events) == 0 {
		return nil, errEventsRequired
	}

	var result Webhook

	path := fmt.Sprintf("/v2/team/%s/webhook", teamID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create webhook: %w", err)
	}

	return &result, nil
}

// Update updates a webhook.
func (s *WebhooksService) Update(ctx context.Context, webhookID string, req UpdateWebhookRequest) (*Webhook, error) {
	if webhookID == "" {
		return nil, errIDRequired
	}

	var result Webhook

	path := fmt.Sprintf("/v2/webhook/%s", webhookID)
	if err := s.client.Put(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("update webhook: %w", err)
	}

	return &result, nil
}

// Delete deletes a webhook.
func (s *WebhooksService) Delete(ctx context.Context, webhookID string) error {
	if webhookID == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/webhook/%s", webhookID)
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("delete webhook: %w", err)
	}

	return nil
}

// --- GoalsService ---

// GoalsService handles goal and key result operations.
type GoalsService struct {
	client *Client
}

// List returns all goals for a team.
func (s *GoalsService) List(ctx context.Context, teamID string, includeCompleted bool) (*GoalsResponse, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/team/%s/goal", teamID)
	if includeCompleted {
		path += "?include_closed=true"
	}

	var result GoalsResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list goals: %w", err)
	}

	return &result, nil
}

// Get returns a single goal with key results.
func (s *GoalsService) Get(ctx context.Context, goalID string) (*Goal, error) {
	if goalID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/goal/%s", goalID)

	var result GoalResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get goal: %w", err)
	}

	return &result.Goal, nil
}

// Create creates a new goal.
func (s *GoalsService) Create(ctx context.Context, teamID string, req CreateGoalRequest) (*Goal, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	if req.Name == "" {
		return nil, errNameRequired
	}

	var result GoalResponse

	path := fmt.Sprintf("/v2/team/%s/goal", teamID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create goal: %w", err)
	}

	return &result.Goal, nil
}

// Update updates a goal.
func (s *GoalsService) Update(ctx context.Context, goalID string, req UpdateGoalRequest) (*Goal, error) {
	if goalID == "" {
		return nil, errIDRequired
	}

	var result GoalResponse

	path := fmt.Sprintf("/v2/goal/%s", goalID)
	if err := s.client.Put(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("update goal: %w", err)
	}

	return &result.Goal, nil
}

// Delete deletes a goal and all its key results.
func (s *GoalsService) Delete(ctx context.Context, goalID string) error {
	if goalID == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/goal/%s", goalID)
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("delete goal: %w", err)
	}

	return nil
}

// CreateKeyResult creates a key result for a goal.
func (s *GoalsService) CreateKeyResult(ctx context.Context, goalID string, req CreateKeyResultRequest) (*KeyResult, error) {
	if goalID == "" {
		return nil, errIDRequired
	}

	if req.Name == "" {
		return nil, errNameRequired
	}

	if req.Type == "" {
		return nil, errKeyResultTypeRequired
	}

	var result KeyResultResponse

	path := fmt.Sprintf("/v2/goal/%s/key_result", goalID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create key result: %w", err)
	}

	return &result.KeyResult, nil
}

// UpdateKeyResult updates a key result's progress.
func (s *GoalsService) UpdateKeyResult(ctx context.Context, keyResultID string, req EditKeyResultRequest) (*KeyResult, error) {
	if keyResultID == "" {
		return nil, errIDRequired
	}

	var result KeyResultResponse

	path := fmt.Sprintf("/v2/key_result/%s", keyResultID)
	if err := s.client.Put(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("update key result: %w", err)
	}

	return &result.KeyResult, nil
}

// DeleteKeyResult deletes a key result.
func (s *GoalsService) DeleteKeyResult(ctx context.Context, keyResultID string) error {
	if keyResultID == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/key_result/%s", keyResultID)
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("delete key result: %w", err)
	}

	return nil
}

// --- UsersService ---

// UsersService handles workspace user operations.
type UsersService struct {
	client *Client
}

// Get returns a user's details.
func (s *UsersService) Get(ctx context.Context, teamID string, userID int) (*UserDetail, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v2/team/%s/user/%d", teamID, userID)

	var result UserResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	return &result.User, nil
}

// Invite invites a user to the workspace by email.
func (s *UsersService) Invite(ctx context.Context, teamID string, req InviteUserRequest) (*UserDetail, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	if req.Email == "" {
		return nil, errEmailRequired
	}

	var result UserResponse

	path := fmt.Sprintf("/v2/team/%s/user", teamID)
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("invite user: %w", err)
	}

	return &result.User, nil
}

// Update updates a user's role on the workspace.
func (s *UsersService) Update(ctx context.Context, teamID string, userID int, req EditUserRequest) (*UserDetail, error) {
	if teamID == "" {
		return nil, errIDRequired
	}

	var result UserResponse

	path := fmt.Sprintf("/v2/team/%s/user/%d", teamID, userID)
	if err := s.client.Put(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	return &result.User, nil
}

// Remove removes a user from the workspace.
func (s *UsersService) Remove(ctx context.Context, teamID string, userID int) error {
	if teamID == "" {
		return errIDRequired
	}

	path := fmt.Sprintf("/v2/team/%s/user/%d", teamID, userID)
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("remove user: %w", err)
	}

	return nil
}
