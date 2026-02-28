# Build Mode

## Completion Promise

**You MUST complete every task fully.** Do not leave TODOs, placeholder implementations, stub functions, or skip "similar" items. If a task says "create 5 endpoints," implement all 5 — not 2 with a comment saying "remaining follow the same pattern." Every line of code you write must be production-ready. Re-read every modified file before marking a task complete.

---

Implement the next task from the implementation plan. You may delegate subtasks to agents, but focus on completing **ONE parent task** per iteration.

## Startup Sequence

Every iteration:

1. **Sync repo and check PR status**:
   ```bash
   git fetch origin
   
   # Check if current branch's PR is merged
   CURRENT_BRANCH=$(git branch --show-current)
   PR_STATE=$(gh pr list --head "$CURRENT_BRANCH" --json state --jq '.[0].state' 2>/dev/null || echo "none")
   
   if [ "$PR_STATE" = "MERGED" ]; then
     # PR was merged - sync from base and continue or start new phase
     git checkout fix/plain-output  # or your base branch
     git pull origin fix/plain-output
     # Continue working on this base - next task will create its own branch if needed
   elif [ "$PR_STATE" = "OPEN" ]; then
     # PR exists - pull and continue
     git pull origin "$CURRENT_BRANCH"
   else
     # No PR - check if we need to create a branch or just pull
     git pull origin "$CURRENT_BRANCH" 2>/dev/null || true
   fi
   ```
2. Read `IMPLEMENTATION_PLAN.md` for task status
3. Read `AGENTS.md` for build commands, patterns, and design mandates
4. Read `progress.txt` for context from previous iterations
5. Check current branch — ensure you're on the correct phase branch
6. Find the highest-priority task with status "pending" whose dependencies are all "complete"

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

### 4. Stage, Commit, Push, PR

After implementing and verifying the task:

```bash
# Stage all changes
git add -A

# Commit with task reference
git commit -m "feat: [task description] (task N)"

# Push to remote
git push -u origin $(git branch --show-current)
```

Then create or update a PR for this iteration's changes:

```bash
# Check if PR already exists for this branch
gh pr list --head $(git branch --show-current)

# If no PR exists, create one
gh pr create --base [base-branch] --title "[Task Name] (task N)" --body "## Summary
[What was implemented]

## Files Changed
[List of modified files]

## Testing
- \`make ci\` passes

Closes #[issue-number] (if applicable)"
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
   - PR URL: [link to PR]
   - Agent(s) used (if any)
   - Learnings for future iterations
   ---
   ```
3. If this was the last task in a phase, note it in `progress.txt` for final review

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
