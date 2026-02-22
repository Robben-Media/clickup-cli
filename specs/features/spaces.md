# Spaces

## Overview

Extend the Spaces domain to full CRUD parity. The CLI currently supports listing spaces. This spec adds get, create, update, and delete.

**Why**: Spaces are the top level of ClickUp's hierarchy. Full CRUD enables workspace bootstrapping and teardown from the CLI.

## API Endpoints

| Status | Method | Path | Summary | Operation ID | Version |
|--------|--------|------|---------|--------------|---------|
| impl | GET | /team/{}/space | Get Spaces | GetSpaces | v2 |
| missing | GET | /space/{} | Get Space | GetSpace | v2 |
| missing | POST | /team/{}/space | Create Space | CreateSpace | v2 |
| missing | PUT | /space/{} | Update Space | UpdateSpace | v2 |
| missing | DELETE | /space/{} | Delete Space | DeleteSpace | v2 |

## User Stories

### US-001: Get Space Details

**CLI Command:** `clickup spaces get <space_id>`

**JSON Output:**
```json
{
  "id": "789",
  "name": "Engineering",
  "private": false,
  "statuses": [
    {"status": "open", "color": "#d3d3d3", "orderindex": 0},
    {"status": "in progress", "color": "#4194f6", "orderindex": 1},
    {"status": "done", "color": "#6bc950", "orderindex": 2}
  ],
  "multiple_assignees": true,
  "features": {
    "due_dates": {"enabled": true},
    "time_tracking": {"enabled": true},
    "tags": {"enabled": true},
    "checklists": {"enabled": true}
  }
}
```

**Plain Output (TSV):**
Headers: `ID	NAME	PRIVATE	STATUS_COUNT`
```
789	Engineering	false	3
```

**Human-Readable Output:**
```
ID: 789
Name: Engineering
Private: false
Statuses: open, in progress, done
Features: due_dates, time_tracking, tags, checklists
```

**Acceptance Criteria:**
- [ ] Returns full space object including statuses and features

### US-002: Create Space

**CLI Command:** `clickup spaces create <team_id> <name> [--private] [--color <hex>]`

**Acceptance Criteria:**
- [ ] Creates a space in the specified workspace
- [ ] Returns the created space
- [ ] Supports `--private` flag and `--color` option

**JSON/Plain/Human Output:** Same format as get.

**Plain Output (TSV):**
Headers: `ID	NAME	PRIVATE`
```
789	Engineering	false
```

### US-003: Update Space

**CLI Command:** `clickup spaces update <space_id> [--name "..."] [--color <hex>] [--private] [--multiple-assignees]`

**Acceptance Criteria:**
- [ ] Updates specified fields
- [ ] Returns updated space

### US-004: Delete Space

**CLI Command:** `clickup spaces delete <space_id>`

**JSON Output:** `{"status": "success", "message": "Space deleted", "space_id": "789"}`

**Plain Output (TSV):** Headers: `STATUS	SPACE_ID` → `success	789`

**Human-Readable:** `Space 789 deleted`

**Acceptance Criteria:**
- [ ] Destructive — deletes space and all folders/lists/tasks within
- [ ] Should warn user

## Request/Response Types

```go
type SpaceDetail struct {
    ID                string        `json:"id"`
    Name              string        `json:"name"`
    Private           bool          `json:"private"`
    Color             string        `json:"color,omitempty"`
    Statuses          []SpaceStatus `json:"statuses,omitempty"`
    MultipleAssignees bool          `json:"multiple_assignees"`
    Features          SpaceFeatures `json:"features,omitempty"`
}

type SpaceStatus struct {
    Status     string `json:"status"`
    Color      string `json:"color"`
    OrderIndex int    `json:"orderindex"`
}

type SpaceFeatures struct {
    DueDates     FeatureToggle `json:"due_dates"`
    TimeTracking FeatureToggle `json:"time_tracking"`
    Tags         FeatureToggle `json:"tags"`
    Checklists   FeatureToggle `json:"checklists"`
}

type FeatureToggle struct {
    Enabled bool `json:"enabled"`
}

type CreateSpaceRequest struct {
    Name              string `json:"name"`
    MultipleAssignees bool   `json:"multiple_assignees,omitempty"`
    Features          *SpaceFeatures `json:"features,omitempty"`
}

type UpdateSpaceRequest struct {
    Name              string `json:"name,omitempty"`
    Color             string `json:"color,omitempty"`
    Private           *bool  `json:"private,omitempty"`
    MultipleAssignees *bool  `json:"multiple_assignees,omitempty"`
}
```

## Edge Cases

- Deleting a space is highly destructive — all child content is deleted
- Creating a private space limits visibility to invited members only
- Space statuses define defaults for all lists in the space
- Features can be enabled/disabled but the API structure is nested

## Feedback Loops

### Unit Tests
```go
func TestSpacesService_Get(t *testing.T)    { /* full space details with statuses/features */ }
func TestSpacesService_Create(t *testing.T) { /* create with name and options */ }
func TestSpacesService_Update(t *testing.T) { /* update fields */ }
func TestSpacesService_Delete(t *testing.T) { /* delete */ }
```

## Technical Requirements

- `SpacesService` already exists — extend with Get, Create, Update, Delete methods
- Space features JSON is deeply nested — define proper structs
- Delete returns empty body with 200
