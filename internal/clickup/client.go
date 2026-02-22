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
	errMembersRequired     = errors.New("members are required")
	errParentRequired      = errors.New("parent_type and parent_id are required")
	errContentRequired     = errors.New("content is required")
	errReactionRequired    = errors.New("reaction is required")
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

// Chat provides methods for the Chat v3 API.
func (c *Client) Chat() *ChatService {
	return &ChatService{client: c}
}

// Docs provides methods for the Docs v3 API.
func (c *Client) Docs() *DocsService {
	return &DocsService{client: c}
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

// --- ChatService (v3) ---

// ChatService handles Chat v3 API operations.
type ChatService struct {
	client *Client
}

// ListChannels retrieves all chat channels in a workspace.
func (s *ChatService) ListChannels(ctx context.Context) (*ChatChannelsResponse, error) {
	path, err := s.client.v3Path("/chat/channels")
	if err != nil {
		return nil, err
	}

	var result ChatChannelsResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list chat channels: %w", err)
	}

	return &result, nil
}

// GetChannel retrieves a single chat channel by ID.
func (s *ChatService) GetChannel(ctx context.Context, channelID string) (*ChatChannel, error) {
	if channelID == "" {
		return nil, errIDRequired
	}

	path, err := s.client.v3Path(fmt.Sprintf("/chat/channels/%s", channelID))
	if err != nil {
		return nil, err
	}

	var result ChatChannel
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get chat channel: %w", err)
	}

	return &result, nil
}

// GetChannelFollowers retrieves followers of a channel.
func (s *ChatService) GetChannelFollowers(ctx context.Context, channelID string) (*ChatUsersResponse, error) {
	if channelID == "" {
		return nil, errIDRequired
	}

	path, err := s.client.v3Path(fmt.Sprintf("/chat/channels/%s/followers", channelID))
	if err != nil {
		return nil, err
	}

	var result ChatUsersResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get channel followers: %w", err)
	}

	return &result, nil
}

// GetChannelMembers retrieves members of a channel.
func (s *ChatService) GetChannelMembers(ctx context.Context, channelID string) (*ChatUsersResponse, error) {
	if channelID == "" {
		return nil, errIDRequired
	}

	path, err := s.client.v3Path(fmt.Sprintf("/chat/channels/%s/members", channelID))
	if err != nil {
		return nil, err
	}

	var result ChatUsersResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get channel members: %w", err)
	}

	return &result, nil
}

// ListMessages retrieves messages from a channel.
func (s *ChatService) ListMessages(ctx context.Context, channelID string, limit int, cursor string) (*ChatMessagesResponse, error) {
	if channelID == "" {
		return nil, errIDRequired
	}

	path, err := s.client.v3Path(fmt.Sprintf("/chat/channels/%s/messages", channelID))
	if err != nil {
		return nil, err
	}

	// Build query parameters
	params := url.Values{}

	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}

	if cursor != "" {
		params.Set("cursor", cursor)
	}

	if len(params) > 0 {
		path = path + "?" + params.Encode()
	}

	var result ChatMessagesResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list chat messages: %w", err)
	}

	return &result, nil
}

// CreateChannel creates a new chat channel.
func (s *ChatService) CreateChannel(ctx context.Context, req CreateChatChannelRequest) (*ChatChannel, error) {
	if req.Name == "" {
		return nil, errNameRequired
	}

	path, err := s.client.v3Path("/chat/channels")
	if err != nil {
		return nil, err
	}

	var result ChatChannel
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create chat channel: %w", err)
	}

	return &result, nil
}

// CreateDirectMessage creates a direct message channel.
func (s *ChatService) CreateDirectMessage(ctx context.Context, req CreateDMRequest) (*ChatChannel, error) {
	if len(req.Members) == 0 {
		return nil, errMembersRequired
	}

	path, err := s.client.v3Path("/chat/channels/direct_message")
	if err != nil {
		return nil, err
	}

	var result ChatChannel
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create direct message: %w", err)
	}

	return &result, nil
}

// CreateLocationChannel creates a channel tied to a space/folder/list.
func (s *ChatService) CreateLocationChannel(ctx context.Context, req CreateLocationChannelRequest) (*ChatChannel, error) {
	if req.Name == "" {
		return nil, errNameRequired
	}

	if req.ParentType == "" || req.ParentID == "" {
		return nil, errParentRequired
	}

	path, err := s.client.v3Path("/chat/channels/location")
	if err != nil {
		return nil, err
	}

	var result ChatChannel
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create location channel: %w", err)
	}

	return &result, nil
}

