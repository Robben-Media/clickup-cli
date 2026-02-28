# Comments

## Overview

Extend the existing Comments domain to full API parity. The CLI already supports listing task comments and creating task comments. This spec adds delete, update, threaded replies, list-level comments, view-level comments, and v3 post subtype retrieval.

**Why**: Comments are a core collaboration primitive. Supporting threaded replies and multi-context comments (list, view) enables richer automation workflows.

## API Endpoints

| Status | Method | Path | Summary | Operation ID | Version |
|--------|--------|------|---------|--------------|---------|
| impl | GET | /task/{}/comment | Get Task Comments | GetTaskComments | v2 |
| impl | POST | /task/{}/comment | Create Task Comment | CreateTaskComment | v2 |
| missing | DELETE | /comment/{} | Delete Comment | DeleteComment | v2 |
| missing | PUT | /comment/{} | Update Comment | UpdateComment | v2 |
| missing | GET | /comment/{}/reply | Get Threaded Comments | GetThreadedComments | v2 |
| missing | POST | /comment/{}/reply | Create Threaded Comment | CreateThreadedComment | v2 |
| missing | GET | /list/{}/comment | Get List Comments | GetListComments | v2 |
| missing | POST | /list/{}/comment | Create List Comment | CreateListComment | v2 |
| missing | GET | /view/{}/comment | Get Chat View Comments | GetChatViewComments | v2 |
| missing | POST | /view/{}/comment | Create Chat View Comment | CreateChatViewComment | v2 |
| missing | GET | /api/v3/workspaces/{}/comments/types/{}/subtypes | Get Post Subtype IDs | getSubtypes | v3 |

## User Stories

### US-001: Delete a Comment

As a CLI user,
I want to delete a comment,
so that I can remove outdated or incorrect information.

**Acceptance Criteria:**
- [ ] `clickup comments delete <comment_id>` deletes the comment
- [ ] Returns success confirmation

**CLI Command:** `clickup comments delete <comment_id>`

**JSON Output:**
```json
{"status": "success", "message": "Comment deleted", "comment_id": "456"}
```

**Plain Output (TSV):**
Headers: `STATUS	COMMENT_ID`
```
success	456
```

**Human-Readable Output:**
```
Comment 456 deleted
```

### US-002: Update a Comment

As a CLI user,
I want to edit an existing comment,
so that I can correct or update information.

**Acceptance Criteria:**
- [ ] `clickup comments update <comment_id> --text "..."` updates the comment text
- [ ] Supports `--resolved` flag to mark as resolved
- [ ] Supports `--assignee <user_id>` to reassign

**CLI Command:** `clickup comments update <comment_id> [--text "..."] [--resolved] [--assignee <user_id>]`

**JSON Output:**
```json
{"status": "success", "message": "Comment updated", "comment_id": "456"}
```

**Plain Output (TSV):**
Headers: `STATUS	COMMENT_ID`
```
success	456
```

**Human-Readable Output:**
```
Comment 456 updated
```

### US-003: View Threaded Replies

As a CLI user,
I want to see replies to a comment,
so that I can follow conversation threads.

**Acceptance Criteria:**
- [ ] `clickup comments replies <comment_id>` lists all replies
- [ ] Shows reply author, date, and text

**CLI Command:** `clickup comments replies <comment_id>`

**JSON Output:**
```json
{
  "comments": [
    {"id": "789", "comment_text": "I agree", "user": {"id": 1, "username": "alice"}, "date": "1700000000000"}
  ]
}
```

**Plain Output (TSV):**
Headers: `ID	USER	DATE	TEXT`
```
789	alice	1700000000000	I agree
```

**Human-Readable Output:**
```
Replies to comment 456

ID: 789
  User: alice
  Date: 1700000000000
  Text: I agree
```

### US-004: Create Threaded Reply

As a CLI user,
I want to reply to an existing comment,
so that I can participate in discussions.

**Acceptance Criteria:**
- [ ] `clickup comments reply <comment_id> --text "..."` creates a reply
- [ ] Returns the reply ID

**CLI Command:** `clickup comments reply <comment_id> --text "..."`

**JSON Output:**
```json
{"id": "789", "comment_text": "My reply", "user": {"id": 1, "username": "jeremy"}}
```

**Plain Output (TSV):**
Headers: `ID	TEXT`
```
789	My reply
```

