# AGENTS.md — ClickUp CLI

Project context for autonomous agents (Ralph Wiggum).

## What We're Building

A Go CLI for the ClickUp API. Currently 13/168 endpoints implemented (7.7%). Target: full v2+v3 API parity (155 remaining endpoints across 29 API domains).

## Specs

All specifications live in `specs/`:
- `specs/README.md` — Index with priority tiers (P0/P1/P2)
- `specs/architecture/v3-client-support.md` — Dynamic v2/v3 base URL design
- `specs/features/*.md` — 29 feature specs, one per API domain

## Repository Guidelines

## Project Structure

- `cmd/clickup/`: CLI entrypoint
- `internal/`: implementation packages
  - `cmd/`: Kong CLI commands
  - `api/`: HTTP client with retry/rate limiting
  - `clickup/`: ClickUp API client and types
  - `secrets/`: Keyring-backed credential storage
  - `outfmt/`: JSON/plain output formatting
  - `errfmt/`: User-friendly error formatting
  - `config/`: Platform-aware config paths

## Build, Test, and Development Commands

- `make` / `make build`: build `bin/clickup-cli`
- `make tools`: install pinned dev tools into `.tools/`
- `make fmt` / `make lint` / `make test` / `make ci`: format, lint, test, full local gate
- Hooks: `lefthook install` enables pre-commit checks

## Coding Style & Naming Conventions

- Formatting: `make fmt` (`goimports` local prefix `github.com/builtbyrobben/clickup-cli` + `gofumpt`)
- Output: keep stdout parseable (`--json` / `--plain`); send human hints/progress to stderr

## Testing Guidelines

- Unit tests: stdlib `testing` (and `httptest` where needed)
- All tests should use mocked HTTP servers (no live API calls in CI)

## Commit & Pull Request Guidelines

- Follow Conventional Commits + action-oriented subjects (e.g. `feat(cli): add --verbose flag`)
- Group related changes; avoid bundling unrelated refactors
- PRs should summarize scope, note testing performed, and mention user-facing changes

## Security

- Never commit API keys or credentials
- Use `--stdin` for credential input (avoid shell history leaks)
- Prefer OS keychain backends; use file backend only for headless environments

## Feedback Loops (Required Before Commit)

Run ALL loops. Do NOT commit if any fail.

```bash
make fmt     # Format code (goimports + gofumpt)
make lint    # golangci-lint
make test    # Unit tests
make build   # Build binary
make ci      # All of the above in sequence
```

## Codebase Patterns

### Service Pattern
Each API domain gets a service struct:
```go
type XxxService struct { client *Client }
func (c *Client) Xxx() *XxxService { return &XxxService{client: c} }
```

### Command Pattern
Each domain gets a cmd file (`internal/cmd/<domain>.go`):
```go
type XxxCmd struct {
    List   XxxListCmd   `cmd:"" help:"List xxx"`
    Get    XxxGetCmd    `cmd:"" help:"Get xxx by ID"`
    // ...
}
```

### Output Format Pattern
Every command supports three output modes:
```go
if outfmt.IsJSON(ctx) { return outfmt.WriteJSON(os.Stdout, result) }
if outfmt.IsPlain(ctx) {
    headers := []string{"ID", "NAME", ...}
    rows := [][]string{{item.ID, item.Name, ...}}
    return outfmt.WritePlain(os.Stdout, headers, rows)
}
// Default: human-readable to stderr/stdout
```

### v2 vs v3 Paths
- v2: `/v2/task/{id}` (current, all existing endpoints)
- v3: `/v3/workspaces/{workspace_id}/chat/channels` (new domains)
- See `specs/architecture/v3-client-support.md` for migration design

## Do NOT

- Leave `// TODO` comments or placeholder implementations
- Skip tests for new endpoints
- Use `any` type at API boundaries — define proper structs
- Skip or disable tests
- Leave debug logging
- Make live API calls in tests — use `httptest` mock servers
