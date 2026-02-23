package clickup

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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

// --- TasksService extended tests ---

func TestTasksSearch_ReturnsFilteredTasks(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		if r.URL.Path != "/v2/team/team-1/task" {
			t.Fatalf("expected path /v2/team/team-1/task, got %s", r.URL.Path)
		}

		// Check query parameters
		if r.URL.Query().Get("statuses[]") != "open" {
			t.Fatalf("expected statuses[]=open, got %s", r.URL.Query().Get("statuses[]"))
		}

		if r.URL.Query().Get("include_closed") != "true" {
			t.Fatalf("expected include_closed=true, got %s", r.URL.Query().Get("include_closed"))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(FilteredTeamTasksResponse{
			Tasks: []Task{
				{ID: "task-1", Name: "Fix bug", Status: TaskStatus{Status: "open"}},
				{ID: "task-2", Name: "Write docs", Status: TaskStatus{Status: "open"}},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	params := FilteredTeamTasksParams{
		Statuses:      []string{"open"},
		IncludeClosed: true,
	}

	result, err := client.Tasks().Search(context.Background(), "team-1", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(result.Tasks))
	}

	if result.Tasks[0].Name != "Fix bug" {
		t.Fatalf("expected first task name Fix bug, got %s", result.Tasks[0].Name)
	}
}

func TestTasksSearch_WithPagination(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "2" {
			t.Fatalf("expected page=2, got %s", r.URL.Query().Get("page"))
		}

		if r.URL.Query().Get("order_by") != "due_date" {
			t.Fatalf("expected order_by=due_date, got %s", r.URL.Query().Get("order_by"))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(FilteredTeamTasksResponse{Tasks: []Task{}})
	}))
	defer server.Close()

	client := newTestClient(server)

	params := FilteredTeamTasksParams{
		Page:    2,
		OrderBy: "due_date",
	}

	result, err := client.Tasks().Search(context.Background(), "team-1", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Tasks) != 0 {
		t.Fatalf("expected 0 tasks, got %d", len(result.Tasks))
	}
}

func TestTasksSearch_RequiresTeamID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Tasks().Search(context.Background(), "", FilteredTeamTasksParams{})
	if err == nil {
		t.Fatal("expected error for missing team ID, got nil")
	}
}

func TestTasksTimeInStatus_ReturnsData(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		if r.URL.Path != "/v2/task/task-1/time_in_status" {
			t.Fatalf("expected path /v2/task/task-1/time_in_status, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(TimeInStatusResponse{
			CurrentStatus: &StatusTime{
				Status: "in progress",
				Color:  "#4194f6",
				TotalTime: TimeValue{
					ByMinute: 120,
					Since:    "1700000000000",
				},
			},
			StatusHistory: []StatusTime{
				{
					Status: "open",
					TotalTime: TimeValue{
						ByMinute: 60,
						Since:    "1699900000000",
					},
				},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Tasks().TimeInStatus(context.Background(), "task-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.CurrentStatus == nil {
		t.Fatal("expected current status, got nil")
	}

	if result.CurrentStatus.Status != "in progress" {
		t.Fatalf("expected status in progress, got %s", result.CurrentStatus.Status)
	}

	if result.CurrentStatus.TotalTime.ByMinute != 120 {
		t.Fatalf("expected 120 minutes, got %d", result.CurrentStatus.TotalTime.ByMinute)
	}

	if len(result.StatusHistory) != 1 {
		t.Fatalf("expected 1 status history entry, got %d", len(result.StatusHistory))
	}
}

func TestTasksTimeInStatus_RequiresTaskID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Tasks().TimeInStatus(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing task ID, got nil")
	}
}

func TestTasksBulkTimeInStatus_ReturnsMap(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		if r.URL.Path != "/v2/task/bulk_time_in_status/task_ids" {
			t.Fatalf("expected path /v2/task/bulk_time_in_status/task_ids, got %s", r.URL.Path)
		}

		// Check that task_ids are passed
		taskIDs := r.URL.Query()["task_ids"]
		if len(taskIDs) != 2 {
			t.Fatalf("expected 2 task_ids, got %d", len(taskIDs))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(BulkTimeInStatusResponse{
			"task-1": {
				CurrentStatus: &StatusTime{
					Status:    "in progress",
					TotalTime: TimeValue{ByMinute: 120},
				},
			},
			"task-2": {
				CurrentStatus: &StatusTime{
					Status:    "done",
					TotalTime: TimeValue{ByMinute: 30},
				},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Tasks().BulkTimeInStatus(context.Background(), []string{"task-1", "task-2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 tasks in result, got %d", len(result))
	}

	if result["task-1"].CurrentStatus.Status != "in progress" {
		t.Fatalf("expected task-1 status in progress, got %s", result["task-1"].CurrentStatus.Status)
	}
}

func TestTasksBulkTimeInStatus_RequiresTaskIDs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Tasks().BulkTimeInStatus(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for missing task IDs, got nil")
	}
}

func TestTasksMerge_SendsSourceIDs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/v2/task/target-1/merge" {
			t.Fatalf("expected path /v2/task/target-1/merge, got %s", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		mergedIDs, ok := body["merged_task_ids"].([]any)
		if !ok || len(mergedIDs) != 2 {
			t.Fatalf("expected merged_task_ids with 2 items, got %#v", body["merged_task_ids"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(MergeTasksResponse{ID: "target-1"})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Tasks().Merge(context.Background(), "target-1", []string{"source-1", "source-2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "target-1" {
		t.Fatalf("expected ID target-1, got %s", result.ID)
	}
}

func TestTasksMerge_RequiresTargetID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Tasks().Merge(context.Background(), "", []string{"source-1"})
	if err == nil {
		t.Fatal("expected error for missing target task ID, got nil")
	}
}

func TestTasksMerge_RequiresSourceIDs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Tasks().Merge(context.Background(), "target-1", nil)
	if err == nil {
		t.Fatal("expected error for missing source task IDs, got nil")
	}
}

func TestTasksMove_UsesV3Path(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}

		if r.URL.Path != "/v3/workspaces/workspace-1/tasks/task-1/home_list/list-1" {
			t.Fatalf("expected path /v3/workspaces/workspace-1/tasks/task-1/home_list/list-1, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(MoveTaskResponse{
			Status:  "success",
			TaskID:  "task-1",
			ListID:  "list-1",
			Message: "Task moved",
		})
	}))
	defer server.Close()

	client := newTestClient(server)
	client.workspaceID = "workspace-1"

	result, err := client.Tasks().Move(context.Background(), "task-1", "list-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != "success" {
		t.Fatalf("expected status success, got %s", result.Status)
	}

	if result.TaskID != "task-1" {
		t.Fatalf("expected task ID task-1, got %s", result.TaskID)
	}
}

func TestTasksMove_RequiresWorkspaceID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)
	// No workspaceID set

	_, err := client.Tasks().Move(context.Background(), "task-1", "list-1")
	if err == nil {
		t.Fatal("expected error for missing workspace ID, got nil")
	}
}

func TestTasksMove_RequiresTaskAndListID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)
	client.workspaceID = "workspace-1"

	_, err := client.Tasks().Move(context.Background(), "", "list-1")
	if err == nil {
		t.Fatal("expected error for missing task ID, got nil")
	}

	_, err = client.Tasks().Move(context.Background(), "task-1", "")
	if err == nil {
		t.Fatal("expected error for missing list ID, got nil")
	}
}

