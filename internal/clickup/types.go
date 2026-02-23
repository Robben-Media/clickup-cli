package clickup

import "encoding/json"

// Task represents a ClickUp task.
type Task struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Status      TaskStatus `json:"status"`
	Priority    *Priority  `json:"priority"`
	DueDate     string     `json:"due_date,omitempty"`
	Assignees   []User     `json:"assignees,omitempty"`
	URL         string     `json:"url,omitempty"`
	List        ListRef    `json:"list,omitempty"`
	Folder      FolderRef  `json:"folder,omitempty"`
	Space       SpaceRef   `json:"space"`
	Tags        []Tag      `json:"tags,omitempty"`
}

// TaskStatus represents a task's status.
type TaskStatus struct {
	Status string `json:"status"`
	Color  string `json:"color,omitempty"`
}

// Priority represents a task's priority.
type Priority struct {
	ID    string `json:"id"`
	Name  string `json:"priority"`
	Color string `json:"color,omitempty"`
}

// User represents a ClickUp user.
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email,omitempty"`
}

// UserDetail represents a full user object with role information.
type UserDetail struct {
	ID             int         `json:"id"`
	Username       string      `json:"username"`
	Email          string      `json:"email"`
	Role           int         `json:"role"`                     // 1=owner, 2=admin, 3=member, 4=guest
	ProfilePicture string      `json:"profilePicture,omitempty"` //nolint:tagliatelle // ClickUp API uses camelCase
	InvitedBy      *UserRef    `json:"invited_by,omitempty"`
	DateInvited    string      `json:"date_invited,omitempty"`
	DateJoined     string      `json:"date_joined,omitempty"`
	CustomRole     *CustomRole `json:"custom_role,omitempty"`
}

// UserRef is a reference to a user (used in invited_by).
type UserRef struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email,omitempty"`
}

// UserResponse is the response for a single user.
type UserResponse struct {
	User UserDetail `json:"user"`
}

// InviteUserRequest is the request body for inviting a user.
type InviteUserRequest struct {
	Email           string `json:"email"`
	Admin           bool   `json:"admin,omitempty"`
	CustomRoleIDs   []int  `json:"custom_role_ids,omitempty"`
	Locale          string `json:"locale,omitempty"`
	SendInviteEmail *bool  `json:"send_invite_email,omitempty"`
}

// EditUserRequest is the request body for editing a user.
type EditUserRequest struct {
	Username       string `json:"username,omitempty"`
	Admin          bool   `json:"admin,omitempty"`
	CustomRoleIDs  []int  `json:"custom_role_ids,omitempty"`
	AddCustomRoles []int  `json:"add_custom_role_ids,omitempty"`
	RemCustomRoles []int  `json:"remove_custom_role_ids,omitempty"`
}

// Space represents a ClickUp space.
type Space struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SpaceDetail represents a full space object with statuses and features.
type SpaceDetail struct {
	ID                string        `json:"id"`
	Name              string        `json:"name"`
	Private           bool          `json:"private"`
	Color             string        `json:"color,omitempty"`
	Statuses          []SpaceStatus `json:"statuses,omitempty"`
	MultipleAssignees bool          `json:"multiple_assignees"`
	Features          SpaceFeatures `json:"features,omitempty"`
}

// SpaceStatus represents a status in a space.
type SpaceStatus struct {
	Status     string `json:"status"`
	Color      string `json:"color"`
	OrderIndex int    `json:"orderindex"`
}

// SpaceFeatures contains feature toggles for a space.
type SpaceFeatures struct {
	DueDates     FeatureToggle `json:"due_dates"`
	TimeTracking FeatureToggle `json:"time_tracking"`
	Tags         FeatureToggle `json:"tags"`
	Checklists   FeatureToggle `json:"checklists"`
}

// FeatureToggle represents a feature enabled/disabled state.
type FeatureToggle struct {
	Enabled bool `json:"enabled"`
}

