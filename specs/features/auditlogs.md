# Audit Logs (v3)

## Overview

Implement workspace-level audit log querying.

**Why**: Audit logs provide a compliance and security trail of workspace activity. CLI access enables automated compliance reporting.

**Requires**: v3 base URL support and workspace ID configuration.

## API Endpoints

| Method | Path | Summary | Operation ID | Version |
|--------|------|---------|--------------|---------|
| POST | /api/v3/workspaces/{}/auditlogs | Query Audit Logs | queryAuditLog | v3 |

Note: This is POST (query with filter body), not a create operation.

## User Stories

### US-001: Query Audit Logs

**CLI Command:** `clickup auditlogs query [--workspace <id>] [--start-date <date>] [--end-date <date>] [--event-type <type>] [--user-id <id>] [--limit <n>]`

**JSON Output:**
```json
{
  "audit_logs": [
    {
      "id": "al_123",
      "event_type": "task_created",
      "user_id": "1",
      "timestamp": 1700000000000,
      "resource_type": "task",
      "resource_id": "abc123",
      "details": {}
    }
  ]
}
```

**Plain Output (TSV):** Headers: `ID	EVENT_TYPE	USER_ID	TIMESTAMP	RESOURCE_TYPE	RESOURCE_ID`
```
al_123	task_created	1	1700000000000	task	abc123
```

**Human-Readable:**
```
Audit Logs

  al_123: task_created by user 1 at 2023-11-14T12:00:00Z
    Resource: task abc123
```

**Acceptance Criteria:**
- [ ] Filters by date range, event type, user
- [ ] Returns paginated results
- [ ] POST method with JSON body (query, not create)

## Request/Response Types

```go
type AuditLogQuery struct {
    StartDate  int64  `json:"start_date,omitempty"`
    EndDate    int64  `json:"end_date,omitempty"`
    EventType  string `json:"event_type,omitempty"`
    UserID     string `json:"user_id,omitempty"`
    Limit      int    `json:"limit,omitempty"`
}

type AuditLogEntry struct {
    ID           string      `json:"id"`
    EventType    string      `json:"event_type"`
    UserID       string      `json:"user_id"`
    Timestamp    json.Number `json:"timestamp"`
    ResourceType string      `json:"resource_type"`
    ResourceID   string      `json:"resource_id"`
    Details      interface{} `json:"details,omitempty"`
}

type AuditLogsResponse struct {
    AuditLogs []AuditLogEntry `json:"audit_logs"`
}
```

## Edge Cases

- POST is used for querying (not GET) due to complex filter body
- Enterprise plan may be required for audit logs
- Large date ranges may be paginated

## Feedback Loops

### Unit Tests
```go
func TestAuditLogsService_Query(t *testing.T) { /* query with filters */ }
```

## Technical Requirements

- New `AuditLogsService` on `clickup.Client`
- Uses `v3Path()` helper
- POST with filter body, not a create operation
