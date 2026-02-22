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

// SpaceDetail represents a full space with statuses and features.
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

// SpaceFeatures represents feature toggles for a space.
type SpaceFeatures struct {
	DueDates     FeatureToggle `json:"due_dates"`
	TimeTracking FeatureToggle `json:"time_tracking"`
	Tags         FeatureToggle `json:"tags"`
	Checklists   FeatureToggle `json:"checklists"`
}

// FeatureToggle represents an enabled/disabled feature.
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

// ListDetail represents a full list with all properties.
type ListDetail struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	Content          string     `json:"content,omitempty"`
	DueDate          string     `json:"due_date,omitempty"`
	Priority         *Priority  `json:"priority,omitempty"`
	Assignee         *User      `json:"assignee,omitempty"`
	TaskCount        int        `json:"task_count"`
	Folder           FolderRef  `json:"folder,omitempty"`
	Space            SpaceRef   `json:"space,omitempty"`
	Status           ListStatus `json:"status,omitempty"`
	Permission       string     `json:"permission_level,omitempty"`
	Archived         bool       `json:"archived"`
	OverrideStatuses bool       `json:"override_statuses"`
}

// ListStatus represents a list's status/due date info.
type ListStatus struct {
	Status     string `json:"status,omitempty"`
	Color      string `json:"color,omitempty"`
	OrderIndex int    `json:"orderindex,omitempty"`
}

// CreateListRequest is the request body for creating a list.
type CreateListRequest struct {
	Name     string `json:"name"`
	Content  string `json:"content,omitempty"`
	DueDate  int64  `json:"due_date,omitempty"`
	Priority int    `json:"priority,omitempty"`
	Assignee int    `json:"assignee,omitempty"`
	Status   string `json:"status,omitempty"`
}

// UpdateListRequest is the request body for updating a list.
type UpdateListRequest struct {
	Name          string `json:"name,omitempty"`
	Content       string `json:"content,omitempty"`
	DueDate       int64  `json:"due_date,omitempty"`
	Priority      int    `json:"priority,omitempty"`
	Assignee      int    `json:"assignee,omitempty"`
	UnsetAssignee bool   `json:"unset_assignee,omitempty"`
	UnsetDueDate  bool   `json:"unset_due_date,omitempty"`
}

// CreateListFromTemplateRequest is the request body for creating a list from a template.
type CreateListFromTemplateRequest struct {
	Name string `json:"name,omitempty"`
}

// Folder represents a ClickUp folder (lightweight, from list endpoint).
type Folder struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Lists []List `json:"lists,omitempty"`
}

// FolderDetail represents a full folder with all properties.
type FolderDetail struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	OrderIndex       int      `json:"orderindex"`
	OverrideStatuses bool     `json:"override_statuses"`
	Hidden           bool     `json:"hidden"`
	Space            SpaceRef `json:"space"`
	TaskCount        string   `json:"task_count"`
	Lists            []List   `json:"lists,omitempty"`
	Archived         bool     `json:"archived"`
	PermissionLevel  string   `json:"permission_level,omitempty"`
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

// FilteredTeamTasksParams contains query parameters for the filtered team tasks endpoint.
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

// TimeInStatusResponse represents time-in-status data for a single task.
type TimeInStatusResponse struct {
	CurrentStatus StatusTime   `json:"current_status"`
	StatusHistory []StatusTime `json:"status_history"`
}

// StatusTime represents time spent in a particular status.
type StatusTime struct {
	Status    string    `json:"status"`
	Color     string    `json:"color,omitempty"`
	TotalTime TimeValue `json:"total_time"`
}

// TimeValue represents a time duration with since timestamp.
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

// CreateTaskFromTemplateRequest is the request body for creating a task from template.
type CreateTaskFromTemplateRequest struct {
	Name string `json:"name,omitempty"`
}
