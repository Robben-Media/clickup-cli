# User Groups

## Overview

Implement user group (team) CRUD operations.

**Why**: Groups simplify bulk assignment and permission management.

## API Endpoints

| Method | Path | Summary | Operation ID | Version |
|--------|------|---------|--------------|---------|
| GET | /group | Get Groups | GetTeams1 | v2 |
| POST | /team/{}/group | Create Group | CreateUserGroup | v2 |
| PUT | /group/{} | Update Group | UpdateTeam | v2 |
| DELETE | /group/{} | Delete Group | DeleteTeam | v2 |

## User Stories

### US-001: List Groups

**CLI Command:** `clickup groups list [--team <team_id>]`

**JSON Output:**
```json
{"groups": [{"id": "grp-123", "name": "Engineering", "members": [{"id": 1, "username": "jeremy"}]}]}
```

**Plain Output (TSV):** Headers: `ID	NAME	MEMBER_COUNT`
```
grp-123	Engineering	5
```

### US-002: Create Group

**CLI Command:** `clickup groups create --team <team_id> <name> [--members <user_id1,user_id2>]`

### US-003: Update Group

**CLI Command:** `clickup groups update <group_id> [--name "..."] [--add-members <ids>] [--remove-members <ids>]`

### US-004: Delete Group

**CLI Command:** `clickup groups delete <group_id>`

## Request/Response Types

```go
type UserGroup struct {
    ID      string `json:"id"`
    Name    string `json:"name"`
    Members []User `json:"members,omitempty"`
}

type UserGroupsResponse struct {
    Groups []UserGroup `json:"groups"`
}

type CreateUserGroupRequest struct {
    Name    string `json:"name"`
    Members []int  `json:"members,omitempty"`
}

type UpdateUserGroupRequest struct {
    Name    string `json:"name,omitempty"`
    Members struct {
        Add []int `json:"add,omitempty"`
        Rem []int `json:"rem,omitempty"`
    } `json:"members,omitempty"`
}
```

## Edge Cases

- `GET /group` is workspace-scoped via the team_id query param
- Update members uses add/rem pattern (same as task assignees)
- Group IDs are strings (UUIDs)

## Feedback Loops

### Unit Tests
```go
func TestUserGroupsService_List(t *testing.T)   { /* list groups */ }
func TestUserGroupsService_Create(t *testing.T) { /* create with members */ }
func TestUserGroupsService_Update(t *testing.T) { /* rename, add/remove members */ }
func TestUserGroupsService_Delete(t *testing.T) { /* delete */ }
```
