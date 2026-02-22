# Views

## Overview

Implement views CRUD at all hierarchy levels (workspace, space, folder, list) plus view task retrieval.

**Why**: Views are saved filters/layouts (board, list, calendar, etc.) that organize task visibility. Managing them from CLI enables automated dashboard setup.

## API Endpoints

| Method | Path | Summary | Operation ID | Version |
|--------|------|---------|--------------|---------|
| GET | /team/{}/view | Get Workspace Views | GetTeamViews | v2 |
| GET | /space/{}/view | Get Space Views | GetSpaceViews | v2 |
| GET | /folder/{}/view | Get Folder Views | GetFolderViews | v2 |
| GET | /list/{}/view | Get List Views | GetListViews | v2 |
| GET | /view/{} | Get View | GetView | v2 |
| GET | /view/{}/task | Get View Tasks | GetViewTasks | v2 |
| POST | /team/{}/view | Create Workspace View | CreateTeamView | v2 |
| POST | /space/{}/view | Create Space View | CreateSpaceView | v2 |
| POST | /folder/{}/view | Create Folder View | CreateFolderView | v2 |
| POST | /list/{}/view | Create List View | CreateListView | v2 |
| PUT | /view/{} | Update View | UpdateView | v2 |
| DELETE | /view/{} | Delete View | DeleteView | v2 |

## User Stories

### US-001: List Views

**CLI Command:** `clickup views list --team <id>` / `--space <id>` / `--folder <id>` / `--list <id>`

**JSON Output:**
```json
{
  "views": [
    {"id": "v-123", "name": "Sprint Board", "type": "board", "parent": {"id": "789", "type": 7}}
  ]
}
```

**Plain Output (TSV):** Headers: `ID	NAME	TYPE	PARENT_TYPE	PARENT_ID`
```
v-123	Sprint Board	board	7	789
```

**Acceptance Criteria:**
- [ ] Exactly one scope flag required
- [ ] Returns view ID, name, type, parent info

### US-002: Get View

**CLI Command:** `clickup views get <view_id>`

**JSON Output:** Full view object with filters, columns, settings.

**Plain Output (TSV):** Headers: `ID	NAME	TYPE	PROTECTED`

### US-003: Get View Tasks

**CLI Command:** `clickup views tasks <view_id> [--page <n>]`

**JSON Output:**
```json
{"tasks": [{"id": "abc", "name": "Task 1", "status": {"status": "open"}}]}
```

**Plain Output (TSV):** Headers: `ID	NAME	STATUS	PRIORITY	URL`

**Acceptance Criteria:**
- [ ] Returns tasks matching the view's filters
- [ ] Supports pagination

### US-004: Create View

**CLI Command:** `clickup views create --team <id> <name> --type <type>` (also `--space`, `--folder`, `--list`)

**Acceptance Criteria:**
- [ ] Exactly one scope flag required
- [ ] Type required: list, board, calendar, gantt, activity, map, workload, table
- [ ] Returns created view

### US-005: Update View

**CLI Command:** `clickup views update <view_id> [--name "..."]`

### US-006: Delete View

**CLI Command:** `clickup views delete <view_id>`

## Request/Response Types

```go
type View struct {
    ID        string   `json:"id"`
    Name      string   `json:"name"`
    Type      string   `json:"type"`
    Parent    ViewParent `json:"parent"`
    Protected bool     `json:"protected"`
}

type ViewParent struct {
    ID   string `json:"id"`
    Type int    `json:"type"` // 7=space, 5=folder, 6=list, etc.
}

type ViewsResponse struct {
    Views []View `json:"views"`
}

type ViewResponse struct {
    View View `json:"view"`
}

type CreateViewRequest struct {
    Name string `json:"name"`
    Type string `json:"type"`
}

type UpdateViewRequest struct {
    Name string `json:"name,omitempty"`
}
```

## Edge Cases

- View types vary: list, board, calendar, gantt, activity, map, workload, table, conversation, doc
- Protected views cannot be edited/deleted
- View tasks respect the view's saved filters
- Parent type is an integer enum, not a string

## Feedback Loops

### Unit Tests
```go
func TestViewsService_ListByWorkspace(t *testing.T) { /* workspace views */ }
func TestViewsService_ListBySpace(t *testing.T)     { /* space views */ }
func TestViewsService_ListByFolder(t *testing.T)    { /* folder views */ }
func TestViewsService_ListByList(t *testing.T)      { /* list views */ }
func TestViewsService_Get(t *testing.T)             { /* single view */ }
func TestViewsService_Tasks(t *testing.T)           { /* view tasks */ }
func TestViewsService_Create(t *testing.T)          { /* create at each level */ }
func TestViewsService_Update(t *testing.T)          { /* update */ }
func TestViewsService_Delete(t *testing.T)          { /* delete */ }
```
