# Custom Task Types

## Overview

Implement custom task type discovery for a workspace.

**Why**: Custom task types (e.g., Bug, Feature, Epic) customize the task creation experience. Knowing available types is needed when creating tasks with `custom_item_id`.

## API Endpoints

| Method | Path | Summary | Operation ID | Version |
|--------|------|---------|--------------|---------|
| GET | /team/{}/custom_item | Get Custom Task Types | GetCustomItems | v2 |

## User Stories

### US-001: List Custom Task Types

**CLI Command:** `clickup task-types list --team <team_id>`

**JSON Output:**
```json
{
  "custom_items": [
    {"id": 1, "name": "Task", "name_plural": "Tasks", "description": "Default task type"},
    {"id": 2, "name": "Bug", "name_plural": "Bugs", "description": "Bug reports"}
  ]
}
```

**Plain Output (TSV):** Headers: `ID	NAME	NAME_PLURAL	DESCRIPTION`
```
1	Task	Tasks	Default task type
2	Bug	Bugs	Bug reports
```

**Human-Readable:**
```
Custom Task Types

  1: Task (Tasks)
  2: Bug (Bugs) — Bug reports
```

**Acceptance Criteria:**
- [ ] Returns all custom task types for the workspace
- [ ] Shows ID needed for `custom_item_id` in task creation

## Request/Response Types

```go
type CustomTaskType struct {
    ID          int    `json:"id"`
    Name        string `json:"name"`
    NamePlural  string `json:"name_plural"`
    Description string `json:"description,omitempty"`
}

type CustomTaskTypesResponse struct {
    CustomItems []CustomTaskType `json:"custom_items"`
}
```

## Feedback Loops

### Unit Tests
```go
func TestCustomTaskTypesService_List(t *testing.T) { /* list types */ }
```

## Technical Requirements

- Can be a method on `WorkspacesService` or standalone `CustomTaskTypesService`
- Response key is `custom_items`, not `custom_task_types`
