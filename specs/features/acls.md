# ACLs (v3)

## Overview

Implement access control list management for workspace objects.

**Why**: ACLs control privacy and access at the object level. CLI access enables automated permission management across spaces, folders, and lists.

**Requires**: v3 base URL support and workspace ID configuration.

## API Endpoints

| Method | Path | Summary | Operation ID | Version |
|--------|------|---------|--------------|---------|
| PATCH | /api/v3/workspaces/{}/{object_type}/{object_id}/acls | Update ACLs | publicPatchAcl | v3 |

## User Stories

### US-001: Update Object Access

**CLI Command:** `clickup acls update [--workspace <id>] --type <space|folder|list> --id <object_id> --private [--sharing <open|closed>]`

**JSON Output:**
```json
{"status": "success", "message": "ACL updated"}
```

**Plain Output (TSV):** Headers: `STATUS	OBJECT_TYPE	OBJECT_ID`
```
success	space	789
```

**Human-Readable:**
```
ACL updated for space 789
```

**Acceptance Criteria:**
- [ ] Object type is a path segment (space, folder, list)
- [ ] Can set private/public and sharing mode
- [ ] Uses PATCH method

## Request/Response Types

```go
type UpdateACLRequest struct {
    Private *bool  `json:"private,omitempty"`
    Sharing string `json:"sharing,omitempty"` // "open" or "closed"
}
```

## Edge Cases

- PATCH method (not PUT)
- Object type in URL path: `/{object_type}/{object_id}/acls`
- Changing privacy may cascade to child objects
- Enterprise plan may be required

## Feedback Loops

### Unit Tests
```go
func TestACLsService_Update(t *testing.T) { /* update privacy */ }
```

## Technical Requirements

- New `ACLsService` on `clickup.Client` (or method on a shared v3 service)
- Uses `v3Path()` helper with dynamic object_type/object_id path segments
- Requires PATCH method on `api.Client`
