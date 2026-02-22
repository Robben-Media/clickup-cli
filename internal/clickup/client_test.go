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

// --- UsersService tests ---

func TestUsersGet_ReturnsUserDetails(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		wantPath := "/v2/team/team-1/user/123"
		if r.URL.Path != wantPath {
			t.Fatalf("expected path %s, got %s", wantPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(UserResponse{
			User: UserDetail{
				ID:       123,
				Username: "alice",
				Email:    "alice@example.com",
				Role:     2,
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Users().Get(context.Background(), "team-1", "123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.User.ID != 123 {
		t.Fatalf("expected ID 123, got %d", result.User.ID)
	}

	if result.User.Username != "alice" {
		t.Fatalf("expected username alice, got %s", result.User.Username)
	}

	if result.User.Role != 2 {
		t.Fatalf("expected role 2, got %d", result.User.Role)
	}
}

func TestUsersGet_RequiresTeamID(t *testing.T) {
	t.Parallel()

	// Client without server - we're testing validation only
	_, err := NewClient("test-key").Users().Get(context.Background(), "", "123")
	if err == nil {
		t.Fatal("expected error for missing team ID, got nil")
	}
}

func TestUsersGet_RequiresUserID(t *testing.T) {
	t.Parallel()

	_, err := NewClient("test-key").Users().Get(context.Background(), "team-1", "")
	if err == nil {
		t.Fatal("expected error for missing user ID, got nil")
	}
}

func TestUsersInvite_SendsEmail(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		wantPath := "/v2/team/team-1/user"
		if r.URL.Path != wantPath {
			t.Fatalf("expected path %s, got %s", wantPath, r.URL.Path)
		}

		var body InviteUserRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if body.Email != "new@example.com" {
			t.Fatalf("expected email new@example.com, got %s", body.Email)
		}

		if !body.Admin {
			t.Fatal("expected admin to be true")
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(UserResponse{
			User: UserDetail{
				ID:    456,
				Email: "new@example.com",
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Users().Invite(context.Background(), "team-1", InviteUserRequest{
		Email: "new@example.com",
		Admin: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.User.Email != "new@example.com" {
		t.Fatalf("expected email new@example.com, got %s", result.User.Email)
	}
}

func TestUsersInvite_RequiresEmail(t *testing.T) {
	t.Parallel()

	_, err := NewClient("test-key").Users().Invite(context.Background(), "team-1", InviteUserRequest{})
	if err == nil {
		t.Fatal("expected error for missing email, got nil")
	}
}

func TestUsersUpdate_SendsFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}

		wantPath := "/v2/team/team-1/user/123"
		if r.URL.Path != wantPath {
			t.Fatalf("expected path %s, got %s", wantPath, r.URL.Path)
		}

		var body EditUserRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if body.Username != "newname" {
			t.Fatalf("expected username newname, got %s", body.Username)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(UserResponse{
			User: UserDetail{
				ID:       123,
				Username: "newname",
				Role:     2,
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Users().Update(context.Background(), "team-1", "123", EditUserRequest{
		Username: "newname",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.User.Username != "newname" {
		t.Fatalf("expected username newname, got %s", result.User.Username)
	}
}

func TestUsersRemove_DeletesUser(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}

		wantPath := "/v2/team/team-1/user/123"
		if r.URL.Path != wantPath {
			t.Fatalf("expected path %s, got %s", wantPath, r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.Users().Remove(context.Background(), "team-1", "123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUsersRemove_RequiresIDs(t *testing.T) {
	t.Parallel()

	client := NewClient("test-key")

	err := client.Users().Remove(context.Background(), "", "123")
	if err == nil {
		t.Fatal("expected error for missing team ID, got nil")
	}

	err = client.Users().Remove(context.Background(), "team-1", "")
	if err == nil {
		t.Fatal("expected error for missing user ID, got nil")
	}
}
