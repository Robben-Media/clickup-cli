# Implementation Plan

Generated: 2026-02-21
Last Updated: 2026-02-21

## Summary

Achieve full ClickUp API parity by implementing 155 missing endpoints across v2 and v3 APIs. The CLI currently has 13/168 endpoints (7.7%). This plan organizes 67 tasks across 8 phases, prioritizing foundation changes and P0 endpoints first.

## Architecture Decisions

- **v2/v3 base URL routing**: Change base URL from `https://api.clickup.com/api/v2` to `https://api.clickup.com/api`, with all paths including version prefix
- **Service pattern**: Each API domain gets a `XxxService` struct with methods on `clickup.Client`
- **Output format pattern**: Every command supports `--json`, `--plain` (TSV), and human-readable output
- **Workspace ID config**: v3 endpoints require workspace ID via `--workspace` flag or `CLICKUP_WORKSPACE_ID` env var

## Blockers / Questions

1. None - all specs are complete with sufficient detail for implementation

---

## Phase 1: Foundation Architecture
> Branch: `phase/1-foundation` | PR → fix/plain-output

### 1. Migrate to Dynamic v2/v3 Base URL
- **Status**: complete
- **Depends on**: none
- **Spec**: specs/architecture/v3-client-support.md
- **Description**: 
  - Change `defaultBaseURL` in `clickup/client.go` from `https://api.clickup.com/api/v2` to `https://api.clickup.com/api`
  - Update ALL existing service method paths to include `/v2/` prefix (e.g., `/task/{id}` → `/v2/task/{id}`)
  - Add `workspaceID` field to `Client` struct
  - Add `v3Path(path string) string` helper method
  - Add `Patch()` method to `api.Client` for v3 PATCH endpoints
  - Add `PostMultipart()` method to `api.Client` for file uploads
- **Files**:
  - `internal/clickup/client.go` — modify (base URL, workspaceID, v3Path helper)
  - `internal/api/client.go` — modify (add Patch, PostMultipart methods)
- **Verification**: `make ci` passes; existing tests verify backward compatibility

### 2. Add Workspace ID Configuration
- **Status**: complete
- **Depends on**: 1
- **Spec**: specs/architecture/v3-client-support.md
- **Description**:
  - Add `--workspace` global flag to CLI root command
  - Add `CLICKUP_WORKSPACE_ID` environment variable support
  - Add workspace ID to client initialization
  - Add error handling for missing workspace ID on v3 calls
- **Files**:
  - `internal/cmd/root.go` — modify (add global --workspace flag)
  - `internal/cmd/helpers.go` — modify (add env var lookup)
  - `internal/clickup/client.go` — modify (workspaceID field)
- **Verification**: `make ci` passes; test with/without workspace ID

### 3. Add Contract Tests for v3 Infrastructure
- **Status**: complete
- **Depends on**: 2
- **Spec**: specs/architecture/v3-client-support.md
- **Description**:
  - Add tests for base URL construction (v2 and v3 paths)
  - Add tests for `v3Path()` helper
  - Add tests for `Patch()` method
  - Add tests for `PostMultipart()` method
- **Files**:
  - `internal/api/client_test.go` — modify (add Patch, PostMultipart tests)
  - `internal/clickup/client_test.go` — modify (add v3Path tests)
- **Verification**: `make test` passes with new tests

---

## Phase 2: Workspaces & Authorization (P0)
> Branch: `phase/2-workspaces-auth` | PR → fix/plain-output

### 4. Implement Workspaces Service
- **Status**: complete
- **Depends on**: 1
- **Spec**: specs/features/workspaces.md
- **Description**:
  - Create `WorkspacesService` with List, Plan, Seats methods
  - Implement `GET /v2/team` (list workspaces)
  - Implement `GET /v2/team/{team_id}/plan` (get plan)
  - Implement `GET /v2/team/{team_id}/seats` (get seats)
  - Add CLI commands: `workspaces list`, `workspaces plan`, `workspaces seats`
- **Files**:
  - `internal/clickup/client.go` — modify (add Workspaces service accessor)
  - `internal/clickup/types.go` — modify (add Workspace types)
  - `internal/cmd/workspaces.go` — create (CLI commands)
