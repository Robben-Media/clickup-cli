# Attachments

## Overview

Implement file attachment management — v2 task attachments and v3 entity-level attachments.

**Why**: Attachments enable file sharing on tasks and other entities. CLI upload enables automated file attachment workflows.

**Requires**: v3 base URL support for v3 endpoints. Multipart upload support for file uploads.

## API Endpoints

| Method | Path | Summary | Operation ID | Version |
|--------|------|---------|--------------|---------|
| POST | /task/{}/attachment | Create Task Attachment | CreateTaskAttachment | v2 |
| GET | /api/v3/workspaces/{}/{parent_type}/{parent_id}/attachments | Get Attachments | getParentEntityAttachments | v3 |
| POST | /api/v3/workspaces/{}/{parent_type}/{parent_id}/attachments | Create an Attachment | postEntityAttachment | v3 |

## User Stories

### US-001: Upload Task Attachment (v2)

**CLI Command:** `clickup attachments upload --task <task_id> <file_path>`

**JSON Output:**
```json
{"id": "att_123", "url": "https://...", "title": "report.pdf", "extension": "pdf", "size": 102400}
```

**Plain Output (TSV):** Headers: `ID	TITLE	SIZE	URL`
```
att_123	report.pdf	102400	https://...
```

**Acceptance Criteria:**
- [ ] Uploads file via multipart/form-data
- [ ] Field name is "attachment"
- [ ] Returns attachment metadata

### US-002: List Attachments (v3)

**CLI Command:** `clickup attachments list [--workspace <id>] --parent-type <task|list|folder|space> --parent-id <id>`

**JSON Output:**
```json
{"attachments": [{"id": "att_123", "title": "report.pdf", "url": "https://...", "size": 102400}]}
```

**Plain Output (TSV):** Headers: `ID	TITLE	SIZE	URL`

### US-003: Create Attachment (v3)

**CLI Command:** `clickup attachments create [--workspace <id>] --parent-type <type> --parent-id <id> <file_path>`

## Request/Response Types

```go
type Attachment struct {
    ID        string `json:"id"`
    URL       string `json:"url"`
    Title     string `json:"title"`
    Extension string `json:"extension,omitempty"`
    Size      int64  `json:"size"`
}

type AttachmentsResponse struct {
    Attachments []Attachment `json:"attachments"`
}
```

## Edge Cases

- v2 upload uses `multipart/form-data` with field name "attachment"
- v3 parent types are path segments (task, list, folder, space)
- Max file size depends on workspace plan
- File path must exist and be readable

## Feedback Loops

### Unit Tests
```go
func TestAttachmentsService_Upload(t *testing.T)   { /* v2 multipart upload */ }
func TestAttachmentsService_List(t *testing.T)     { /* v3 list */ }
func TestAttachmentsService_Create(t *testing.T)   { /* v3 multipart upload */ }
```

## Technical Requirements

- `PostMultipart()` method needed on `api.Client` (see architecture spec)
- v2: POST multipart to `/v2/task/{task_id}/attachment`
- v3: POST multipart to `/v3/workspaces/{wid}/{parent_type}/{parent_id}/attachments`
- v3 GET returns attachment list for any parent entity type
