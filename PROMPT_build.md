# Build Mode

## Completion Promise

**You MUST complete every task fully.** Do not leave TODOs, placeholder implementations, stub functions, or skip "similar" items. If a task says "create 5 endpoints," implement all 5 — not 2 with a comment saying "remaining follow the same pattern." Every line of code you write must be production-ready. Re-read every modified file before marking a task complete.

---

Implement the next task from the implementation plan. You may delegate subtasks to agents, but focus on completing **ONE parent task** per iteration.

## Startup Sequence

Every iteration:

1. Read `IMPLEMENTATION_PLAN.md` for task status
2. Read `AGENTS.md` for build commands, patterns, and design mandates
3. Read `progress.txt` for context from previous iterations
4. Check current branch — ensure you're on the correct phase branch
5. Find the highest-priority task with status "pending" whose dependencies are all "complete"

## Branching & PR Strategy

Work happens on phase branches. Each phase gets its own PR.

### Branch Structure
```
feat/initial-template
└── agent/[machine]/[task]                ← long-lived build branch
    ├── phase/1-[name]                    ← PR → build branch
    ├── phase/2-[name]                    ← PR → build branch
    └── phase/N-[name]                    ← PR → build branch

Final: build branch → feat/initial-template  ← final review PR
```

### Phase Management

Phase → PR mappings are defined in `IMPLEMENTATION_PLAN.md`.

**Starting a phase** (if no phase branch exists for the current task):
```bash
git checkout [build-branch]
git checkout -b phase/N-[name]
```

**Closing a phase** (when ALL tasks in a phase are complete):
1. Run full verification (`make ci`)
2. Open PR from phase branch → build branch
3. Note the PR URL in `progress.txt`

## Implementing the Task

### 1. Read the task and its spec

Read the task description from `IMPLEMENTATION_PLAN.md` and the referenced spec file. Understand exactly what needs to be built.

### 2. Decompose into subtasks (if complex)

For simple tasks, implement directly. For complex tasks with multiple files or concerns, break into subtasks and delegate to agents via the Task tool:

| Agent Type | Use For |
|-----------|---------|
| `general-purpose` | File creation, data files, utilities, config, non-visual code |
| `code-architect` | Architecture decisions, complex structures, API design |
| `code-reviewer` | Code quality audits, dead code, cleanup |

### Delegation tips

- Each delegated subtask must be **self-contained** — include exact file paths, spec references, and acceptance criteria
- **Parallelize** independent subtasks by launching multiple agents in a single message
- Run tests after every implementation: `make test`

### 3. Verify

Run ALL feedback loops from AGENTS.md:
```bash
make ci   # runs fmt, lint, test, build
```

Fix failures before committing.

### 4. Commit

```bash
git add [specific files]
git commit -m "feat: [task description] (task N)"
git push
```

One `feat:` commit per task. Keep it clean.

## On Success

1. Update task status to "complete" in `IMPLEMENTATION_PLAN.md`
2. Append to `progress.txt` using this template:
   ```
   ## Task [N] - [Timestamp]
   ### Completed: [Task Name]
   - What was implemented
   - Files changed
   - Agent(s) used (if any)
   - Learnings for future iterations
   ---
   ```
3. If this was the last task in a phase, also:
   - Run full verification
   - Open PR: `gh pr create --base [build-branch] --head phase/N-[name] --title "Phase N: [Name]" --body "..."`
   - Append PR info to `progress.txt`

## Rules

- ONLY work on ONE parent task per iteration
- Delegate subtasks to agents when beneficial, but stay focused on the one task
- NEVER skip verification steps
- NEVER leave TODO comments, placeholder implementations, or debug logging in code
- NEVER skip "similar" items — if a task says to create N things, create all N
- ALWAYS re-read every modified file before marking a task complete
- ALWAYS commit before exiting
- ALWAYS update IMPLEMENTATION_PLAN.md status before exiting
- Before reporting done: confirm all PR comments addressed, CI passing, no merge conflicts

## Completion Signal

If ALL tasks in `IMPLEMENTATION_PLAN.md` have status "complete" and all phase PRs are open/merged:

<promise>BUILD_COMPLETE</promise>