- **Verification**: `make ci` passes; `clickup workspaces list` works

### 5. Implement Authorization Service
- **Status**: complete
- **Depends on**: 1
- **Spec**: specs/features/authorization.md
- **Description**:
  - Create `AuthService` with Whoami, Token methods
  - Implement `GET /v2/user` (get authorized user)
  - Implement `POST /v2/oauth/token` (OAuth token exchange)
  - Add CLI commands: `auth whoami`, `auth token`
- **Files**:
  - `internal/clickup/client.go` — modify (add Auth service accessor)
  - `internal/clickup/types.go` — modify (add Auth types)
  - `internal/cmd/auth.go` — modify (add whoami command)
- **Verification**: `make ci` passes; `clickup auth whoami` works

---

## Phase 3: Core CRUD Extensions (P0)
> Branch: `phase/3-core-crud` | PR → fix/plain-output

### 6. Extend Spaces Service (Get, Create, Update, Delete)
- **Status**: pending
- **Depends on**: 1
- **Spec**: specs/features/spaces.md
- **Description**:
  - Add Get, Create, Update, Delete methods to existing SpacesService
  - Implement `GET /v2/space/{space_id}`
  - Implement `POST /v2/team/{team_id}/space`
  - Implement `PUT /v2/space/{space_id}`
  - Implement `DELETE /v2/space/{space_id}`
  - Add CLI commands: `spaces get`, `spaces create`, `spaces update`, `spaces delete`
- **Files**:
  - `internal/clickup/client.go` — modify (extend SpacesService)
  - `internal/clickup/types.go` — modify (add SpaceDetail, CreateSpaceRequest, etc.)
  - `internal/cmd/spaces.go` — modify (add commands)
- **Verification**: `make ci` passes; CRUD commands work

### 7. Create Folders Service (Full CRUD)
- **Status**: pending
- **Depends on**: 1
- **Spec**: specs/features/folders.md
- **Description**:
  - Create new `FoldersService` (separate from ListsService.ListFolders)
  - Implement `GET /v2/folder/{folder_id}`
  - Implement `POST /v2/space/{space_id}/folder`
  - Implement `PUT /v2/folder/{folder_id}`
  - Implement `DELETE /v2/folder/{folder_id}`
  - Implement `POST /v2/space/{space_id}/folder_template/{template_id}`
  - Add CLI commands: `folders get`, `folders create`, `folders update`, `folders delete`, `folders from-template`
- **Files**:
  - `internal/clickup/client.go` — modify (add Folders service accessor)
  - `internal/clickup/types.go` — modify (add FolderDetail, CreateFolderRequest, etc.)
  - `internal/cmd/folders.go` — create (CLI commands)
- **Verification**: `make ci` passes; all folder commands work

### 8. Extend Lists Service (Get, Create, Update, Delete)
- **Status**: pending
- **Depends on**: 1
- **Spec**: specs/features/lists.md
- **Description**:
  - Add Get, Create, Update, Delete, AddTask, RemoveTask methods to ListsService
  - Implement `GET /v2/list/{list_id}`
  - Implement `POST /v2/folder/{folder_id}/list`
  - Implement `POST /v2/space/{space_id}/list`
  - Implement `PUT /v2/list/{list_id}`
  - Implement `DELETE /v2/list/{list_id}`
  - Implement `POST /v2/list/{list_id}/task/{task_id}`
  - Implement `DELETE /v2/list/{list_id}/task/{task_id}`
  - Add CLI commands: `lists get`, `lists create`, `lists update`, `lists delete`, `lists add-task`, `lists remove-task`
- **Files**:
  - `internal/clickup/client.go` — modify (extend ListsService)
  - `internal/clickup/types.go` — modify (add ListDetail, CreateListRequest, etc.)
  - `internal/cmd/lists.go` — modify (add commands)
- **Verification**: `make ci` passes; all list commands work

