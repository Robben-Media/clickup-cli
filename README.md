# clickup-cli

Command-line interface for ClickUp project management. Manage tasks, spaces, lists, comments, and time tracking from your terminal.

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap builtbyrobben/tap
brew install clickup-cli
```

### Download Binary

Download the latest release from [GitHub Releases](https://github.com/builtbyrobben/clickup-cli/releases).

### Build from Source

```bash
git clone https://github.com/builtbyrobben/clickup-cli.git
cd clickup-cli
make build
```

## Configuration

clickup-cli requires a ClickUp API key and a Team ID.

**Environment variables (recommended for CI/scripts):**

```bash
export CLICKUP_API_KEY="your-api-key"
export CLICKUP_TEAM_ID="your-team-id"
```

**Keyring + config storage (recommended for interactive use):**

```bash
# Store API key in system keyring
clickup-cli auth set-key --stdin

# Store Team ID in config file
clickup-cli auth set-team YOUR_TEAM_ID
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `CLICKUP_API_KEY` | API key (overrides keyring) |
| `CLICKUP_TEAM_ID` | Team ID (overrides config file) |
| `CLICKUP_CLI_COLOR` | Color output: `auto`, `always`, `never` |
| `CLICKUP_CLI_OUTPUT` | Default output mode: `json`, `plain` |

## Global Flags

| Flag | Description |
|------|-------------|
| `--json` | Output JSON to stdout (best for scripting) |
| `--plain` | Output stable, parseable text (TSV; no colors) |
| `--color` | Color output: `auto`, `always`, `never` |
| `--verbose` | Enable verbose logging |
| `--force` | Skip confirmations for destructive commands |
| `--no-input` | Never prompt; fail instead (useful for CI) |

## Commands

### auth

Manage authentication credentials.

```bash
# Store API key in system keyring
clickup-cli auth set-key --stdin

# Store Team ID in config
clickup-cli auth set-team 1234567

# Check authentication status
clickup-cli auth status

# Remove stored credentials
clickup-cli auth remove
```

### tasks

Create, read, update, and delete tasks.

```bash
# List tasks in a list
clickup-cli tasks list --list LIST_ID

# Filter by status or assignee
clickup-cli tasks list --list LIST_ID --status open
clickup-cli tasks list --list LIST_ID --assignee "John"

# Get a specific task
clickup-cli tasks get TASK_ID

# Create a task
clickup-cli tasks create LIST_ID "Fix login bug"

# Create with priority (1=urgent, 2=high, 3=normal, 4=low)
clickup-cli tasks create LIST_ID "Deploy v2" --priority 1

# Create with due date (unix ms)
clickup-cli tasks create LIST_ID "Write docs" --due 1700000000000

# Update a task
clickup-cli tasks update TASK_ID --status done
clickup-cli tasks update TASK_ID --name "New name" --priority 2

# Delete a task
clickup-cli tasks delete TASK_ID
```

### spaces

List spaces in your team.

```bash
# List all spaces
clickup-cli spaces list

# Output as JSON
clickup-cli spaces list --json
```

### lists

List ClickUp lists within spaces or folders.

```bash
# List by space (shows folders + folderless lists)
clickup-cli lists list --space SPACE_ID

# List by folder
clickup-cli lists list --folder FOLDER_ID

# Output as JSON
clickup-cli lists list --space SPACE_ID --json
```

### members

List team members.

```bash
# List all team members
clickup-cli members list

# Output as JSON
clickup-cli members list --json
```

### comments

Manage task comments.

```bash
# List comments on a task
clickup-cli comments list TASK_ID

# Add a comment to a task
clickup-cli comments add TASK_ID "Looks good, deploying now"

# Output as JSON
clickup-cli comments list TASK_ID --json
```

### time

Track time on tasks.

```bash
# Log time to a task (duration in milliseconds)
clickup-cli time log TASK_ID 3600000

# List time entries for a task
clickup-cli time list TASK_ID

# Output as JSON
clickup-cli time list TASK_ID --json
```

### version

Print version information.

```bash
clickup-cli version
```

## License

MIT
