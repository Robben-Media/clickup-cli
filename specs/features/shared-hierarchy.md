# Shared Hierarchy

## Overview

Implement shared hierarchy endpoint that returns all resources shared with the authenticated user.

**Why**: Enables discovery of resources shared across spaces, folders, and lists — useful for guest users and cross-team visibility.

## API Endpoints

| Method | Path | Summary | Operation ID | Version |
|--------|------|---------|--------------|---------|
| GET | /team/{}/shared | Shared Hierarchy | SharedHierarchy | v2 |

## User Stories

### US-001: List Shared Resources

**CLI Command:** `clickup shared-hierarchy list --team <team_id>`

**JSON Output:**
```json
{
  "shared": {
    "tasks": [{"id": "abc", "name": "Shared Task"}],
    "lists": [{"id": "901", "name": "Shared List"}],
    "folders": [{"id": "456", "name": "Shared Folder"}]
  }
}
```

**Plain Output (TSV):** Headers: `TYPE	ID	NAME`
```
task	abc	Shared Task
list	901	Shared List
folder	456	Shared Folder
```

**Human-Readable:**
```
Shared Hierarchy

Tasks:
  abc: Shared Task

Lists:
  901: Shared List

Folders:
  456: Shared Folder
```

## Request/Response Types

```go
type SharedHierarchyResponse struct {
    Shared SharedResources `json:"shared"`
}

type SharedResources struct {
    Tasks   []TaskRef  `json:"tasks,omitempty"`
    Lists   []ListRef  `json:"lists,omitempty"`
    Folders []FolderRef `json:"folders,omitempty"`
}
```

## Feedback Loops

### Unit Tests
```go
func TestSharedHierarchyService_List(t *testing.T) { /* list shared resources */ }
```
