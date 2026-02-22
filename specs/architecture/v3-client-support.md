# v3 API Client Support Architecture

## Overview

The ClickUp CLI currently supports only v2 API endpoints via a hardcoded base URL (`https://api.clickup.com/api/v2`). To support v3 endpoints (Chat, Docs, Attachments, Audit Logs, ACLs), the client needs dynamic base URL routing so service methods can target either `/api/v2/...` or `/api/v3/workspaces/{workspace_id}/...` paths.

**Why**: 33 of 168 total API operations (20%) are v3-only, including the entire Chat and Docs domains. Without v3 support, the CLI cannot achieve full API parity.

## System Components

1. **`api.Client`** (`internal/api/client.go`)
   - Change `baseURL` from `https://api.clickup.com/api/v2` to `https://api.clickup.com/api`
   - All existing service methods prepend `/v2/` to their paths
   - New v3 service methods prepend `/v3/workspaces/{workspace_id}/` to their paths
   - Add `Patch()` method for v3 PATCH endpoints
   - Add `PostMultipart()` method for file upload endpoints (Attachments)

2. **`clickup.Client`** (`internal/clickup/client.go`)
   - Add `workspaceID` field (configurable, stored in config or passed via `--workspace` flag)
   - Add helper `v3Path(path string) string` that builds `/v3/workspaces/{workspaceID}/...`
   - Add service accessors for new v3 domains: `Chat()`, `Docs()`, `Attachments()`, `AuditLogs()`, `Acls()`

3. **Config** (`internal/cmd/helpers.go`)
   - Add `CLICKUP_WORKSPACE_ID` environment variable / config key
   - The workspace ID is required for v3 calls; error early if missing
   - Can be auto-detected via `GET /team` (Get Authorized Workspaces) if not configured

## Architecture Diagram

```
┌──────────────────────────────────────────────────────┐
│                    CLI Commands                       │
│  internal/cmd/*.go                                   │
│  (tasks.go, chat.go, docs.go, ...)                   │
└────────────────────┬─────────────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────┐
│               clickup.Client                          │
│  internal/clickup/client.go                          │
│                                                      │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────────┐ │
│  │TasksService │  │ChatService  │  │DocsService   │ │
│  │(v2 paths)   │  │(v3 paths)   │  │(v3 paths)    │ │
│  └──────┬──────┘  └──────┬──────┘  └──────┬───────┘ │
│         │                │                │          │
│    /v2/task/...    /v3/workspaces/    /v3/workspaces/ │
│                    {wid}/chat/...    {wid}/docs/...   │
└────────────────────┬─────────────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────┐
│                  api.Client                           │
│  internal/api/client.go                              │
│  baseURL: https://api.clickup.com/api                │
│                                                      │
│  Methods: Get, Post, Put, Delete, Patch, PostMultipart│
│  URL = baseURL + path (path includes /v2/ or /v3/)   │
└──────────────────────────────────────────────────────┘
```

## Key Interfaces

```go
// internal/api/client.go — new methods

// Patch sends a PATCH request (used by v3 update endpoints).
func (c *Client) Patch(ctx context.Context, path string, body, result any) error {
    return c.doJSON(ctx, Request{Method: http.MethodPatch, Path: path, Body: body}, result)
}

// PostMultipart sends a multipart/form-data POST request (file uploads).
func (c *Client) PostMultipart(ctx context.Context, path string, fieldName string, reader io.Reader, fileName string, result any) error {
    // Build multipart body, set Content-Type header with boundary
}
```

```go
// internal/clickup/client.go — v3 helper

// v3Path builds a v3 API path with the workspace ID prefix.
func (c *Client) v3Path(path string) string {
    return fmt.Sprintf("/v3/workspaces/%s%s", c.workspaceID, path)
}

// Example usage in ChatService:
func (s *ChatService) ListChannels(ctx context.Context) (*ChatChannelsResponse, error) {
    var result ChatChannelsResponse
    path := s.client.v3Path("/chat/channels")
    if err := s.client.Get(ctx, path, &result); err != nil {
        return nil, fmt.Errorf("list chat channels: %w", err)
    }
    return &result, nil
}
```