// CreateSpaceRequest is the request body for creating a space.
type CreateSpaceRequest struct {
	Name              string         `json:"name"`
	MultipleAssignees bool           `json:"multiple_assignees,omitempty"`
	Features          *SpaceFeatures `json:"features,omitempty"`
}

// UpdateSpaceRequest is the request body for updating a space.
type UpdateSpaceRequest struct {
	Name              string `json:"name,omitempty"`
	Color             string `json:"color,omitempty"`
	Private           *bool  `json:"private,omitempty"`
	MultipleAssignees *bool  `json:"multiple_assignees,omitempty"`
}

// List represents a ClickUp list.
type List struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ListDetail represents a full list object with task count and references.
type ListDetail struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Content   string    `json:"content,omitempty"`
	DueDate   string    `json:"due_date,omitempty"`
	Priority  *Priority `json:"priority,omitempty"`
	Assignee  *User     `json:"assignee,omitempty"`
	TaskCount int       `json:"task_count"`
	Folder    FolderRef `json:"folder"`
	Space     SpaceRef  `json:"space"`
}

// CreateListRequest is the request body for creating a list.
type CreateListRequest struct {
	Name     string `json:"name"`
	Content  string `json:"content,omitempty"`
	DueDate  int64  `json:"due_date,omitempty"`
	Priority int    `json:"priority,omitempty"`
	Assignee int    `json:"assignee,omitempty"`
}

// UpdateListRequest is the request body for updating a list.
type UpdateListRequest struct {
	Name          string `json:"name,omitempty"`
	Content       string `json:"content,omitempty"`
	DueDate       int64  `json:"due_date,omitempty"`
	Priority      int    `json:"priority,omitempty"`
	Assignee      int    `json:"assignee,omitempty"`
	UnsetAssignee bool   `json:"unset_assignee,omitempty"`
}

// CreateListFromTemplateRequest is the request body for creating a list from template.
type CreateListFromTemplateRequest struct {
	Name string `json:"name"`
}

// Folder represents a ClickUp folder.
type Folder struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Lists []List `json:"lists,omitempty"`
}

// FolderDetail represents a full folder object with task count and lists.
type FolderDetail struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	OrderIndex       int      `json:"orderindex"`
	OverrideStatuses bool     `json:"override_statuses"`
	Hidden           bool     `json:"hidden"`
	Space            SpaceRef `json:"space"`
	TaskCount        string   `json:"task_count"`
	Lists            []List   `json:"lists,omitempty"`
}

// CreateFolderRequest is the request body for creating a folder.
type CreateFolderRequest struct {
	Name string `json:"name"`
}

// UpdateFolderRequest is the request body for updating a folder.
type UpdateFolderRequest struct {
	Name string `json:"name,omitempty"`
}

// CreateFolderFromTemplateRequest is the request body for creating a folder from template.
type CreateFolderFromTemplateRequest struct {
	Name string `json:"name,omitempty"`
}

// Comment represents a ClickUp comment.
type Comment struct {
	ID   json.Number `json:"id"`
	Text string      `json:"comment_text"`
	User User        `json:"user"`
	Date string      `json:"date"`
}

// TimeEntry represents a ClickUp time entry.
type TimeEntry struct {
	ID       json.Number `json:"id"`
	Task     TaskRef     `json:"task"`
	Duration json.Number `json:"duration"`
	Start    json.Number `json:"start"`
	End      json.Number `json:"end"`
}

// Tag represents a ClickUp tag.
type Tag struct {
	Name string `json:"name"`
}

// ListRef is a reference to a list within a task.
type ListRef struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

// FolderRef is a reference to a folder within a task.
type FolderRef struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

// SpaceRef is a reference to a space within a task.
type SpaceRef struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

// TaskRef is a reference to a task within a time entry.
type TaskRef struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

// Member represents a team member in ClickUp.
type Member struct {
	User User `json:"user"`
}

// SharedHierarchyResponse is the response from the shared hierarchy endpoint.
type SharedHierarchyResponse struct {
	Shared SharedResources `json:"shared"`
}

