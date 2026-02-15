# clickup-cli

A CLI tool for ClickUp project management built with Go.

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

## Authentication

### Set API Key

```bash
# Interactive (secure, recommended)
clickup-cli auth set-key --stdin

# From environment variable
echo $CLICKUP_API_KEY | clickup-cli auth set-key --stdin

# From argument (discouraged - exposes in shell history)
clickup-cli auth set-key YOUR_API_KEY
```

### Set Team ID

```bash
clickup-cli auth set-team YOUR_TEAM_ID
```

### Check Status

```bash
clickup-cli auth status
```

### Remove Credentials

```bash
clickup-cli auth remove
```

### Environment Variables

- `CLICKUP_API_KEY` - Override stored API key
- `CLICKUP_TEAM_ID` - Override stored Team ID
- `CLICKUP_CLI_KEYRING_BACKEND` - Force keyring backend (auto/keychain/file)
- `CLICKUP_CLI_KEYRING_PASS` - Password for file backend (headless systems)

## Usage

### Tasks

```bash
# List tasks in a list
clickup-cli tasks list --list=LIST_ID

# Get a specific task
clickup-cli tasks get TASK_ID

# Create a task
clickup-cli tasks create LIST_ID "My Task Name" --priority=2

# Update a task
clickup-cli tasks update TASK_ID --status=done --name="Updated Name"

# Delete a task
clickup-cli tasks delete TASK_ID
```

### Spaces

```bash
clickup-cli spaces list
```

### Lists

```bash
# List by space (shows folders + folderless lists)
clickup-cli lists list --space=SPACE_ID

# List by folder
clickup-cli lists list --folder=FOLDER_ID
```

### Members

```bash
clickup-cli members list
```

### Comments

```bash
# List comments on a task
clickup-cli comments list TASK_ID

# Add a comment
clickup-cli comments add TASK_ID "This is my comment"
```

### Time Tracking

```bash
# Log time (in milliseconds)
clickup-cli time log TASK_ID 3600000

# List time entries
clickup-cli time list TASK_ID
```

### JSON Output

All commands support `--json` for machine-readable output:

```bash
clickup-cli tasks list --list=LIST_ID --json
clickup-cli spaces list --json
```

## Development

### Prerequisites

- Go 1.25+
- Make

### Commands

```bash
make build        # Build binary
make test         # Run tests
make lint         # Run linter
make ci           # Run full CI suite
make tools        # Install dev tools
```

## License

MIT

## Contributing

Contributions are welcome! Please read our contributing guidelines before submitting PRs.