## Data Flow

### v2 Request Flow (existing, path change only)
1. CLI command calls service method (e.g., `client.Tasks().Get(ctx, "abc123")`)
2. Service builds path: `/v2/task/abc123`
3. `api.Client.Get()` constructs URL: `https://api.clickup.com/api` + `/v2/task/abc123`
4. Response decoded into typed struct

### v3 Request Flow (new)
1. CLI command calls service method (e.g., `client.Chat().ListChannels(ctx)`)
2. Service calls `c.v3Path("/chat/channels")` → `/v3/workspaces/12345/chat/channels`
3. `api.Client.Get()` constructs URL: `https://api.clickup.com/api` + `/v3/workspaces/12345/chat/channels`
4. Response decoded into typed struct

## Migration Plan

### Phase 1: Base URL Change (non-breaking)
1. Change `defaultBaseURL` in `clickup/client.go` from `https://api.clickup.com/api/v2` to `https://api.clickup.com/api`
2. Update ALL existing service method paths to include `/v2/` prefix:
   - `/task/{id}` → `/v2/task/{id}`
   - `/list/{id}/task` → `/v2/list/{id}/task`
   - `/team/{id}/space` → `/v2/team/{id}/space`
   - etc.
3. Run all existing tests to verify nothing breaks

### Phase 2: v3 Infrastructure
1. Add `workspaceID` field to `clickup.Client`
2. Add `v3Path()` helper
3. Add `Patch()` method to `api.Client`
4. Add `--workspace` global flag to CLI
5. Add `CLICKUP_WORKSPACE_ID` env var support
6. Add workspace auto-detection fallback

### Phase 3: v3 Service Implementation
1. Implement v3 service types (Chat, Docs, Attachments, AuditLogs, Acls)
2. Each service uses `v3Path()` for all API calls
3. Add CLI commands for each v3 domain

## Error Handling

- **Missing workspace ID**: Return clear error `"workspace ID required for v3 API; set CLICKUP_WORKSPACE_ID or use --workspace flag"`
- **v3 auth errors**: Same `APIError` struct, v3 may return different error shapes — handle gracefully
- **PATCH method**: v3 uses PATCH for updates (not PUT) — ensure `api.Client.Patch()` handles same error patterns

## Feedback Loops

### Migration Contract Tests
```go
// internal/api/client_test.go
func TestBaseURLConstruction(t *testing.T) {
    c := NewClient("test-key", WithBaseURL("https://api.clickup.com/api"))

    tests := []struct {
        name string
        path string
        want string
    }{
        {"v2 path", "/v2/task/abc", "https://api.clickup.com/api/v2/task/abc"},
        {"v3 path", "/v3/workspaces/123/chat/channels", "https://api.clickup.com/api/v3/workspaces/123/chat/channels"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Verify URL construction
        })
    }
}
```

### v3Path Helper Tests
```go
// internal/clickup/client_test.go
func TestV3Path(t *testing.T) {
    c := &Client{workspaceID: "12345"}

    got := c.v3Path("/chat/channels")
    want := "/v3/workspaces/12345/chat/channels"
    if got != want {
        t.Errorf("v3Path() = %q, want %q", got, want)
    }
}

func TestV3PathMissingWorkspace(t *testing.T) {
    c := &Client{}
    // Should panic or return error
}
```

### Patch Method Tests
```go
func TestClientPatch(t *testing.T) {
    // Verify PATCH method sends correct HTTP method
    // Verify request body is JSON-encoded
    // Verify response is decoded
    // Verify error handling matches existing patterns
}
```

## Constraints

- Backward compatible: All existing v2 commands must work identically after migration
- No breaking config changes: Existing `CLICKUP_API_KEY` config unchanged
- Workspace ID is optional for v2-only usage (not required until a v3 command is invoked)
- The `Patch()` method follows the same pattern as `Put()` for consistency