### 9. Extend Tasks Service (Search, Time-in-Status, Merge, Move)
- **Status**: pending
- **Depends on**: 1
- **Spec**: specs/features/tasks.md
- **Description**:
  - Add Search, TimeInStatus, BulkTimeInStatus, Merge, Move methods to TasksService
  - Implement `GET /v2/team/{team_id}/task` (filtered search)
  - Implement `GET /v2/task/{task_id}/time_in_status`
  - Implement `GET /v2/task/bulk_time_in_status/task_ids` (bulk)
  - Implement `POST /v2/task/{task_id}/merge`
  - Implement `PUT /v3/workspaces/{wid}/tasks/{task_id}/home_list/{list_id}` (move - v3)
  - Implement `POST /v2/list/{list_id}/taskTemplate/{template_id}` (from template)
  - Add CLI commands: `tasks search`, `tasks time-in-status`, `tasks bulk-time-in-status`, `tasks merge`, `tasks move`, `tasks from-template`
- **Files**:
  - `internal/clickup/client.go` — modify (extend TasksService)
  - `internal/clickup/types.go` — modify (add FilteredTeamTasksParams, TimeInStatusResponse, etc.)
  - `internal/cmd/tasks.go` — modify (add commands)
- **Verification**: `make ci` passes; all task commands work

### 10. Extend Comments Service (Full CRUD + Threaded + List/View)
- **Status**: pending
- **Depends on**: 1
- **Spec**: specs/features/comments.md
- **Description**:
  - Add Delete, Update, Replies, Reply, ListComments, AddList, ViewComments, AddView, Subtypes methods
  - Implement `DELETE /v2/comment/{comment_id}`
  - Implement `PUT /v2/comment/{comment_id}`
  - Implement `GET /v2/comment/{comment_id}/reply`
  - Implement `POST /v2/comment/{comment_id}/reply`
  - Implement `GET /v2/list/{list_id}/comment`
  - Implement `POST /v2/list/{list_id}/comment`
  - Implement `GET /v2/view/{view_id}/comment`
  - Implement `POST /v2/view/{view_id}/comment`
  - Implement `GET /v3/workspaces/{wid}/comments/types/{type_id}/subtypes` (v3)
  - Add CLI commands: `comments delete`, `comments update`, `comments replies`, `comments reply`, `comments list-comments`, `comments add-list`, `comments view-comments`, `comments add-view`, `comments subtypes`
- **Files**:
  - `internal/clickup/client.go` — modify (extend CommentsService)
  - `internal/clickup/types.go` — modify (add UpdateCommentRequest, ThreadedCommentsResponse, etc.)
  - `internal/cmd/comments.go` — modify (add commands)
- **Verification**: `make ci` passes; all comment commands work

---

## Phase 4: Time Tracking (P0)
> Branch: `phase/4-time-tracking` | PR → fix/plain-output

### 11. Extend Time Service (Timer + CRUD + Tags + History)
- **Status**: pending
- **Depends on**: 1
- **Spec**: specs/features/time-tracking.md
- **Description**:
  - Add Get, Current, Start, Stop, Update, Delete, History, Tags methods to TimeService
  - Implement `GET /v2/team/{team_id}/time_entries/{entry_id}`
  - Implement `GET /v2/team/{team_id}/time_entries/current`
  - Implement `GET /v2/team/{team_id}/time_entries/{entry_id}/history`
  - Implement `GET /v2/team/{team_id}/time_entries/tags`
  - Implement `POST /v2/team/{team_id}/time_entries/start`
  - Implement `POST /v2/team/{team_id}/time_entries/stop`
  - Implement `POST /v2/team/{team_id}/time_entries/tags`
  - Implement `PUT /v2/team/{team_id}/time_entries/{entry_id}`
  - Implement `PUT /v2/team/{team_id}/time_entries/tags`
  - Implement `DELETE /v2/team/{team_id}/time_entries/{entry_id}`
  - Implement `DELETE /v2/team/{team_id}/time_entries/tags`
  - Add CLI commands: `time get`, `time current`, `time start`, `time stop`, `time update`, `time delete`, `time history`, `time tags`, `time add-tags`, `time remove-tags`, `time rename-tag`
