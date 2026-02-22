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

// --- ViewsService Tests ---

func TestViewsListByTeam_ReturnsViews(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/team/team-1/view" {
			t.Fatalf("expected path /v2/team/team-1/view, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ViewsResponse{
			Views: []View{
				{ID: "view-1", Name: "Board View", Type: "board"},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Views().ListByTeam(context.Background(), "team-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Views) != 1 {
		t.Fatalf("expected 1 view, got %d", len(result.Views))
	}

	if result.Views[0].Name != "Board View" {
		t.Fatalf("expected name Board View, got %s", result.Views[0].Name)
	}
}

func TestViewsListBySpace_ReturnsViews(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/space/space-1/view" {
			t.Fatalf("expected path /v2/space/space-1/view, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ViewsResponse{
			Views: []View{
				{ID: "view-1", Name: "List View", Type: "list"},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Views().ListBySpace(context.Background(), "space-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Views) != 1 {
		t.Fatalf("expected 1 view, got %d", len(result.Views))
	}
}

func TestViewsListByFolder_ReturnsViews(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/folder/folder-1/view" {
			t.Fatalf("expected path /v2/folder/folder-1/view, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ViewsResponse{
			Views: []View{
				{ID: "view-1", Name: "Calendar View", Type: "calendar"},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Views().ListByFolder(context.Background(), "folder-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Views) != 1 {
		t.Fatalf("expected 1 view, got %d", len(result.Views))
	}
}

func TestViewsListByList_ReturnsViews(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/list/list-1/view" {
			t.Fatalf("expected path /v2/list/list-1/view, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ViewsResponse{
			Views: []View{
				{ID: "view-1", Name: "Gantt View", Type: "gantt"},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Views().ListByList(context.Background(), "list-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Views) != 1 {
		t.Fatalf("expected 1 view, got %d", len(result.Views))
	}
}

func TestViewsListByTeam_RequiresTeamID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This should never be called
		t.Fatal("unexpected HTTP request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Views().ListByTeam(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty team ID, got nil")
	}
}

func TestViewsGet_ReturnsView(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/view/view-1" {
			t.Fatalf("expected path /v2/view/view-1, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ViewResponse{
			View: View{
				ID:        "view-1",
				Name:      "My View",
				Type:      "board",
				Protected: false,
				Parent:    ViewParent{ID: "space-1", Type: 7},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Views().Get(context.Background(), "view-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Name != "My View" {
		t.Fatalf("expected name My View, got %s", result.Name)
	}
}

func TestViewsGet_RequiresViewID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This should never be called
		t.Fatal("unexpected HTTP request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Views().Get(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty view ID, got nil")
	}
}

func TestViewsGetTasks_ReturnsTasks(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/view/view-1/task" {
			t.Fatalf("expected path /v2/view/view-1/task, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(TasksListResponse{
			Tasks: []Task{
				{ID: "task-1", Name: "Task 1", Status: TaskStatus{Status: "open"}},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Views().GetTasks(context.Background(), "view-1", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(result.Tasks))
	}
}

func TestViewsGetTasks_WithPagination(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/view/view-1/task" {
			t.Fatalf("expected path /v2/view/view-1/task, got %s", r.URL.Path)
		}

		if r.URL.Query().Get("page") != "2" {
			t.Fatalf("expected page=2, got %s", r.URL.Query().Get("page"))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(TasksListResponse{
			Tasks: []Task{},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Views().GetTasks(context.Background(), "view-1", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestViewsCreateByTeam_ReturnsView(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/v2/team/team-1/view" {
			t.Fatalf("expected path /v2/team/team-1/view, got %s", r.URL.Path)
		}

		var req CreateViewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if req.Name != "New View" {
			t.Fatalf("expected name New View, got %s", req.Name)
		}

		if req.Type != "board" {
			t.Fatalf("expected type board, got %s", req.Type)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ViewResponse{
			View: View{ID: "view-1", Name: "New View", Type: "board"},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Views().CreateByTeam(context.Background(), "team-1", CreateViewRequest{
		Name: "New View",
		Type: "board",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Name != "New View" {
		t.Fatalf("expected name New View, got %s", result.Name)
	}
}

func TestViewsCreateByTeam_RequiresTeamID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This should never be called
		t.Fatal("unexpected HTTP request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Views().CreateByTeam(context.Background(), "", CreateViewRequest{Name: "Test", Type: "board"})
	if err == nil {
		t.Fatal("expected error for empty team ID, got nil")
	}
}

func TestViewsCreateByTeam_RequiresName(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This should never be called
		t.Fatal("unexpected HTTP request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Views().CreateByTeam(context.Background(), "team-1", CreateViewRequest{Type: "board"})
	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
}

func TestViewsCreateBySpace_ReturnsView(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/v2/space/space-1/view" {
			t.Fatalf("expected path /v2/space/space-1/view, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ViewResponse{
			View: View{ID: "view-1", Name: "New View", Type: "list"},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Views().CreateBySpace(context.Background(), "space-1", CreateViewRequest{
		Name: "New View",
		Type: "list",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Type != "list" {
		t.Fatalf("expected type list, got %s", result.Type)
	}
}

func TestViewsCreateByFolder_ReturnsView(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/v2/folder/folder-1/view" {
			t.Fatalf("expected path /v2/folder/folder-1/view, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ViewResponse{
			View: View{ID: "view-1", Name: "New View", Type: "gantt"},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Views().CreateByFolder(context.Background(), "folder-1", CreateViewRequest{
		Name: "New View",
		Type: "gantt",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestViewsCreateByList_ReturnsView(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/v2/list/list-1/view" {
			t.Fatalf("expected path /v2/list/list-1/view, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ViewResponse{
			View: View{ID: "view-1", Name: "New View", Type: "calendar"},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Views().CreateByList(context.Background(), "list-1", CreateViewRequest{
		Name: "New View",
		Type: "calendar",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestViewsUpdate_ReturnsView(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}

		if r.URL.Path != "/v2/view/view-1" {
			t.Fatalf("expected path /v2/view/view-1, got %s", r.URL.Path)
		}

		var req UpdateViewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if req.Name != "Updated Name" {
			t.Fatalf("expected name Updated Name, got %s", req.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ViewResponse{
			View: View{ID: "view-1", Name: "Updated Name", Type: "board"},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Views().Update(context.Background(), "view-1", UpdateViewRequest{Name: "Updated Name"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Name != "Updated Name" {
		t.Fatalf("expected name Updated Name, got %s", result.Name)
	}
}

func TestViewsUpdate_RequiresViewID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This should never be called
		t.Fatal("unexpected HTTP request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Views().Update(context.Background(), "", UpdateViewRequest{Name: "Test"})
	if err == nil {
		t.Fatal("expected error for empty view ID, got nil")
	}
}

func TestViewsDelete_SendsDeleteRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}

		if r.URL.Path != "/v2/view/view-1" {
			t.Fatalf("expected path /v2/view/view-1, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.Views().Delete(context.Background(), "view-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestViewsDelete_RequiresViewID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This should never be called
		t.Fatal("unexpected HTTP request")
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.Views().Delete(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty view ID, got nil")
	}
}