// SharedResources contains shared tasks, lists, and folders.
type SharedResources struct {
	Tasks   []TaskRef   `json:"tasks,omitempty"`
	Lists   []ListRef   `json:"lists,omitempty"`
	Folders []FolderRef `json:"folders,omitempty"`
}

// --- Request/Response types ---

// CreateTaskRequest is the request body for creating a task.
type CreateTaskRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Assignees   []int  `json:"assignees,omitempty"`
	Priority    *int   `json:"priority,omitempty"`
	DueDate     string `json:"due_date,omitempty"`
}

// TaskAssigneesUpdate is the assignee update payload for the task update endpoint.
// ClickUp expects assignees as an object with add/rem arrays.
type TaskAssigneesUpdate struct {
	Add []int `json:"add,omitempty"`
	Rem []int `json:"rem,omitempty"`
}

// UpdateTaskRequest is the request body for updating a task.
type UpdateTaskRequest struct {
	Name      string               `json:"name,omitempty"`
	Status    string               `json:"status,omitempty"`
	Assignees *TaskAssigneesUpdate `json:"assignees,omitempty"`
	Priority  *int                 `json:"priority,omitempty"`
}

// CreateCommentRequest is the request body for creating a comment.
type CreateCommentRequest struct {
	CommentText string `json:"comment_text"`
}

// CreateTimeEntryRequest is the request body for creating a time entry.
type CreateTimeEntryRequest struct {
	Duration int64 `json:"duration"`
}

// TasksListResponse is the response for listing tasks.
type TasksListResponse struct {
	Tasks []Task `json:"tasks"`
}

// SpacesListResponse is the response for listing spaces.
type SpacesListResponse struct {
	Spaces []Space `json:"spaces"`
}

// FoldersListResponse is the response for listing folders.
type FoldersListResponse struct {
	Folders []Folder `json:"folders"`
}

// ListsListResponse is the response for listing lists in a folder.
type ListsListResponse struct {
	Lists []List `json:"lists"`
}

// FolderlessListsResponse is the response for listing folderless lists.
type FolderlessListsResponse struct {
	Lists []List `json:"lists"`
}

// CommentsListResponse is the response for listing comments.
type CommentsListResponse struct {
	Comments []Comment `json:"comments"`
}

// TimeEntriesListResponse is the response for listing time entries.
type TimeEntriesListResponse struct {
	Data []TimeEntry `json:"data"`
}

// MembersListResponse is the response for listing team members.
type MembersListResponse struct {
	Members []Member `json:"members"`
}

// --- User Group types ---

// UserGroup represents a ClickUp user group.
type UserGroup struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Members []User `json:"members,omitempty"`
}

// CreateUserGroupRequest is the request body for creating a user group.
type CreateUserGroupRequest struct {
	Name    string `json:"name"`
	Members []int  `json:"members,omitempty"`
}

// UserGroupMembersUpdate represents add/remove member operations.
type UserGroupMembersUpdate struct {
	Add []int `json:"add,omitempty"`
	Rem []int `json:"rem,omitempty"`
}

// UpdateUserGroupRequest is the request body for updating a user group.
type UpdateUserGroupRequest struct {
	Name    string                  `json:"name,omitempty"`
	Members *UserGroupMembersUpdate `json:"members,omitempty"`
}

// UserGroupsResponse is the response for listing user groups.
type UserGroupsResponse struct {
	Groups []UserGroup `json:"groups"`
}

// CustomRole represents a ClickUp custom role.
type CustomRole struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions,omitempty"`
}

// CustomRolesResponse is the response for listing custom roles.
type CustomRolesResponse struct {
	CustomRoles []CustomRole `json:"custom_roles"`
}

// --- Guest types ---

// Guest represents a ClickUp guest user.
type Guest struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	TasksCount   int    `json:"tasks_count,omitempty"`
	ListsCount   int    `json:"lists_count,omitempty"`
	FoldersCount int    `json:"folders_count,omitempty"`
}

