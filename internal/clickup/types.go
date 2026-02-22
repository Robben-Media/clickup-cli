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

// --- Space Tags ---

// SpaceTag represents a tag in a space.
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

// --- Checklists ---

// Checklist represents a ClickUp checklist.
type Checklist struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	OrderIndex int             `json:"orderindex"`
	Items      []ChecklistItem `json:"items,omitempty"`
}

// ChecklistItem represents an item in a ClickUp checklist.
type ChecklistItem struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Resolved   bool   `json:"resolved"`
	Assignee   *User  `json:"assignee,omitempty"`
	Parent     string `json:"parent,omitempty"`
	OrderIndex int    `json:"orderindex"`
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

// --- Relationships ---

// AddDependencyRequest is the request body for adding a task dependency.
type AddDependencyRequest struct {
	DependsOn    string `json:"depends_on,omitempty"`
	DependencyOf string `json:"dependency_of,omitempty"`
}
