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

// Space represents a ClickUp space.
type Space struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// List represents a ClickUp list.
type List struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Folder represents a ClickUp folder.
type Folder struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Lists []List `json:"lists,omitempty"`
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
	ID string `json:"id"`
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
