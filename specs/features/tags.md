# Tags

## Overview

Implement the Tags domain for managing space-level tags and tag-task associations.

**Why**: Tags enable cross-list categorization and filtering. Essential for `GetFilteredTeamTasks` tag filters to work meaningfully.

## API Endpoints

| Method | Path | Summary | Operation ID | Version |
|--------|------|---------|--------------|---------|
| GET | /space/{}/tag | Get Space Tags | GetSpaceTags | v2 |
| POST | /space/{}/tag | Create Space Tag | CreateSpaceTag | v2 |
| PUT | /space/{}/tag/{} | Edit Space Tag | EditSpaceTag | v2 |
| DELETE | /space/{}/tag/{} | Delete Space Tag | DeleteSpaceTag | v2 |
| POST | /task/{}/tag/{} | Add Tag To Task | AddTagToTask | v2 |
| DELETE | /task/{}/tag/{} | Remove Tag From Task | RemoveTagFromTask | v2 |

## User Stories

### US-001: List Space Tags

**CLI Command:** `clickup tags list --space <space_id>`

**JSON Output:**
```json
{"tags": [{"name": "bug", "tag_fg": "#fff", "tag_bg": "#f44336"}, {"name": "feature", "tag_fg": "#fff", "tag_bg": "#4caf50"}]}
```

**Plain Output (TSV):** Headers: `NAME	FG_COLOR	BG_COLOR`
```
bug	#fff	#f44336
feature	#fff	#4caf50
```

**Human-Readable:**
```
Tags in space 789

  bug (bg: #f44336)
  feature (bg: #4caf50)
```

### US-002: Create Space Tag

**CLI Command:** `clickup tags create --space <space_id> <name> [--bg <color>] [--fg <color>]`

**JSON Output:** `{"status": "success", "message": "Tag created"}`
**Plain Output (TSV):** Headers: `STATUS	NAME` → `success	bug`

### US-003: Edit Space Tag

**CLI Command:** `clickup tags update --space <space_id> <name> [--new-name <name>] [--bg <color>] [--fg <color>]`

### US-004: Delete Space Tag

**CLI Command:** `clickup tags delete --space <space_id> <name>`

### US-005: Add Tag to Task

**CLI Command:** `clickup tags add --task <task_id> <tag_name>`

**Acceptance Criteria:**
- [ ] Tag must exist in the task's space
- [ ] Returns success

### US-006: Remove Tag from Task

**CLI Command:** `clickup tags remove --task <task_id> <tag_name>`

## Request/Response Types

```go
type SpaceTag struct {
    Name  string `json:"name"`
    TagFg string `json:"tag_fg,omitempty"`
    TagBg string `json:"tag_bg,omitempty"`
}

type SpaceTagsResponse struct {
    Tags []SpaceTag `json:"tags"`
}

type CreateSpaceTagRequest struct {
    Tag SpaceTag `json:"tag"`
}

type EditSpaceTagRequest struct {
    Tag SpaceTag `json:"tag"`
}
```

## Edge Cases

- Tag names are case-sensitive in the API
- Tag paths use the tag name (URL-encoded), not an ID
- Adding a non-existent tag to a task creates it in the space
- Editing wraps the tag in a `{"tag": {...}}` envelope

## Feedback Loops

### Unit Tests
```go
func TestTagsService_List(t *testing.T)       { /* space tags */ }
func TestTagsService_Create(t *testing.T)     { /* create with colors */ }
func TestTagsService_Update(t *testing.T)     { /* rename/recolor */ }
func TestTagsService_Delete(t *testing.T)     { /* delete */ }
func TestTagsService_AddToTask(t *testing.T)  { /* add to task */ }
func TestTagsService_RemoveFromTask(t *testing.T) { /* remove from task */ }
```

## Technical Requirements

- New `TagsService` on `clickup.Client`
- Tag CRUD uses `space_id` in path
- Task tag operations use `task_id` and URL-encoded `tag_name` in path
- Create/Edit wrap body in `{"tag": {...}}` envelope