// GuestResponse is the response for getting a guest.
type GuestResponse struct {
	Guest Guest `json:"guest"`
}

// InviteGuestRequest is the request body for inviting a guest.
type InviteGuestRequest struct {
	Email               string `json:"email"`
	CanEditTags         bool   `json:"can_edit_tags,omitempty"`
	CanSeeTimeSpent     bool   `json:"can_see_time_spent,omitempty"`
	CanSeeTimeEstimated bool   `json:"can_see_time_estimated,omitempty"`
}

// EditGuestRequest is the request body for editing a guest.
type EditGuestRequest struct {
	CanEditTags         *bool `json:"can_edit_tags,omitempty"`
	CanSeeTimeSpent     *bool `json:"can_see_time_spent,omitempty"`
	CanSeeTimeEstimated *bool `json:"can_see_time_estimated,omitempty"`
}

// AddGuestToResourceRequest is the request body for adding a guest to a resource.
type AddGuestToResourceRequest struct {
	PermissionLevel string `json:"permission_level"` // "read", "comment", "edit", "create"
}

// --- Task Template types ---

// TaskTemplate represents a ClickUp task template.
type TaskTemplate struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// TaskTemplatesResponse is the response for listing task templates.
type TaskTemplatesResponse struct {
	Templates []TaskTemplate `json:"templates"`
}

// --- Custom Task Type types ---

// CustomTaskType represents a ClickUp custom task type.
type CustomTaskType struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	NamePlural  string `json:"name_plural"`
	Description string `json:"description,omitempty"`
}

// CustomTaskTypesResponse is the response for listing custom task types.
type CustomTaskTypesResponse struct {
	CustomItems []CustomTaskType `json:"custom_items"`
}

// --- Legacy Time Tracking types ---

// LegacyTimeInterval represents a time interval in the legacy time tracking system.
type LegacyTimeInterval struct {
	ID        string `json:"id"`
	Start     int64  `json:"start"`
	End       int64  `json:"end"`
	Time      int64  `json:"time"`
	Source    string `json:"source,omitempty"`
	DateAdded string `json:"date_added,omitempty"`
}

// LegacyTimeResponse is the response for listing legacy time intervals.
type LegacyTimeResponse struct {
	Data []LegacyTimeInterval `json:"data"`
}

// TrackTimeRequest is the request body for tracking time.
type TrackTimeRequest struct {
	Start int64 `json:"start,omitempty"`
	End   int64 `json:"end,omitempty"`
	Time  int64 `json:"time"`
}

// EditTimeRequest is the request body for editing time.
type EditTimeRequest struct {
	Start int64 `json:"start,omitempty"`
	End   int64 `json:"end,omitempty"`
	Time  int64 `json:"time,omitempty"`
}

// TrackTimeResponse is the response for tracking time.
type TrackTimeResponse struct {
	ID string `json:"id"`
}

// --- Audit Log types ---

