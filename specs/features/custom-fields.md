# Custom Fields

## Overview

Implement custom field discovery and value management. Custom fields can be scoped to list, folder, space, or workspace level.

**Why**: Custom fields extend tasks with domain-specific data (budgets, priorities, dropdowns). Setting field values from CLI enables data-driven automation.

## API Endpoints

| Method | Path | Summary | Operation ID | Version |
|--------|------|---------|--------------|---------|
| GET | /list/{}/field | Get List Custom Fields | GetAccessibleCustomFields | v2 |
| GET | /folder/{}/field | Get Folder Custom Fields | getFolderAvailableFields | v2 |
| GET | /space/{}/field | Get Space Custom Fields | getSpaceAvailableFields | v2 |
| GET | /team/{}/field | Get Workspace Custom Fields | getTeamAvailableFields | v2 |
| POST | /task/{}/field/{} | Set Custom Field Value | SetCustomFieldValue | v2 |
| DELETE | /task/{}/field/{} | Remove Custom Field Value | RemoveCustomFieldValue | v2 |

## User Stories

### US-001: List Custom Fields

**CLI Command:** `clickup fields list --list <id>` / `--folder <id>` / `--space <id>` / `--team <id>`

**JSON Output:**
```json
{
  "fields": [
    {
      "id": "cf-123",
      "name": "Budget",
      "type": "currency",
      "type_config": {"precision": 2, "currency_type": "USD"},
      "required": false
    }
  ]
}
```

**Plain Output (TSV):** Headers: `ID	NAME	TYPE	REQUIRED`
```
cf-123	Budget	currency	false
```

**Human-Readable:**
```
Custom Fields

  cf-123: Budget (currency, optional)
  cf-456: Priority Level (dropdown, required)
```

**Acceptance Criteria:**
- [ ] Exactly one scope flag required
- [ ] Returns field ID, name, type, config, and required status

### US-002: Set Custom Field Value

**CLI Command:** `clickup fields set --task <task_id> --field <field_id> --value <value>`

**Acceptance Criteria:**
- [ ] Value format depends on field type (number, string, array, object)
- [ ] For dropdowns, value is the option UUID
- [ ] Returns success

**JSON Output:** `{"status": "success"}`
**Plain Output (TSV):** Headers: `STATUS	TASK_ID	FIELD_ID` → `success	abc	cf-123`

### US-003: Remove Custom Field Value

**CLI Command:** `clickup fields remove --task <task_id> --field <field_id>`

## Request/Response Types

```go
type CustomField struct {
    ID         string      `json:"id"`
    Name       string      `json:"name"`
    Type       string      `json:"type"`
    TypeConfig interface{} `json:"type_config,omitempty"`
    Required   bool        `json:"required"`
}

type CustomFieldsResponse struct {
    Fields []CustomField `json:"fields"`
}

type SetCustomFieldRequest struct {
    Value interface{} `json:"value"`
}
```

## Edge Cases

- Field types: text, number, currency, dropdown, checkbox, date, email, phone, url, labels, automatic_progress, manual_progress, tasks, users, emoji, location
- Dropdown values require the option UUID, not the label text
- Some field types require nested objects (e.g., location: `{"lat": 0, "lng": 0}`)
- Setting a value on a task that doesn't have the field — API auto-associates

## Feedback Loops

### Unit Tests
```go
func TestCustomFieldsService_ListByList(t *testing.T)      { /* list-scoped */ }
func TestCustomFieldsService_ListByFolder(t *testing.T)    { /* folder-scoped */ }
func TestCustomFieldsService_ListBySpace(t *testing.T)     { /* space-scoped */ }
func TestCustomFieldsService_ListByWorkspace(t *testing.T) { /* workspace-scoped */ }
func TestCustomFieldsService_Set(t *testing.T)             { /* set value */ }
func TestCustomFieldsService_Remove(t *testing.T)          { /* remove value */ }
```

## Technical Requirements

- New `CustomFieldsService` on `clickup.Client`
- All list endpoints return `{"fields": [...]}`
- Set value: POST with `{"value": <any>}` body — field ID in URL path
- Remove value: DELETE — field ID in URL path, no body
