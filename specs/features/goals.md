# Goals

## Overview

Implement goals and key results (OKR) CRUD operations.

**Why**: Goals with key results enable OKR-style tracking. CLI access enables automated progress reporting and goal management.

## API Endpoints

| Method | Path | Summary | Operation ID | Version |
|--------|------|---------|--------------|---------|
| GET | /team/{}/goal | Get Goals | GetGoals | v2 |
| GET | /goal/{} | Get Goal | GetGoal | v2 |
| POST | /team/{}/goal | Create Goal | CreateGoal | v2 |
| PUT | /goal/{} | Update Goal | UpdateGoal | v2 |
| DELETE | /goal/{} | Delete Goal | DeleteGoal | v2 |
| POST | /goal/{}/key_result | Create Key Result | CreateKeyResult | v2 |
| PUT | /key_result/{} | Edit Key Result | EditKeyResult | v2 |
| DELETE | /key_result/{} | Delete Key Result | DeleteKeyResult | v2 |

## User Stories

### US-001: List Goals

**CLI Command:** `clickup goals list --team <team_id> [--include-completed]`

**JSON Output:**
```json
{
  "goals": [
    {"id": "g-123", "name": "Q1 Revenue", "description": "Hit $100k", "date_created": "1700000000000", "due_date": "1710000000000", "percent_completed": 65, "key_results": []}
  ]
}
```

**Plain Output (TSV):** Headers: `ID	NAME	PERCENT_COMPLETE	DUE_DATE`
```
g-123	Q1 Revenue	65	1710000000000
```

**Human-Readable:**
```
Goals

  g-123: Q1 Revenue (65% complete)
    Due: 2024-03-09
    Description: Hit $100k
```

### US-002: Get Goal

**CLI Command:** `clickup goals get <goal_id>`

**JSON Output:** Full goal object with key results.

**Plain Output (TSV):** Headers: `ID	NAME	PERCENT_COMPLETE	KEY_RESULT_COUNT`

### US-003: Create Goal

**CLI Command:** `clickup goals create --team <team_id> <name> [--due-date <ms>] [--description "..."] [--owners <user_ids>] [--color <hex>]`

### US-004: Update Goal

**CLI Command:** `clickup goals update <goal_id> [--name "..."] [--description "..."] [--due-date <ms>] [--color <hex>] [--add-owners <ids>] [--remove-owners <ids>]`

### US-005: Delete Goal

**CLI Command:** `clickup goals delete <goal_id>`

### US-006: Create Key Result

**CLI Command:** `clickup goals add-key-result <goal_id> --name <name> --type <type> [--steps-start <n>] [--steps-end <n>] [--unit <unit>] [--owners <user_ids>]`

Key result types: number, currency, boolean, percentage, automatic

**Acceptance Criteria:**
- [ ] Type is required
- [ ] `steps_start` and `steps_end` define the range for number/currency/percentage types
- [ ] Returns the created key result

### US-007: Edit Key Result

**CLI Command:** `clickup goals update-key-result <key_result_id> [--steps-current <n>] [--note "..."]`

### US-008: Delete Key Result

**CLI Command:** `clickup goals delete-key-result <key_result_id>`

## Request/Response Types

```go
type Goal struct {
    ID               string      `json:"id"`
    Name             string      `json:"name"`
    Description      string      `json:"description,omitempty"`
    DateCreated      string      `json:"date_created,omitempty"`
    DueDate          string      `json:"due_date,omitempty"`
    PercentCompleted int         `json:"percent_completed"`
    Color            string      `json:"color,omitempty"`
    KeyResults       []KeyResult `json:"key_results,omitempty"`
    Owners           []User      `json:"owners,omitempty"`
}

type KeyResult struct {
    ID           string `json:"id"`
    Name         string `json:"name"`
    Type         string `json:"type"` // number, currency, boolean, percentage, automatic
    StepsStart   int    `json:"steps_start"`
    StepsEnd     int    `json:"steps_end"`
    StepsCurrent int    `json:"steps_current"`
    Unit         string `json:"unit,omitempty"`
    Note         string `json:"note,omitempty"`
}

type GoalsResponse struct {
    Goals []Goal `json:"goals"`
}

type GoalResponse struct {
    Goal Goal `json:"goal"`
}

type CreateGoalRequest struct {
    Name        string `json:"name"`
    DueDate     int64  `json:"due_date,omitempty"`
    Description string `json:"description,omitempty"`
    Owners      []int  `json:"owners,omitempty"`
    Color       string `json:"color,omitempty"`
}

type UpdateGoalRequest struct {
    Name        string `json:"name,omitempty"`
    Description string `json:"description,omitempty"`
    DueDate     int64  `json:"due_date,omitempty"`
    Color       string `json:"color,omitempty"`
    AddOwners   []int  `json:"add_owners,omitempty"`
    RemOwners   []int  `json:"rem_owners,omitempty"`
}

type CreateKeyResultRequest struct {
    Name       string `json:"name"`
    Type       string `json:"type"`
    StepsStart int    `json:"steps_start,omitempty"`
    StepsEnd   int    `json:"steps_end,omitempty"`
    Unit       string `json:"unit,omitempty"`
    Owners     []int  `json:"owners,omitempty"`
}

type EditKeyResultRequest struct {
    StepsCurrent int    `json:"steps_current,omitempty"`
    Note         string `json:"note,omitempty"`
}
```

## Edge Cases

- `percent_completed` is calculated from key results — read-only
- Boolean key results have steps_start=0, steps_end=1
- Automatic key results derive progress from linked tasks
- Deleting a goal deletes all its key results

## Feedback Loops

### Unit Tests
```go
func TestGoalsService_List(t *testing.T)            { /* list goals */ }
func TestGoalsService_Get(t *testing.T)             { /* get with key results */ }
func TestGoalsService_Create(t *testing.T)          { /* create goal */ }
func TestGoalsService_Update(t *testing.T)          { /* update goal */ }
func TestGoalsService_Delete(t *testing.T)          { /* delete goal */ }
func TestGoalsService_CreateKeyResult(t *testing.T) { /* create KR */ }
func TestGoalsService_EditKeyResult(t *testing.T)   { /* update KR progress */ }
func TestGoalsService_DeleteKeyResult(t *testing.T) { /* delete KR */ }
```