// AuditLogQuery is the request body for querying audit logs.
type AuditLogQuery struct {
	StartDate int64  `json:"start_date,omitempty"`
	EndDate   int64  `json:"end_date,omitempty"`
	EventType string `json:"event_type,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

// AuditLogEntry represents a single audit log entry.
type AuditLogEntry struct {
	ID           string      `json:"id"`
	EventType    string      `json:"event_type"`
	UserID       string      `json:"user_id"`
	Timestamp    json.Number `json:"timestamp"`
	ResourceType string      `json:"resource_type"`
	ResourceID   string      `json:"resource_id"`
	Details      any         `json:"details,omitempty"`
}

// AuditLogsResponse is the response for querying audit logs.
type AuditLogsResponse struct {
	AuditLogs []AuditLogEntry `json:"audit_logs"`
}

// --- ACL types ---

// UpdateACLRequest is the request body for updating ACLs.
type UpdateACLRequest struct {
	Private *bool  `json:"private,omitempty"`
	Sharing string `json:"sharing,omitempty"` // "open" or "closed"
}

// --- Workspace types ---

// WorkspacesResponse is the response for listing workspaces.
type WorkspacesResponse struct {
	Teams []Workspace `json:"teams"`
}

// Workspace represents a ClickUp workspace (team).
type Workspace struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Color   string   `json:"color,omitempty"`
	Members []Member `json:"members,omitempty"`
}

// WorkspacePlanResponse is the response for getting workspace plan.
type WorkspacePlanResponse struct {
	TeamID   string `json:"team_id"`
	PlanID   int    `json:"plan_id"`
	PlanName string `json:"plan_name"`
}

// WorkspaceSeatsResponse is the response for getting workspace seats.
type WorkspaceSeatsResponse struct {
	Members SeatInfo `json:"members"`
	Guests  SeatInfo `json:"guests"`
}

// SeatInfo contains seat usage information.
type SeatInfo struct {
	FilledSeats int `json:"filled_member_seats"`
	TotalSeats  int `json:"total_member_seats"`
	EmptySeats  int `json:"empty_member_seats"`
}

// --- Auth types ---

// AuthorizedUserResponse is the response for getting the authorized user.
type AuthorizedUserResponse struct {
	User AuthUser `json:"user"`
}

// AuthUser represents the authenticated user.
type AuthUser struct {
	ID             int    `json:"id"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	Color          string `json:"color,omitempty"`
	ProfilePicture string `json:"profilePicture,omitempty"` //nolint:tagliatelle // ClickUp API uses camelCase
}

// OAuthTokenRequest is the request body for OAuth token exchange.
type OAuthTokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
}

// OAuthTokenResponse is the response from OAuth token exchange.
type OAuthTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// --- Comment extension types ---

// UpdateCommentRequest is the request body for updating a comment.
type UpdateCommentRequest struct {
	CommentText string `json:"comment_text,omitempty"`
	Assignee    int    `json:"assignee,omitempty"`
	Resolved    *bool  `json:"resolved,omitempty"`
}

// CreateListCommentRequest is the request body for creating a list comment.
type CreateListCommentRequest struct {
	CommentText string `json:"comment_text"`
	Assignee    int    `json:"assignee,omitempty"`
}

// CreateViewCommentRequest is the request body for creating a view comment.
type CreateViewCommentRequest struct {
	CommentText string `json:"comment_text"`
	Assignee    int    `json:"assignee,omitempty"`
}

// ThreadedCommentsResponse is the response for getting threaded replies.
type ThreadedCommentsResponse struct {
	Comments []Comment `json:"comments"`
}

// PostSubtype represents a comment post subtype.
type PostSubtype struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// PostSubtypesResponse is the response for getting post subtypes.
type PostSubtypesResponse struct {
	Subtypes []PostSubtype `json:"subtypes"`
}

// --- Task extension types ---

// FilteredTeamTasksParams contains query parameters for searching tasks across a workspace.
type FilteredTeamTasksParams struct {
	Page          int      `url:"page,omitempty"`
	OrderBy       string   `url:"order_by,omitempty"`
	Reverse       bool     `url:"reverse,omitempty"`
	Subtasks      bool     `url:"subtasks,omitempty"`
	Statuses      []string `url:"statuses[],omitempty"`
	IncludeClosed bool     `url:"include_closed,omitempty"`
	Assignees     []int    `url:"assignees[],omitempty"`
	Tags          []string `url:"tags[],omitempty"`
	DueDateGt     int64    `url:"due_date_gt,omitempty"`
	DueDateLt     int64    `url:"due_date_lt,omitempty"`
	DateCreatedGt int64    `url:"date_created_gt,omitempty"`
	DateCreatedLt int64    `url:"date_created_lt,omitempty"`
	DateUpdatedGt int64    `url:"date_updated_gt,omitempty"`
	DateUpdatedLt int64    `url:"date_updated_lt,omitempty"`
}

// FilteredTeamTasksResponse is the response for filtered team tasks search.
type FilteredTeamTasksResponse struct {
	Tasks []Task `json:"tasks"`
}

