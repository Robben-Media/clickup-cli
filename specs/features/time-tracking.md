# Time Tracking

## Overview

Extend the Time Tracking domain to full parity. The CLI currently supports listing time entries and creating entries. This spec adds get single entry, running timer management (start/stop/current), entry CRUD, tag management, and entry history.

**Why**: Time tracking is critical for billing, reporting, and productivity analysis. Start/stop timer is the most requested missing feature for interactive CLI usage.

## API Endpoints

| Status | Method | Path | Summary | Operation ID | Version |
|--------|--------|------|---------|--------------|---------|
| impl | GET | /team/{}/time_entries | Get time entries within a date range | Gettimeentrieswithinadaterange | v2 |
| impl | POST | /team/{}/time_entries | Create a time entry | Createatimeentry | v2 |
| missing | GET | /team/{}/time_entries/{} | Get singular time entry | Getsingulartimeentry | v2 |
| missing | GET | /team/{}/time_entries/current | Get running time entry | Getrunningtimeentry | v2 |
| missing | GET | /team/{}/time_entries/{}/history | Get time entry history | Gettimeentryhistory | v2 |
| missing | GET | /team/{}/time_entries/tags | Get all tags from time entries | Getalltagsfromtimeentries | v2 |
| missing | POST | /team/{}/time_entries/start | Start a time entry | StartatimeEntry | v2 |
| missing | POST | /team/{}/time_entries/stop | Stop a time entry | StopatimeEntry | v2 |
| missing | POST | /team/{}/time_entries/tags | Add tags from time entries | Addtagsfromtimeentries | v2 |
| missing | PUT | /team/{}/time_entries/{} | Update a time entry | UpdateatimeEntry | v2 |
| missing | PUT | /team/{}/time_entries/tags | Change tag names from time entries | Changetagnamesfromtimeentries | v2 |
| missing | DELETE | /team/{}/time_entries/{} | Delete a time entry | DeleteatimeEntry | v2 |
| missing | DELETE | /team/{}/time_entries/tags | Remove tags from time entries | Removetagsfromtimeentries | v2 |

## User Stories

### US-001: Get Single Time Entry

**CLI Command:** `clickup time get <entry_id> --team <team_id>`

**JSON Output:**
```json
{
  "data": {
    "id": "123",
    "task": {"id": "abc", "name": "Fix bug"},
    "wid": "789",
    "user": {"id": 1, "username": "jeremy"},
    "billable": true,
    "start": "1700000000000",
    "end": "1700003600000",
    "duration": "3600000",
    "description": "Debugging login issue",
    "tags": [{"name": "billable"}]
  }
}
```

**Plain Output (TSV):**
Headers: `ID	TASK	USER	START	END	DURATION_MS	DESCRIPTION`
```
123	abc	jeremy	1700000000000	1700003600000	3600000	Debugging login issue
```

**Human-Readable Output:**
```
Time Entry 123
  Task: Fix bug (abc)
  User: jeremy
  Duration: 1h 0m
  Start: 2023-11-14T12:00:00Z
  End: 2023-11-14T13:00:00Z
  Description: Debugging login issue
  Tags: billable
```

**Acceptance Criteria:**
- [ ] Returns full time entry with task ref, user, duration, tags

### US-002: Get Running Timer

**CLI Command:** `clickup time current --team <team_id>`

**JSON Output:**
```json
{
  "data": {
    "id": "124",
    "task": {"id": "abc", "name": "Fix bug"},
    "start": "1700000000000",
    "duration": "-1700000000000",
    "description": "Working on it"
  }
}
```

**Plain Output (TSV):**
Headers: `ID	TASK	START	DESCRIPTION`
```
124	abc	1700000000000	Working on it
```

**Human-Readable Output:**
```
Running Timer
  ID: 124
  Task: Fix bug (abc)
  Started: 2023-11-14T12:00:00Z
  Elapsed: 45m 30s
  Description: Working on it
```

**Acceptance Criteria:**
- [ ] Returns currently running timer or "No timer running" message
- [ ] Running entries have negative duration (ClickUp convention)

### US-003: Start Timer

**CLI Command:** `clickup time start --team <team_id> --task <task_id> [--description "..."] [--billable] [--tags <tag1,tag2>]`

**Acceptance Criteria:**
- [ ] Starts a new running timer
- [ ] Optionally associates with a task
- [ ] Returns the started entry

**JSON/Plain/Human Output:** Same as get single entry.

### US-004: Stop Timer

**CLI Command:** `clickup time stop --team <team_id>`

**Acceptance Criteria:**
- [ ] Stops the currently running timer
- [ ] Returns the completed entry with final duration

**JSON/Plain/Human Output:** Same as get single entry.

### US-005: Update Time Entry

**CLI Command:** `clickup time update <entry_id> --team <team_id> [--description "..."] [--duration <ms>] [--start <ms>] [--end <ms>] [--billable] [--tag-action <add|remove>] [--tags <names>]`

**Acceptance Criteria:**
- [ ] Updates specified fields
- [ ] Can modify duration, description, billable status
- [ ] Can add/remove tags via tag_action + tags