**Human-Readable Output:**
```
Reply created

ID: 789
Text: My reply
```

### US-005: List and Create List Comments

As a CLI user,
I want to manage comments at the list level,
so that I can discuss list-wide topics.

**Acceptance Criteria:**
- [ ] `clickup comments list-comments <list_id>` returns list-level comments
- [ ] `clickup comments add-list <list_id> --text "..."` creates a list comment
- [ ] Supports `--assignee <user_id>` on creation

**CLI Commands:**
- `clickup comments list-comments <list_id>`
- `clickup comments add-list <list_id> --text "..." [--assignee <user_id>]`

**JSON/Plain/Human Output:** Same format as task comments (see existing implementation).

### US-006: List and Create View Comments

As a CLI user,
I want to manage comments on views,
so that I can annotate dashboards and reports.

**Acceptance Criteria:**
- [ ] `clickup comments view-comments <view_id>` returns view-level comments
- [ ] `clickup comments add-view <view_id> --text "..."` creates a view comment
- [ ] Supports `--start` and `--start_id` for pagination

**CLI Commands:**
- `clickup comments view-comments <view_id> [--start <int>] [--start-id <string>]`
- `clickup comments add-view <view_id> --text "..." [--assignee <user_id>]`

**JSON/Plain/Human Output:** Same format as task comments.

### US-007: Get Post Subtype IDs (v3)

As a CLI user,
I want to retrieve available comment post subtypes,
so that I can create properly typed comments.

**Acceptance Criteria:**
- [ ] `clickup comments subtypes <type_id>` returns subtype IDs
- [ ] Requires `--workspace` flag or env var

**CLI Command:** `clickup comments subtypes <type_id> [--workspace <workspace_id>]`

**JSON Output:**
```json
{"subtypes": [{"id": "1", "name": "announcement"}, {"id": "2", "name": "question"}]}
```

**Plain Output (TSV):**
Headers: `ID	NAME`
```
1	announcement
2	question
```

**Human-Readable Output:**
```
Post Subtypes for type 1

  1: announcement
  2: question
```

## Request/Response Types

```go
// UpdateCommentRequest
type UpdateCommentRequest struct {
    CommentText string `json:"comment_text,omitempty"`
    Assignee    int    `json:"assignee,omitempty"`
    Resolved    *bool  `json:"resolved,omitempty"`
}

// CreateListCommentRequest
type CreateListCommentRequest struct {
    CommentText string `json:"comment_text"`
    Assignee    int    `json:"assignee,omitempty"`
}

// CreateViewCommentRequest
type CreateViewCommentRequest struct {
    CommentText string `json:"comment_text"`
    Assignee    int    `json:"assignee,omitempty"`
}

// ThreadedCommentsResponse
type ThreadedCommentsResponse struct {
    Comments []Comment `json:"comments"`
}

// PostSubtypesResponse
type PostSubtypesResponse struct {
    Subtypes []PostSubtype `json:"subtypes"`
}

type PostSubtype struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}
```

## Edge Cases

- Deleting a comment that has replies — ClickUp may cascade or error
- Updating a comment you didn't author — API may return 403
- View comments use `start` (integer offset) and `start_id` (string) for pagination, not page numbers
- Comment text may contain markdown — pass through as-is
- v3 subtypes endpoint requires workspace ID

## Feedback Loops

### Unit Tests
```go
func TestCommentsService_Delete(t *testing.T) {
    // Test successful delete
    // Test delete with invalid ID returns error
}

func TestCommentsService_Update(t *testing.T) {
    // Test update text only
    // Test update resolved flag
    // Test update assignee
}

func TestCommentsService_ThreadedReplies(t *testing.T) {
    // Test listing replies
    // Test creating a reply
    // Test empty replies list
}

func TestCommentsService_ListComments(t *testing.T) {
    // Test list-level comment listing
    // Test list-level comment creation
}

func TestCommentsService_ViewComments(t *testing.T) {
    // Test view-level comment listing with pagination
    // Test view-level comment creation
}
```

## Technical Requirements

- Comment IDs are returned as `json.Number` from ClickUp API (existing pattern)
- Delete returns empty body with 200 status
- Update returns empty body with 200 status
- Reply endpoint uses same path structure as get replies
- View comments support `start` and `start_id` query params for cursor pagination