// TimeInStatusResponse contains time-in-status data for a single task.
type TimeInStatusResponse struct {
	CurrentStatus *StatusTime  `json:"current_status,omitempty"`
	StatusHistory []StatusTime `json:"status_history,omitempty"`
}

// StatusTime represents time spent in a status.
type StatusTime struct {
	Status    string    `json:"status"`
	Color     string    `json:"color,omitempty"`
	TotalTime TimeValue `json:"total_time"`
}

// TimeValue represents a duration and start time.
type TimeValue struct {
	ByMinute int64  `json:"by_minute"`
	Since    string `json:"since,omitempty"`
}

// BulkTimeInStatusResponse maps task IDs to their time-in-status data.
type BulkTimeInStatusResponse map[string]TimeInStatusResponse

// MergeTasksRequest is the request body for merging tasks.
type MergeTasksRequest struct {
	MergedTaskIDs []string `json:"merged_task_ids"`
}

// MergeTasksResponse is the response from merging tasks.
type MergeTasksResponse struct {
	ID string `json:"id"`
}

// MoveTaskResponse is the response from moving a task.
type MoveTaskResponse struct {
	Status  string `json:"status"`
	TaskID  string `json:"task_id"`
	ListID  string `json:"list_id"`
	Message string `json:"message,omitempty"`
}

// CreateTaskFromTemplateRequest is the request body for creating a task from template.
type CreateTaskFromTemplateRequest struct {
	Name string `json:"name,omitempty"`
}

// --- Time Tracking extension types ---

// TimeEntryDetail is a full time entry with all fields.
type TimeEntryDetail struct {
	ID          json.Number `json:"id"`
	Task        TaskRef     `json:"task"`
	Wid         string      `json:"wid"`
	User        User        `json:"user"`
	Billable    bool        `json:"billable"`
	Start       json.Number `json:"start"`
	End         json.Number `json:"end"`
	Duration    json.Number `json:"duration"`
	Description string      `json:"description"`
	Tags        []Tag       `json:"tags"`
}

// TimeEntryDetailResponse is the response for get/current time entry.
type TimeEntryDetailResponse struct {
	Data TimeEntryDetail `json:"data"`
}

// StartTimeEntryRequest is the request body for starting a timer.
type StartTimeEntryRequest struct {
	TaskID      string `json:"tid,omitempty"`
	Description string `json:"description,omitempty"`
	Billable    bool   `json:"billable,omitempty"`
	Tags        []Tag  `json:"tags,omitempty"`
}

// UpdateTimeEntryRequest is the request body for updating a time entry.
type UpdateTimeEntryRequest struct {
	Description string `json:"description,omitempty"`
	Duration    int64  `json:"duration,omitempty"`
	Start       int64  `json:"start,omitempty"`
	End         int64  `json:"end,omitempty"`
	Billable    *bool  `json:"billable,omitempty"`
	TagAction   string `json:"tag_action,omitempty"` // "add" or "remove"
	Tags        []Tag  `json:"tags,omitempty"`
}

// TimeEntryTagsRequest is the request body for tag operations on time entries.
type TimeEntryTagsRequest struct {
	TimeEntryIDs []string `json:"time_entry_ids"`
	Tags         []Tag    `json:"tags"`
}

// RenameTimeEntryTagRequest is the request body for renaming a tag.
type RenameTimeEntryTagRequest struct {
	Name    string `json:"name"`
	NewName string `json:"new_name"`
}

// TimeEntryHistoryResponse is the response for time entry history.
type TimeEntryHistoryResponse struct {
	Data []TimeEntryHistoryItem `json:"data"`
}

// TimeEntryHistoryItem represents a single change in time entry history.
type TimeEntryHistoryItem struct {
	ID     string `json:"id"`
	Field  string `json:"field"`
	Before string `json:"before"`
	After  string `json:"after"`
	Date   string `json:"date"`
	User   User   `json:"user"`
}

// TimeEntryTag represents a time entry tag.
type TimeEntryTag struct {
	Name string `json:"name"`
}