// UpdateChannel updates a chat channel.
func (s *ChatService) UpdateChannel(ctx context.Context, channelID string, req UpdateChannelRequest) (*ChatChannel, error) {
	if channelID == "" {
		return nil, errIDRequired
	}

	path, err := s.client.v3Path(fmt.Sprintf("/chat/channels/%s", channelID))
	if err != nil {
		return nil, err
	}

	var result ChatChannel
	if err := s.client.Patch(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("update chat channel: %w", err)
	}

	return &result, nil
}

// DeleteChannel deletes a chat channel.
func (s *ChatService) DeleteChannel(ctx context.Context, channelID string) error {
	if channelID == "" {
		return errIDRequired
	}

	path, err := s.client.v3Path(fmt.Sprintf("/chat/channels/%s", channelID))
	if err != nil {
		return err
	}

	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("delete chat channel: %w", err)
	}

	return nil
}

// SendMessage sends a message to a channel.
func (s *ChatService) SendMessage(ctx context.Context, channelID string, req SendMessageRequest) (*ChatMessage, error) {
	if channelID == "" {
		return nil, errIDRequired
	}

	if req.Content == "" {
		return nil, errContentRequired
	}

	path, err := s.client.v3Path(fmt.Sprintf("/chat/channels/%s/messages", channelID))
	if err != nil {
		return nil, err
	}

	var result ChatMessage
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("send chat message: %w", err)
	}

	return &result, nil
}

// GetReactions retrieves reactions on a message.
func (s *ChatService) GetReactions(ctx context.Context, messageID string) (*ChatReactionsResponse, error) {
	if messageID == "" {
		return nil, errIDRequired
	}

	path, err := s.client.v3Path(fmt.Sprintf("/chat/messages/%s/reactions", messageID))
	if err != nil {
		return nil, err
	}

	var result ChatReactionsResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get chat reactions: %w", err)
	}

	return &result, nil
}

// GetReplies retrieves replies to a message.
func (s *ChatService) GetReplies(ctx context.Context, messageID string) (*ChatMessagesResponse, error) {
	if messageID == "" {
		return nil, errIDRequired
	}

	path, err := s.client.v3Path(fmt.Sprintf("/chat/messages/%s/replies", messageID))
	if err != nil {
		return nil, err
	}

	var result ChatMessagesResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get chat replies: %w", err)
	}

	return &result, nil
}

// GetTaggedUsers retrieves users tagged in a message.
func (s *ChatService) GetTaggedUsers(ctx context.Context, messageID string) (*ChatUsersResponse, error) {
	if messageID == "" {
		return nil, errIDRequired
	}

	path, err := s.client.v3Path(fmt.Sprintf("/chat/messages/%s/tagged_users", messageID))
	if err != nil {
		return nil, err
	}

	var result ChatUsersResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get tagged users: %w", err)
	}

	return &result, nil
}

// CreateReaction adds a reaction to a message.
func (s *ChatService) CreateReaction(ctx context.Context, messageID string, req CreateReactionRequest) (*ChatReaction, error) {
	if messageID == "" {
		return nil, errIDRequired
	}

	if req.Reaction == "" {
		return nil, errReactionRequired
	}

	path, err := s.client.v3Path(fmt.Sprintf("/chat/messages/%s/reactions", messageID))
	if err != nil {
		return nil, err
	}

	var result ChatReaction
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create chat reaction: %w", err)
	}

	return &result, nil
}

// CreateReply creates a reply to a message.
func (s *ChatService) CreateReply(ctx context.Context, messageID string, req SendMessageRequest) (*ChatMessage, error) {
	if messageID == "" {
		return nil, errIDRequired
	}

	if req.Content == "" {
		return nil, errContentRequired
	}

	path, err := s.client.v3Path(fmt.Sprintf("/chat/messages/%s/replies", messageID))
	if err != nil {
		return nil, err
	}

	var result ChatMessage
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create chat reply: %w", err)
	}

	return &result, nil
}

// UpdateMessage updates a chat message.
func (s *ChatService) UpdateMessage(ctx context.Context, messageID string, req UpdateMessageRequest) (*ChatMessage, error) {
	if messageID == "" {
		return nil, errIDRequired
	}

	path, err := s.client.v3Path(fmt.Sprintf("/chat/messages/%s", messageID))
	if err != nil {
		return nil, err
	}

	var result ChatMessage
	if err := s.client.Patch(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("update chat message: %w", err)
	}

	return &result, nil
}