- **Files**:
  - `internal/clickup/client.go` — modify (extend TimeService)
  - `internal/clickup/types.go` — modify (add TimeEntryDetail, StartTimeEntryRequest, etc.)
  - `internal/cmd/time.go` — modify (add commands)
- **Verification**: `make ci` passes; timer start/stop works

---

## Phase 5: Task Extensions (P1)
> Branch: `phase/5-task-extensions` | PR → fix/plain-output

### 12. Create Tags Service
- **Status**: pending
- **Depends on**: 1
- **Spec**: specs/features/tags.md
- **Description**:
  - Create `TagsService` with List, Create, Update, Delete, AddToTask, RemoveFromTask methods
  - Implement `GET /v2/space/{space_id}/tag`
  - Implement `POST /v2/space/{space_id}/tag`
  - Implement `PUT /v2/space/{space_id}/tag/{tag_name}`
  - Implement `DELETE /v2/space/{space_id}/tag/{tag_name}`
  - Implement `POST /v2/task/{task_id}/tag/{tag_name}`
  - Implement `DELETE /v2/task/{task_id}/tag/{tag_name}`
  - Add CLI commands: `tags list`, `tags create`, `tags update`, `tags delete`, `tags add`, `tags remove`
- **Files**:
  - `internal/clickup/client.go` — modify (add Tags service accessor)
  - `internal/clickup/types.go` — modify (add SpaceTag types)
  - `internal/cmd/tags.go` — create (CLI commands)
- **Verification**: `make ci` passes; tag CRUD works

### 13. Create Checklists Service
- **Status**: pending
- **Depends on**: 1
- **Spec**: specs/features/task-checklists.md
- **Description**:
  - Create `ChecklistsService` with Create, Update, Delete, AddItem, UpdateItem, DeleteItem methods
  - Implement `POST /v2/task/{task_id}/checklist`
  - Implement `PUT /v2/checklist/{checklist_id}`
  - Implement `DELETE /v2/checklist/{checklist_id}`
  - Implement `POST /v2/checklist/{checklist_id}/checklist_item`
  - Implement `PUT /v2/checklist/{checklist_id}/checklist_item/{item_id}`
  - Implement `DELETE /v2/checklist/{checklist_id}/checklist_item/{item_id}`
  - Add CLI commands: `checklists create`, `checklists update`, `checklists delete`, `checklists add-item`, `checklists update-item`, `checklists delete-item`
- **Files**:
  - `internal/clickup/client.go` — modify (add Checklists service accessor)
  - `internal/clickup/types.go` — modify (add Checklist types)
  - `internal/cmd/checklists.go` — create (CLI commands)
- **Verification**: `make ci` passes; checklist CRUD works

### 14. Create Relationships Service
- **Status**: pending
- **Depends on**: 1
- **Spec**: specs/features/task-relationships.md
- **Description**:
  - Create `RelationshipsService` with AddDependency, RemoveDependency, Link, Unlink methods
  - Implement `POST /v2/task/{task_id}/dependency`
  - Implement `DELETE /v2/task/{task_id}/dependency`
  - Implement `POST /v2/task/{task_id}/link/{link_task_id}`
  - Implement `DELETE /v2/task/{task_id}/link/{link_task_id}`
  - Add CLI commands: `relationships add-dep`, `relationships remove-dep`, `relationships link`, `relationships unlink`
- **Files**:
  - `internal/clickup/client.go` — modify (add Relationships service accessor)
  - `internal/clickup/types.go` — modify (add Relationship types)
  - `internal/cmd/relationships.go` — create (CLI commands)
- **Verification**: `make ci` passes; dependency/link CRUD works

### 15. Create Custom Fields Service
- **Status**: pending
- **Depends on**: 1
- **Spec**: specs/features/custom-fields.md
- **Description**:
  - Create `CustomFieldsService` with List, Set, Remove methods
  - Implement `GET /v2/list/{list_id}/field`
  - Implement `GET /v2/folder/{folder_id}/field`
  - Implement `GET /v2/space/{space_id}/field`
  - Implement `GET /v2/team/{team_id}/field`
  - Implement `POST /v2/task/{task_id}/field/{field_id}`
  - Implement `DELETE /v2/task/{task_id}/field/{field_id}`
  - Add CLI commands: `fields list`, `fields set`, `fields remove`