### US-006: Delete Time Entry

**CLI Command:** `clickup time delete <entry_id> --team <team_id>`

**JSON Output:** `{"status": "success", "message": "Time entry deleted", "entry_id": "123"}`
**Plain Output (TSV):** Headers: `STATUS	ENTRY_ID` → `success	123`
**Human-Readable:** `Time entry 123 deleted`

### US-007: Time Entry History

**CLI Command:** `clickup time history <entry_id> --team <team_id>`

**JSON Output:**
```json
{
  "data": [
    {"id": "1", "field": "duration", "before": "3600000", "after": "7200000", "date": "1700100000000", "user": {"id": 1, "username": "jeremy"}}
  ]
}
```

**Plain Output (TSV):**
Headers: `ID	FIELD	BEFORE	AFTER	DATE	USER`
```
1	duration	3600000	7200000	1700100000000	jeremy
```

### US-008: Tag Management

**CLI Commands:**
- `clickup time tags --team <team_id>` — list all time entry tags
- `clickup time add-tags --team <team_id> --entry <id1,id2> --tag <name1,name2>` — add tags to entries
- `clickup time remove-tags --team <team_id> --entry <id1,id2> --tag <name1,name2>` — remove tags
- `clickup time rename-tag --team <team_id> --old <name> --new <name>` — rename a tag

**JSON Output (list tags):**
```json
{"data": [{"name": "billable"}, {"name": "internal"}]}
```

**Plain Output (TSV):**
Headers: `NAME`
```
billable
internal
```

**Acceptance Criteria:**
- [ ] List all unique tags used across time entries
- [ ] Add/remove tags operates on multiple entries at once
- [ ] Rename changes the tag name across all entries

## Request/Response Types

```go
type TimeEntryDetail struct {
    ID          json.Number `json:"id"`
    Task        TaskRef     `json:"task"`
    Wid         string      `json:"wid"`
    User        User        `json:"user"`
    Billable    bool        `json:"billable"`
    Start       json.Number `json:"start"`
    End         json.Number `json:"end"`
    Duration    json.Number `json:"duration"`
    Description string      `json:"description"`
    Tags        []Tag       `json:"tags"`
}

type StartTimeEntryRequest struct {
    TaskID      string `json:"tid,omitempty"`
    Description string `json:"description,omitempty"`
    Billable    bool   `json:"billable,omitempty"`
    Tags        []Tag  `json:"tags,omitempty"`
}

type UpdateTimeEntryRequest struct {
    Description string `json:"description,omitempty"`
    Duration    int64  `json:"duration,omitempty"`
    Start       int64  `json:"start,omitempty"`
    End         int64  `json:"end,omitempty"`
    Billable    *bool  `json:"billable,omitempty"`
    TagAction   string `json:"tag_action,omitempty"` // "add" or "remove"
    Tags        []Tag  `json:"tags,omitempty"`
}

type TimeEntryTagsRequest struct {
    TimeEntryIDs []string `json:"time_entry_ids"`
    Tags         []Tag    `json:"tags"`
}

type RenameTimeEntryTagRequest struct {
    Name    string `json:"name"`
    NewName string `json:"new_name"`
}

type TimeEntryHistoryResponse struct {
    Data []TimeEntryHistoryItem `json:"data"`
}

type TimeEntryHistoryItem struct {
    ID     string `json:"id"`
    Field  string `json:"field"`
    Before string `json:"before"`
    After  string `json:"after"`
    Date   string `json:"date"`
    User   User   `json:"user"`
}
```

## Edge Cases

- Running timer has negative duration value (convention: `-start_timestamp`)
- Stopping when no timer is running returns error
- Starting when a timer is already running — ClickUp may auto-stop the existing one
- Tag operations on non-existent entries — partial success possible
- Duration in milliseconds; display should convert to human-readable

## Feedback Loops

### Unit Tests
```go
func TestTimeService_Get(t *testing.T)         { /* single entry */ }
func TestTimeService_Current(t *testing.T)     { /* running timer or empty */ }
func TestTimeService_Start(t *testing.T)       { /* start with task and description */ }
func TestTimeService_Stop(t *testing.T)        { /* stop running timer */ }
func TestTimeService_Update(t *testing.T)      { /* update fields */ }
func TestTimeService_Delete(t *testing.T)      { /* delete entry */ }
func TestTimeService_History(t *testing.T)     { /* entry change history */ }
func TestTimeService_ListTags(t *testing.T)    { /* all tags */ }
func TestTimeService_AddTags(t *testing.T)     { /* add tags to entries */ }
func TestTimeService_RemoveTags(t *testing.T)  { /* remove tags from entries */ }
func TestTimeService_RenameTag(t *testing.T)   { /* rename tag */ }
```

## Technical Requirements

- `TimeService` already exists — extend with new methods
- All time endpoints require `team_id` parameter
- Tag operations use request body with arrays (not query params)
- Rename tag uses PUT method
- Remove tags uses DELETE method with request body
