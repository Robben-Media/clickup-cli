package clickup

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/builtbyrobben/clickup-cli/internal/api"
)

func newTestClient(server *httptest.Server) *Client {
	return &Client{
		Client: api.NewClient("test-api-key",
			api.WithBaseURL(server.URL),
			api.WithUserAgent("clickup-cli/test"),
		),
	}
}

func TestV3Path_BuildsCorrectPath(t *testing.T) {
	t.Parallel()

	client := &Client{
		Client:      api.NewClient("test-key"),
		workspaceID: "workspace-123",
	}

	path, err := client.v3Path("/chat/channels")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "/v3/workspaces/workspace-123/chat/channels"
	if path != want {
		t.Fatalf("expected path %q, got %q", want, path)
	}
}

func TestV3Path_MissingWorkspaceID(t *testing.T) {
	t.Parallel()

	client := &Client{
		Client: api.NewClient("test-key"),
	}

	_, err := client.v3Path("/chat/channels")
	if err == nil {
		t.Fatal("expected error for missing workspace ID, got nil")
	}

	if !errors.Is(err, errWorkspaceIDRequired) {
		t.Fatalf("expected errWorkspaceIDRequired, got %v", err)
	}
}

func TestTasksList_EncodesFilters(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/list/list-1/task" {
			t.Fatalf("expected path /v2/list/list-1/task, got %s", r.URL.Path)
		}

		if r.URL.Query().Get("include_closed") != "true" {
			t.Fatalf("expected include_closed=true, got %s", r.URL.Query().Get("include_closed"))
		}

		if r.URL.Query().Get("statuses[]") != "in progress" {
			t.Fatalf("expected statuses[]=in progress, got %s", r.URL.Query().Get("statuses[]"))
		}

		if r.URL.Query().Get("assignees[]") != "user+1@example.com" {
			t.Fatalf("expected assignees[]=user+1@example.com, got %s", r.URL.Query().Get("assignees[]"))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(TasksListResponse{
			Tasks: []Task{{ID: "task-1", Name: "Task 1"}},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Tasks().List(context.Background(), "list-1", "in progress", "user+1@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Tasks) != 1 {
		t.Fatalf("expected one task, got %d", len(result.Tasks))
	}
}

func TestMembersList_ExtractsNestedMembers(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/team/team-1" {
			t.Fatalf("expected path /v2/team/team-1, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"team": map[string]any{
				"members": []map[string]any{
					{"user": map[string]any{"id": 1, "username": "alice"}},
				},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Members().List(context.Background(), "team-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Members) != 1 {
		t.Fatalf("expected one member, got %d", len(result.Members))
	}

	if result.Members[0].User.Username != "alice" {
		t.Fatalf("expected username alice, got %s", result.Members[0].User.Username)
	}
}

func TestCommentsAdd_ReturnsIDAndText(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/v2/task/task-1/comment" {
			t.Fatalf("expected path /v2/task/task-1/comment, got %s", r.URL.Path)
		}

		var req CreateCommentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if req.CommentText != "hello from test" {
			t.Fatalf("expected comment_text hello from test, got %s", req.CommentText)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]int{"id": 123})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Comments().Add(context.Background(), "task-1", "hello from test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID.String() != "123" {
		t.Fatalf("expected ID 123, got %s", result.ID.String())
	}

	if result.Text != "hello from test" {
		t.Fatalf("expected text hello from test, got %s", result.Text)
	}
}

func TestTimeList_EscapesTaskID(t *testing.T) {
	t.Parallel()

	taskID := "task/with?chars"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/team/team-1/time_entries" {
			t.Fatalf("expected path /v2/team/team-1/time_entries, got %s", r.URL.Path)
		}

		if r.URL.Query().Get("task_id") != taskID {
			t.Fatalf("expected task_id=%s, got %s", taskID, r.URL.Query().Get("task_id"))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(TimeEntriesListResponse{
			Data: []TimeEntry{{ID: json.Number("1"), Task: TaskRef{ID: taskID}}},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Time().List(context.Background(), "team-1", taskID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Data) != 1 {
		t.Fatalf("expected one time entry, got %d", len(result.Data))
	}
}

func TestTasksUpdate_SendsAssigneesAddRem(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}

		if r.URL.Path != "/v2/task/task-1" {
			t.Fatalf("expected path /v2/task/task-1, got %s", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		assignees, ok := body["assignees"].(map[string]any)
		if !ok {
			t.Fatalf("expected assignees object in request")
		}

		add, ok := assignees["add"].([]any)
		if !ok || len(add) != 1 || int(add[0].(float64)) != 123 {
			t.Fatalf("expected assignees.add [123], got %#v", assignees["add"])
		}

		rem, ok := assignees["rem"].([]any)
		if !ok || len(rem) != 1 || int(rem[0].(float64)) != 456 {
			t.Fatalf("expected assignees.rem [456], got %#v", assignees["rem"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Task{ID: "task-1", Name: "Task 1"})
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Tasks().Update(context.Background(), "task-1", UpdateTaskRequest{
		Assignees: &TaskAssigneesUpdate{Add: []int{123}, Rem: []int{456}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Chat Service Tests ---

func newTestClientWithWorkspace(server *httptest.Server) *Client {
	return &Client{
		Client: api.NewClient("test-api-key",
			api.WithBaseURL(server.URL),
			api.WithUserAgent("clickup-cli/test"),
		),
		workspaceID: "workspace-123",
	}
}

func TestChatService_ListChannels(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/v3/workspaces/workspace-123/chat/channels"
		if r.URL.Path != expectedPath {
			t.Fatalf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ChatChannelsResponse{
			Channels: []ChatChannel{
				{ID: "chan-1", Name: "General", Type: "public", MemberCount: 10},
			},
		})
	}))
	defer server.Close()

	client := newTestClientWithWorkspace(server)

	result, err := client.Chat().ListChannels(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Channels) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(result.Channels))
	}

	if result.Channels[0].Name != "General" {
		t.Fatalf("expected channel name General, got %s", result.Channels[0].Name)
	}
}

func TestChatService_GetChannel(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/v3/workspaces/workspace-123/chat/channels/chan-1"
		if r.URL.Path != expectedPath {
			t.Fatalf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ChatChannel{
			ID: "chan-1", Name: "Engineering", Type: "private", MemberCount: 5,
		})
	}))
	defer server.Close()

	client := newTestClientWithWorkspace(server)

	result, err := client.Chat().GetChannel(context.Background(), "chan-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Name != "Engineering" {
		t.Fatalf("expected channel name Engineering, got %s", result.Name)
	}
}

func TestChatService_CreateChannel(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		expectedPath := "/v3/workspaces/workspace-123/chat/channels"
		if r.URL.Path != expectedPath {
			t.Fatalf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		var req CreateChatChannelRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if req.Name != "New Channel" {
			t.Fatalf("expected name New Channel, got %s", req.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ChatChannel{
			ID: "chan-new", Name: "New Channel", Type: "public",
		})
	}))
	defer server.Close()

	client := newTestClientWithWorkspace(server)

	result, err := client.Chat().CreateChannel(context.Background(), CreateChatChannelRequest{Name: "New Channel"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "chan-new" {
		t.Fatalf("expected ID chan-new, got %s", result.ID)
	}
}

func TestChatService_UpdateChannel(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("expected PATCH, got %s", r.Method)
		}

		expectedPath := "/v3/workspaces/workspace-123/chat/channels/chan-1"
		if r.URL.Path != expectedPath {
			t.Fatalf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ChatChannel{
			ID: "chan-1", Name: "Updated Name", Type: "public",
		})
	}))
	defer server.Close()

	client := newTestClientWithWorkspace(server)

	result, err := client.Chat().UpdateChannel(context.Background(), "chan-1", UpdateChannelRequest{Name: "Updated Name"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Name != "Updated Name" {
		t.Fatalf("expected name Updated Name, got %s", result.Name)
	}
}

func TestChatService_DeleteChannel(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}

		expectedPath := "/v3/workspaces/workspace-123/chat/channels/chan-1"
		if r.URL.Path != expectedPath {
			t.Fatalf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClientWithWorkspace(server)

	err := client.Chat().DeleteChannel(context.Background(), "chan-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestChatService_SendMessage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		expectedPath := "/v3/workspaces/workspace-123/chat/channels/chan-1/messages"
		if r.URL.Path != expectedPath {
			t.Fatalf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		var req SendMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if req.Content != "Hello world" {
			t.Fatalf("expected content Hello world, got %s", req.Content)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ChatMessage{
			ID: "msg-1", Content: "Hello world", ParentChannel: "chan-1",
		})
	}))
	defer server.Close()

	client := newTestClientWithWorkspace(server)

	result, err := client.Chat().SendMessage(context.Background(), "chan-1", SendMessageRequest{Content: "Hello world"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "msg-1" {
		t.Fatalf("expected ID msg-1, got %s", result.ID)
	}
}

func TestChatService_ListMessages(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/v3/workspaces/workspace-123/chat/channels/chan-1/messages"
		if r.URL.Path != expectedPath {
			t.Fatalf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Check query params
		if r.URL.Query().Get("limit") != "10" {
			t.Fatalf("expected limit=10, got %s", r.URL.Query().Get("limit"))
		}

		if r.URL.Query().Get("cursor") != "abc123" {
			t.Fatalf("expected cursor=abc123, got %s", r.URL.Query().Get("cursor"))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ChatMessagesResponse{
			Data: []ChatMessage{
				{ID: "msg-1", Content: "First message", UserID: "user-1"},
			},
			Pagination: &ChatPagination{NextPageToken: "next-token"},
		})
	}))
	defer server.Close()

	client := newTestClientWithWorkspace(server)

	result, err := client.Chat().ListMessages(context.Background(), "chan-1", 10, "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Data) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result.Data))
	}

	if result.Pagination.NextPageToken != "next-token" {
		t.Fatalf("expected next-token, got %s", result.Pagination.NextPageToken)
	}
}