- **Files**:
  - `internal/clickup/client.go` — modify (add CustomFields service accessor)
  - `internal/clickup/types.go` — modify (add CustomField types)
  - `internal/cmd/fields.go` — create (CLI commands)
- **Verification**: `make ci` passes; field CRUD works

---

## Phase 6: Views, Webhooks, Goals (P1)
> Branch: `phase/6-views-webhooks-goals` | PR → fix/plain-output

### 16. Create Views Service
- **Status**: pending
- **Depends on**: 1
- **Spec**: specs/features/views.md
- **Description**:
  - Create `ViewsService` with List, Get, Tasks, Create, Update, Delete methods
  - Implement `GET /v2/team/{team_id}/view`
  - Implement `GET /v2/space/{space_id}/view`
  - Implement `GET /v2/folder/{folder_id}/view`
  - Implement `GET /v2/list/{list_id}/view`
  - Implement `GET /v2/view/{view_id}`
  - Implement `GET /v2/view/{view_id}/task`
  - Implement `POST /v2/team/{team_id}/view`
  - Implement `POST /v2/space/{space_id}/view`
  - Implement `POST /v2/folder/{folder_id}/view`
  - Implement `POST /v2/list/{list_id}/view`
  - Implement `PUT /v2/view/{view_id}`
  - Implement `DELETE /v2/view/{view_id}`
  - Add CLI commands: `views list`, `views get`, `views tasks`, `views create`, `views update`, `views delete`
- **Files**:
  - `internal/clickup/client.go` — modify (add Views service accessor)
  - `internal/clickup/types.go` — modify (add View types)
  - `internal/cmd/views.go` — create (CLI commands)
- **Verification**: `make ci` passes; view CRUD works

### 17. Create Webhooks Service
- **Status**: pending
- **Depends on**: 1
- **Spec**: specs/features/webhooks.md
- **Description**:
  - Create `WebhooksService` with List, Create, Update, Delete methods
  - Implement `GET /v2/team/{team_id}/webhook`
  - Implement `POST /v2/team/{team_id}/webhook`
  - Implement `PUT /v2/webhook/{webhook_id}`
  - Implement `DELETE /v2/webhook/{webhook_id}`
  - Add CLI commands: `webhooks list`, `webhooks create`, `webhooks update`, `webhooks delete`
- **Files**:
  - `internal/clickup/client.go` — modify (add Webhooks service accessor)
  - `internal/clickup/types.go` — modify (add Webhook types)
  - `internal/cmd/webhooks.go` — create (CLI commands)
- **Verification**: `make ci` passes; webhook CRUD works

### 18. Create Goals Service
- **Status**: pending
- **Depends on**: 1
- **Spec**: specs/features/goals.md
- **Description**:
  - Create `GoalsService` with List, Get, Create, Update, Delete, CreateKeyResult, EditKeyResult, DeleteKeyResult methods
  - Implement `GET /v2/team/{team_id}/goal`
  - Implement `GET /v2/goal/{goal_id}`
  - Implement `POST /v2/team/{team_id}/goal`
  - Implement `PUT /v2/goal/{goal_id}`
  - Implement `DELETE /v2/goal/{goal_id}`
  - Implement `POST /v2/goal/{goal_id}/key_result`
  - Implement `PUT /v2/key_result/{key_result_id}`
  - Implement `DELETE /v2/key_result/{key_result_id}`
  - Add CLI commands: `goals list`, `goals get`, `goals create`, `goals update`, `goals delete`, `goals add-key-result`, `goals update-key-result`, `goals delete-key-result`
- **Files**:
  - `internal/clickup/client.go` — modify (add Goals service accessor)
  - `internal/clickup/types.go` — modify (add Goal types)
  - `internal/cmd/goals.go` — create (CLI commands)
- **Verification**: `make ci` passes; goal CRUD works

---

## Phase 7: Users, Members, Attachments (P1)
> Branch: `phase/7-users-members` | PR → fix/plain-output

