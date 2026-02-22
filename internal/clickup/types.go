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

// --- Chat v3 types ---

// ChatChannel represents a ClickUp chat channel.
type ChatChannel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"` // public, private, direct_message
	MemberCount int    `json:"member_count,omitempty"`
}

// ChatMessage represents a chat message.
type ChatMessage struct {
	ID            string      `json:"id"`
	Content       string      `json:"content"`
	UserID        string      `json:"user_id"`
	Type          string      `json:"type"` // message, post
	DateCreated   json.Number `json:"date_created"`
	DateUpdated   json.Number `json:"date_updated"`
	ParentChannel string      `json:"parent_channel"`
	ParentMessage string      `json:"parent_message,omitempty"`
	Resolved      bool        `json:"resolved"`
	RepliesCount  int         `json:"replies_count"`
}

// ChatReaction represents a reaction on a message.
type ChatReaction struct {
	ID          string      `json:"id"`
	MessageID   string      `json:"message_id"`
	UserID      string      `json:"user_id"`
	Reaction    string      `json:"reaction"`
	DateCreated json.Number `json:"date_created"`
}

// ChatUser represents a user in the chat system.
type ChatUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

// ChatPagination represents cursor-based pagination.
type ChatPagination struct {
	NextPageToken string `json:"next_page_token,omitempty"`
}

// --- Chat Request types ---

// CreateChatChannelRequest is the request body for creating a channel.
type CreateChatChannelRequest struct {
	Name string `json:"name"`
}

// CreateDMRequest is the request body for creating a direct message.
type CreateDMRequest struct {
	Members []string `json:"members"` // user IDs
}

// CreateLocationChannelRequest is the request body for creating a location channel.
type CreateLocationChannelRequest struct {
	Name       string `json:"name"`
	ParentType string `json:"parent_type"` // space, folder, list
	ParentID   string `json:"parent_id"`
}

// SendMessageRequest is the request body for sending a message.
type SendMessageRequest struct {
	Content string `json:"content"`
}

// CreateReactionRequest is the request body for creating a reaction.
type CreateReactionRequest struct {
	Reaction string `json:"reaction"`
}

// UpdateChannelRequest is the request body for updating a channel.
type UpdateChannelRequest struct {
	Name string `json:"name,omitempty"`
}

// UpdateMessageRequest is the request body for updating a message.
type UpdateMessageRequest struct {
	Content string `json:"content,omitempty"`
}

// --- Chat Response types ---

// ChatChannelsResponse is the response for listing channels.
type ChatChannelsResponse struct {
	Channels []ChatChannel `json:"channels"`
}

// ChatMessagesResponse is the response for listing messages.
type ChatMessagesResponse struct {
	Data       []ChatMessage   `json:"data"`
	Pagination *ChatPagination `json:"pagination,omitempty"`
}

// ChatReactionsResponse is the response for listing reactions.
type ChatReactionsResponse struct {
	Reactions []ChatReaction `json:"reactions"`
}

// ChatUsersResponse is the response for listing users (followers, members, tagged).
type ChatUsersResponse struct {
	Users []ChatUser `json:"users"`
}

// --- Docs v3 types ---

// Doc represents a ClickUp doc.
type Doc struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DateCreated int64  `json:"date_created,omitempty"`
	Creator     *User  `json:"creator,omitempty"`
}

// DocPage represents a page within a doc.
type DocPage struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content,omitempty"`
	Order   int    `json:"order,omitempty"`
}

// --- Docs Request types ---

// CreateDocRequest is the request body for creating a doc.
type CreateDocRequest struct {
	Name       string `json:"name"`
	ParentType string `json:"parent_type,omitempty"`
	ParentID   string `json:"parent_id,omitempty"`
}

// CreatePageRequest is the request body for creating a page.
type CreatePageRequest struct {
	Name          string `json:"name"`
	Content       string `json:"content,omitempty"`
	ContentFormat string `json:"content_format,omitempty"` // "md" or "html"
}

// EditPageRequest is the request body for editing a page.
type EditPageRequest struct {
	Name          string `json:"name,omitempty"`
	Content       string `json:"content,omitempty"`
	ContentFormat string `json:"content_format,omitempty"`
}

// --- Docs Response types ---

// DocsResponse is the response for listing/searching docs.
type DocsResponse struct {
	Docs []Doc `json:"docs"`
}

// DocPagesResponse is the response for listing pages.
type DocPagesResponse struct {
	Pages []DocPage `json:"pages"`
}
