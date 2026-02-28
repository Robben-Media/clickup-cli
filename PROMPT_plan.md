# Planning Mode

Study the specifications and create an implementation plan organized into phases.

## Your Task

1. Read all files in `specs/` directory
2. Read `AGENTS.md` for project context, build commands, and design mandates
3. Read the existing codebase to understand what's already built
4. Create `IMPLEMENTATION_PLAN.md` with ordered tasks grouped into phases

## Implementation Plan Format

```markdown
# Implementation Plan

Generated: [timestamp]
Last Updated: [timestamp]

## Summary
[1-2 sentence overview of what's being built]

## Architecture Decisions
- [Key technical decisions]

## Blockers / Questions
1. [Any missing information or decisions needed]

---

## Phase 1: Foundation
> Branch: `phase/1-foundation` | PR → build branch

### 1. [Task Name]
- **Status**: pending
- **Depends on**: none | [task numbers]
- **Spec**: specs/[filename].md
- **Description**: [Detailed — what to build, file paths to create/modify]
- **Files**: `path/to/file.ts` — [create/modify]
- **Verification**: [What "done" looks like + commands to run]

### 2. [Next Task]
...

---

## Phase 2: [Name]
> Branch: `phase/2-[name]` | PR → build branch

### N. [Task]
...

---

## Recommended Starting Point
[Which tasks have zero dependencies]
```

## Ordering Principles

1. **Data layer first**: config/shared data before anything else
2. **Dependencies first**: If B needs A, A comes first
3. **Risk first**: Architectural uncertainty early
4. **Foundation → Components → Pages → Integration → Polish**

## Phase Structure

| Phase | Purpose |
|-------|---------|
| Foundation | API client refactor (v3 base URL), new service scaffolds, type definitions |
| Core CRUD | Spaces, Folders, Lists, Tasks, Comments — complete CRUD for existing domains |
| Time Tracking | Timer start/stop, entry CRUD, tags, history |
| Task Extensions | Tags, Checklists, Relationships, Custom Fields |
| Workspace Management | Workspaces, Auth, Users, Groups, Members, Roles, Guests |
| Views & Goals | Views CRUD, Webhooks, Goals + Key Results |
| v3 Domains | Chat, Docs, Attachments, Audit Logs, ACLs |
| Polish | Verification passes, test coverage, documentation |

## Task Requirements

- Each task completable in ONE iteration (one `claude -p` invocation)
- Include exact file paths to create or modify
- Reference specific spec files
- Include verification commands
- Detailed enough to implement with optional agent delegation, but do NOT pre-assign agent types or pre-plan subtasks

## Guidelines

- Flag ambiguities in the Blockers section
- Never put all polish into one mega-task
- Keep tasks focused — one concern per task

## Output

After creating `IMPLEMENTATION_PLAN.md`, summarize:
- Total tasks and phase breakdown
- Any blockers or questions
- Recommended starting point

## Completion Signal

After you have:
1. Read all specs in `specs/`
2. Created/updated `IMPLEMENTATION_PLAN.md` with tasks for every spec
3. Provided your summary

Output exactly:

<promise>PLAN_COMPLETE</promise>

This signals the planning loop to exit. Do NOT continue iterating once all specs are planned.
