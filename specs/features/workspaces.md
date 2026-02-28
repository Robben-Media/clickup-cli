# Workspaces

## Overview

Implement workspace discovery and metadata endpoints.

**Why**: Workspace listing is the entry point for all operations. Plan and seats info is needed for quota management and billing awareness.

## API Endpoints

| Method | Path | Summary | Operation ID | Version |
|--------|------|---------|--------------|---------|
| GET | /team | Get Authorized Workspaces | GetAuthorizedTeams | v2 |
| GET | /team/{}/plan | Get Workspace Plan | GetWorkspaceplan | v2 |
| GET | /team/{}/seats | Get Workspace Seats | GetWorkspaceseats | v2 |

## User Stories

### US-001: List Workspaces

**CLI Command:** `clickup workspaces list`

**JSON Output:**
```json
{"teams": [{"id": "789", "name": "Robben Media", "color": "#4194f6", "members": [{"user": {"id": 1, "username": "jeremy"}}]}]}
```

**Plain Output (TSV):** Headers: `ID	NAME	MEMBER_COUNT`
```
789	Robben Media	5
```

**Human-Readable:**
```
Workspaces

  789: Robben Media (5 members)
```

**Acceptance Criteria:**
- [ ] Lists all workspaces the API token has access to
- [ ] Shows workspace ID (needed as `--team` in other commands)

### US-002: Get Workspace Plan

**CLI Command:** `clickup workspaces plan --team <team_id>`

**JSON Output:**
```json
{"team_id": "789", "plan_id": 3, "plan_name": "Business"}
```

**Plain Output (TSV):** Headers: `TEAM_ID	PLAN_ID	PLAN_NAME`

### US-003: Get Workspace Seats

**CLI Command:** `clickup workspaces seats --team <team_id>`

**JSON Output:**
```json
{"members": {"filled_member_seats": 5, "total_member_seats": 10, "empty_member_seats": 5}, "guests": {"filled_guest_seats": 2, "total_guest_seats": 5}}
```

**Plain Output (TSV):** Headers: `TYPE	FILLED	TOTAL	EMPTY`
```
members	5	10	5
guests	2	5	3
```

## Request/Response Types

```go
type WorkspacesResponse struct {
    Teams []Workspace `json:"teams"`
}

type Workspace struct {
    ID      string   `json:"id"`
    Name    string   `json:"name"`
    Color   string   `json:"color,omitempty"`
    Members []Member `json:"members,omitempty"`
}

type WorkspacePlanResponse struct {
    TeamID   string `json:"team_id"`
    PlanID   int    `json:"plan_id"`
    PlanName string `json:"plan_name"`
}

type WorkspaceSeatsResponse struct {
    Members SeatInfo `json:"members"`
    Guests  SeatInfo `json:"guests"`
}

type SeatInfo struct {
    FilledSeats int `json:"filled_member_seats"`
    TotalSeats  int `json:"total_member_seats"`
    EmptySeats  int `json:"empty_member_seats"`
}
```

## Feedback Loops

### Unit Tests
```go
func TestWorkspacesService_List(t *testing.T)  { /* list workspaces */ }
func TestWorkspacesService_Plan(t *testing.T)  { /* get plan */ }
func TestWorkspacesService_Seats(t *testing.T) { /* get seats */ }
```

## Technical Requirements

- New `WorkspacesService` on `clickup.Client`
- `GET /team` has no path params — just the base URL
- This is the endpoint to auto-detect workspace ID for v3 calls