// TimeEntryTagsResponse is the response for listing time entry tags.
type TimeEntryTagsResponse struct {
	Data []TimeEntryTag `json:"data"`
}

// --- Space Tag types ---

// SpaceTag represents a tag in a space with colors.
type SpaceTag struct {
	Name  string `json:"name"`
	TagFg string `json:"tag_fg,omitempty"`
	TagBg string `json:"tag_bg,omitempty"`
}

// SpaceTagsResponse is the response for listing space tags.
type SpaceTagsResponse struct {
	Tags []SpaceTag `json:"tags"`
}

// CreateSpaceTagRequest is the request body for creating a space tag.
type CreateSpaceTagRequest struct {
	Tag SpaceTag `json:"tag"`
}

// EditSpaceTagRequest is the request body for editing a space tag.
type EditSpaceTagRequest struct {
	Tag SpaceTag `json:"tag"`
}

// --- Checklist types ---

// Checklist represents a task checklist.
type Checklist struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	OrderIndex int             `json:"orderindex"`
	Items      []ChecklistItem `json:"items,omitempty"`
}

// ChecklistItem represents an item in a checklist.
type ChecklistItem struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Resolved   bool   `json:"resolved"`
	Assignee   *User  `json:"assignee,omitempty"`
	Parent     string `json:"parent,omitempty"`
	OrderIndex int    `json:"orderindex"`
}

// ChecklistResponse wraps a single checklist response.
type ChecklistResponse struct {
	Checklist Checklist `json:"checklist"`
}

// CreateChecklistRequest is the request body for creating a checklist.
type CreateChecklistRequest struct {
	Name string `json:"name"`
}

// EditChecklistRequest is the request body for editing a checklist.
type EditChecklistRequest struct {
	Name     string `json:"name,omitempty"`
	Position int    `json:"position,omitempty"`
}

// CreateChecklistItemRequest is the request body for creating a checklist item.
type CreateChecklistItemRequest struct {
	Name     string `json:"name"`
	Assignee int    `json:"assignee,omitempty"`
}

// EditChecklistItemRequest is the request body for editing a checklist item.
type EditChecklistItemRequest struct {
	Name     string `json:"name,omitempty"`
	Resolved *bool  `json:"resolved,omitempty"`
	Assignee int    `json:"assignee,omitempty"`
	Parent   string `json:"parent,omitempty"`
}

// --- Task Relationship types ---

// AddDependencyRequest is the request body for adding a task dependency.
type AddDependencyRequest struct {
	DependsOn    string `json:"depends_on,omitempty"`
	DependencyOf string `json:"dependency_of,omitempty"`
}

// --- Custom Field types ---

// CustomField represents a custom field definition.
type CustomField struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Type       string      `json:"type"`
	TypeConfig interface{} `json:"type_config,omitempty"`
	Required   bool        `json:"required"`
}

// CustomFieldsResponse is the response for listing custom fields.
type CustomFieldsResponse struct {
	Fields []CustomField `json:"fields"`
}

// SetCustomFieldRequest is the request body for setting a custom field value.
type SetCustomFieldRequest struct {
	Value interface{} `json:"value"`
}

// --- View types ---

// View represents a saved view configuration.
type View struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Type      string     `json:"type"`
	Parent    ViewParent `json:"parent"`
	Protected bool       `json:"protected"`
}

// ViewParent represents the parent resource of a view.
type ViewParent struct {
	ID   string `json:"id"`
	Type int    `json:"type"` // 7=space, 5=folder, 6=list, etc.
}

// ViewsResponse is the response for listing views.
type ViewsResponse struct {
	Views []View `json:"views"`
}

// ViewResponse wraps a single view response.
type ViewResponse struct {
	View View `json:"view"`
}

// ViewTasksResponse is the response for getting tasks in a view.
type ViewTasksResponse struct {
	Tasks []Task `json:"tasks"`
}

// CreateViewRequest is the request body for creating a view.
type CreateViewRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// UpdateViewRequest is the request body for updating a view.
type UpdateViewRequest struct {
	Name string `json:"name,omitempty"`
}

