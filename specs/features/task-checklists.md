# Task Checklists

## Overview

Implement checklist management for tasks — create/edit/delete checklists and their items.

**Why**: Checklists are a common task decomposition tool. CLI support enables template-based task setup and progress tracking.

## API Endpoints

| Method | Path | Summary | Operation ID | Version |
|--------|------|---------|--------------|---------|
| POST | /task/{}/checklist | Create Checklist | CreateChecklist | v2 |
| PUT | /checklist/{} | Edit Checklist | EditChecklist | v2 |
| DELETE | /checklist/{} | Delete Checklist | DeleteChecklist | v2 |
| POST | /checklist/{}/checklist_item | Create Checklist Item | CreateChecklistItem | v2 |
| PUT | /checklist/{}/checklist_item/{} | Edit Checklist Item | EditChecklistItem | v2 |
| DELETE | /checklist/{}/checklist_item/{} | Delete Checklist Item | DeleteChecklistItem | v2 |

## User Stories

### US-001: Create Checklist

**CLI Command:** `clickup checklists create --task <task_id> <name>`

**JSON Output:**
```json
{"checklist": {"id": "cl-123", "name": "QA Steps", "orderindex": 0, "items": []}}
```

**Plain Output (TSV):** Headers: `ID	NAME	ITEM_COUNT`
```
cl-123	QA Steps	0
```

### US-002: Edit Checklist

**CLI Command:** `clickup checklists update <checklist_id> [--name "..."] [--position <n>]`

### US-003: Delete Checklist

**CLI Command:** `clickup checklists delete <checklist_id>`

**JSON Output:** `{"status": "success", "message": "Checklist deleted"}`
**Plain Output (TSV):** Headers: `STATUS	CHECKLIST_ID` → `success	cl-123`

### US-004: Create Checklist Item

**CLI Command:** `clickup checklists add-item <checklist_id> <name> [--assignee <user_id>]`

**JSON Output:**
```json
{"checklist": {"id": "cl-123", "items": [{"id": "ci-456", "name": "Step 1", "resolved": false}]}}
```

**Plain Output (TSV):** Headers: `CHECKLIST_ID	ITEM_ID	NAME	RESOLVED`
```
cl-123	ci-456	Step 1	false
```

### US-005: Edit Checklist Item

**CLI Command:** `clickup checklists update-item <checklist_id> <item_id> [--name "..."] [--resolved] [--assignee <user_id>] [--parent <item_id>]`

**Acceptance Criteria:**
- [ ] Can mark item as resolved/unresolved
- [ ] Can nest under another item via `--parent`

### US-006: Delete Checklist Item

**CLI Command:** `clickup checklists delete-item <checklist_id> <item_id>`

## Request/Response Types

```go
type Checklist struct {
    ID         string          `json:"id"`
    Name       string          `json:"name"`
    OrderIndex int             `json:"orderindex"`
    Items      []ChecklistItem `json:"items,omitempty"`
}

type ChecklistItem struct {
    ID         string `json:"id"`
    Name       string `json:"name"`
    Resolved   bool   `json:"resolved"`
    Assignee   *User  `json:"assignee,omitempty"`
    Parent     string `json:"parent,omitempty"`
    OrderIndex int    `json:"orderindex"`
}

type CreateChecklistRequest struct {
    Name string `json:"name"`
}

type EditChecklistRequest struct {
    Name     string `json:"name,omitempty"`
    Position int    `json:"position,omitempty"`
}

type CreateChecklistItemRequest struct {
    Name     string `json:"name"`
    Assignee int    `json:"assignee,omitempty"`
}

type EditChecklistItemRequest struct {
    Name     string `json:"name,omitempty"`
    Resolved *bool  `json:"resolved,omitempty"`
    Assignee int    `json:"assignee,omitempty"`
    Parent   string `json:"parent,omitempty"`
}
```

## Edge Cases

- Checklist response wraps in `{"checklist": {...}}` envelope
- Items can be nested (parent-child via `parent` field)
- Resolved is a boolean, not a status string
- Position for reordering uses integer index

## Feedback Loops

### Unit Tests
```go
func TestChecklistsService_Create(t *testing.T)     { /* create checklist */ }
func TestChecklistsService_Edit(t *testing.T)       { /* rename/reorder */ }
func TestChecklistsService_Delete(t *testing.T)     { /* delete */ }
func TestChecklistsService_AddItem(t *testing.T)    { /* add item */ }
func TestChecklistsService_EditItem(t *testing.T)   { /* edit item, resolve */ }
func TestChecklistsService_DeleteItem(t *testing.T) { /* delete item */ }
```

## Technical Requirements

- New `ChecklistsService` on `clickup.Client`
- Response wraps in `{"checklist": {...}}` — unwrap in service layer
- Checklist IDs are UUIDs (not numeric like task IDs)