func TestChatService_DeleteMessage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}

		expectedPath := "/v3/workspaces/workspace-123/chat/messages/msg-1"
		if r.URL.Path != expectedPath {
			t.Fatalf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClientWithWorkspace(server)

	err := client.Chat().DeleteMessage(context.Background(), "msg-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestChatService_CreateReaction(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		expectedPath := "/v3/workspaces/workspace-123/chat/messages/msg-1/reactions"
		if r.URL.Path != expectedPath {
			t.Fatalf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		var req CreateReactionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if req.Reaction != "👍" {
			t.Fatalf("expected reaction 👍, got %s", req.Reaction)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ChatReaction{
			ID: "react-1", MessageID: "msg-1", Reaction: "👍",
		})
	}))
	defer server.Close()

	client := newTestClientWithWorkspace(server)

	result, err := client.Chat().CreateReaction(context.Background(), "msg-1", CreateReactionRequest{Reaction: "👍"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Reaction != "👍" {
		t.Fatalf("expected reaction 👍, got %s", result.Reaction)
	}
}

func TestChatService_DeleteReaction(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}

		expectedPath := "/v3/workspaces/workspace-123/chat/messages/msg-1/reactions/react-1"
		if r.URL.Path != expectedPath {
			t.Fatalf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClientWithWorkspace(server)

	err := client.Chat().DeleteReaction(context.Background(), "msg-1", "react-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestChatService_MissingWorkspaceID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should not be made without workspace ID")
	}))
	defer server.Close()

	client := newTestClient(server) // no workspace ID

	_, err := client.Chat().ListChannels(context.Background())
	if err == nil {
		t.Fatal("expected error for missing workspace ID")
	}

	if !errors.Is(err, errWorkspaceIDRequired) {
		t.Fatalf("expected errWorkspaceIDRequired, got %v", err)
	}
}

