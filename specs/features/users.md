# Users

## Overview

Implement workspace user management — get, invite, edit, and remove users.

**Why**: User management from CLI enables automated onboarding/offboarding workflows.

## API Endpoints

| Method | Path | Summary | Operation ID | Version |
|--------|------|---------|--------------|---------|
| GET | /team/{}/user/{} | Get User | GetUser | v2 |
| POST | /team/{}/user | Invite User To Workspace | InviteUserToWorkspace | v2 |
| PUT | /team/{}/user/{} | Edit User On Workspace | EditUserOnWorkspace | v2 |
| DELETE | /team/{}/user/{} | Remove User From Workspace | RemoveUserFromWorkspace | v2 |

## User Stories

### US-001: Get User

**CLI Command:** `clickup users get --team <team_id> <user_id>`

**JSON Output:**
```json
{"user": {"id": 1, "username": "jeremy", "email": "jeremy@example.com", "role": 1}}
```

**Plain Output (TSV):** Headers: `ID	USERNAME	EMAIL	ROLE`
```
1	jeremy	jeremy@example.com	1
```

**Human-Readable:**
```
User 1
  Username: jeremy
  Email: jeremy@example.com
  Role: Owner (1)
```

### US-002: Invite User

**CLI Command:** `clickup users invite --team <team_id> --email <email> [--admin]`

**JSON Output:**
```json
{"user": {"id": 2, "username": "alice", "email": "alice@example.com", "invited_by": {"id": 1}}}
```

**Plain Output (TSV):** Headers: `ID	EMAIL	INVITED_BY`

### US-003: Edit User Role

**CLI Command:** `clickup users update --team <team_id> <user_id> --role <role_id> [--username <name>]`

### US-004: Remove User

**CLI Command:** `clickup users remove --team <team_id> <user_id>`

**JSON Output:** `{"status": "success", "message": "User removed"}`
**Plain Output (TSV):** Headers: `STATUS	USER_ID` → `success	2`

## Request/Response Types

```go
type UserResponse struct {
    User UserDetail `json:"user"`
}

type UserDetail struct {
    ID       int    `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Role     int    `json:"role"` // 1=owner, 2=admin, 3=member, 4=guest
}

type InviteUserRequest struct {
    Email string `json:"email"`
    Admin bool   `json:"admin,omitempty"`
}

type EditUserRequest struct {
    Username string `json:"username,omitempty"`
    Admin    bool   `json:"admin,omitempty"`
}
```

## Edge Cases

- Cannot remove yourself from workspace
- Role IDs: 1=owner, 2=admin, 3=member, 4=guest
- Inviting existing user returns error
- Removing last admin may fail

## Feedback Loops

### Unit Tests
```go
func TestUsersService_Get(t *testing.T)    { /* get user details */ }
func TestUsersService_Invite(t *testing.T) { /* invite by email */ }
func TestUsersService_Edit(t *testing.T)   { /* change role */ }
func TestUsersService_Remove(t *testing.T) { /* remove from workspace */ }
```
