# ClickUp CLI — API Parity Specs

Specifications for achieving full ClickUp API parity (155 missing endpoints across v2 and v3).

**Source**: Gap analysis from `docs/gaps/clickup-api-gap-2026-02-22/`

## Architecture

| Spec | Description |
|------|-------------|
| [v3-client-support](architecture/v3-client-support.md) | Dynamic v2/v3 base URL routing, Patch method, multipart upload, workspace ID config |

## Feature Specs — Core Domains (partially implemented)

| Spec | Endpoints | Missing | Version | Priority |
|------|-----------|---------|---------|----------|
| [tasks](features/tasks.md) | 11 | 6 | v2+v3 | P0 |
| [comments](features/comments.md) | 12 | 9 | v2+v3 | P0 |
| [folders](features/folders.md) | 6 | 5 | v2 | P0 |
| [lists](features/lists.md) | 11 | 9 | v2 | P0 |
| [spaces](features/spaces.md) | 5 | 4 | v2 | P0 |

## Feature Specs — Time Tracking

| Spec | Endpoints | Missing | Version | Priority |
|------|-----------|---------|---------|----------|
| [time-tracking](features/time-tracking.md) | 13 | 11 | v2 | P0 |
| [time-tracking-legacy](features/time-tracking-legacy.md) | 4 | 4 | v2 | P2 |

## Feature Specs — Task Extensions

| Spec | Endpoints | Missing | Version | Priority |
|------|-----------|---------|---------|----------|
| [tags](features/tags.md) | 6 | 6 | v2 | P1 |
| [task-checklists](features/task-checklists.md) | 6 | 6 | v2 | P1 |
| [task-relationships](features/task-relationships.md) | 4 | 4 | v2 | P1 |
| [custom-fields](features/custom-fields.md) | 6 | 6 | v2 | P1 |
| [custom-task-types](features/custom-task-types.md) | 1 | 1 | v2 | P2 |
| [templates](features/templates.md) | 1 | 1 | v2 | P2 |

## Feature Specs — Workspace Management

| Spec | Endpoints | Missing | Version | Priority |
|------|-----------|---------|---------|----------|
| [workspaces](features/workspaces.md) | 3 | 3 | v2 | P0 |
| [authorization](features/authorization.md) | 2 | 2 | v2 | P0 |
| [users](features/users.md) | 4 | 4 | v2 | P1 |
| [user-groups](features/user-groups.md) | 4 | 4 | v2 | P2 |
| [members](features/members.md) | 3 | 2 | v2 | P1 |
| [roles](features/roles.md) | 1 | 1 | v2 | P2 |
| [guests](features/guests.md) | 10 | 10 | v2 | P2 |
| [shared-hierarchy](features/shared-hierarchy.md) | 1 | 1 | v2 | P2 |

## Feature Specs — Views, Webhooks, Goals

| Spec | Endpoints | Missing | Version | Priority |
|------|-----------|---------|---------|----------|
| [views](features/views.md) | 12 | 12 | v2 | P1 |
| [webhooks](features/webhooks.md) | 4 | 4 | v2 | P1 |
| [goals](features/goals.md) | 8 | 8 | v2 | P1 |

## Feature Specs — v3 Domains

| Spec | Endpoints | Missing | Version | Priority |
|------|-----------|---------|---------|----------|
| [chat](features/chat.md) | 19 | 19 | v3 | P1 |
| [docs](features/docs.md) | 8 | 8 | v3 | P1 |
| [attachments](features/attachments.md) | 3 | 3 | v2+v3 | P1 |
| [auditlogs](features/auditlogs.md) | 1 | 1 | v3 | P2 |
| [acls](features/acls.md) | 1 | 1 | v3 | P2 |

## Implementation Order

### Phase 1: Foundation (P0) — 39 endpoints
1. **Architecture**: v3 base URL migration (change base, add `/v2/` prefix to all existing paths)
2. **Workspaces**: List workspaces (needed for workspace ID auto-detection)
3. **Authorization**: whoami (key validation)
4. **Core CRUD**: Tasks (search + time-in-status), Spaces, Folders, Lists, Comments
5. **Time Tracking**: Start/stop timer, entry CRUD, tags

### Phase 2: Extensions (P1) — 84 endpoints
6. **Task Extensions**: Tags, Checklists, Relationships, Custom Fields
7. **Views**: Full CRUD at all hierarchy levels
8. **Webhooks**: CRUD
9. **Goals**: Goals + Key Results CRUD
10. **Members/Users**: List/task members, user CRUD
11. **v3 Domains**: Chat, Docs, Attachments

### Phase 3: Completeness (P2) — 32 endpoints
12. **Remaining**: User Groups, Roles, Guests, Shared Hierarchy, Templates, Custom Task Types, Time Tracking Legacy, Audit Logs, ACLs

## Conventions

All specs follow these patterns:
- **CLI**: `clickup <domain> <verb> [args] [--flags]`
- **Output**: Every command supports `--json`, `--plain` (TSV), and human-readable (default)
- **Service pattern**: `XxxService` struct with methods on `clickup.Client`
- **Types**: Defined in `internal/clickup/types.go`
- **Commands**: Defined in `internal/cmd/<domain>.go`
- **Tests**: `internal/clickup/client_test.go` and `internal/cmd/<domain>_test.go`
