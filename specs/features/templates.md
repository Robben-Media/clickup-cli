# Templates

## Overview

Implement task template discovery for a workspace.

**Why**: Templates enable standardized task creation. Listing available templates is a prerequisite for `tasks from-template` and `lists from-template` commands.

## API Endpoints

| Method | Path | Summary | Operation ID | Version |
|--------|------|---------|--------------|---------|
| GET | /team/{}/taskTemplate | Get Task Templates | GetTaskTemplates | v2 |

## User Stories

### US-001: List Task Templates

**CLI Command:** `clickup templates list --team <team_id> [--page <n>]`

**JSON Output:**
```json
{
  "templates": [
    {"id": "tpl-123", "name": "Weekly Report", "task": {"name": "Weekly Report", "description": "Fill out weekly status"}}
  ]
}
```

**Plain Output (TSV):** Headers: `ID	NAME`
```
tpl-123	Weekly Report
```

**Human-Readable:**
```
Task Templates

  tpl-123: Weekly Report
```

**Acceptance Criteria:**
- [ ] Returns paginated template list
- [ ] Shows template ID and name

## Request/Response Types

```go
type TaskTemplate struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

type TaskTemplatesResponse struct {
    Templates []TaskTemplate `json:"templates"`
}
```

## Feedback Loops

### Unit Tests
```go
func TestTemplatesService_List(t *testing.T) { /* list templates with pagination */ }
```

## Technical Requirements

- Supports `page` query param (0-indexed)
- Can be a method on `WorkspacesService` or standalone
