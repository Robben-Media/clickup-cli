# Tasks

## Overview

Extend the existing Tasks domain to full API parity. The CLI already supports CRUD (get, list, create, update, delete). This spec covers the 5 remaining v2 operations and 1 v3 operation: workspace-level task search, time-in-status queries, task merging, template creation, and moving tasks between lists.

**Why**: `GetFilteredTeamTasks` is the single most impactful missing endpoint — it enables workspace-wide task search with filters, which is critical for automation and reporting workflows.

## API Endpoints

| Status | Method | Path | Summary | Operation ID | Version |
|--------|--------|------|---------|--------------|---------|
| missing | GET | /team/{}/task | Get Filtered Team Tasks | GetFilteredTeamTasks | v2 |
| missing | GET | /task/{}/time_in_status | Get Task's Time in Status | GetTask'sTimeinStatus | v2 |
| missing | GET | /task/bulk_time_in_status/task_ids | Get Bulk Tasks' Time in Status | GetBulkTasks'TimeinStatus | v2 |
| missing | POST | /task/{}/merge | Merge Tasks | mergeTasks | v2 |
| missing | POST | /list/{}/taskTemplate/{} | Create Task From Template | CreateTaskFromTemplate | v2 |
| missing | PUT | /api/v3/workspaces/{}/tasks/{}/home_list/{} | Move a task to a new List | moveTask | v3 |

## User Stories

### US-001: Search Tasks Across Workspace

As a CLI user,
I want to search for tasks across my entire workspace with filters,
so that I can find tasks without knowing which list they're in.

**Acceptance Criteria:**
- [ ] `clickup tasks search --team <id>` returns tasks from all lists
- [ ] Supports `--status`, `--assignee`, `--tag`, `--due-date-gt`, `--due-date-lt` filters
- [ ] Supports `--page` and `--order-by` for pagination
- [ ] `--include-closed` flag includes closed tasks
- [ ] Returns up to 100 tasks per page (ClickUp API limit)

**CLI Command:** `clickup tasks search --team <team_id> [--status <status>] [--assignee <user_id>] [--tag <tag>] [--due-date-gt <ms>] [--due-date-lt <ms>] [--include-closed] [--page <n>] [--order-by <field>]`

**JSON Output:**
```json
{
  "tasks": [
    {
      "id": "abc123",
      "name": "Fix login bug",
      "status": {"status": "in progress"},
      "priority": {"id": "2", "priority": "high"},
      "due_date": "1700000000000",
      "assignees": [{"id": 123, "username": "jeremy"}],
      "url": "https://app.clickup.com/t/abc123",
      "list": {"id": "901", "name": "Sprint 1"},
      "folder": {"id": "456"},
      "space": {"id": "789"}
    }
  ]
}
```

**Plain Output (TSV):**
Headers: `ID	NAME	STATUS	PRIORITY	ASSIGNEES	LIST	URL`
```
abc123	Fix login bug	in progress	high	jeremy	Sprint 1	https://app.clickup.com/t/abc123
```

**Human-Readable Output:**
```
Found 12 tasks

ID: abc123
  Name: Fix login bug
  Status: in progress
  Priority: high
  Assignees: jeremy
  List: Sprint 1
  URL: https://app.clickup.com/t/abc123
```

### US-002: Get Task Time in Status

As a CLI user,
I want to see how long a task has been in each status,
so that I can identify bottlenecks in my workflow.

**Acceptance Criteria:**
- [ ] `clickup tasks time-in-status <task_id>` returns time spent per status
- [ ] Times are in milliseconds and human-readable duration
- [ ] Current status shows "current_status" marker

**CLI Command:** `clickup tasks time-in-status <task_id>`

**JSON Output:**
```json
{
  "current_status": {
    "status": "in progress",
    "color": "#4194f6",
    "total_time": {
      "by_minute": 120,
      "since": "1700000000000"
    }
  },
  "status_history": [
    {
      "status": "open",
      "total_time": {"by_minute": 60, "since": "1699900000000"}
    }
  ]
}
```

**Plain Output (TSV):**
Headers: `STATUS	MINUTES	CURRENT`
```
open	60	false
in progress	120	true
```

**Human-Readable Output:**
```
Time in Status for task abc123

  open: 1h 0m
  in progress: 2h 0m (current)
```

### US-003: Bulk Time in Status

As a CLI user,
I want to get time-in-status for multiple tasks at once,
so that I can build reports efficiently.

**Acceptance Criteria:**
- [ ] `clickup tasks bulk-time-in-status <task_id>...` accepts multiple task IDs
- [ ] Returns per-task time-in-status data

**CLI Command:** `clickup tasks bulk-time-in-status <task_id> [<task_id>...]`

**JSON Output:**
```json
{
  "abc123": {
    "current_status": {"status": "in progress", "total_time": {"by_minute": 120}},
    "status_history": [{"status": "open", "total_time": {"by_minute": 60}}]
  },
  "def456": {
    "current_status": {"status": "done", "total_time": {"by_minute": 30}},
    "status_history": [{"status": "open", "total_time": {"by_minute": 45}}]
  }
}
```

**Plain Output (TSV):**
Headers: `TASK_ID	STATUS	MINUTES	CURRENT`
```
abc123	open	60	false
abc123	in progress	120	true
def456	open	45	false
def456	done	30	true
```

**Human-Readable Output:**
```
Task abc123:
  open: 1h 0m
  in progress: 2h 0m (current)

Task def456:
  open: 0h 45m
  done: 0h 30m (current)
```

### US-004: Merge Tasks

As a CLI user,
I want to merge duplicate tasks into one,
so that I can consolidate work items.