// --- Docs Service Tests ---

func TestDocsService_Search(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/v3/workspaces/workspace-123/docs"
		if r.URL.Path != expectedPath {
			t.Fatalf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		if r.URL.Query().Get("query") != "test" {
			t.Fatalf("expected query=test, got %s", r.URL.Query().Get("query"))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(DocsResponse{
			Docs: []Doc{
				{ID: "doc-1", Name: "Test Doc", DateCreated: 1700000000},
			},
		})
	}))
	defer server.Close()

	client := newTestClientWithWorkspace(server)

	result, err := client.Docs().Search(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Docs) != 1 {
		t.Fatalf("expected 1 doc, got %d", len(result.Docs))
	}

	if result.Docs[0].Name != "Test Doc" {
		t.Fatalf("expected doc name Test Doc, got %s", result.Docs[0].Name)
	}
}

func TestDocsService_Get(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/v3/workspaces/workspace-123/docs/doc-1"
		if r.URL.Path != expectedPath {
			t.Fatalf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Doc{
			ID: "doc-1", Name: "API Guide", DateCreated: 1700000000,
		})
	}))
	defer server.Close()

	client := newTestClientWithWorkspace(server)

	result, err := client.Docs().Get(context.Background(), "doc-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Name != "API Guide" {
		t.Fatalf("expected doc name API Guide, got %s", result.Name)
	}
}

