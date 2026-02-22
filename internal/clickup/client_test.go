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

// SharedHierarchy Service Tests

func TestSharedHierarchyList_ReturnsResources(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		if r.URL.Path != "/v2/team/team-1/shared" {
			t.Fatalf("expected path /v2/team/team-1/shared, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(SharedHierarchyResponse{
			Shared: SharedResources{
				Tasks: []TaskRef{
					{ID: "task-1", Name: "Shared Task"},
				},
				Lists: []ListRef{
					{ID: "list-1", Name: "Shared List"},
				},
				Folders: []FolderRef{
					{ID: "folder-1", Name: "Shared Folder"},
				},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.SharedHierarchy().List(context.Background(), "team-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Shared.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(result.Shared.Tasks))
	}

	if result.Shared.Tasks[0].Name != "Shared Task" {
		t.Fatalf("expected task name Shared Task, got %s", result.Shared.Tasks[0].Name)
	}

	if len(result.Shared.Lists) != 1 {
		t.Fatalf("expected 1 list, got %d", len(result.Shared.Lists))
	}

	if result.Shared.Lists[0].Name != "Shared List" {
		t.Fatalf("expected list name Shared List, got %s", result.Shared.Lists[0].Name)
	}

	if len(result.Shared.Folders) != 1 {
		t.Fatalf("expected 1 folder, got %d", len(result.Shared.Folders))
	}

	if result.Shared.Folders[0].Name != "Shared Folder" {
		t.Fatalf("expected folder name Shared Folder, got %s", result.Shared.Folders[0].Name)
	}
}

func TestSharedHierarchyList_RequiresTeamID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.SharedHierarchy().List(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing team ID, got nil")
	}
}

// Templates Service Tests

func TestTemplatesList_ReturnsTemplates(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		if r.URL.Path != "/v2/team/team-1/taskTemplate" {
			t.Fatalf("expected path /v2/team/team-1/taskTemplate, got %s", r.URL.Path)
		}

		if r.URL.Query().Get("page") != "0" {
			t.Fatalf("expected page=0, got %s", r.URL.Query().Get("page"))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(TaskTemplatesResponse{
			Templates: []TaskTemplate{
				{ID: "tpl-1", Name: "Weekly Report"},
				{ID: "tpl-2", Name: "Bug Report"},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Templates().List(context.Background(), "team-1", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Templates) != 2 {
		t.Fatalf("expected 2 templates, got %d", len(result.Templates))
	}

	if result.Templates[0].Name != "Weekly Report" {
		t.Fatalf("expected first template name Weekly Report, got %s", result.Templates[0].Name)
	}
}

func TestTemplatesList_WithPage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "2" {
			t.Fatalf("expected page=2, got %s", r.URL.Query().Get("page"))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(TaskTemplatesResponse{
			Templates: []TaskTemplate{},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Templates().List(context.Background(), "team-1", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Templates) != 0 {
		t.Fatalf("expected 0 templates, got %d", len(result.Templates))
	}
}

func TestTemplatesList_RequiresTeamID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Templates().List(context.Background(), "", 0)
	if err == nil {
		t.Fatal("expected error for missing team ID, got nil")
	}
}

// --- CustomTaskTypesService tests ---

func TestCustomTaskTypesList_ReturnsTypes(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		if r.URL.Path != "/v2/team/team-1/custom_item" {
			t.Fatalf("expected path /v2/team/team-1/custom_item, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CustomTaskTypesResponse{
			CustomItems: []CustomTaskType{
				{ID: 1, Name: "Task", NamePlural: "Tasks", Description: "Default task type"},
				{ID: 2, Name: "Bug", NamePlural: "Bugs", Description: "Bug reports"},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.CustomTaskTypes().List(context.Background(), "team-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.CustomItems) != 2 {
		t.Fatalf("expected 2 custom task types, got %d", len(result.CustomItems))
	}

	if result.CustomItems[0].Name != "Task" {
		t.Fatalf("expected first type name Task, got %s", result.CustomItems[0].Name)
	}
}

func TestCustomTaskTypesList_RequiresTeamID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.CustomTaskTypes().List(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing team ID, got nil")
	}
}

// --- LegacyTimeService tests ---

func TestLegacyTimeList_ReturnsIntervals(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		if r.URL.Path != "/v2/task/task-1/time" {
			t.Fatalf("expected path /v2/task/task-1/time, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LegacyTimeResponse{
			Data: []LegacyTimeInterval{
				{ID: "567", Start: 1567780450202, End: 1508369194377, Time: 8640000, Source: "clickup"},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.LegacyTime().List(context.Background(), "task-1", false, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Data) != 1 {
		t.Fatalf("expected 1 interval, got %d", len(result.Data))
	}

	if result.Data[0].ID != "567" {
		t.Fatalf("expected interval ID 567, got %s", result.Data[0].ID)
	}
}

func TestLegacyTimeList_WithQueryParams(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("custom_task_ids") != "true" {
			t.Fatalf("expected custom_task_ids=true, got %s", r.URL.Query().Get("custom_task_ids"))
		}

		if r.URL.Query().Get("team_id") != "team-123" {
			t.Fatalf("expected team_id=team-123, got %s", r.URL.Query().Get("team_id"))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LegacyTimeResponse{Data: []LegacyTimeInterval{}})
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.LegacyTime().List(context.Background(), "task-1", true, "team-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLegacyTimeList_RequiresTaskID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.LegacyTime().List(context.Background(), "", false, "")
	if err == nil {
		t.Fatal("expected error for missing task ID, got nil")
	}
}

func TestLegacyTimeTrack_CreatesInterval(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/v2/task/task-1/time" {
			t.Fatalf("expected path /v2/task/task-1/time, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(TrackTimeResponse{ID: "567"})
	}))
	defer server.Close()

	client := newTestClient(server)

	req := TrackTimeRequest{Time: 8640000}

	result, err := client.LegacyTime().Track(context.Background(), "task-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "567" {
		t.Fatalf("expected interval ID 567, got %s", result.ID)
	}
}

func TestLegacyTimeEdit_UpdatesInterval(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}

		if r.URL.Path != "/v2/task/task-1/time/567" {
			t.Fatalf("expected path /v2/task/task-1/time/567, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	req := EditTimeRequest{Time: 3600000}

	if err := client.LegacyTime().Edit(context.Background(), "task-1", "567", req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLegacyTimeDelete_RemovesInterval(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}

		if r.URL.Path != "/v2/task/task-1/time/567" {
			t.Fatalf("expected path /v2/task/task-1/time/567, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	if err := client.LegacyTime().Delete(context.Background(), "task-1", "567"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLegacyTimeEdit_RequiresIDs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	if err := client.LegacyTime().Edit(context.Background(), "", "567", EditTimeRequest{}); err == nil {
		t.Fatal("expected error for missing task ID, got nil")
	}

	if err := client.LegacyTime().Edit(context.Background(), "task-1", "", EditTimeRequest{}); err == nil {
		t.Fatal("expected error for missing interval ID, got nil")
	}
}

// --- AuditLogsService tests ---

func TestAuditLogsQuery_ReturnsLogs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/v3/workspaces/workspace-1/auditlogs" {
			t.Fatalf("expected path /v3/workspaces/workspace-1/auditlogs, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(AuditLogsResponse{
			AuditLogs: []AuditLogEntry{
				{ID: "al_123", EventType: "task_created", UserID: "1", Timestamp: "1700000000000", ResourceType: "task", ResourceID: "abc123"},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)
	client.workspaceID = "workspace-1"

	req := AuditLogQuery{Limit: 100}

	result, err := client.AuditLogs().Query(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.AuditLogs) != 1 {
		t.Fatalf("expected 1 audit log, got %d", len(result.AuditLogs))
	}

	if result.AuditLogs[0].EventType != "task_created" {
		t.Fatalf("expected event type task_created, got %s", result.AuditLogs[0].EventType)
	}
}

func TestAuditLogsQuery_RequiresWorkspaceID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)
	// No workspaceID set

	req := AuditLogQuery{Limit: 100}

	_, err := client.AuditLogs().Query(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for missing workspace ID, got nil")
	}
}

// --- ACLsService tests ---

func TestACLsUpdate_SendsPatchRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("expected PATCH, got %s", r.Method)
		}

		if r.URL.Path != "/v3/workspaces/workspace-1/space/space-789/acls" {
			t.Fatalf("expected path /v3/workspaces/workspace-1/space/space-789/acls, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)
	client.workspaceID = "workspace-1"

	privateTrue := true
	req := UpdateACLRequest{Private: &privateTrue}

	if err := client.ACLs().Update(context.Background(), "space", "space-789", req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestACLsUpdate_RequiresWorkspaceID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)
	// No workspaceID set

	req := UpdateACLRequest{}

	if err := client.ACLs().Update(context.Background(), "space", "space-789", req); err == nil {
		t.Fatal("expected error for missing workspace ID, got nil")
	}
}

func TestACLsUpdate_RequiresObjectTypeAndID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)
	client.workspaceID = "workspace-1"

	req := UpdateACLRequest{}

	if err := client.ACLs().Update(context.Background(), "", "space-789", req); err == nil {
		t.Fatal("expected error for missing object type, got nil")
	}

	if err := client.ACLs().Update(context.Background(), "space", "", req); err == nil {
		t.Fatal("expected error for missing object ID, got nil")
	}
}

// --- WorkspacesService tests ---

func TestWorkspacesList_ReturnsWorkspaces(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		if r.URL.Path != "/v2/team" {
			t.Fatalf("expected path /v2/team, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(WorkspacesResponse{
			Teams: []Workspace{
				{ID: "team-1", Name: "Robben Media", Color: "#4194f6", Members: []Member{{User: User{ID: 1, Username: "alice"}}}},
				{ID: "team-2", Name: "Client Workspace", Color: "#e74c3c", Members: []Member{{User: User{ID: 2, Username: "bob"}}}},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Workspaces().List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Teams) != 2 {
		t.Fatalf("expected 2 workspaces, got %d", len(result.Teams))
	}

	if result.Teams[0].Name != "Robben Media" {
		t.Fatalf("expected first workspace name Robben Media, got %s", result.Teams[0].Name)
	}
}

func TestWorkspacesPlan_ReturnsPlan(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		if r.URL.Path != "/v2/team/team-1/plan" {
			t.Fatalf("expected path /v2/team/team-1/plan, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(WorkspacePlanResponse{
			TeamID:   "team-1",
			PlanID:   3,
			PlanName: "Business",
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Workspaces().Plan(context.Background(), "team-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TeamID != "team-1" {
		t.Fatalf("expected team ID team-1, got %s", result.TeamID)
	}

	if result.PlanName != "Business" {
		t.Fatalf("expected plan name Business, got %s", result.PlanName)
	}

	if result.PlanID != 3 {
		t.Fatalf("expected plan ID 3, got %d", result.PlanID)
	}
}

func TestWorkspacesSeats_ReturnsSeats(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		if r.URL.Path != "/v2/team/team-1/seats" {
			t.Fatalf("expected path /v2/team/team-1/seats, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(WorkspaceSeatsResponse{
			Members: SeatInfo{FilledSeats: 5, TotalSeats: 10, EmptySeats: 5},
			Guests:  SeatInfo{FilledSeats: 2, TotalSeats: 5, EmptySeats: 3},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Workspaces().Seats(context.Background(), "team-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Members.FilledSeats != 5 {
		t.Fatalf("expected members filled seats 5, got %d", result.Members.FilledSeats)
	}

	if result.Members.TotalSeats != 10 {
		t.Fatalf("expected members total seats 10, got %d", result.Members.TotalSeats)
	}

	if result.Guests.FilledSeats != 2 {
		t.Fatalf("expected guests filled seats 2, got %d", result.Guests.FilledSeats)
	}
}

func TestWorkspacesPlan_RequiresTeamID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Workspaces().Plan(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing team ID, got nil")
	}
}

func TestWorkspacesSeats_RequiresTeamID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Workspaces().Seats(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing team ID, got nil")
	}
}

// --- AuthService tests ---

func TestAuthWhoami_ReturnsUser(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		if r.URL.Path != "/v2/user" {
			t.Fatalf("expected path /v2/user, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(AuthorizedUserResponse{
			User: AuthUser{
				ID:             1,
				Username:       "jeremy",
				Email:          "jeremy@example.com",
				Color:          "#4194f6",
				ProfilePicture: "https://example.com/avatar.png",
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Auth().Whoami(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.User.ID != 1 {
		t.Fatalf("expected user ID 1, got %d", result.User.ID)
	}

	if result.User.Username != "jeremy" {
		t.Fatalf("expected username jeremy, got %s", result.User.Username)
	}

	if result.User.Email != "jeremy@example.com" {
		t.Fatalf("expected email jeremy@example.com, got %s", result.User.Email)
	}
}

func TestAuthToken_ExchangeCodeForToken(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/v2/oauth/token" {
			t.Fatalf("expected path /v2/oauth/token, got %s", r.URL.Path)
		}

		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if body["client_id"] != "MYID" {
			t.Fatalf("expected client_id MYID, got %s", body["client_id"])
		}

		if body["client_secret"] != "MYSECRET" {
			t.Fatalf("expected client_secret MYSECRET, got %s", body["client_secret"])
		}

		if body["code"] != "MYCODE" {
			t.Fatalf("expected code MYCODE, got %s", body["code"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(OAuthTokenResponse{
			AccessToken: "MYTOKEN",
			TokenType:   "Bearer",
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	req := OAuthTokenRequest{
		ClientID:     "MYID",
		ClientSecret: "MYSECRET",
		Code:         "MYCODE",
	}

	result, err := client.Auth().Token(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.AccessToken != "MYTOKEN" {
		t.Fatalf("expected access token MYTOKEN, got %s", result.AccessToken)
	}

	if result.TokenType != "Bearer" {
		t.Fatalf("expected token type Bearer, got %s", result.TokenType)
	}
}

func TestAuthToken_RequiresAllFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	// Missing ClientID
	_, err := client.Auth().Token(context.Background(), OAuthTokenRequest{
		ClientSecret: "secret",
		Code:         "code",
	})
	if err == nil {
		t.Fatal("expected error for missing client_id, got nil")
	}

	// Missing ClientSecret
	_, err = client.Auth().Token(context.Background(), OAuthTokenRequest{
		ClientID: "id",
		Code:     "code",
	})
	if err == nil {
		t.Fatal("expected error for missing client_secret, got nil")
	}

	// Missing Code
	_, err = client.Auth().Token(context.Background(), OAuthTokenRequest{
		ClientID:     "id",
		ClientSecret: "secret",
	})
	if err == nil {
		t.Fatal("expected error for missing code, got nil")
	}
}

// --- SpacesService extended tests ---

func TestSpacesGet_ReturnsDetails(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		if r.URL.Path != "/v2/space/space-1" {
			t.Fatalf("expected path /v2/space/space-1, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(SpaceDetail{
			ID:      "space-1",
			Name:    "Engineering",
			Private: false,
			Statuses: []SpaceStatus{
				{Status: "open", Color: "#d3d3d3", OrderIndex: 0},
				{Status: "in progress", Color: "#4194f6", OrderIndex: 1},
			},
			MultipleAssignees: true,
			Features: SpaceFeatures{
				DueDates:     FeatureToggle{Enabled: true},
				TimeTracking: FeatureToggle{Enabled: true},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Spaces().Get(context.Background(), "space-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "space-1" {
		t.Fatalf("expected ID space-1, got %s", result.ID)
	}

	if result.Name != "Engineering" {
		t.Fatalf("expected name Engineering, got %s", result.Name)
	}

	if len(result.Statuses) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(result.Statuses))
	}

	if !result.Features.DueDates.Enabled {
		t.Fatal("expected due_dates feature to be enabled")
	}
}

func TestSpacesCreate_SendsNameAndOptions(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/v2/team/team-1/space" {
			t.Fatalf("expected path /v2/team/team-1/space, got %s", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if body["name"] != "New Space" {
			t.Fatalf("expected name New Space, got %s", body["name"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(SpaceDetail{
			ID:      "space-new",
			Name:    "New Space",
			Private: false,
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	req := CreateSpaceRequest{Name: "New Space"}

	result, err := client.Spaces().Create(context.Background(), "team-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "space-new" {
		t.Fatalf("expected ID space-new, got %s", result.ID)
	}
}

func TestSpacesUpdate_SendsFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}

		if r.URL.Path != "/v2/space/space-1" {
			t.Fatalf("expected path /v2/space/space-1, got %s", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if body["name"] != "Updated Name" {
			t.Fatalf("expected name Updated Name, got %s", body["name"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(SpaceDetail{
			ID:      "space-1",
			Name:    "Updated Name",
			Private: true,
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	req := UpdateSpaceRequest{Name: "Updated Name"}

	result, err := client.Spaces().Update(context.Background(), "space-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Name != "Updated Name" {
		t.Fatalf("expected name Updated Name, got %s", result.Name)
	}
}

func TestSpacesDelete_SendsCorrectRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}

		if r.URL.Path != "/v2/space/space-1" {
			t.Fatalf("expected path /v2/space/space-1, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	if err := client.Spaces().Delete(context.Background(), "space-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSpacesCreate_RequiresTeamID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Spaces().Create(context.Background(), "", CreateSpaceRequest{Name: "Test"})
	if err == nil {
		t.Fatal("expected error for missing team ID, got nil")
	}
}

func TestSpacesCreate_RequiresName(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Spaces().Create(context.Background(), "team-1", CreateSpaceRequest{})
	if err == nil {
		t.Fatal("expected error for missing name, got nil")
	}
}

func TestSpacesGet_RequiresSpaceID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Spaces().Get(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing space ID, got nil")
	}
}

func TestSpacesUpdate_RequiresSpaceID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Spaces().Update(context.Background(), "", UpdateSpaceRequest{})
	if err == nil {
		t.Fatal("expected error for missing space ID, got nil")
	}
}

func TestSpacesDelete_RequiresSpaceID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	if err := client.Spaces().Delete(context.Background(), ""); err == nil {
		t.Fatal("expected error for missing space ID, got nil")
	}
}

// --- ListsService extended tests ---

func TestListsGet_ReturnsDetails(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		if r.URL.Path != "/v2/list/list-1" {
			t.Fatalf("expected path /v2/list/list-1, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ListDetail{
			ID:        "list-1",
			Name:      "Sprint 1",
			TaskCount: 15,
			Folder:    FolderRef{ID: "folder-1", Name: "Backlog"},
			Space:     SpaceRef{ID: "space-1", Name: "Engineering"},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Lists().Get(context.Background(), "list-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "list-1" {
		t.Fatalf("expected ID list-1, got %s", result.ID)
	}

	if result.Name != "Sprint 1" {
		t.Fatalf("expected name Sprint 1, got %s", result.Name)
	}

	if result.TaskCount != 15 {
		t.Fatalf("expected task count 15, got %d", result.TaskCount)
	}
}

func TestListsCreateInFolder_SendsName(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/v2/folder/folder-1/list" {
			t.Fatalf("expected path /v2/folder/folder-1/list, got %s", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if body["name"] != "New List" {
			t.Fatalf("expected name New List, got %s", body["name"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ListDetail{
			ID:   "list-new",
			Name: "New List",
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	req := CreateListRequest{Name: "New List"}

	result, err := client.Lists().CreateInFolder(context.Background(), "folder-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "list-new" {
		t.Fatalf("expected ID list-new, got %s", result.ID)
	}
}

func TestListsCreateFolderless_SendsName(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/v2/space/space-1/list" {
			t.Fatalf("expected path /v2/space/space-1/list, got %s", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if body["name"] != "New List" {
			t.Fatalf("expected name New List, got %s", body["name"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ListDetail{
			ID:   "list-new",
			Name: "New List",
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	req := CreateListRequest{Name: "New List"}

	result, err := client.Lists().CreateFolderless(context.Background(), "space-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "list-new" {
		t.Fatalf("expected ID list-new, got %s", result.ID)
	}
}

func TestListsUpdate_SendsFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}

		if r.URL.Path != "/v2/list/list-1" {
			t.Fatalf("expected path /v2/list/list-1, got %s", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if body["name"] != "Updated Name" {
			t.Fatalf("expected name Updated Name, got %s", body["name"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ListDetail{
			ID:   "list-1",
			Name: "Updated Name",
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	req := UpdateListRequest{Name: "Updated Name"}

	result, err := client.Lists().Update(context.Background(), "list-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Name != "Updated Name" {
		t.Fatalf("expected name Updated Name, got %s", result.Name)
	}
}

func TestListsDelete_SendsCorrectRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}

		if r.URL.Path != "/v2/list/list-1" {
			t.Fatalf("expected path /v2/list/list-1, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	if err := client.Lists().Delete(context.Background(), "list-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListsAddTask_SendsCorrectRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/v2/list/list-1/task/task-1" {
			t.Fatalf("expected path /v2/list/list-1/task/task-1, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	if err := client.Lists().AddTask(context.Background(), "list-1", "task-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListsRemoveTask_SendsCorrectRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}

		if r.URL.Path != "/v2/list/list-1/task/task-1" {
			t.Fatalf("expected path /v2/list/list-1/task/task-1, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	if err := client.Lists().RemoveTask(context.Background(), "list-1", "task-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListsCreateInFolder_RequiresName(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Lists().CreateInFolder(context.Background(), "folder-1", CreateListRequest{})
	if err == nil {
		t.Fatal("expected error for missing name, got nil")
	}
}

func TestListsGet_RequiresListID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Lists().Get(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing list ID, got nil")
	}
}
