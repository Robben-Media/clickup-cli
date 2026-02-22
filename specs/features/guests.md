# Guests

## Overview

Implement guest management — workspace-level CRUD plus guest-to-resource (task/list/folder) assignment.

**Why**: Guests are external collaborators with limited permissions. Managing them from CLI enables automated client onboarding.

## API Endpoints

| Method | Path | Summary | Operation ID | Version |
|--------|------|---------|--------------|---------|
| GET | /team/{}/guest/{} | Get Guest | GetGuest | v2 |
| POST | /team/{}/guest | Invite Guest To Workspace | InviteGuestToWorkspace | v2 |
| PUT | /team/{}/guest/{} | Edit Guest On Workspace | EditGuestOnWorkspace | v2 |
| DELETE | /team/{}/guest/{} | Remove Guest From Workspace | RemoveGuestFromWorkspace | v2 |
| POST | /task/{}/guest/{} | Add Guest To Task | AddGuestToTask | v2 |
| DELETE | /task/{}/guest/{} | Remove Guest From Task | RemoveGuestFromTask | v2 |
| POST | /list/{}/guest/{} | Add Guest To List | AddGuestToList | v2 |
| DELETE | /list/{}/guest/{} | Remove Guest From List | RemoveGuestFromList | v2 |
| POST | /folder/{}/guest/{} | Add Guest To Folder | AddGuestToFolder | v2 |
| DELETE | /folder/{}/guest/{} | Remove Guest From Folder | RemoveGuestFromFolder | v2 |

## User Stories

### US-001: Get Guest

**CLI Command:** `clickup guests get --team <team_id> <guest_id>`

**JSON Output:**
```json
{"guest": {"id": 100, "username": "client-bob", "email": "bob@client.com", "tasks_count": 3, "lists_count": 1, "folders_count": 0}}
```

**Plain Output (TSV):** Headers: `ID	USERNAME	EMAIL	TASKS	LISTS	FOLDERS`

### US-002: Invite Guest

**CLI Command:** `clickup guests invite --team <team_id> --email <email> [--can-edit-tags] [--can-see-time-spent] [--can-see-time-estimated]`

### US-003: Edit Guest

**CLI Command:** `clickup guests update --team <team_id> <guest_id> [--can-edit-tags] [--can-see-time-spent]`

### US-004: Remove Guest

**CLI Command:** `clickup guests remove --team <team_id> <guest_id>`

### US-005: Add/Remove Guest from Resources

**CLI Commands:**
- `clickup guests add-to-task <task_id> <guest_id> --permission <read|comment|edit|create>`
- `clickup guests remove-from-task <task_id> <guest_id>`
- `clickup guests add-to-list <list_id> <guest_id> --permission <read|comment|edit|create>`
- `clickup guests remove-from-list <list_id> <guest_id>`
- `clickup guests add-to-folder <folder_id> <guest_id> --permission <read|comment|edit|create>`
- `clickup guests remove-from-folder <folder_id> <guest_id>`

**Acceptance Criteria:**
- [ ] Permission level is required when adding guest to a resource
- [ ] Permission levels: read, comment, edit, create

## Request/Response Types

```go
type Guest struct {
    ID           int    `json:"id"`
    Username     string `json:"username"`
    Email        string `json:"email"`
    TasksCount   int    `json:"tasks_count,omitempty"`
    ListsCount   int    `json:"lists_count,omitempty"`
    FoldersCount int    `json:"folders_count,omitempty"`
}

type GuestResponse struct {
    Guest Guest `json:"guest"`
}

type InviteGuestRequest struct {
    Email               string `json:"email"`
    CanEditTags         bool   `json:"can_edit_tags,omitempty"`
    CanSeeTimeSpent     bool   `json:"can_see_time_spent,omitempty"`
    CanSeeTimeEstimated bool   `json:"can_see_time_estimated,omitempty"`
}

type EditGuestRequest struct {
    CanEditTags         *bool `json:"can_edit_tags,omitempty"`
    CanSeeTimeSpent     *bool `json:"can_see_time_spent,omitempty"`
    CanSeeTimeEstimated *bool `json:"can_see_time_estimated,omitempty"`
}

type AddGuestToResourceRequest struct {
    PermissionLevel string `json:"permission_level"` // "read", "comment", "edit", "create"
}
```

## Edge Cases

- Guest ID is in the URL path for add/remove from resources
- Permission level is required for add operations
- Removing a guest from workspace removes them from all resources
- Guests require Business plan or higher

## Feedback Loops

### Unit Tests
```go
func TestGuestsService_Get(t *testing.T)              { /* get guest */ }
func TestGuestsService_Invite(t *testing.T)           { /* invite */ }
func TestGuestsService_Edit(t *testing.T)             { /* edit permissions */ }
func TestGuestsService_Remove(t *testing.T)           { /* remove from workspace */ }
func TestGuestsService_AddToTask(t *testing.T)        { /* add to task */ }
func TestGuestsService_RemoveFromTask(t *testing.T)   { /* remove from task */ }
func TestGuestsService_AddToList(t *testing.T)        { /* add to list */ }
func TestGuestsService_RemoveFromList(t *testing.T)   { /* remove from list */ }
func TestGuestsService_AddToFolder(t *testing.T)      { /* add to folder */ }
func TestGuestsService_RemoveFromFolder(t *testing.T) { /* remove from folder */ }
```
