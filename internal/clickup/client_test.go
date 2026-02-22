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

func TestWorkspacesList_ReturnsTeams(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/team" {
			t.Fatalf("expected path /v2/team, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(WorkspacesResponse{
			Teams: []Workspace{
				{ID: "team-1", Name: "Team One", Members: []Member{{User: User{ID: 1, Username: "alice"}}}},
				{ID: "team-2", Name: "Team Two", Members: []Member{{User: User{ID: 2, Username: "bob"}}}},
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
		t.Fatalf("expected two teams, got %d", len(result.Teams))
	}

	if result.Teams[0].ID != "team-1" || result.Teams[0].Name != "Team One" {
		t.Fatalf("expected team-1/Team One, got %s/%s", result.Teams[0].ID, result.Teams[0].Name)
	}
}

func TestWorkspacesPlan_ReturnsPlanInfo(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		t.Fatalf("expected team-1, got %s", result.TeamID)
	}

	if result.PlanID != 3 {
		t.Fatalf("expected plan ID 3, got %d", result.PlanID)
	}

	if result.PlanName != "Business" {
		t.Fatalf("expected plan name Business, got %s", result.PlanName)
	}
}

func TestWorkspacesSeats_ReturnsSeatInfo(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		t.Fatalf("expected 5 filled member seats, got %d", result.Members.FilledSeats)
	}

	if result.Members.TotalSeats != 10 {
		t.Fatalf("expected 10 total member seats, got %d", result.Members.TotalSeats)
	}

	if result.Guests.FilledSeats != 2 {
		t.Fatalf("expected 2 filled guest seats, got %d", result.Guests.FilledSeats)
	}
}

func TestWorkspacesPlan_RequiresTeamID(t *testing.T) {
	t.Parallel()

	// Create client without server since this test doesn't make HTTP requests
	client := &Client{
		Client: api.NewClient("test-api-key"),
	}

	_, err := client.Workspaces().Plan(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty team ID, got nil")
	}

	if !errors.Is(err, errIDRequired) {
		t.Fatalf("expected errIDRequired, got %v", err)
	}
}

func TestWorkspacesSeats_RequiresTeamID(t *testing.T) {
	t.Parallel()

	// Create client without server since this test doesn't make HTTP requests
	client := &Client{
		Client: api.NewClient("test-api-key"),
	}

	_, err := client.Workspaces().Seats(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty team ID, got nil")
	}

	if !errors.Is(err, errIDRequired) {
		t.Fatalf("expected errIDRequired, got %v", err)
	}
}

func TestAuthServiceWhoami_ReturnsUser(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/user" {
			t.Fatalf("expected path /v2/user, got %s", r.URL.Path)
		}

		// Verify Authorization header is set
		if r.Header.Get("Authorization") != "test-api-key" {
			t.Fatalf("expected Authorization header, got %s", r.Header.Get("Authorization"))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(AuthorizedUserResponse{
			User: AuthUser{
				ID:             123,
				Username:       "testuser",
				Email:          "test@example.com",
				Color:          "#4194f6",
				ProfilePicture: "https://example.com/pic.jpg",
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Auth().Whoami(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.User.ID != 123 {
		t.Fatalf("expected user ID 123, got %d", result.User.ID)
	}

	if result.User.Username != "testuser" {
		t.Fatalf("expected username testuser, got %s", result.User.Username)
	}

	if result.User.Email != "test@example.com" {
		t.Fatalf("expected email test@example.com, got %s", result.User.Email)
	}
}

func TestAuthServiceToken_ReturnsToken(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/v2/oauth/token" {
			t.Fatalf("expected path /v2/oauth/token, got %s", r.URL.Path)
		}

		// Verify Authorization header is NOT set
		if r.Header.Get("Authorization") != "" {
			t.Fatalf("expected no Authorization header, got %s", r.Header.Get("Authorization"))
		}

		// Decode and verify request body
		var req OAuthTokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if req.ClientID != "client-123" {
			t.Fatalf("expected client_id client-123, got %s", req.ClientID)
		}

		if req.ClientSecret != "mock_client_secret_xyz" {
			t.Fatalf("expected client_secret mock_client_secret_xyz, got %s", req.ClientSecret)
		}

		if req.Code != "auth-code-789" {
			t.Fatalf("expected code auth-code-789, got %s", req.Code)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(OAuthTokenResponse{
			AccessToken: "mock_access_xyz",
			TokenType:   "Bearer",
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Auth().Token(context.Background(), OAuthTokenRequest{
		ClientID:     "client-123",
		ClientSecret: "mock_client_secret_xyz",
		Code:         "auth-code-789",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.AccessToken != "mock_access_xyz" {
		t.Fatalf("expected access token mock_access_xyz, got %s", result.AccessToken)
	}

	if result.TokenType != "Bearer" {
		t.Fatalf("expected token type Bearer, got %s", result.TokenType)
	}
}

func TestAuthServiceToken_RequiresAllFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		req     OAuthTokenRequest
		wantErr bool
	}{
		{
			name:    "missing client_id",
			req:     OAuthTokenRequest{ClientSecret: "mock_value_xyz", Code: "code"},
			wantErr: true,
		},
		{
			name:    "missing client_secret",
			req:     OAuthTokenRequest{ClientID: "id", Code: "code"},
			wantErr: true,
		},
		{
			name:    "missing code",
			req:     OAuthTokenRequest{ClientID: "id", ClientSecret: "mock_value_xyz"},
			wantErr: true,
		},
		{
			name:    "all fields present",
			req:     OAuthTokenRequest{ClientID: "id", ClientSecret: "mock_value_xyz", Code: "code"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// We need a server for the valid case, but won't reach it for validation errors
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(OAuthTokenResponse{})
			}))
			defer server.Close()

			testClient := newTestClient(server)

			_, err := testClient.Auth().Token(context.Background(), tt.req)

			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
