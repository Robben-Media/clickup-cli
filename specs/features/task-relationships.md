# Task Relationships

## Overview

Implement dependency and link management between tasks.

**Why**: Dependencies define execution order; links define related-but-not-blocking relationships. Both are critical for project management workflows.

## API Endpoints

| Method | Path | Summary | Operation ID | Version |
|--------|------|---------|--------------|---------|
| POST | /task/{}/dependency | Add Dependency | AddDependency | v2 |
| DELETE | /task/{}/dependency | Delete Dependency | DeleteDependency | v2 |
| POST | /task/{}/link/{} | Add Task Link | AddTaskLink | v2 |
| DELETE | /task/{}/link/{} | Delete Task Link | DeleteTaskLink | v2 |

## User Stories

### US-001: Add Dependency

**CLI Command:** `clickup relationships add-dep <task_id> --depends-on <other_task_id>` or `--blocking <other_task_id>`

**JSON Output:** `{"status": "success", "message": "Dependency added"}`
**Plain Output (TSV):** Headers: `STATUS	TASK_ID	DEPENDS_ON` → `success	abc	def`

**Acceptance Criteria:**
- [ ] Supports both "waiting on" and "blocking" dependency types
- [ ] `depends_on` = this task waits for other; `dependency_of` = this task blocks other

### US-002: Remove Dependency

**CLI Command:** `clickup relationships remove-dep <task_id> --depends-on <other_task_id>`

### US-003: Add Task Link

**CLI Command:** `clickup relationships link <task_id> <linked_task_id>`

**JSON Output:** `{"status": "success", "message": "Task link added"}`
**Plain Output (TSV):** Headers: `STATUS	TASK_ID	LINKED_TO` → `success	abc	def`

### US-004: Remove Task Link

**CLI Command:** `clickup relationships unlink <task_id> <linked_task_id>`

## Request/Response Types

```go
type AddDependencyRequest struct {
    DependsOn    string `json:"depends_on,omitempty"`
    DependencyOf string `json:"dependency_of,omitempty"`
}
```

## Edge Cases

- Circular dependencies — API may reject or allow
- Linking to a task in a different workspace — may fail
- Delete dependency uses query params (`depends_on` or `dependency_of`), not request body
- Delete link uses the linked task ID in the URL path

## Feedback Loops

### Unit Tests
```go
func TestRelationshipsService_AddDep(t *testing.T)    { /* add depends_on */ }
func TestRelationshipsService_RemoveDep(t *testing.T)  { /* remove dependency */ }
func TestRelationshipsService_Link(t *testing.T)       { /* add link */ }
func TestRelationshipsService_Unlink(t *testing.T)     { /* remove link */ }
```

## Technical Requirements

- New `RelationshipsService` on `clickup.Client`
- Add dependency: POST body with `depends_on` or `dependency_of`
- Delete dependency: DELETE with query params `depends_on` or `dependency_of`
- Links: POST/DELETE with linked task ID in URL path
