# Time Tracking (Legacy)

## Overview

Implement the legacy time tracking endpoints that operate at the task level with intervals. These are separate from the main Time Tracking API (workspace-level entries).

**Why**: Some integrations still use legacy time tracking. Full parity requires both systems.

## API Endpoints

| Method | Path | Summary | Operation ID | Version |
|--------|------|---------|--------------|---------|
| GET | /task/{}/time | Get tracked time | Gettrackedtime | v2 |
| POST | /task/{}/time | Track time | Tracktime | v2 |
| PUT | /task/{}/time/{} | Edit time tracked | Edittimetracked | v2 |
| DELETE | /task/{}/time/{} | Delete time tracked | Deletetimetracked | v2 |

## User Stories

### US-001: Get Tracked Time

**CLI Command:** `clickup time-legacy list <task_id> [--custom-task-ids --team <team_id>]`

**JSON Output:**
```json
{
  "data": [
    {
      "id": "567",
      "start": 1567780450202,
      "end": 1508369194377,
      "time": 8640000,
      "source": "clickup",
      "date_added": "1567780450202"
    }
  ]
}
```

**Plain Output (TSV):**
Headers: `ID	START	END	TIME_MS	SOURCE`
```
567	1567780450202	1508369194377	8640000	clickup
```

**Human-Readable Output:**
```
Tracked Time for task abc123

ID: 567
  Duration: 2h 24m
  Start: 2019-09-06T12:00:00Z
  End: 2017-10-18T18:00:00Z
  Source: clickup
```

**Acceptance Criteria:**
- [ ] Returns all time intervals for the task
- [ ] Supports `custom_task_ids` and `team_id` query params

### US-002: Track Time

**CLI Command:** `clickup time-legacy track <task_id> --time <ms> [--start <ms>] [--end <ms>]`

**Acceptance Criteria:**
- [ ] Creates a time interval on the task
- [ ] Accepts total `time` OR `start`+`end` pair
- [ ] Returns the created interval ID

**JSON Output:**
```json
{"id": "567"}
```

**Plain Output (TSV):**
Headers: `ID`
```
567
```

### US-003: Edit Tracked Time

**CLI Command:** `clickup time-legacy update <task_id> <interval_id> [--time <ms>] [--start <ms>] [--end <ms>]`

**Acceptance Criteria:**
- [ ] Updates the specified interval
- [ ] Returns success

### US-004: Delete Tracked Time

**CLI Command:** `clickup time-legacy delete <task_id> <interval_id>`

**JSON Output:** `{"status": "success", "message": "Time interval deleted"}`
**Plain Output (TSV):** Headers: `STATUS	INTERVAL_ID` → `success	567`
**Human-Readable:** `Time interval 567 deleted`

## Request/Response Types

```go
type LegacyTimeInterval struct {
    ID        string `json:"id"`
    Start     int64  `json:"start"`
    End       int64  `json:"end"`
    Time      int64  `json:"time"`
    Source    string `json:"source,omitempty"`
    DateAdded string `json:"date_added,omitempty"`
}

type LegacyTimeResponse struct {
    Data []LegacyTimeInterval `json:"data"`
}

type TrackTimeRequest struct {
    Start int64 `json:"start,omitempty"`
    End   int64 `json:"end,omitempty"`
    Time  int64 `json:"time"`
}

type EditTimeRequest struct {
    Start int64 `json:"start,omitempty"`
    End   int64 `json:"end,omitempty"`
    Time  int64 `json:"time,omitempty"`
}
```

## Edge Cases

- `time` and `start/end` are partially redundant — API accepts either
- Legacy intervals don't have user association in the response
- `custom_task_ids` param changes task ID interpretation
- ClickUp recommends using modern Time Tracking API instead

## Feedback Loops

### Unit Tests
```go
func TestLegacyTimeService_List(t *testing.T)   { /* list intervals */ }
func TestLegacyTimeService_Track(t *testing.T)  { /* create interval */ }
func TestLegacyTimeService_Edit(t *testing.T)   { /* edit interval */ }
func TestLegacyTimeService_Delete(t *testing.T) { /* delete interval */ }
```

## Technical Requirements

- New `LegacyTimeService` or add to existing `TimeService` with `Legacy` prefix methods
- These endpoints use `/task/{task_id}/time` path (task-level), not `/team/{team_id}/time_entries`
- POST returns only `{"id": "..."}` — minimal response
