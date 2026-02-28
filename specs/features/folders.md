# Folders

## Overview

Extend the Folders domain to full CRUD parity. The CLI currently supports listing folders via `ListsService.ListFolders`. This spec adds get, create, update, delete, and template-based folder creation.

**Why**: Folders are the middle tier of ClickUp's Space > Folder > List hierarchy. Full CRUD enables workspace scaffolding and cleanup from the CLI.

## API Endpoints

| Status | Method | Path | Summary | Operation ID | Version |
|--------|--------|------|---------|--------------|---------|
| impl | GET | /space/{}/folder | Get Folders | GetFolders | v2 |
| missing | GET | /folder/{} | Get Folder | GetFolder | v2 |
| missing | POST | /space/{}/folder | Create Folder | CreateFolder | v2 |
| missing | PUT | /folder/{} | Update Folder | UpdateFolder | v2 |
| missing | DELETE | /folder/{} | Delete Folder | DeleteFolder | v2 |
| missing | POST | /space/{}/folder_template/{} | Create Folder from template | CreateFolderFromTemplate | v2 |

## User Stories

### US-001: Get Folder Details

As a CLI user,
I want to view a specific folder's details,
so that I can see its properties and contained lists.

**Acceptance Criteria:**
- [ ] `clickup folders get <folder_id>` returns folder details
- [ ] Shows folder name, ID, and contained lists

**CLI Command:** `clickup folders get <folder_id>`

**JSON Output:**
```json
{
  "id": "456",
  "name": "Sprint Backlog",
  "orderindex": 0,
  "override_statuses": false,
  "hidden": false,
  "space": {"id": "789"},
  "task_count": "12",
  "lists": [
    {"id": "901", "name": "Sprint 1"}
  ]
}
```

**Plain Output (TSV):**
Headers: `ID	NAME	TASK_COUNT	LIST_COUNT`
```
456	Sprint Backlog	12	1
```

**Human-Readable Output:**
```
ID: 456
Name: Sprint Backlog
Task Count: 12
Lists:
  901: Sprint 1
```

### US-002: Create Folder

As a CLI user,
I want to create a folder in a space,
so that I can organize my lists.

**Acceptance Criteria:**
- [ ] `clickup folders create <space_id> <name>` creates a folder
- [ ] Returns the created folder

**CLI Command:** `clickup folders create <space_id> <name>`

**JSON Output:** Full folder object (same as get).

**Plain Output (TSV):**
Headers: `ID	NAME	SPACE_ID`
```
456	Sprint Backlog	789
```

**Human-Readable Output:**
```
Created folder

ID: 456
Name: Sprint Backlog
Space: 789
```

### US-003: Update Folder

As a CLI user,
I want to rename or modify a folder,
so that I can keep my workspace organized.

**Acceptance Criteria:**
- [ ] `clickup folders update <folder_id> --name "..."` updates the folder name

**CLI Command:** `clickup folders update <folder_id> [--name "..."]`

**JSON Output:** Full folder object.

**Plain Output (TSV):**
Headers: `ID	NAME`
```
456	New Name
```

**Human-Readable Output:**
```
Updated folder

ID: 456
Name: New Name
```

### US-004: Delete Folder

As a CLI user,
I want to delete a folder,
so that I can clean up my workspace.

**Acceptance Criteria:**
- [ ] `clickup folders delete <folder_id>` deletes the folder
- [ ] Destructive — should confirm

**CLI Command:** `clickup folders delete <folder_id>`

**JSON Output:**
```json
{"status": "success", "message": "Folder deleted", "folder_id": "456"}
```

**Plain Output (TSV):**
Headers: `STATUS	FOLDER_ID`
```
success	456
```

**Human-Readable Output:**
```
Folder 456 deleted
```

### US-005: Create Folder from Template

As a CLI user,
I want to create a folder from a template,
so that I can replicate standard project structures.

**Acceptance Criteria:**
- [ ] `clickup folders from-template <space_id> <template_id>` creates a folder
- [ ] Supports `--name` override

**CLI Command:** `clickup folders from-template <space_id> <template_id> [--name "..."]`

**JSON Output:** Full folder object.

**Plain Output (TSV):**
Headers: `ID	NAME	SPACE_ID`
```
456	Project Alpha	789
```

**Human-Readable Output:**
```
Created folder from template

ID: 456
Name: Project Alpha
Space: 789
```

## Request/Response Types

```go
// Folder (extended from existing)
type FolderDetail struct {
    ID               string `json:"id"`
    Name             string `json:"name"`
    OrderIndex       int    `json:"orderindex"`
    OverrideStatuses bool   `json:"override_statuses"`
    Hidden           bool   `json:"hidden"`
    Space            SpaceRef `json:"space"`
    TaskCount        string `json:"task_count"`
    Lists            []List `json:"lists,omitempty"`
}

type CreateFolderRequest struct {
    Name string `json:"name"`
}

type UpdateFolderRequest struct {
    Name string `json:"name,omitempty"`
}

type CreateFolderFromTemplateRequest struct {
    Name string `json:"name,omitempty"`
}
```

## Edge Cases

- Deleting a folder with lists inside — ClickUp moves lists to folderless
- Creating folder with duplicate name — API allows it
- Template may not exist — 404 error
- Folder `task_count` is returned as string, not number

## Feedback Loops

### Unit Tests
```go
func TestFoldersService_Get(t *testing.T)        { /* Test get by ID */ }
func TestFoldersService_Create(t *testing.T)     { /* Test create with name */ }
func TestFoldersService_Update(t *testing.T)     { /* Test update name */ }
func TestFoldersService_Delete(t *testing.T)     { /* Test delete */ }
func TestFoldersService_FromTemplate(t *testing.T) { /* Test template creation */ }
```

## Technical Requirements

- FoldersService should be a new top-level service on `clickup.Client`
- Currently folder listing is on `ListsService.ListFolders` — keep for backward compat, add `FoldersService` for new ops
- Delete returns empty body with 200
