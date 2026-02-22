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

// UserGroups Service Tests

func TestUserGroupsList_ReturnsGroups(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		if r.URL.Path != "/v2/group" {
			t.Fatalf("expected path /v2/group, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(UserGroupsResponse{
			Groups: []UserGroup{
				{ID: "group-1", Name: "Engineering", Members: []User{{ID: 1, Username: "alice"}}},
				{ID: "group-2", Name: "Design", Members: []User{{ID: 2, Username: "bob"}}},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.UserGroups().List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(result.Groups))
	}

	if result.Groups[0].Name != "Engineering" {
		t.Fatalf("expected first group name Engineering, got %s", result.Groups[0].Name)
	}
}

func TestUserGroupsCreate_SendsNameAndMembers(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/v2/team/team-1/group" {
			t.Fatalf("expected path /v2/team/team-1/group, got %s", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if body["name"] != "New Group" {
			t.Fatalf("expected name New Group, got %s", body["name"])
		}

		members, ok := body["members"].([]any)
		if !ok || len(members) != 2 {
			t.Fatalf("expected members [1, 2], got %#v", body["members"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(UserGroup{ID: "group-new", Name: "New Group"})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.UserGroups().Create(context.Background(), "team-1", CreateUserGroupRequest{
		Name:    "New Group",
		Members: []int{1, 2},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "group-new" {
		t.Fatalf("expected ID group-new, got %s", result.ID)
	}
}

func TestUserGroupsUpdate_SendsMembersAddRem(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}

		if r.URL.Path != "/v2/group/group-1" {
			t.Fatalf("expected path /v2/group/group-1, got %s", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if body["name"] != "Updated Name" {
			t.Fatalf("expected name Updated Name, got %s", body["name"])
		}

		members, ok := body["members"].(map[string]any)
		if !ok {
			t.Fatalf("expected members object in request")
		}

		add, ok := members["add"].([]any)
		if !ok || len(add) != 1 || int(add[0].(float64)) != 10 {
			t.Fatalf("expected members.add [10], got %#v", members["add"])
		}

		rem, ok := members["rem"].([]any)
		if !ok || len(rem) != 1 || int(rem[0].(float64)) != 20 {
			t.Fatalf("expected members.rem [20], got %#v", members["rem"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(UserGroup{ID: "group-1", Name: "Updated Name"})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.UserGroups().Update(context.Background(), "group-1", UpdateUserGroupRequest{
		Name: "Updated Name",
		Members: &UserGroupMembersUpdate{
			Add: []int{10},
			Rem: []int{20},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Name != "Updated Name" {
		t.Fatalf("expected name Updated Name, got %s", result.Name)
	}
}

func TestUserGroupsDelete_SendsCorrectRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}

		if r.URL.Path != "/v2/group/group-1" {
			t.Fatalf("expected path /v2/group/group-1, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.UserGroups().Delete(context.Background(), "group-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserGroupsCreate_RequiresTeamID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.UserGroups().Create(context.Background(), "", CreateUserGroupRequest{Name: "Test"})
	if err == nil {
		t.Fatal("expected error for missing team ID, got nil")
	}
}

func TestUserGroupsCreate_RequiresName(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.UserGroups().Create(context.Background(), "team-1", CreateUserGroupRequest{})
	if err == nil {
		t.Fatal("expected error for missing name, got nil")
	}
}

func TestRolesList_ReturnsCustomRoles(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		if r.URL.Path != "/v2/team/team-1/customroles" {
			t.Fatalf("expected path /v2/team/team-1/customroles, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CustomRolesResponse{
			CustomRoles: []CustomRole{
				{ID: 1, Name: "Project Manager", Permissions: []string{"task_create", "task_delete"}},
				{ID: 2, Name: "Developer", Permissions: []string{"task_create", "task_edit"}},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Roles().List(context.Background(), "team-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.CustomRoles) != 2 {
		t.Fatalf("expected 2 custom roles, got %d", len(result.CustomRoles))
	}

	if result.CustomRoles[0].Name != "Project Manager" {
		t.Fatalf("expected name Project Manager, got %s", result.CustomRoles[0].Name)
	}

	if len(result.CustomRoles[0].Permissions) != 2 {
		t.Fatalf("expected 2 permissions, got %d", len(result.CustomRoles[0].Permissions))
	}
}

func TestRolesList_RequiresTeamID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Roles().List(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing team ID, got nil")
	}
}

// Guests Service Tests

func TestGuestsGet_ReturnsGuest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		if r.URL.Path != "/v2/team/team-1/guest/123" {
			t.Fatalf("expected path /v2/team/team-1/guest/123, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(GuestResponse{
			Guest: Guest{ID: 123, Username: "client-bob", Email: "bob@client.com", TasksCount: 3},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Guests().Get(context.Background(), "team-1", 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != 123 {
		t.Fatalf("expected ID 123, got %d", result.ID)
	}

	if result.Username != "client-bob" {
		t.Fatalf("expected username client-bob, got %s", result.Username)
	}
}

func TestGuestsInvite_SendsEmailAndPermissions(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/v2/team/team-1/guest" {
			t.Fatalf("expected path /v2/team/team-1/guest, got %s", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if body["email"] != "bob@client.com" {
			t.Fatalf("expected email bob@client.com, got %s", body["email"])
		}

		if body["can_edit_tags"] != true {
			t.Fatalf("expected can_edit_tags true, got %v", body["can_edit_tags"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(GuestResponse{
			Guest: Guest{ID: 123, Username: "client-bob", Email: "bob@client.com"},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Guests().Invite(context.Background(), "team-1", InviteGuestRequest{
		Email:       "bob@client.com",
		CanEditTags: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != 123 {
		t.Fatalf("expected ID 123, got %d", result.ID)
	}
}

func TestGuestsUpdate_SendsPermissions(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}

		if r.URL.Path != "/v2/team/team-1/guest/123" {
			t.Fatalf("expected path /v2/team/team-1/guest/123, got %s", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if body["can_edit_tags"] != true {
			t.Fatalf("expected can_edit_tags true, got %v", body["can_edit_tags"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(GuestResponse{
			Guest: Guest{ID: 123, Username: "client-bob", Email: "bob@client.com"},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	canEdit := true

	result, err := client.Guests().Update(context.Background(), "team-1", 123, EditGuestRequest{
		CanEditTags: &canEdit,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != 123 {
		t.Fatalf("expected ID 123, got %d", result.ID)
	}
}

func TestGuestsRemove_SendsCorrectRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}

		if r.URL.Path != "/v2/team/team-1/guest/123" {
			t.Fatalf("expected path /v2/team/team-1/guest/123, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.Guests().Remove(context.Background(), "team-1", 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGuestsAddToTask_SendsPermission(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/v2/task/task-1/guest/123" {
			t.Fatalf("expected path /v2/task/task-1/guest/123, got %s", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if body["permission_level"] != "edit" {
			t.Fatalf("expected permission_level edit, got %s", body["permission_level"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(GuestResponse{
			Guest: Guest{ID: 123, Username: "client-bob", Email: "bob@client.com", TasksCount: 1},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Guests().AddToTask(context.Background(), "task-1", 123, "edit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TasksCount != 1 {
		t.Fatalf("expected tasks_count 1, got %d", result.TasksCount)
	}
}

func TestGuestsRemoveFromTask_SendsCorrectRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}

		if r.URL.Path != "/v2/task/task-1/guest/123" {
			t.Fatalf("expected path /v2/task/task-1/guest/123, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.Guests().RemoveFromTask(context.Background(), "task-1", 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGuestsAddToList_SendsPermission(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/v2/list/list-1/guest/123" {
			t.Fatalf("expected path /v2/list/list-1/guest/123, got %s", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if body["permission_level"] != "read" {
			t.Fatalf("expected permission_level read, got %s", body["permission_level"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(GuestResponse{
			Guest: Guest{ID: 123, Username: "client-bob", Email: "bob@client.com", ListsCount: 1},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Guests().AddToList(context.Background(), "list-1", 123, "read")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ListsCount != 1 {
		t.Fatalf("expected lists_count 1, got %d", result.ListsCount)
	}
}

func TestGuestsRemoveFromList_SendsCorrectRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}

		if r.URL.Path != "/v2/list/list-1/guest/123" {
			t.Fatalf("expected path /v2/list/list-1/guest/123, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.Guests().RemoveFromList(context.Background(), "list-1", 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGuestsAddToFolder_SendsPermission(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/v2/folder/folder-1/guest/123" {
			t.Fatalf("expected path /v2/folder/folder-1/guest/123, got %s", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if body["permission_level"] != "comment" {
			t.Fatalf("expected permission_level comment, got %s", body["permission_level"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(GuestResponse{
			Guest: Guest{ID: 123, Username: "client-bob", Email: "bob@client.com", FoldersCount: 1},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Guests().AddToFolder(context.Background(), "folder-1", 123, "comment")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.FoldersCount != 1 {
		t.Fatalf("expected folders_count 1, got %d", result.FoldersCount)
	}
}

func TestGuestsRemoveFromFolder_SendsCorrectRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}

		if r.URL.Path != "/v2/folder/folder-1/guest/123" {
			t.Fatalf("expected path /v2/folder/folder-1/guest/123, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.Guests().RemoveFromFolder(context.Background(), "folder-1", 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGuestsInvite_RequiresEmail(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Guests().Invite(context.Background(), "team-1", InviteGuestRequest{})
	if err == nil {
		t.Fatal("expected error for missing email, got nil")
	}
}