// DeleteMessage deletes a chat message.
func (s *ChatService) DeleteMessage(ctx context.Context, messageID string) error {
	if messageID == "" {
		return errIDRequired
	}

	path, err := s.client.v3Path(fmt.Sprintf("/chat/messages/%s", messageID))
	if err != nil {
		return err
	}

	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("delete chat message: %w", err)
	}

	return nil
}

// DeleteReaction removes a reaction from a message.
func (s *ChatService) DeleteReaction(ctx context.Context, messageID, reactionID string) error {
	if messageID == "" || reactionID == "" {
		return errIDRequired
	}

	path, err := s.client.v3Path(fmt.Sprintf("/chat/messages/%s/reactions/%s", messageID, reactionID))
	if err != nil {
		return err
	}

	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("delete chat reaction: %w", err)
	}

	return nil
}

// --- DocsService (v3) ---

// DocsService handles Docs v3 API operations.
type DocsService struct {
	client *Client
}

// Search searches for docs in a workspace.
func (s *DocsService) Search(ctx context.Context, query string) (*DocsResponse, error) {
	path, err := s.client.v3Path("/docs")
	if err != nil {
		return nil, err
	}

	if query != "" {
		path = path + "?query=" + url.QueryEscape(query)
	}

	var result DocsResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("search docs: %w", err)
	}

	return &result, nil
}

// Get retrieves a single doc by ID.
func (s *DocsService) Get(ctx context.Context, docID string) (*Doc, error) {
	if docID == "" {
		return nil, errIDRequired
	}

	path, err := s.client.v3Path(fmt.Sprintf("/docs/%s", docID))
	if err != nil {
		return nil, err
	}

	var result Doc
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get doc: %w", err)
	}

	return &result, nil
}

// GetPageListing retrieves the page listing for a doc.
func (s *DocsService) GetPageListing(ctx context.Context, docID string) (*DocPagesResponse, error) {
	if docID == "" {
		return nil, errIDRequired
	}

	path, err := s.client.v3Path(fmt.Sprintf("/docs/%s/page_listing", docID))
	if err != nil {
		return nil, err
	}

	var result DocPagesResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get page listing: %w", err)
	}

	return &result, nil
}

// GetPages retrieves all pages in a doc.
func (s *DocsService) GetPages(ctx context.Context, docID string) (*DocPagesResponse, error) {
	if docID == "" {
		return nil, errIDRequired
	}

	path, err := s.client.v3Path(fmt.Sprintf("/docs/%s/pages", docID))
	if err != nil {
		return nil, err
	}

	var result DocPagesResponse
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get pages: %w", err)
	}

	return &result, nil
}

// GetPage retrieves a single page from a doc.
func (s *DocsService) GetPage(ctx context.Context, docID, pageID string) (*DocPage, error) {
	if docID == "" || pageID == "" {
		return nil, errIDRequired
	}

	path, err := s.client.v3Path(fmt.Sprintf("/docs/%s/pages/%s", docID, pageID))
	if err != nil {
		return nil, err
	}

	var result DocPage
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get page: %w", err)
	}

	return &result, nil
}

// Create creates a new doc.
func (s *DocsService) Create(ctx context.Context, req CreateDocRequest) (*Doc, error) {
	if req.Name == "" {
		return nil, errNameRequired
	}

	path, err := s.client.v3Path("/docs")
	if err != nil {
		return nil, err
	}

	var result Doc
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create doc: %w", err)
	}

	return &result, nil
}

// CreatePage creates a new page in a doc.
func (s *DocsService) CreatePage(ctx context.Context, docID string, req CreatePageRequest) (*DocPage, error) {
	if docID == "" {
		return nil, errIDRequired
	}

	if req.Name == "" {
		return nil, errNameRequired
	}

	path, err := s.client.v3Path(fmt.Sprintf("/docs/%s/pages", docID))
	if err != nil {
		return nil, err
	}

	var result DocPage
	if err := s.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("create page: %w", err)
	}

	return &result, nil
}

// EditPage updates a page in a doc.
func (s *DocsService) EditPage(ctx context.Context, docID, pageID string, req EditPageRequest) (*DocPage, error) {
	if docID == "" || pageID == "" {
		return nil, errIDRequired
	}

	path, err := s.client.v3Path(fmt.Sprintf("/docs/%s/pages/%s", docID, pageID))
	if err != nil {
		return nil, err
	}

	var result DocPage
	if err := s.client.Put(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("edit page: %w", err)
	}

	return &result, nil
}
