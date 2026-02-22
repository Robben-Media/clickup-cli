# Members

## Overview

Extend member listing to list and task levels. Workspace-level member listing already works.

**Why**: Knowing who has access to a specific list or task enables permission auditing and smart assignee suggestions.

## API Endpoints

| Status | Method | Path | Summary | Operation ID | Version |
|--------|--------|------|---------|--------------|---------|
| impl | GET | /team/{} | List workspace members (via team endpoint) | — | v2 |
| missing | GET | /list/{}/member | Get List Members | GetListMembers | v2 |
| missing | GET | /task/{}/member | Get Task Members | GetTaskMembers | v2 |

## User Stories

### US-001: List Members

**CLI Command:** `clickup members list-members --list <list_id>` or `clickup members task-members --task <task_id>`

**JSON Output:**
```json
{"members": [{"id": 1, "username": "jeremy", "email": "jeremy@example.com"}]}
```

**Plain Output (TSV):** Headers: `ID	USERNAME	EMAIL`
```
1	jeremy	jeremy@example.com
```

**Acceptance Criteria:**
- [ ] List members returns users with access to the list
- [ ] Task members returns users involved with the task (assignees + watchers)

## Request/Response Types

```go
// Reuses existing MembersListResponse but response shape differs:
// List/Task members return {"members": [{"id": ..., "username": ..., "email": ...}]}
// (flat user objects, not wrapped in {"user": {...}})
type MemberUser struct {
    ID       int    `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email,omitempty"`
}

type MemberUsersResponse struct {
    Members []MemberUser `json:"members"`
}
```

## Edge Cases

- List members response is flat `[{id, username, email}]`, NOT wrapped in `{"user": {}}` like workspace members
- Task members may include watchers, not just assignees

## Feedback Loops

### Unit Tests
```go
func TestMembersService_ListMembers(t *testing.T) { /* list-level */ }
func TestMembersService_TaskMembers(t *testing.T) { /* task-level */ }
```
