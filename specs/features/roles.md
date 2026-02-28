# Roles

## Overview

Implement custom role discovery for a workspace.

**Why**: Custom roles extend the default role system. Needed for user management and permission auditing.

## API Endpoints

| Method | Path | Summary | Operation ID | Version |
|--------|------|---------|--------------|---------|
| GET | /team/{}/customroles | Get Custom Roles | GetCustomRoles | v2 |

## User Stories

### US-001: List Custom Roles

**CLI Command:** `clickup roles list --team <team_id>`

**JSON Output:**
```json
{"custom_roles": [{"id": 1, "name": "Project Manager", "permissions": ["task_create", "task_delete"]}]}
```

**Plain Output (TSV):** Headers: `ID	NAME	PERMISSION_COUNT`
```
1	Project Manager	2
```

## Request/Response Types

```go
type CustomRole struct {
    ID          int      `json:"id"`
    Name        string   `json:"name"`
    Permissions []string `json:"permissions,omitempty"`
}

type CustomRolesResponse struct {
    CustomRoles []CustomRole `json:"custom_roles"`
}
```

## Feedback Loops

### Unit Tests
```go
func TestRolesService_List(t *testing.T) { /* list custom roles */ }
```
