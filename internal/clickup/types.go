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

// --- View types ---

// View represents a ClickUp view.
type View struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Type      string     `json:"type"`
	Parent    ViewParent `json:"parent"`
	Protected bool       `json:"protected"`
}

// ViewParent is the parent reference for a view.
type ViewParent struct {
	ID   string `json:"id"`
	Type int    `json:"type"` // 7=space, 5=folder, 6=list, etc.
}

// ViewsResponse is the response for listing views.
type ViewsResponse struct {
	Views []View `json:"views"`
}

// ViewResponse is the response for a single view.
type ViewResponse struct {
	View View `json:"view"`
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
	Status   string   `json:"status,omitempty"`
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

// KeyResult represents a key result within a goal.
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