func TestDocsService_Create(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		expectedPath := "/v3/workspaces/workspace-123/docs"
		if r.URL.Path != expectedPath {
			t.Fatalf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		var req CreateDocRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if req.Name != "New Doc" {
			t.Fatalf("expected name New Doc, got %s", req.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Doc{
			ID: "doc-new", Name: "New Doc",
		})
	}))
	defer server.Close()

	client := newTestClientWithWorkspace(server)

	result, err := client.Docs().Create(context.Background(), CreateDocRequest{Name: "New Doc"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "doc-new" {
		t.Fatalf("expected ID doc-new, got %s", result.ID)
	}
}

func TestDocsService_GetPages(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/v3/workspaces/workspace-123/docs/doc-1/pages"
		if r.URL.Path != expectedPath {
			t.Fatalf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(DocPagesResponse{
			Pages: []DocPage{
				{ID: "page-1", Name: "Introduction", Order: 0},
				{ID: "page-2", Name: "Getting Started", Order: 1},
			},
		})
	}))
	defer server.Close()

	client := newTestClientWithWorkspace(server)

	result, err := client.Docs().GetPages(context.Background(), "doc-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Pages) != 2 {
		t.Fatalf("expected 2 pages, got %d", len(result.Pages))
	}

	if result.Pages[0].Name != "Introduction" {
		t.Fatalf("expected page name Introduction, got %s", result.Pages[0].Name)
	}
}

func TestDocsService_GetPage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/v3/workspaces/workspace-123/docs/doc-1/pages/page-1"
		if r.URL.Path != expectedPath {
			t.Fatalf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(DocPage{
			ID: "page-1", Name: "Introduction", Content: "# Hello World", Order: 0,
		})
	}))
	defer server.Close()

	client := newTestClientWithWorkspace(server)

	result, err := client.Docs().GetPage(context.Background(), "doc-1", "page-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Content != "# Hello World" {
		t.Fatalf("expected content # Hello World, got %s", result.Content)
	}
}

func TestDocsService_CreatePage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		expectedPath := "/v3/workspaces/workspace-123/docs/doc-1/pages"
		if r.URL.Path != expectedPath {
			t.Fatalf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		var req CreatePageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if req.Name != "New Page" {
			t.Fatalf("expected name New Page, got %s", req.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(DocPage{
			ID: "page-new", Name: "New Page",
		})
	}))
	defer server.Close()

	client := newTestClientWithWorkspace(server)

	result, err := client.Docs().CreatePage(context.Background(), "doc-1", CreatePageRequest{Name: "New Page"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "page-new" {
		t.Fatalf("expected ID page-new, got %s", result.ID)
	}
}

func TestDocsService_EditPage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}

		expectedPath := "/v3/workspaces/workspace-123/docs/doc-1/pages/page-1"
		if r.URL.Path != expectedPath {
			t.Fatalf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(DocPage{
			ID: "page-1", Name: "Updated Page",
		})
	}))
	defer server.Close()

	client := newTestClientWithWorkspace(server)

	result, err := client.Docs().EditPage(context.Background(), "doc-1", "page-1", EditPageRequest{Name: "Updated Page"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Name != "Updated Page" {
		t.Fatalf("expected name Updated Page, got %s", result.Name)
	}
}

func TestDocsService_MissingWorkspaceID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should not be made without workspace ID")
	}))
	defer server.Close()

	client := newTestClient(server) // no workspace ID

	_, err := client.Docs().Search(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing workspace ID")
	}

	if !errors.Is(err, errWorkspaceIDRequired) {
		t.Fatalf("expected errWorkspaceIDRequired, got %v", err)
	}
}