### 19. Create Users Service
- **Status**: pending
- **Depends on**: 1
- **Spec**: specs/features/users.md
- **Description**:
  - Create `UsersService` with Get, Invite, Update, Remove methods
  - Implement `GET /v2/team/{team_id}/user/{user_id}`
  - Implement `POST /v2/team/{team_id}/user`
  - Implement `PUT /v2/team/{team_id}/user/{user_id}`
  - Implement `DELETE /v2/team/{team_id}/user/{user_id}`
  - Add CLI commands: `users get`, `users invite`, `users update`, `users remove`
- **Files**:
  - `internal/clickup/client.go` — modify (add Users service accessor)
  - `internal/clickup/types.go` — modify (add UserDetail types)
  - `internal/cmd/users.go` — create (CLI commands)
- **Verification**: `make ci` passes; user CRUD works

### 20. Extend Members Service (List/Task Members)
- **Status**: pending
- **Depends on**: 1
- **Spec**: specs/features/members.md
- **Description**:
  - Add ListMembers, TaskMembers methods to existing MembersService
  - Implement `GET /v2/list/{list_id}/member`
  - Implement `GET /v2/task/{task_id}/member`
  - Add CLI commands: `members list-members`, `members task-members`
- **Files**:
  - `internal/clickup/client.go` — modify (extend MembersService)
  - `internal/clickup/types.go` — modify (add MemberUser types)
  - `internal/cmd/members.go` — modify (add commands)
- **Verification**: `make ci` passes; member listing works

### 21. Create Attachments Service
- **Status**: pending
- **Depends on**: 1, 2
- **Spec**: specs/features/attachments.md
- **Description**:
  - Create `AttachmentsService` with Upload, List, Create methods
  - Implement `POST /v2/task/{task_id}/attachment` (v2 upload via multipart)
  - Implement `GET /v3/workspaces/{wid}/{parent_type}/{parent_id}/attachments` (v3 list)
  - Implement `POST /v3/workspaces/{wid}/{parent_type}/{parent_id}/attachments` (v3 upload)
  - Add CLI commands: `attachments upload`, `attachments list`, `attachments create`
- **Files**:
  - `internal/clickup/client.go` — modify (add Attachments service accessor)
  - `internal/clickup/types.go` — modify (add Attachment types)
  - `internal/cmd/attachments.go` — create (CLI commands)
- **Verification**: `make ci` passes; file upload works

---

## Phase 8: v3 Domains (P1)
> Branch: `phase/8-v3-domains` | PR → fix/plain-output

### 22. Create Chat Service (v3)
- **Status**: pending
- **Depends on**: 2
- **Spec**: specs/features/chat.md
- **Description**:
  - Create `ChatService` with all channel and message methods
  - Implement channels: ListChannels, GetChannel, GetFollowers, GetMembers, CreateChannel, CreateDM, CreateLocationChannel, UpdateChannel, DeleteChannel
  - Implement messages: ListMessages, SendMessage, UpdateMessage, DeleteMessage
  - Implement reactions: ListReactions, CreateReaction, DeleteReaction
  - Implement replies: ListReplies, CreateReply
  - Implement: GetTaggedUsers
  - All endpoints use v3 paths with workspace ID
  - Add CLI commands: `chat channels`, `chat channel`, `chat create-channel`, `chat create-dm`, `chat messages`, `chat send`, etc.
- **Files**:
  - `internal/clickup/client.go` — modify (add Chat service accessor)
  - `internal/clickup/types.go` — modify (add Chat types)
  - `internal/cmd/chat.go` — create (CLI commands)
- **Verification**: `make ci` passes; chat commands work