func TestTasksCreateFromTemplate_ReturnsTask(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/v2/list/list-1/taskTemplate/template-1" {
			t.Fatalf("expected path /v2/list/list-1/taskTemplate/template-1, got %s", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if name, ok := body["name"]; ok && name != "Custom Name" {
			t.Fatalf("expected name Custom Name, got %s", name)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Task{
			ID:     "task-new",
			Name:   "Custom Name",
			Status: TaskStatus{Status: "open"},
			URL:    "https://app.clickup.com/t/task-new",
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Tasks().CreateFromTemplate(context.Background(), "list-1", "template-1", CreateTaskFromTemplateRequest{
		Name: "Custom Name",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "task-new" {
		t.Fatalf("expected ID task-new, got %s", result.ID)
	}

	if result.Name != "Custom Name" {
		t.Fatalf("expected name Custom Name, got %s", result.Name)
	}
}

func TestTasksCreateFromTemplate_RequiresIDs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Tasks().CreateFromTemplate(context.Background(), "", "template-1", CreateTaskFromTemplateRequest{})
	if err == nil {
		t.Fatal("expected error for missing list ID, got nil")
	}

	_, err = client.Tasks().CreateFromTemplate(context.Background(), "list-1", "", CreateTaskFromTemplateRequest{})
	if err == nil {
		t.Fatal("expected error for missing template ID, got nil")
	}
}

// --- FoldersService tests ---

func TestFoldersGet_ReturnsDetails(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		if r.URL.Path != "/v2/folder/folder-1" {
			t.Fatalf("expected path /v2/folder/folder-1, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(FolderDetail{
			ID:         "folder-1",
			Name:       "Sprint Backlog",
			TaskCount:  "12",
			OrderIndex: 0,
			Space:      SpaceRef{ID: "space-1"},
			Lists:      []List{{ID: "list-1", Name: "Sprint 1"}},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Folders().Get(context.Background(), "folder-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "folder-1" {
		t.Fatalf("expected ID folder-1, got %s", result.ID)
	}

	if result.Name != "Sprint Backlog" {
		t.Fatalf("expected name Sprint Backlog, got %s", result.Name)
	}

	if result.TaskCount != "12" {
		t.Fatalf("expected task count 12, got %s", result.TaskCount)
	}

	if len(result.Lists) != 1 {
		t.Fatalf("expected 1 list, got %d", len(result.Lists))
	}
}

func TestFoldersCreate_SendsName(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/v2/space/space-1/folder" {
			t.Fatalf("expected path /v2/space/space-1/folder, got %s", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if body["name"] != "New Folder" {
			t.Fatalf("expected name New Folder, got %s", body["name"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(FolderDetail{
			ID:    "folder-new",
			Name:  "New Folder",
			Space: SpaceRef{ID: "space-1"},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	req := CreateFolderRequest{Name: "New Folder"}

	result, err := client.Folders().Create(context.Background(), "space-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "folder-new" {
		t.Fatalf("expected ID folder-new, got %s", result.ID)
	}
}

func TestFoldersUpdate_SendsName(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}

		if r.URL.Path != "/v2/folder/folder-1" {
			t.Fatalf("expected path /v2/folder/folder-1, got %s", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if body["name"] != "Updated Name" {
			t.Fatalf("expected name Updated Name, got %s", body["name"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(FolderDetail{
			ID:   "folder-1",
			Name: "Updated Name",
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	req := UpdateFolderRequest{Name: "Updated Name"}

	result, err := client.Folders().Update(context.Background(), "folder-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Name != "Updated Name" {
		t.Fatalf("expected name Updated Name, got %s", result.Name)
	}
}

func TestFoldersDelete_SendsCorrectRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}

		if r.URL.Path != "/v2/folder/folder-1" {
			t.Fatalf("expected path /v2/folder/folder-1, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	if err := client.Folders().Delete(context.Background(), "folder-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFoldersCreate_RequiresName(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Folders().Create(context.Background(), "space-1", CreateFolderRequest{})
	if err == nil {
		t.Fatal("expected error for missing name, got nil")
	}
}

func TestFoldersGet_RequiresFolderID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Folders().Get(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing folder ID, got nil")
	}
}

// --- CommentsService extended tests ---

func TestCommentsDelete_SendsCorrectRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}

		if r.URL.Path != "/v2/comment/comment-1" {
			t.Fatalf("expected path /v2/comment/comment-1, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	if err := client.Comments().Delete(context.Background(), "comment-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCommentsDelete_RequiresCommentID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	if err := client.Comments().Delete(context.Background(), ""); err == nil {
		t.Fatal("expected error for missing comment ID, got nil")
	}
}

func TestCommentsUpdate_SendsFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}

		if r.URL.Path != "/v2/comment/comment-1" {
			t.Fatalf("expected path /v2/comment/comment-1, got %s", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if body["comment_text"] != "Updated text" {
			t.Fatalf("expected comment_text Updated text, got %s", body["comment_text"])
		}

		if body["resolved"] != true {
			t.Fatalf("expected resolved true, got %v", body["resolved"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	resolved := true
	req := UpdateCommentRequest{
		CommentText: "Updated text",
		Resolved:    &resolved,
	}

	if err := client.Comments().Update(context.Background(), "comment-1", req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCommentsUpdate_RequiresCommentID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	if err := client.Comments().Update(context.Background(), "", UpdateCommentRequest{}); err == nil {
		t.Fatal("expected error for missing comment ID, got nil")
	}
}

func TestCommentsReplies_ReturnsThreadedComments(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		if r.URL.Path != "/v2/comment/comment-1/reply" {
			t.Fatalf("expected path /v2/comment/comment-1/reply, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ThreadedCommentsResponse{
			Comments: []Comment{
				{ID: json.Number("456"), Text: "I agree", User: User{ID: 1, Username: "alice"}, Date: "1700000000000"},
				{ID: json.Number("789"), Text: "Me too", User: User{ID: 2, Username: "bob"}, Date: "1700000001000"},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Comments().Replies(context.Background(), "comment-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Comments) != 2 {
		t.Fatalf("expected 2 replies, got %d", len(result.Comments))
	}

	if result.Comments[0].Text != "I agree" {
		t.Fatalf("expected first reply text 'I agree', got %s", result.Comments[0].Text)
	}
}

func TestCommentsReplies_RequiresCommentID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Comments().Replies(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing comment ID, got nil")
	}
}

func TestCommentsReply_CreatesThreadedReply(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/v2/comment/comment-1/reply" {
			t.Fatalf("expected path /v2/comment/comment-1/reply, got %s", r.URL.Path)
		}

		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if body["comment_text"] != "My reply" {
			t.Fatalf("expected comment_text My reply, got %s", body["comment_text"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Comment{
			ID:   json.Number("456"),
			Text: "My reply",
			User: User{ID: 1, Username: "testuser"},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Comments().Reply(context.Background(), "comment-1", "My reply")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID.String() != "456" {
		t.Fatalf("expected ID 456, got %s", result.ID.String())
	}

	if result.Text != "My reply" {
		t.Fatalf("expected text My reply, got %s", result.Text)
	}
}

func TestCommentsReply_RequiresText(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Comments().Reply(context.Background(), "comment-1", "")
	if err == nil {
		t.Fatal("expected error for missing text, got nil")
	}
}

func TestCommentsListComments_ReturnsListComments(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		if r.URL.Path != "/v2/list/list-1/comment" {
			t.Fatalf("expected path /v2/list/list-1/comment, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CommentsListResponse{
			Comments: []Comment{
				{ID: json.Number("123"), Text: "List comment", User: User{ID: 1, Username: "alice"}, Date: "1700000000000"},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Comments().ListComments(context.Background(), "list-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(result.Comments))
	}

	if result.Comments[0].Text != "List comment" {
		t.Fatalf("expected comment text 'List comment', got %s", result.Comments[0].Text)
	}
}

func TestCommentsListComments_RequiresListID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Comments().ListComments(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing list ID, got nil")
	}
}

func TestCommentsAddList_CreatesListComment(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/v2/list/list-1/comment" {
			t.Fatalf("expected path /v2/list/list-1/comment, got %s", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if body["comment_text"] != "List comment" {
			t.Fatalf("expected comment_text List comment, got %s", body["comment_text"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]int{"id": 456})
	}))
	defer server.Close()

	client := newTestClient(server)

	req := CreateListCommentRequest{CommentText: "List comment"}

	result, err := client.Comments().AddList(context.Background(), "list-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID.String() != "456" {
		t.Fatalf("expected ID 456, got %s", result.ID.String())
	}
}

func TestCommentsAddList_RequiresText(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Comments().AddList(context.Background(), "list-1", CreateListCommentRequest{})
	if err == nil {
		t.Fatal("expected error for missing text, got nil")
	}
}

func TestCommentsViewComments_ReturnsViewComments(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		if r.URL.Path != "/v2/view/view-1/comment" {
			t.Fatalf("expected path /v2/view/view-1/comment, got %s", r.URL.Path)
		}

		if r.URL.Query().Get("start") != "10" {
			t.Fatalf("expected start=10, got %s", r.URL.Query().Get("start"))
		}

		if r.URL.Query().Get("start_id") != "cursor-abc" {
			t.Fatalf("expected start_id=cursor-abc, got %s", r.URL.Query().Get("start_id"))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CommentsListResponse{
			Comments: []Comment{
				{ID: json.Number("123"), Text: "View comment", User: User{ID: 1, Username: "alice"}, Date: "1700000000000"},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Comments().ViewComments(context.Background(), "view-1", 10, "cursor-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(result.Comments))
	}

	if result.Comments[0].Text != "View comment" {
		t.Fatalf("expected comment text 'View comment', got %s", result.Comments[0].Text)
	}
}

func TestCommentsViewComments_RequiresViewID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Comments().ViewComments(context.Background(), "", 0, "")
	if err == nil {
		t.Fatal("expected error for missing view ID, got nil")
	}
}

func TestCommentsAddView_CreatesViewComment(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/v2/view/view-1/comment" {
			t.Fatalf("expected path /v2/view/view-1/comment, got %s", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if body["comment_text"] != "View comment" {
			t.Fatalf("expected comment_text View comment, got %s", body["comment_text"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]int{"id": 456})
	}))
	defer server.Close()

	client := newTestClient(server)

	req := CreateViewCommentRequest{CommentText: "View comment"}

	result, err := client.Comments().AddView(context.Background(), "view-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID.String() != "456" {
		t.Fatalf("expected ID 456, got %s", result.ID.String())
	}
}

func TestCommentsAddView_RequiresText(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Comments().AddView(context.Background(), "view-1", CreateViewCommentRequest{})
	if err == nil {
		t.Fatal("expected error for missing text, got nil")
	}
}

func TestCommentsSubtypes_ReturnsSubtypes(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		if r.URL.Path != "/v3/workspaces/workspace-1/comments/types/type-1/subtypes" {
			t.Fatalf("expected path /v3/workspaces/workspace-1/comments/types/type-1/subtypes, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(PostSubtypesResponse{
			Subtypes: []PostSubtype{
				{ID: "1", Name: "announcement"},
				{ID: "2", Name: "question"},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)
	client.workspaceID = "workspace-1"

	result, err := client.Comments().Subtypes(context.Background(), "type-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Subtypes) != 2 {
		t.Fatalf("expected 2 subtypes, got %d", len(result.Subtypes))
	}

	if result.Subtypes[0].Name != "announcement" {
		t.Fatalf("expected first subtype name announcement, got %s", result.Subtypes[0].Name)
	}
}

func TestCommentsSubtypes_RequiresTypeID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)
	client.workspaceID = "workspace-1"

	_, err := client.Comments().Subtypes(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing type ID, got nil")
	}
}

func TestCommentsSubtypes_RequiresWorkspaceID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)
	// No workspaceID set

	_, err := client.Comments().Subtypes(context.Background(), "type-1")
	if err == nil {
		t.Fatal("expected error for missing workspace ID, got nil")
	}
}

// Time Service extension tests

func TestTimeGet_ReturnsEntry(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/team/team-1/time_entries/123" {
			t.Fatalf("expected path /v2/team/team-1/time_entries/123, got %s", r.URL.Path)
		}

		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(TimeEntryDetailResponse{
			Data: TimeEntryDetail{
				ID:          json.Number("123"),
				Task:        TaskRef{ID: "task-1", Name: "Test Task"},
				Wid:         "workspace-1",
				User:        User{ID: 1, Username: "testuser"},
				Billable:    true,
				Start:       json.Number("1700000000000"),
				End:         json.Number("1700003600000"),
				Duration:    json.Number("3600000"),
				Description: "Test description",
				Tags:        []Tag{{Name: "billable"}},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Time().Get(context.Background(), "team-1", "123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID.String() != "123" {
		t.Fatalf("expected ID 123, got %s", result.ID)
	}

	if result.Task.ID != "task-1" {
		t.Fatalf("expected task ID task-1, got %s", result.Task.ID)
	}

	if !result.Billable {
		t.Fatal("expected billable to be true")
	}
}

func TestTimeGet_RequiresIDs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Time().Get(context.Background(), "", "entry-1")
	if err == nil {
		t.Fatal("expected error for missing team ID, got nil")
	}

	_, err = client.Time().Get(context.Background(), "team-1", "")
	if err == nil {
		t.Fatal("expected error for missing entry ID, got nil")
	}
}

func TestTimeCurrent_ReturnsRunningTimer(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/team/team-1/time_entries/current" {
			t.Fatalf("expected path /v2/team/team-1/time_entries/current, got %s", r.URL.Path)
		}

		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(TimeEntryDetailResponse{
			Data: TimeEntryDetail{
				ID:          json.Number("124"),
				Task:        TaskRef{ID: "task-1", Name: "Test Task"},
				Start:       json.Number("1700000000000"),
				Duration:    json.Number("-1700000000000"),
				Description: "Working on it",
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Time().Current(context.Background(), "team-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID.String() != "124" {
		t.Fatalf("expected ID 124, got %s", result.ID)
	}
}

func TestTimeCurrent_RequiresTeamID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Time().Current(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing team ID, got nil")
	}
}

func TestTimeStart_SendsRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/team/team-1/time_entries/start" {
			t.Fatalf("expected path /v2/team/team-1/time_entries/start, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		var req StartTimeEntryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.TaskID != "task-1" {
			t.Fatalf("expected task ID task-1, got %s", req.TaskID)
		}

		if req.Description != "Working on it" {
			t.Fatalf("expected description 'Working on it', got %s", req.Description)
		}

		if !req.Billable {
			t.Fatal("expected billable to be true")
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(TimeEntryDetailResponse{
			Data: TimeEntryDetail{
				ID:          json.Number("125"),
				Task:        TaskRef{ID: "task-1", Name: "Test Task"},
				Start:       json.Number("1700000000000"),
				Duration:    json.Number("-1700000000000"),
				Description: "Working on it",
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	req := StartTimeEntryRequest{
		TaskID:      "task-1",
		Description: "Working on it",
		Billable:    true,
	}

	result, err := client.Time().Start(context.Background(), "team-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID.String() != "125" {
		t.Fatalf("expected ID 125, got %s", result.ID)
	}
}

func TestTimeStart_RequiresTeamID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Time().Start(context.Background(), "", StartTimeEntryRequest{})
	if err == nil {
		t.Fatal("expected error for missing team ID, got nil")
	}
}

func TestTimeStop_SendsRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/team/team-1/time_entries/stop" {
			t.Fatalf("expected path /v2/team/team-1/time_entries/stop, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(TimeEntryDetailResponse{
			Data: TimeEntryDetail{
				ID:       json.Number("126"),
				Task:     TaskRef{ID: "task-1", Name: "Test Task"},
				Start:    json.Number("1700000000000"),
				End:      json.Number("1700003600000"),
				Duration: json.Number("3600000"),
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Time().Stop(context.Background(), "team-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID.String() != "126" {
		t.Fatalf("expected ID 126, got %s", result.ID)
	}
}

func TestTimeStop_RequiresTeamID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Time().Stop(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing team ID, got nil")
	}
}

func TestTimeUpdate_SendsRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/team/team-1/time_entries/entry-1" {
			t.Fatalf("expected path /v2/team/team-1/time_entries/entry-1, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}

		var req UpdateTimeEntryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Description != "Updated description" {
			t.Fatalf("expected description 'Updated description', got %s", req.Description)
		}

		if req.Duration != 7200000 {
			t.Fatalf("expected duration 7200000, got %d", req.Duration)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(TimeEntryDetailResponse{
			Data: TimeEntryDetail{
				ID:          json.Number("123"),
				Task:        TaskRef{ID: "task-1", Name: "Test Task"},
				Start:       json.Number("1700000000000"),
				End:         json.Number("1700007200000"),
				Duration:    json.Number("7200000"),
				Description: "Updated description",
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	billable := true
	req := UpdateTimeEntryRequest{
		Description: "Updated description",
		Duration:    7200000,
		Billable:    &billable,
	}

	result, err := client.Time().Update(context.Background(), "team-1", "entry-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Duration.String() != "7200000" {
		t.Fatalf("expected duration 7200000, got %s", result.Duration)
	}
}

func TestTimeUpdate_RequiresIDs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Time().Update(context.Background(), "", "entry-1", UpdateTimeEntryRequest{})
	if err == nil {
		t.Fatal("expected error for missing team ID, got nil")
	}

	_, err = client.Time().Update(context.Background(), "team-1", "", UpdateTimeEntryRequest{})
	if err == nil {
		t.Fatal("expected error for missing entry ID, got nil")
	}
}

func TestTimeDelete_SendsRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/team/team-1/time_entries/entry-1" {
			t.Fatalf("expected path /v2/team/team-1/time_entries/entry-1, got %s", r.URL.Path)
		}

		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	if err := client.Time().Delete(context.Background(), "team-1", "entry-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTimeDelete_RequiresIDs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.Time().Delete(context.Background(), "", "entry-1")
	if err == nil {
		t.Fatal("expected error for missing team ID, got nil")
	}

	err = client.Time().Delete(context.Background(), "team-1", "")
	if err == nil {
		t.Fatal("expected error for missing entry ID, got nil")
	}
}

func TestTimeHistory_ReturnsHistory(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/team/team-1/time_entries/entry-1/history" {
			t.Fatalf("expected path /v2/team/team-1/time_entries/entry-1/history, got %s", r.URL.Path)
		}

		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(TimeEntryHistoryResponse{
			Data: []TimeEntryHistoryItem{
				{
					ID:     "1",
					Field:  "duration",
					Before: "3600000",
					After:  "7200000",
					Date:   "1700100000000",
					User:   User{ID: 1, Username: "testuser"},
				},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Time().History(context.Background(), "team-1", "entry-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Data) != 1 {
		t.Fatalf("expected 1 history item, got %d", len(result.Data))
	}

	if result.Data[0].Field != "duration" {
		t.Fatalf("expected field duration, got %s", result.Data[0].Field)
	}
}

func TestTimeHistory_RequiresIDs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Time().History(context.Background(), "", "entry-1")
	if err == nil {
		t.Fatal("expected error for missing team ID, got nil")
	}

	_, err = client.Time().History(context.Background(), "team-1", "")
	if err == nil {
		t.Fatal("expected error for missing entry ID, got nil")
	}
}

func TestTimeListTags_ReturnsTags(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/team/team-1/time_entries/tags" {
			t.Fatalf("expected path /v2/team/team-1/time_entries/tags, got %s", r.URL.Path)
		}

		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(TimeEntryTagsResponse{
			Data: []TimeEntryTag{
				{Name: "billable"},
				{Name: "internal"},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Time().ListTags(context.Background(), "team-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Data) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(result.Data))
	}

	if result.Data[0].Name != "billable" {
		t.Fatalf("expected first tag billable, got %s", result.Data[0].Name)
	}
}

func TestTimeListTags_RequiresTeamID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Time().ListTags(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing team ID, got nil")
	}
}

func TestTimeAddTags_SendsRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/team/team-1/time_entries/tags" {
			t.Fatalf("expected path /v2/team/team-1/time_entries/tags, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		var req TimeEntryTagsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if len(req.TimeEntryIDs) != 2 {
			t.Fatalf("expected 2 entry IDs, got %d", len(req.TimeEntryIDs))
		}

		if len(req.Tags) != 1 {
			t.Fatalf("expected 1 tag, got %d", len(req.Tags))
		}

		if req.Tags[0].Name != "billable" {
			t.Fatalf("expected tag name billable, got %s", req.Tags[0].Name)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	req := TimeEntryTagsRequest{
		TimeEntryIDs: []string{"entry-1", "entry-2"},
		Tags:         []Tag{{Name: "billable"}},
	}

	if err := client.Time().AddTags(context.Background(), "team-1", req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTimeAddTags_RequiresIDsAndTags(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.Time().AddTags(context.Background(), "team-1", TimeEntryTagsRequest{})
	if err == nil {
		t.Fatal("expected error for missing entry IDs and tags, got nil")
	}

	err = client.Time().AddTags(context.Background(), "", TimeEntryTagsRequest{TimeEntryIDs: []string{"1"}, Tags: []Tag{{Name: "tag"}}})
	if err == nil {
		t.Fatal("expected error for missing team ID, got nil")
	}
}

func TestTimeRemoveTags_SendsRequestWithBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/team/team-1/time_entries/tags" {
			t.Fatalf("expected path /v2/team/team-1/time_entries/tags, got %s", r.URL.Path)
		}

		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}

		var req TimeEntryTagsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if len(req.TimeEntryIDs) != 1 {
			t.Fatalf("expected 1 entry ID, got %d", len(req.TimeEntryIDs))
		}

		if len(req.Tags) != 1 {
			t.Fatalf("expected 1 tag, got %d", len(req.Tags))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	req := TimeEntryTagsRequest{
		TimeEntryIDs: []string{"entry-1"},
		Tags:         []Tag{{Name: "billable"}},
	}

	if err := client.Time().RemoveTags(context.Background(), "team-1", req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTimeRemoveTags_RequiresIDsAndTags(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.Time().RemoveTags(context.Background(), "team-1", TimeEntryTagsRequest{})
	if err == nil {
		t.Fatal("expected error for missing entry IDs and tags, got nil")
	}

	err = client.Time().RemoveTags(context.Background(), "", TimeEntryTagsRequest{TimeEntryIDs: []string{"1"}, Tags: []Tag{{Name: "tag"}}})
	if err == nil {
		t.Fatal("expected error for missing team ID, got nil")
	}
}

func TestTimeRenameTag_SendsRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/team/team-1/time_entries/tags" {
			t.Fatalf("expected path /v2/team/team-1/time_entries/tags, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}

		var req RenameTimeEntryTagRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Name != "old-tag" {
			t.Fatalf("expected old name old-tag, got %s", req.Name)
		}

		if req.NewName != "new-tag" {
			t.Fatalf("expected new name new-tag, got %s", req.NewName)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	req := RenameTimeEntryTagRequest{
		Name:    "old-tag",
		NewName: "new-tag",
	}

	if err := client.Time().RenameTag(context.Background(), "team-1", req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTimeRenameTag_RequiresNames(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.Time().RenameTag(context.Background(), "team-1", RenameTimeEntryTagRequest{})
	if err == nil {
		t.Fatal("expected error for missing names, got nil")
	}

	err = client.Time().RenameTag(context.Background(), "team-1", RenameTimeEntryTagRequest{Name: "old"})
	if err == nil {
		t.Fatal("expected error for missing new name, got nil")
	}

	err = client.Time().RenameTag(context.Background(), "team-1", RenameTimeEntryTagRequest{NewName: "new"})
	if err == nil {
		t.Fatal("expected error for missing old name, got nil")
	}

	err = client.Time().RenameTag(context.Background(), "", RenameTimeEntryTagRequest{Name: "old", NewName: "new"})
	if err == nil {
		t.Fatal("expected error for missing team ID, got nil")
	}
}

// Tags Service tests

func TestTagsList_ReturnsTags(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/space/space-1/tag" {
			t.Fatalf("expected path /v2/space/space-1/tag, got %s", r.URL.Path)
		}

		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
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
		t.Fatalf("expected 2 tags, got %d", len(result.Tags))
	}

	if result.Tags[0].Name != "bug" {
		t.Fatalf("expected first tag bug, got %s", result.Tags[0].Name)
	}

	if result.Tags[0].TagBg != "#f44336" {
		t.Fatalf("expected background #f44336, got %s", result.Tags[0].TagBg)
	}
}

func TestTagsList_RequiresSpaceID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Tags().List(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing space ID, got nil")
	}
}

func TestTagsCreate_SendsRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/space/space-1/tag" {
			t.Fatalf("expected path /v2/space/space-1/tag, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		var req CreateSpaceTagRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Tag.Name != "urgent" {
			t.Fatalf("expected tag name urgent, got %s", req.Tag.Name)
		}

		if req.Tag.TagBg != "#ff0000" {
			t.Fatalf("expected background #ff0000, got %s", req.Tag.TagBg)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	req := CreateSpaceTagRequest{
		Tag: SpaceTag{
			Name:  "urgent",
			TagBg: "#ff0000",
			TagFg: "#ffffff",
		},
	}

	if err := client.Tags().Create(context.Background(), "space-1", req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTagsCreate_RequiresSpaceIDAndName(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.Tags().Create(context.Background(), "", CreateSpaceTagRequest{Tag: SpaceTag{Name: "test"}})
	if err == nil {
		t.Fatal("expected error for missing space ID, got nil")
	}

	err = client.Tags().Create(context.Background(), "space-1", CreateSpaceTagRequest{Tag: SpaceTag{}})
	if err == nil {
		t.Fatal("expected error for missing tag name, got nil")
	}
}

func TestTagsUpdate_SendsRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Tag name should be URL encoded in the path
		if r.URL.Path != "/v2/space/space-1/tag/old+tag" {
			t.Fatalf("expected path /v2/space/space-1/tag/old+tag, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}

		var req EditSpaceTagRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Tag.Name != "new tag" {
			t.Fatalf("expected tag name 'new tag', got %s", req.Tag.Name)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	req := EditSpaceTagRequest{
		Tag: SpaceTag{
			Name:  "new tag",
			TagBg: "#00ff00",
		},
	}

	if err := client.Tags().Update(context.Background(), "space-1", "old tag", req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTagsUpdate_RequiresIDs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.Tags().Update(context.Background(), "", "tag", EditSpaceTagRequest{})
	if err == nil {
		t.Fatal("expected error for missing space ID, got nil")
	}

	err = client.Tags().Update(context.Background(), "space-1", "", EditSpaceTagRequest{})
	if err == nil {
		t.Fatal("expected error for missing tag name, got nil")
	}
}

func TestTagsDelete_SendsRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/space/space-1/tag/bug" {
			t.Fatalf("expected path /v2/space/space-1/tag/bug, got %s", r.URL.Path)
		}

		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	if err := client.Tags().Delete(context.Background(), "space-1", "bug"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTagsDelete_RequiresIDs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.Tags().Delete(context.Background(), "", "tag")
	if err == nil {
		t.Fatal("expected error for missing space ID, got nil")
	}

	err = client.Tags().Delete(context.Background(), "space-1", "")
	if err == nil {
		t.Fatal("expected error for missing tag name, got nil")
	}
}

func TestTagsAddToTask_SendsRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/task/task-1/tag/bug" {
			t.Fatalf("expected path /v2/task/task-1/tag/bug, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	if err := client.Tags().AddToTask(context.Background(), "task-1", "bug"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTagsAddToTask_RequiresIDs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.Tags().AddToTask(context.Background(), "", "tag")
	if err == nil {
		t.Fatal("expected error for missing task ID, got nil")
	}

	err = client.Tags().AddToTask(context.Background(), "task-1", "")
	if err == nil {
		t.Fatal("expected error for missing tag name, got nil")
	}
}

func TestTagsRemoveFromTask_SendsRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/task/task-1/tag/bug" {
			t.Fatalf("expected path /v2/task/task-1/tag/bug, got %s", r.URL.Path)
		}

		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	if err := client.Tags().RemoveFromTask(context.Background(), "task-1", "bug"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTagsRemoveFromTask_RequiresIDs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.Tags().RemoveFromTask(context.Background(), "", "tag")
	if err == nil {
		t.Fatal("expected error for missing task ID, got nil")
	}

	err = client.Tags().RemoveFromTask(context.Background(), "task-1", "")
	if err == nil {
		t.Fatal("expected error for missing tag name, got nil")
	}
}

// Checklists Service tests

func TestChecklistsCreate_ReturnsChecklist(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/task/task-1/checklist" {
			t.Fatalf("expected path /v2/task/task-1/checklist, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		var req CreateChecklistRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Name != "QA Steps" {
			t.Fatalf("expected name QA Steps, got %s", req.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ChecklistResponse{
			Checklist: Checklist{
				ID:         "cl-123",
				Name:       "QA Steps",
				OrderIndex: 0,
				Items:      []ChecklistItem{},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	req := CreateChecklistRequest{Name: "QA Steps"}

	result, err := client.Checklists().Create(context.Background(), "task-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "cl-123" {
		t.Fatalf("expected ID cl-123, got %s", result.ID)
	}

	if result.Name != "QA Steps" {
		t.Fatalf("expected name QA Steps, got %s", result.Name)
	}
}

func TestChecklistsCreate_RequiresTaskIDAndName(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Checklists().Create(context.Background(), "", CreateChecklistRequest{Name: "test"})
	if err == nil {
		t.Fatal("expected error for missing task ID, got nil")
	}

	_, err = client.Checklists().Create(context.Background(), "task-1", CreateChecklistRequest{})
	if err == nil {
		t.Fatal("expected error for missing name, got nil")
	}
}

func TestChecklistsUpdate_SendsRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/checklist/cl-123" {
			t.Fatalf("expected path /v2/checklist/cl-123, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ChecklistResponse{
			Checklist: Checklist{
				ID:         "cl-123",
				Name:       "Updated Name",
				OrderIndex: 1,
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	req := EditChecklistRequest{Name: "Updated Name", Position: 1}

	result, err := client.Checklists().Update(context.Background(), "cl-123", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Name != "Updated Name" {
		t.Fatalf("expected name Updated Name, got %s", result.Name)
	}
}

func TestChecklistsUpdate_RequiresChecklistID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Checklists().Update(context.Background(), "", EditChecklistRequest{})
	if err == nil {
		t.Fatal("expected error for missing checklist ID, got nil")
	}
}

func TestChecklistsDelete_SendsRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/checklist/cl-123" {
			t.Fatalf("expected path /v2/checklist/cl-123, got %s", r.URL.Path)
		}

		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	if err := client.Checklists().Delete(context.Background(), "cl-123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestChecklistsDelete_RequiresChecklistID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.Checklists().Delete(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing checklist ID, got nil")
	}
}

func TestChecklistsAddItem_ReturnsChecklist(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/checklist/cl-123/checklist_item" {
			t.Fatalf("expected path /v2/checklist/cl-123/checklist_item, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		var req CreateChecklistItemRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Name != "Step 1" {
			t.Fatalf("expected name Step 1, got %s", req.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ChecklistResponse{
			Checklist: Checklist{
				ID:   "cl-123",
				Name: "QA Steps",
				Items: []ChecklistItem{
					{ID: "ci-456", Name: "Step 1", Resolved: false},
				},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	req := CreateChecklistItemRequest{Name: "Step 1"}

	result, err := client.Checklists().AddItem(context.Background(), "cl-123", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}

	if result.Items[0].ID != "ci-456" {
		t.Fatalf("expected item ID ci-456, got %s", result.Items[0].ID)
	}
}

func TestChecklistsAddItem_RequiresIDsAndName(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Checklists().AddItem(context.Background(), "", CreateChecklistItemRequest{Name: "test"})
	if err == nil {
		t.Fatal("expected error for missing checklist ID, got nil")
	}

	_, err = client.Checklists().AddItem(context.Background(), "cl-123", CreateChecklistItemRequest{})
	if err == nil {
		t.Fatal("expected error for missing name, got nil")
	}
}

func TestChecklistsUpdateItem_SendsRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/checklist/cl-123/checklist_item/ci-456" {
			t.Fatalf("expected path /v2/checklist/cl-123/checklist_item/ci-456, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}

		resolved := true

		var req EditChecklistItemRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if *req.Resolved != resolved {
			t.Fatalf("expected resolved true, got %v", *req.Resolved)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ChecklistResponse{
			Checklist: Checklist{
				ID:   "cl-123",
				Name: "QA Steps",
				Items: []ChecklistItem{
					{ID: "ci-456", Name: "Step 1", Resolved: true},
				},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	resolved := true
	req := EditChecklistItemRequest{Resolved: &resolved}

	result, err := client.Checklists().UpdateItem(context.Background(), "cl-123", "ci-456", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Items[0].Resolved {
		t.Fatal("expected item to be resolved")
	}
}

func TestChecklistsUpdateItem_RequiresIDs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Checklists().UpdateItem(context.Background(), "", "ci-456", EditChecklistItemRequest{})
	if err == nil {
		t.Fatal("expected error for missing checklist ID, got nil")
	}

	_, err = client.Checklists().UpdateItem(context.Background(), "cl-123", "", EditChecklistItemRequest{})
	if err == nil {
		t.Fatal("expected error for missing item ID, got nil")
	}
}

func TestChecklistsDeleteItem_SendsRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/checklist/cl-123/checklist_item/ci-456" {
			t.Fatalf("expected path /v2/checklist/cl-123/checklist_item/ci-456, got %s", r.URL.Path)
		}

		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	if err := client.Checklists().DeleteItem(context.Background(), "cl-123", "ci-456"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestChecklistsDeleteItem_RequiresIDs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.Checklists().DeleteItem(context.Background(), "", "ci-456")
	if err == nil {
		t.Fatal("expected error for missing checklist ID, got nil")
	}

	err = client.Checklists().DeleteItem(context.Background(), "cl-123", "")
	if err == nil {
		t.Fatal("expected error for missing item ID, got nil")
	}
}

// Relationships Service tests

func TestRelationshipsAddDep_SendsRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/task/task-1/dependency" {
			t.Fatalf("expected path /v2/task/task-1/dependency, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		var req AddDependencyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.DependsOn != "task-2" {
			t.Fatalf("expected depends_on task-2, got %s", req.DependsOn)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	req := AddDependencyRequest{DependsOn: "task-2"}
	if err := client.Relationships().AddDependency(context.Background(), "task-1", req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRelationshipsAddDep_RequiresIDs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.Relationships().AddDependency(context.Background(), "", AddDependencyRequest{DependsOn: "task-2"})
	if err == nil {
		t.Fatal("expected error for missing task ID, got nil")
	}

	err = client.Relationships().AddDependency(context.Background(), "task-1", AddDependencyRequest{})
	if err == nil {
		t.Fatal("expected error for missing dependency, got nil")
	}
}

func TestRelationshipsDeleteDep_SendsRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/task/task-1/dependency" {
			// URL might have query params
			if !strings.HasPrefix(r.URL.Path, "/v2/task/task-1/dependency?") {
				t.Fatalf("expected path /v2/task/task-1/dependency, got %s", r.URL.Path)
			}
		}

		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}

		// Check query params
		dependsOn := r.URL.Query().Get("depends_on")
		if dependsOn != "task-2" {
			t.Fatalf("expected depends_on=task-2, got %s", dependsOn)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	req := AddDependencyRequest{DependsOn: "task-2"}
	if err := client.Relationships().DeleteDependency(context.Background(), "task-1", req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRelationshipsDeleteDep_RequiresIDs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.Relationships().DeleteDependency(context.Background(), "", AddDependencyRequest{DependsOn: "task-2"})
	if err == nil {
		t.Fatal("expected error for missing task ID, got nil")
	}

	err = client.Relationships().DeleteDependency(context.Background(), "task-1", AddDependencyRequest{})
	if err == nil {
		t.Fatal("expected error for missing dependency, got nil")
	}
}

func TestRelationshipsAddLink_SendsRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/task/task-1/link/task-2" {
			t.Fatalf("expected path /v2/task/task-1/link/task-2, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	if err := client.Relationships().AddLink(context.Background(), "task-1", "task-2"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRelationshipsAddLink_RequiresIDs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.Relationships().AddLink(context.Background(), "", "task-2")
	if err == nil {
		t.Fatal("expected error for missing task ID, got nil")
	}

	err = client.Relationships().AddLink(context.Background(), "task-1", "")
	if err == nil {
		t.Fatal("expected error for missing linked task ID, got nil")
	}
}

func TestRelationshipsDeleteLink_SendsRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/task/task-1/link/task-2" {
			t.Fatalf("expected path /v2/task/task-1/link/task-2, got %s", r.URL.Path)
		}

		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	if err := client.Relationships().DeleteLink(context.Background(), "task-1", "task-2"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRelationshipsDeleteLink_RequiresIDs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.Relationships().DeleteLink(context.Background(), "", "task-2")
	if err == nil {
		t.Fatal("expected error for missing task ID, got nil")
	}

	err = client.Relationships().DeleteLink(context.Background(), "task-1", "")
	if err == nil {
		t.Fatal("expected error for missing linked task ID, got nil")
	}
}

// CustomFields Service tests

func TestCustomFieldsListByList_ReturnsFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/list/list-1/field" {
			t.Fatalf("expected path /v2/list/list-1/field, got %s", r.URL.Path)
		}

		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CustomFieldsResponse{
			Fields: []CustomField{
				{ID: "cf-123", Name: "Budget", Type: "currency", Required: false},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.CustomFields().ListByList(context.Background(), "list-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(result.Fields))
	}

	if result.Fields[0].Name != "Budget" {
		t.Fatalf("expected field name Budget, got %s", result.Fields[0].Name)
	}
}

func TestCustomFieldsListByList_RequiresID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.CustomFields().ListByList(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing list ID, got nil")
	}
}

func TestCustomFieldsListByFolder_ReturnsFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/folder/folder-1/field" {
			t.Fatalf("expected path /v2/folder/folder-1/field, got %s", r.URL.Path)
		}

		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CustomFieldsResponse{
			Fields: []CustomField{
				{ID: "cf-123", Name: "Priority", Type: "dropdown", Required: true},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.CustomFields().ListByFolder(context.Background(), "folder-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Fields[0].Name != "Priority" {
		t.Fatalf("expected field name Priority, got %s", result.Fields[0].Name)
	}
}

func TestCustomFieldsListByFolder_RequiresID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.CustomFields().ListByFolder(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing folder ID, got nil")
	}
}

func TestCustomFieldsListBySpace_ReturnsFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/space/space-1/field" {
			t.Fatalf("expected path /v2/space/space-1/field, got %s", r.URL.Path)
		}

		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CustomFieldsResponse{
			Fields: []CustomField{},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.CustomFields().ListBySpace(context.Background(), "space-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Fields) != 0 {
		t.Fatalf("expected 0 fields, got %d", len(result.Fields))
	}
}

func TestCustomFieldsListBySpace_RequiresID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.CustomFields().ListBySpace(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing space ID, got nil")
	}
}

func TestCustomFieldsListByTeam_ReturnsFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/team/team-1/field" {
			t.Fatalf("expected path /v2/team/team-1/field, got %s", r.URL.Path)
		}

		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CustomFieldsResponse{
			Fields: []CustomField{},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.CustomFields().ListByTeam(context.Background(), "team-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

func TestCustomFieldsListByTeam_RequiresID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.CustomFields().ListByTeam(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing team ID, got nil")
	}
}

func TestCustomFieldsSet_SendsRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/task/task-1/field/cf-123" {
			t.Fatalf("expected path /v2/task/task-1/field/cf-123, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		var req SetCustomFieldRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		// Value is passed as-is (could be any type)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	if err := client.CustomFields().Set(context.Background(), "task-1", "cf-123", "test value"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCustomFieldsSet_RequiresIDs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.CustomFields().Set(context.Background(), "", "cf-123", "value")
	if err == nil {
		t.Fatal("expected error for missing task ID, got nil")
	}

	err = client.CustomFields().Set(context.Background(), "task-1", "", "value")
	if err == nil {
		t.Fatal("expected error for missing field ID, got nil")
	}
}

func TestCustomFieldsRemove_SendsRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/task/task-1/field/cf-123" {
			t.Fatalf("expected path /v2/task/task-1/field/cf-123, got %s", r.URL.Path)
		}

		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	if err := client.CustomFields().Remove(context.Background(), "task-1", "cf-123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCustomFieldsRemove_RequiresIDs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.CustomFields().Remove(context.Background(), "", "cf-123")
	if err == nil {
		t.Fatal("expected error for missing task ID, got nil")
	}

	err = client.CustomFields().Remove(context.Background(), "task-1", "")
	if err == nil {
		t.Fatal("expected error for missing field ID, got nil")
	}
}

// Views Service tests

func TestViewsListByTeam_ReturnsViews(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/team/team-1/view" {
			t.Fatalf("expected path /v2/team/team-1/view, got %s", r.URL.Path)
		}

		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ViewsResponse{
			Views: []View{
				{ID: "v-123", Name: "Sprint Board", Type: "board", Parent: ViewParent{ID: "789", Type: 7}},
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

	if result.Views[0].Name != "Sprint Board" {
		t.Fatalf("expected view name Sprint Board, got %s", result.Views[0].Name)
	}
}

func TestViewsListByTeam_RequiresID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Views().ListByTeam(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing team ID, got nil")
	}
}

func TestViewsListBySpace_ReturnsViews(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/space/space-1/view" {
			t.Fatalf("expected path /v2/space/space-1/view, got %s", r.URL.Path)
		}

		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ViewsResponse{
			Views: []View{},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Views().ListBySpace(context.Background(), "space-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

func TestViewsListBySpace_RequiresID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Views().ListBySpace(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing space ID, got nil")
	}
}

func TestViewsListByFolder_ReturnsViews(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/folder/folder-1/view" {
			t.Fatalf("expected path /v2/folder/folder-1/view, got %s", r.URL.Path)
		}

		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ViewsResponse{
			Views: []View{},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Views().ListByFolder(context.Background(), "folder-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

func TestViewsListByFolder_RequiresID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Views().ListByFolder(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing folder ID, got nil")
	}
}

func TestViewsListByList_ReturnsViews(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/list/list-1/view" {
			t.Fatalf("expected path /v2/list/list-1/view, got %s", r.URL.Path)
		}

		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ViewsResponse{
			Views: []View{},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Views().ListByList(context.Background(), "list-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

func TestViewsListByList_RequiresID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Views().ListByList(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing list ID, got nil")
	}
}

func TestViewsGet_ReturnsView(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/view/v-123" {
			t.Fatalf("expected path /v2/view/v-123, got %s", r.URL.Path)
		}

		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ViewResponse{
			View: View{ID: "v-123", Name: "My Board", Type: "board", Protected: false},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Views().Get(context.Background(), "v-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Name != "My Board" {
		t.Fatalf("expected view name My Board, got %s", result.Name)
	}
}

func TestViewsGet_RequiresID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Views().Get(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing view ID, got nil")
	}
}

func TestViewsTasks_ReturnsTasks(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/view/v-123/task" {
			t.Fatalf("expected path /v2/view/v-123/task, got %s", r.URL.Path)
		}

		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ViewTasksResponse{
			Tasks: []Task{
				{ID: "task-1", Name: "Task 1"},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Views().Tasks(context.Background(), "v-123", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(result.Tasks))
	}
}

func TestViewsTasks_WithPage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/view/v-123/task" {
			t.Fatalf("expected path /v2/view/v-123/task, got %s", r.URL.Path)
		}

		if r.URL.Query().Get("page") != "2" {
			t.Fatalf("expected page=2, got %s", r.URL.Query().Get("page"))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ViewTasksResponse{Tasks: []Task{}})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Views().Tasks(context.Background(), "v-123", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

func TestViewsTasks_RequiresID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Views().Tasks(context.Background(), "", 0)
	if err == nil {
		t.Fatal("expected error for missing view ID, got nil")
	}
}

func TestViewsCreateInTeam_SendsRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/team/team-1/view" {
			t.Fatalf("expected path /v2/team/team-1/view, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		var req CreateViewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Name != "New View" {
			t.Fatalf("expected name New View, got %s", req.Name)
		}

		if req.Type != "board" {
			t.Fatalf("expected type board, got %s", req.Type)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ViewResponse{
			View: View{ID: "v-new", Name: "New View", Type: "board"},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	req := CreateViewRequest{Name: "New View", Type: "board"}

	result, err := client.Views().CreateInTeam(context.Background(), "team-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "v-new" {
		t.Fatalf("expected ID v-new, got %s", result.ID)
	}
}

func TestViewsCreateInTeam_RequiresFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Views().CreateInTeam(context.Background(), "", CreateViewRequest{Name: "Test", Type: "list"})
	if err == nil {
		t.Fatal("expected error for missing team ID, got nil")
	}

	_, err = client.Views().CreateInTeam(context.Background(), "team-1", CreateViewRequest{Type: "list"})
	if err == nil {
		t.Fatal("expected error for missing name, got nil")
	}

	_, err = client.Views().CreateInTeam(context.Background(), "team-1", CreateViewRequest{Name: "Test"})
	if err == nil {
		t.Fatal("expected error for missing type, got nil")
	}
}

func TestViewsCreateInSpace_SendsRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/space/space-1/view" {
			t.Fatalf("expected path /v2/space/space-1/view, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ViewResponse{
			View: View{ID: "v-new", Name: "New View", Type: "calendar"},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	req := CreateViewRequest{Name: "New View", Type: "calendar"}

	result, err := client.Views().CreateInSpace(context.Background(), "space-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Type != "calendar" {
		t.Fatalf("expected type calendar, got %s", result.Type)
	}
}

func TestViewsCreateInFolder_SendsRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/folder/folder-1/view" {
			t.Fatalf("expected path /v2/folder/folder-1/view, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ViewResponse{
			View: View{ID: "v-new", Name: "New View", Type: "gantt"},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	req := CreateViewRequest{Name: "New View", Type: "gantt"}

	result, err := client.Views().CreateInFolder(context.Background(), "folder-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Type != "gantt" {
		t.Fatalf("expected type gantt, got %s", result.Type)
	}
}

func TestViewsCreateInList_SendsRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/list/list-1/view" {
			t.Fatalf("expected path /v2/list/list-1/view, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ViewResponse{
			View: View{ID: "v-new", Name: "New View", Type: "list"},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	req := CreateViewRequest{Name: "New View", Type: "list"}

	result, err := client.Views().CreateInList(context.Background(), "list-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Type != "list" {
		t.Fatalf("expected type list, got %s", result.Type)
	}
}

func TestViewsUpdate_SendsRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/view/v-123" {
			t.Fatalf("expected path /v2/view/v-123, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}

		var req UpdateViewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Name != "Updated Name" {
			t.Fatalf("expected name Updated Name, got %s", req.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ViewResponse{
			View: View{ID: "v-123", Name: "Updated Name", Type: "board"},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	req := UpdateViewRequest{Name: "Updated Name"}

	result, err := client.Views().Update(context.Background(), "v-123", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Name != "Updated Name" {
		t.Fatalf("expected name Updated Name, got %s", result.Name)
	}
}

func TestViewsUpdate_RequiresID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Views().Update(context.Background(), "", UpdateViewRequest{Name: "Test"})
	if err == nil {
		t.Fatal("expected error for missing view ID, got nil")
	}
}

func TestViewsDelete_SendsRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/view/v-123" {
			t.Fatalf("expected path /v2/view/v-123, got %s", r.URL.Path)
		}

		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(server)

	if err := client.Views().Delete(context.Background(), "v-123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestViewsDelete_RequiresID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected request")
	}))
	defer server.Close()

	client := newTestClient(server)

	err := client.Views().Delete(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing view ID, got nil")
	}
}
