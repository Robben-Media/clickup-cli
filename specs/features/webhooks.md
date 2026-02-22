# Webhooks

## Overview

Implement webhook CRUD for workspace-level event subscriptions.

**Why**: Webhooks enable real-time integration with external systems. CLI management enables automated webhook provisioning.

## API Endpoints

| Method | Path | Summary | Operation ID | Version |
|--------|------|---------|--------------|---------|
| GET | /team/{}/webhook | Get Webhooks | GetWebhooks | v2 |
| POST | /team/{}/webhook | Create Webhook | CreateWebhook | v2 |
| PUT | /webhook/{} | Update Webhook | UpdateWebhook | v2 |
| DELETE | /webhook/{} | Delete Webhook | DeleteWebhook | v2 |

## User Stories

### US-001: List Webhooks

**CLI Command:** `clickup webhooks list --team <team_id>`

**JSON Output:**
```json
{
  "webhooks": [
    {"id": "wh-123", "userid": 1, "team_id": "789", "endpoint": "https://example.com/hook", "events": ["taskCreated", "taskUpdated"], "status": "active"}
  ]
}
```

**Plain Output (TSV):** Headers: `ID	ENDPOINT	STATUS	EVENTS`
```
wh-123	https://example.com/hook	active	taskCreated,taskUpdated
```

### US-002: Create Webhook

**CLI Command:** `clickup webhooks create --team <team_id> --endpoint <url> --events <event1,event2> [--space <id>] [--folder <id>] [--list <id>] [--task <id>]`

**Acceptance Criteria:**
- [ ] Endpoint URL required
- [ ] At least one event required
- [ ] Can scope to space/folder/list/task level
- [ ] Returns the created webhook

**Events:** taskCreated, taskUpdated, taskDeleted, taskPriorityUpdated, taskStatusUpdated, taskAssigneeUpdated, taskDueDateUpdated, taskTagUpdated, taskMoved, taskCommentPosted, taskCommentUpdated, taskTimeEstimateUpdated, taskTimeTrackedUpdated, listCreated, listUpdated, listDeleted, folderCreated, folderUpdated, folderDeleted, spaceCreated, spaceUpdated, spaceDeleted, goalCreated, goalUpdated, goalDeleted, keyResultCreated, keyResultUpdated, keyResultDeleted

### US-003: Update Webhook

**CLI Command:** `clickup webhooks update <webhook_id> [--endpoint <url>] [--events <events>] [--status <active|inactive>]`

### US-004: Delete Webhook

**CLI Command:** `clickup webhooks delete <webhook_id>`

## Request/Response Types

```go
type Webhook struct {
    ID       string   `json:"id"`
    UserID   int      `json:"userid"`
    TeamID   string   `json:"team_id"`
    Endpoint string   `json:"endpoint"`
    Events   []string `json:"events"`
    Status   string   `json:"status"`
    SpaceID  string   `json:"space_id,omitempty"`
    FolderID string   `json:"folder_id,omitempty"`
    ListID   string   `json:"list_id,omitempty"`
    TaskID   string   `json:"task_id,omitempty"`
}

type WebhooksResponse struct {
    Webhooks []Webhook `json:"webhooks"`
}

type CreateWebhookRequest struct {
    Endpoint string   `json:"endpoint"`
    Events   []string `json:"events"`
    SpaceID  string   `json:"space_id,omitempty"`
    FolderID string   `json:"folder_id,omitempty"`
    ListID   string   `json:"list_id,omitempty"`
    TaskID   string   `json:"task_id,omitempty"`
}

type UpdateWebhookRequest struct {
    Endpoint string   `json:"endpoint,omitempty"`
    Events   []string `json:"events,omitempty"`
    Status   string   `json:"status,omitempty"` // "active" or "inactive"
}
```

## Edge Cases

- Webhook health: ClickUp disables webhooks after repeated failures
- Events are strings, not enums — typos won't error on creation
- Scoping to space/folder/list/task narrows which events fire

## Feedback Loops

### Unit Tests
```go
func TestWebhooksService_List(t *testing.T)   { /* list webhooks */ }
func TestWebhooksService_Create(t *testing.T) { /* create with events */ }
func TestWebhooksService_Update(t *testing.T) { /* update endpoint/events/status */ }
func TestWebhooksService_Delete(t *testing.T) { /* delete */ }
```