### 23. Create Docs Service (v3)
- **Status**: pending
- **Depends on**: 2
- **Spec**: specs/features/docs.md
- **Description**:
  - Create `DocsService` with Search, Get, PageListing, Pages, Page, Create, CreatePage, EditPage methods
  - Implement `GET /v3/workspaces/{wid}/docs`
  - Implement `GET /v3/workspaces/{wid}/docs/{doc_id}`
  - Implement `GET /v3/workspaces/{wid}/docs/{doc_id}/page_listing`
  - Implement `GET /v3/workspaces/{wid}/docs/{doc_id}/pages`
  - Implement `GET /v3/workspaces/{wid}/docs/{doc_id}/pages/{page_id}`
  - Implement `POST /v3/workspaces/{wid}/docs`
  - Implement `POST /v3/workspaces/{wid}/docs/{doc_id}/pages`
  - Implement `PUT /v3/workspaces/{wid}/docs/{doc_id}/pages/{page_id}`
  - Add CLI commands: `docs search`, `docs get`, `docs page-listing`, `docs pages`, `docs page`, `docs create`, `docs create-page`, `docs edit-page`
- **Files**:
  - `internal/clickup/client.go` — modify (add Docs service accessor)
  - `internal/clickup/types.go` — modify (add Doc types)
  - `internal/cmd/docs.go` — create (CLI commands)
- **Verification**: `make ci` passes; docs commands work

---

## Phase 9: P2 Completeness
> Branch: `phase/9-p2-completeness` | PR → fix/plain-output

### 24. Create User Groups Service
- **Status**: complete
- **Depends on**: 1
- **Spec**: specs/features/user-groups.md
- **Description**:
  - Create `UserGroupsService` with List, Create, Update, Delete methods
  - Implement `GET /v2/group`
  - Implement `POST /v2/team/{team_id}/group`
  - Implement `PUT /v2/group/{group_id}`
  - Implement `DELETE /v2/group/{group_id}`
  - Add CLI commands: `groups list`, `groups create`, `groups update`, `groups delete`
- **Files**:
  - `internal/clickup/client.go` — modify (add UserGroups service accessor)
  - `internal/clickup/types.go` — modify (add UserGroup types)
  - `internal/cmd/groups.go` — create (CLI commands)
- **Verification**: `make ci` passes

### 25. Create Roles Service
- **Status**: complete
- **Depends on**: 1
- **Spec**: specs/features/roles.md
- **Description**:
  - Create `RolesService` with List method
  - Implement `GET /v2/team/{team_id}/customroles`
  - Add CLI command: `roles list`
- **Files**:
  - `internal/clickup/client.go` — modify (add Roles service accessor)
  - `internal/clickup/types.go` — modify (add CustomRole types)
  - `internal/cmd/roles.go` — create (CLI commands)
- **Verification**: `make ci` passes

### 26. Create Guests Service
- **Status**: complete
- **Depends on**: 1
- **Spec**: specs/features/guests.md
- **Description**:
  - Create `GuestsService` with Get, Invite, Update, Remove, AddToTask, RemoveFromTask, AddToList, RemoveFromList, AddToFolder, RemoveFromFolder methods
  - Implement workspace-level guest CRUD
  - Implement guest-to-resource assignments
  - Add CLI commands: `guests get`, `guests invite`, `guests update`, `guests remove`, `guests add-to-task`, `guests remove-from-task`, etc.
- **Files**:
  - `internal/clickup/client.go` — modify (add Guests service accessor)
  - `internal/clickup/types.go` — modify (add Guest types)
  - `internal/cmd/guests.go` — create (CLI commands)
- **Verification**: `make ci` passes

### 27. Create Shared Hierarchy Service
- **Status**: pending
- **Depends on**: 1
- **Spec**: specs/features/shared-hierarchy.md
- **Description**:
  - Create `SharedHierarchyService` with List method
  - Implement `GET /v2/team/{team_id}/shared`
  - Add CLI command: `shared-hierarchy list`
- **Files**:
  - `internal/clickup/client.go` — modify (add SharedHierarchy service accessor)
  - `internal/clickup/types.go` — modify (add SharedResources types)
  - `internal/cmd/shared.go` — create (CLI commands)
- **Verification**: `make ci` passes

### 28. Create Templates Service
- **Status**: pending
- **Depends on**: 1
- **Spec**: specs/features/templates.md
- **Description**:
  - Create `TemplatesService` with List method
  - Implement `GET /v2/team/{team_id}/taskTemplate`
  - Add CLI command: `templates list`
