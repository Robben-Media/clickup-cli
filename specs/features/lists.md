# Lists

## Overview

Extend the Lists domain to full parity. The CLI currently supports listing by folder and listing folderless. This spec adds get, create (in folder and folderless), update, delete, template creation, and task-list membership management.

**Why**: Lists are the primary container for tasks. Create/update/delete and task-list management are essential for workspace automation.

## API Endpoints

| Status | Method | Path | Summary | Operation ID | Version |
|--------|--------|------|---------|--------------|---------|
| impl | GET | /folder/{}/list | Get Lists | GetLists | v2 |
| impl | GET | /space/{}/list | Get Folderless Lists | GetFolderlessLists | v2 |
| missing | GET | /list/{} | Get List | GetList | v2 |
| missing | POST | /folder/{}/list | Create List | CreateList | v2 |
| missing | POST | /space/{}/list | Create Folderless List | CreateFolderlessList | v2 |
| missing | PUT | /list/{} | Update List | UpdateList | v2 |
| missing | DELETE | /list/{} | Delete List | DeleteList | v2 |
| missing | POST | /folder/{}/list_template/{} | Create List From Template in Folder | CreateFolderListFromTemplate | v2 |
| missing | POST | /space/{}/list_template/{} | Create List From Template in Space | CreateSpaceListFromTemplate | v2 |
| missing | POST | /list/{}/task/{} | Add Task To List | AddTaskToList | v2 |
| missing | DELETE | /list/{}/task/{} | Remove Task From List | RemoveTaskFromList | v2 |

## User Stories

### US-001: Get List Details

**CLI Command:** `clickup lists get <list_id>`

**JSON Output:**
```json
{
  "id": "901",
  "name": "Sprint 1",
  "content": "Sprint 1 tasks",
  "due_date": "1700000000000",
  "priority": {"id": "2", "priority": "high"},
  "assignee": {"id": 1, "username": "jeremy"},
  "task_count": 15,
  "folder": {"id": "456", "name": "Backlog"},
  "space": {"id": "789", "name": "Engineering"}
}
```

**Plain Output (TSV):**
Headers: `ID	NAME	TASK_COUNT	FOLDER	SPACE`
```
901	Sprint 1	15	Backlog	Engineering
```

**Human-Readable Output:**
```
ID: 901
Name: Sprint 1
Task Count: 15
Folder: Backlog (456)
Space: Engineering (789)
```

**Acceptance Criteria:**
- [ ] Returns full list object with task count, folder/space refs

### US-002: Create List

**CLI Commands:**
- `clickup lists create --folder <folder_id> <name>` — creates in folder
- `clickup lists create --space <space_id> <name>` — creates folderless

**Acceptance Criteria:**
- [ ] Exactly one of `--folder` or `--space` is required
- [ ] Returns created list
- [ ] Supports `--content`, `--due-date`, `--priority`, `--assignee` flags

**JSON/Plain/Human Output:** Same as get.

### US-003: Update List

**CLI Command:** `clickup lists update <list_id> [--name "..."] [--content "..."] [--due-date <ms>] [--priority <n>] [--unset-assignee]`

**Acceptance Criteria:**
- [ ] Updates specified fields only
- [ ] Returns updated list

### US-004: Delete List

**CLI Command:** `clickup lists delete <list_id>`

**JSON Output:** `{"status": "success", "message": "List deleted", "list_id": "901"}`

**Plain Output (TSV):** Headers: `STATUS	LIST_ID` → `success	901`

**Human-Readable:** `List 901 deleted`

**Acceptance Criteria:**
- [ ] Destructive operation — deletes list and all tasks

### US-005: Create List from Template

**CLI Commands:**
- `clickup lists from-template --folder <folder_id> <template_id>`
- `clickup lists from-template --space <space_id> <template_id>`

**Acceptance Criteria:**
- [ ] Exactly one of `--folder` or `--space` required
- [ ] Returns created list

### US-006: Manage Task-List Membership

**CLI Commands:**
- `clickup lists add-task <list_id> <task_id>`
- `clickup lists remove-task <list_id> <task_id>`

**Acceptance Criteria:**
- [ ] Add task enables "Tasks in Multiple Lists" feature
- [ ] Remove task removes the task from that list (task still exists in other lists)

**JSON Output (add):** `{"status": "success", "message": "Task added to list"}`
**JSON Output (remove):** `{"status": "success", "message": "Task removed from list"}`

**Plain Output (TSV):** Headers: `STATUS	LIST_ID	TASK_ID`

## Request/Response Types

```go
type ListDetail struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Content   string    `json:"content,omitempty"`
    DueDate   string    `json:"due_date,omitempty"`
    Priority  *Priority `json:"priority,omitempty"`
    Assignee  *User     `json:"assignee,omitempty"`
    TaskCount int       `json:"task_count"`
    Folder    FolderRef `json:"folder"`
    Space     SpaceRef  `json:"space"`
}

type CreateListRequest struct {
    Name     string `json:"name"`
    Content  string `json:"content,omitempty"`
    DueDate  int64  `json:"due_date,omitempty"`
    Priority int    `json:"priority,omitempty"`
    Assignee int    `json:"assignee,omitempty"`
}

type UpdateListRequest struct {
    Name      string `json:"name,omitempty"`
    Content   string `json:"content,omitempty"`
    DueDate   int64  `json:"due_date,omitempty"`
    Priority  int    `json:"priority,omitempty"`
    Assignee  int    `json:"assignee,omitempty"`
    UnsetAssignee bool `json:"unset_assignee,omitempty"`
}
```

## Edge Cases

- Deleting a list deletes all tasks in it — destructive
- Removing a task from its only list — API may error or make it orphaned
- Creating a list with `--folder` and `--space` both set — CLI should reject
- Template may reference statuses not in target space

## Feedback Loops

### Unit Tests
```go
func TestListsService_Get(t *testing.T)           { /* full list details */ }
func TestListsService_CreateInFolder(t *testing.T) { /* create in folder */ }
func TestListsService_CreateFolderless(t *testing.T) { /* create folderless */ }
func TestListsService_Update(t *testing.T)         { /* update fields */ }
func TestListsService_Delete(t *testing.T)         { /* delete */ }
func TestListsService_AddTask(t *testing.T)        { /* add task to list */ }
func TestListsService_RemoveTask(t *testing.T)     { /* remove task from list */ }
```

## Technical Requirements

- `ListsService` already exists — extend it with new methods
- Add/Remove task endpoints return empty body with 200
- List `task_count` is returned as integer (unlike folder's string)
