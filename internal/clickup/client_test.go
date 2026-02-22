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

// --- TagsService Tests ---

func TestTagsList_ReturnsSpaceTags(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/space/space-1/tag" {
			t.Fatalf("expected path /v2/space/space-1/tag, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(SpaceTagsResponse{
			Tags: []SpaceTag{
				{Name: "bug", TagFg: "#fff", TagBg: "#f44336"},
				{Name: "feature", TagFg: "#fff", TagBg: "#4caf50"},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Tags().List(context.Background(), "space-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Tags) != 2 {
		t.Fatalf("expected two tags, got %d", len(result.Tags))
	}

	if result.Tags[0].Name != "bug" {
		t.Fatalf("expected first tag name bug, got %s", result.Tags[0].Name)
	}
}

func TestTagsList_RequiresSpaceID(t *testing.T) {
	t.Parallel()

	client := &Client{Client: api.NewClient("test-key")}

	_, err := client.Tags().List(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing space ID, got nil")
	}
}

func TestTagsCreate_SendsTagEnvelope(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/v2/space/space-1/tag" {
			t.Fatalf("expected path /v2/space/space-1/tag, got %s", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		tag, ok := body["tag"].(map[string]any)
		if !ok {
			t.Fatalf("expected tag object in request")
		}

		if tag["name"] != "urgent" {
			t.Fatalf("expected tag.name urgent, got %s", tag["name"])
		}

		if tag["tag_bg"] != "#ff0000" {
			t.Fatalf("expected tag.tag_bg #ff0000, got %s", tag["tag_bg"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.Tags().Create(context.Background(), "space-1", CreateSpaceTagRequest{
		Tag: SpaceTag{Name: "urgent", TagBg: "#ff0000"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTagsCreate_RequiresSpaceID(t *testing.T) {
	t.Parallel()

	client := &Client{Client: api.NewClient("test-key")}

	err := client.Tags().Create(context.Background(), "", CreateSpaceTagRequest{Tag: SpaceTag{Name: "test"}})
	if err == nil {
		t.Fatal("expected error for missing space ID, got nil")
	}
}

func TestTagsCreate_RequiresName(t *testing.T) {
	t.Parallel()

	client := &Client{Client: api.NewClient("test-key")}

	err := client.Tags().Create(context.Background(), "space-1", CreateSpaceTagRequest{Tag: SpaceTag{}})
	if err == nil {
		t.Fatal("expected error for missing name, got nil")
	}
}

func TestTagsUpdate_SendsPutRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}

		// Use RequestURI to check the raw (escaped) request path
		expectedPath := "/v2/space/space-1/tag/old%20tag"
		if r.RequestURI != expectedPath {
			t.Fatalf("expected path %s, got %s", expectedPath, r.RequestURI)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		tag, ok := body["tag"].(map[string]any)
		if !ok {
			t.Fatalf("expected tag object in request")
		}

		if tag["name"] != "new tag" {
			t.Fatalf("expected tag.name new tag, got %s", tag["name"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.Tags().Update(context.Background(), "space-1", "old tag", EditSpaceTagRequest{
		Tag: SpaceTag{Name: "new tag"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTagsUpdate_RequiresSpaceID(t *testing.T) {
	t.Parallel()

	client := &Client{Client: api.NewClient("test-key")}

	err := client.Tags().Update(context.Background(), "", "tag", EditSpaceTagRequest{})
	if err == nil {
		t.Fatal("expected error for missing space ID, got nil")
	}
}

func TestTagsUpdate_RequiresTagName(t *testing.T) {
	t.Parallel()

	client := &Client{Client: api.NewClient("test-key")}

	err := client.Tags().Update(context.Background(), "space-1", "", EditSpaceTagRequest{})
	if err == nil {
		t.Fatal("expected error for missing tag name, got nil")
	}
}

func TestTagsDelete_SendsDeleteRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}

		// Use RequestURI to check the raw (escaped) request path
		expectedPath := "/v2/space/space-1/tag/test%20tag"
		if r.RequestURI != expectedPath {
			t.Fatalf("expected path %s, got %s", expectedPath, r.RequestURI)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.Tags().Delete(context.Background(), "space-1", "test tag")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTagsDelete_RequiresSpaceID(t *testing.T) {
	t.Parallel()

	client := &Client{Client: api.NewClient("test-key")}

	err := client.Tags().Delete(context.Background(), "", "tag")
	if err == nil {
		t.Fatal("expected error for missing space ID, got nil")
	}
}

func TestTagsDelete_RequiresTagName(t *testing.T) {
	t.Parallel()

	client := &Client{Client: api.NewClient("test-key")}

	err := client.Tags().Delete(context.Background(), "space-1", "")
	if err == nil {
		t.Fatal("expected error for missing tag name, got nil")
	}
}

func TestTagsAddToTask_SendsPostRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		expectedPath := "/v2/task/task-1/tag/bug"
		if r.URL.Path != expectedPath {
			t.Fatalf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.Tags().AddToTask(context.Background(), "task-1", "bug")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTagsAddToTask_EscapesTagName(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use RequestURI to check the raw (escaped) request path
		expectedPath := "/v2/task/task-1/tag/needs%20review"
		if r.RequestURI != expectedPath {
			t.Fatalf("expected path %s, got %s", expectedPath, r.RequestURI)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.Tags().AddToTask(context.Background(), "task-1", "needs review")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTagsAddToTask_RequiresTaskID(t *testing.T) {
	t.Parallel()

	client := &Client{Client: api.NewClient("test-key")}

	err := client.Tags().AddToTask(context.Background(), "", "tag")
	if err == nil {
		t.Fatal("expected error for missing task ID, got nil")
	}
}

func TestTagsAddToTask_RequiresTagName(t *testing.T) {
	t.Parallel()

	client := &Client{Client: api.NewClient("test-key")}

	err := client.Tags().AddToTask(context.Background(), "task-1", "")
	if err == nil {
		t.Fatal("expected error for missing tag name, got nil")
	}
}

func TestTagsRemoveFromTask_SendsDeleteRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}

		expectedPath := "/v2/task/task-1/tag/bug"
		if r.URL.Path != expectedPath {
			t.Fatalf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.Tags().RemoveFromTask(context.Background(), "task-1", "bug")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTagsRemoveFromTask_RequiresTaskID(t *testing.T) {
	t.Parallel()

	client := &Client{Client: api.NewClient("test-key")}

	err := client.Tags().RemoveFromTask(context.Background(), "", "tag")
	if err == nil {
		t.Fatal("expected error for missing task ID, got nil")
	}
}

func TestTagsRemoveFromTask_RequiresTagName(t *testing.T) {
	t.Parallel()

	client := &Client{Client: api.NewClient("test-key")}

	err := client.Tags().RemoveFromTask(context.Background(), "task-1", "")
	if err == nil {
		t.Fatal("expected error for missing tag name, got nil")
	}
}