- **Files**:
  - `internal/clickup/client.go` — modify (add Templates service accessor)
  - `internal/clickup/types.go` — modify (add TaskTemplate types)
  - `internal/cmd/templates.go` — create (CLI commands)
- **Verification**: `make ci` passes

### 29. Create Custom Task Types Service
- **Status**: pending
- **Depends on**: 1
- **Spec**: specs/features/custom-task-types.md
- **Description**:
  - Create `CustomTaskTypesService` with List method
  - Implement `GET /v2/team/{team_id}/custom_item`
  - Add CLI command: `task-types list`
- **Files**:
  - `internal/clickup/client.go` — modify (add CustomTaskTypes service accessor)
  - `internal/clickup/types.go` — modify (add CustomTaskType types)
  - `internal/cmd/tasktypes.go` — create (CLI commands)
- **Verification**: `make ci` passes

### 30. Create Legacy Time Service
- **Status**: pending
- **Depends on**: 1
- **Spec**: specs/features/time-tracking-legacy.md
- **Description**:
  - Create `LegacyTimeService` with List, Track, Edit, Delete methods
  - Implement task-level time tracking endpoints
  - Implement `GET /v2/task/{task_id}/time`
  - Implement `POST /v2/task/{task_id}/time`
  - Implement `PUT /v2/task/{task_id}/time/{interval_id}`
  - Implement `DELETE /v2/task/{task_id}/time/{interval_id}`
  - Add CLI commands: `time-legacy list`, `time-legacy track`, `time-legacy update`, `time-legacy delete`
- **Files**:
  - `internal/clickup/client.go` — modify (add LegacyTime service accessor)
  - `internal/clickup/types.go` — modify (add LegacyTimeInterval types)
  - `internal/cmd/time_legacy.go` — create (CLI commands)
- **Verification**: `make ci` passes

### 31. Create Audit Logs Service (v3)
- **Status**: pending
- **Depends on**: 2
- **Spec**: specs/features/auditlogs.md
- **Description**:
  - Create `AuditLogsService` with Query method
  - Implement `POST /v3/workspaces/{wid}/auditlogs`
  - Add CLI command: `auditlogs query`
- **Files**:
  - `internal/clickup/client.go` — modify (add AuditLogs service accessor)
  - `internal/clickup/types.go` — modify (add AuditLog types)
  - `internal/cmd/auditlogs.go` — create (CLI commands)
- **Verification**: `make ci` passes

### 32. Create ACLs Service (v3)
- **Status**: pending
- **Depends on**: 2
- **Spec**: specs/features/acls.md
- **Description**:
  - Create `ACLsService` with Update method
  - Implement `PATCH /v3/workspaces/{wid}/{object_type}/{object_id}/acls`
  - Add CLI command: `acls update`
- **Files**:
  - `internal/clickup/client.go` — modify (add ACLs service accessor)
  - `internal/clickup/types.go` — modify (add ACL types)
  - `internal/cmd/acls.go` — create (CLI commands)
- **Verification**: `make ci` passes

---

## Summary

| Phase | Tasks | Endpoints | Priority |
|-------|-------|-----------|----------|
| 1. Foundation | 3 | 0 (arch) | - |
| 2. Workspaces & Auth | 2 | 5 | P0 |
| 3. Core CRUD | 5 | 37 | P0 |
| 4. Time Tracking | 1 | 11 | P0 |
| 5. Task Extensions | 4 | 22 | P1 |
| 6. Views/Webhooks/Goals | 3 | 24 | P1 |
| 7. Users/Members/Attachments | 3 | 10 | P1 |
| 8. v3 Domains | 2 | 27 | P1 |
| 9. P2 Completeness | 9 | 32 | P2 |
| **Total** | **32** | **168** | - |

## Recommended Starting Point

Start with **Task 1: Migrate to Dynamic v2/v3 Base URL** - this is the foundational change that enables all subsequent v3 endpoint implementations. It has zero dependencies and is required before any v3 work can begin.

After completing Phase 1 (tasks 1-3), proceed with Phase 2 (Workspaces & Authorization) to enable workspace ID auto-detection for v3 calls, then continue through phases in order.