**Acceptance Criteria:**
- [ ] `clickup tasks merge <task_id> --into <target_id>` merges source into target
- [ ] Confirm before merge (destructive operation)
- [ ] Returns the merged task

**CLI Command:** `clickup tasks merge <task_id> --into <target_task_id>`

**JSON Output:**
```json
{"status": "success", "message": "Task merged", "task_id": "abc123", "merged_into": "def456"}
```

**Plain Output (TSV):**
Headers: `STATUS	TASK_ID	MERGED_INTO`
```
success	abc123	def456
```

**Human-Readable Output:**
```
Task abc123 merged into def456
```

### US-005: Create Task from Template

As a CLI user,
I want to create a task from an existing template,
so that I can standardize task creation.

**Acceptance Criteria:**
- [ ] `clickup tasks from-template <list_id> <template_id>` creates a task
- [ ] Supports `--name` override
- [ ] Returns the created task

**CLI Command:** `clickup tasks from-template <list_id> <template_id> [--name <name>]`

**JSON Output:** Same as `tasks create` — full Task object.

**Plain Output (TSV):**
Headers: `ID	NAME	STATUS	URL`
```
abc123	Weekly Report	open	https://app.clickup.com/t/abc123
```

**Human-Readable Output:**
```
Created task from template

ID: abc123
Name: Weekly Report
Status: open
URL: https://app.clickup.com/t/abc123
```

### US-006: Move Task to New List (v3)

As a CLI user,
I want to move a task to a different list,
so that I can reorganize my work.

**Acceptance Criteria:**
- [ ] `clickup tasks move <task_id> --list <list_id>` moves the task
- [ ] Requires `--workspace` flag or `CLICKUP_WORKSPACE_ID` env var
- [ ] Returns success confirmation

**CLI Command:** `clickup tasks move <task_id> --list <list_id> [--workspace <workspace_id>]`

**JSON Output:**
```json
{"status": "success", "message": "Task moved", "task_id": "abc123", "list_id": "901"}
```

**Plain Output (TSV):**
Headers: `STATUS	TASK_ID	LIST_ID`
```
success	abc123	901
```

**Human-Readable Output:**
```
Task abc123 moved to list 901
```

## Request/Response Types

```go
// GetFilteredTeamTasks query parameters
type FilteredTeamTasksParams struct {
    Page           int      `url:"page,omitempty"`
    OrderBy        string   `url:"order_by,omitempty"`
    Reverse        bool     `url:"reverse,omitempty"`
    Subtasks       bool     `url:"subtasks,omitempty"`
    Statuses       []string `url:"statuses[],omitempty"`
    IncludeClosed  bool     `url:"include_closed,omitempty"`
    Assignees      []int    `url:"assignees[],omitempty"`
    Tags           []string `url:"tags[],omitempty"`
    DueDateGt      int64    `url:"due_date_gt,omitempty"`
    DueDateLt      int64    `url:"due_date_lt,omitempty"`
    DateCreatedGt  int64    `url:"date_created_gt,omitempty"`
    DateCreatedLt  int64    `url:"date_created_lt,omitempty"`
    DateUpdatedGt  int64    `url:"date_updated_gt,omitempty"`
    DateUpdatedLt  int64    `url:"date_updated_lt,omitempty"`
}

// TimeInStatusResponse for single task
type TimeInStatusResponse struct {
    CurrentStatus StatusTime            `json:"current_status"`
    StatusHistory []StatusTime          `json:"status_history"`
}

type StatusTime struct {
    Status    string    `json:"status"`
    Color     string    `json:"color,omitempty"`
    TotalTime TimeValue `json:"total_time"`
}

type TimeValue struct {
    ByMinute int64  `json:"by_minute"`
    Since    string `json:"since,omitempty"`
}

// BulkTimeInStatusResponse maps task IDs to their time-in-status
type BulkTimeInStatusResponse map[string]TimeInStatusResponse

// MergeTasksRequest
type MergeTasksRequest struct {
    MergedTaskIDs []string `json:"merged_task_ids"`
}
```

## Edge Cases

- `GetFilteredTeamTasks` with no filters returns all tasks (up to page limit)
- Task IDs for bulk operations may include invalid IDs — partial results expected
- Merge is irreversible — warn user
- Move task (v3) requires workspace ID — error clearly if missing
- Templates may reference custom fields or statuses not in the target list

## Feedback Loops

### Unit Tests
```go
// internal/clickup/client_test.go
func TestTasksService_Search(t *testing.T) {
    // Test with various filter combinations
    // Test empty results
    // Test pagination parameters
}

func TestTasksService_TimeInStatus(t *testing.T) {
    // Test single task time-in-status parsing
    // Test current status marker
}

func TestTasksService_BulkTimeInStatus(t *testing.T) {
    // Test multiple task IDs
    // Test response map parsing
}

func TestTasksService_Merge(t *testing.T) {
    // Test merge request body
    // Test error on invalid task ID
}

func TestTasksService_MoveTask(t *testing.T) {
    // Test v3 path construction
    // Test workspace ID requirement
}
```

### Integration Tests
- Search with real workspace returns valid tasks
- Time-in-status returns valid duration data
- Move task changes the task's list reference

## Technical Requirements

- `GetFilteredTeamTasks` uses query string arrays (`statuses[]=open&statuses[]=closed`)
- Bulk time-in-status sends task IDs as query params: `?task_ids=abc&task_ids=def`
- Move task is v3 only — requires v3 base URL support (see architecture spec)
- Merge tasks is POST with JSON body containing array of task IDs to merge into the target