// --- Webhook types ---

// Webhook represents a ClickUp webhook.
type Webhook struct {
	ID       string   `json:"id"`
	UserID   int      `json:"userid"`
	TeamID   string   `json:"team_id"`
	Endpoint string   `json:"endpoint"`
	Events   []string `json:"events"`
	Status   string   `json:"status"`
	SpaceID  string   `json:"space_id,omitempty"`
	FolderID string   `json:"folder_id,omitempty"`
	ListID   string   `json:"list_id,omitempty"`
	TaskID   string   `json:"task_id,omitempty"`
}

// WebhooksResponse is the response for listing webhooks.
type WebhooksResponse struct {
	Webhooks []Webhook `json:"webhooks"`
}

// CreateWebhookRequest is the request body for creating a webhook.
type CreateWebhookRequest struct {
	Endpoint string   `json:"endpoint"`
	Events   []string `json:"events"`
	SpaceID  string   `json:"space_id,omitempty"`
	FolderID string   `json:"folder_id,omitempty"`
	ListID   string   `json:"list_id,omitempty"`
	TaskID   string   `json:"task_id,omitempty"`
}

// UpdateWebhookRequest is the request body for updating a webhook.
type UpdateWebhookRequest struct {
	Endpoint string   `json:"endpoint,omitempty"`
	Events   []string `json:"events,omitempty"`
	Status   string   `json:"status,omitempty"` // "active" or "inactive"
}

// --- Goal types ---

// Goal represents a ClickUp goal.
type Goal struct {
	ID               string      `json:"id"`
	Name             string      `json:"name"`
	Description      string      `json:"description,omitempty"`
	DateCreated      string      `json:"date_created,omitempty"`
	DueDate          string      `json:"due_date,omitempty"`
	PercentCompleted int         `json:"percent_completed"`
	Color            string      `json:"color,omitempty"`
	KeyResults       []KeyResult `json:"key_results,omitempty"`
	Owners           []User      `json:"owners,omitempty"`
}

// KeyResult represents a key result for a goal.
type KeyResult struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"` // number, currency, boolean, percentage, automatic
	StepsStart   int    `json:"steps_start"`
	StepsEnd     int    `json:"steps_end"`
	StepsCurrent int    `json:"steps_current"`
	Unit         string `json:"unit,omitempty"`
	Note         string `json:"note,omitempty"`
}

// GoalsResponse is the response for listing goals.
type GoalsResponse struct {
	Goals []Goal `json:"goals"`
}

// GoalResponse is the response for a single goal.
type GoalResponse struct {
	Goal Goal `json:"goal"`
}

// KeyResultResponse is the response for a single key result.
type KeyResultResponse struct {
	KeyResult KeyResult `json:"key_result"`
}

// CreateGoalRequest is the request body for creating a goal.
type CreateGoalRequest struct {
	Name        string `json:"name"`
	DueDate     int64  `json:"due_date,omitempty"`
	Description string `json:"description,omitempty"`
	Owners      []int  `json:"owners,omitempty"`
	Color       string `json:"color,omitempty"`
}

// UpdateGoalRequest is the request body for updating a goal.
type UpdateGoalRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	DueDate     int64  `json:"due_date,omitempty"`
	Color       string `json:"color,omitempty"`
	AddOwners   []int  `json:"add_owners,omitempty"`
	RemOwners   []int  `json:"rem_owners,omitempty"`
}

// CreateKeyResultRequest is the request body for creating a key result.
type CreateKeyResultRequest struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	StepsStart int    `json:"steps_start,omitempty"`
	StepsEnd   int    `json:"steps_end,omitempty"`
	Unit       string `json:"unit,omitempty"`
	Owners     []int  `json:"owners,omitempty"`
}

// EditKeyResultRequest is the request body for editing a key result.
type EditKeyResultRequest struct {
	StepsCurrent int    `json:"steps_current,omitempty"`
	Note         string `json:"note,omitempty"`
}
