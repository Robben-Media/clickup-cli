---
name: clickup
description: Use clickup-cli for ClickUp. Install, auth, read, and change ClickUp through the CLI. Do not create live tasks unless the user asked.
---

# ClickUp

Use `clickup-cli` for all ClickUp work. Do not curl the ClickUp API. Do not invent a second tracker.

This skill is for **using** ClickUp. `AGENTS.md` in this repo is for **developing** the CLI. Do not start implementing API endpoints unless that is the task.

## Install

Repo: https://github.com/Robben-Media/clickup-cli

```bash
git clone https://github.com/Robben-Media/clickup-cli.git
cd clickup-cli
make build
```

Binary: `bin/clickup-cli`. Releases: https://github.com/Robben-Media/clickup-cli/releases

To keep the skill globally:

```bash
cp -R .claude/skills/clickup ~/.claude/skills/clickup
```

## Auth

Do not put the API key in a repo `.env`. Do not write it into a file you might commit. The CLI does not read `.env`.

Interactive (key goes in the OS keyring):

```bash
clickup-cli auth set-key --stdin
clickup-cli auth set-team TEAM_ID
clickup-cli auth status
```

CI / scripts:

```bash
export CLICKUP_API_KEY="..."
export CLICKUP_TEAM_ID="..."
```

If no key is present, ask the user to run `auth set-key --stdin` themselves.

## Live ClickUp

Reads are fine when the user asked a ClickUp question.

These are live writes: `create`, `update`, `delete`, `merge`, `move`, `from-template`, `add-dep`, `comments add`, and anything else that mutates ClickUp. Name the workspace, list, or task. Get explicit approval for that operation. "Set up SEO" is not approval to seed a list.

Do not pass `--force` unless asked.

## How ClickUp actually works

Do not build a workflow product on top of ClickUp. Cite [ClickUp help](https://help.clickup.com) for product behavior.

- Dependencies are Waiting On / Blocking. Closing a blocker does not change the dependent's status or due date. That needs a ClickUp Automation. This CLI cannot create automations.
- Statuses belong to the space or list. `N/A` only resolves Waiting On if it is a closed-type status.
- Folders are not tasks. Parent tasks are tasks.
- Recurring tasks are not the same object as a one-shot card.
- If the CLI cannot do it, say so. Do not fake it with tags and extra views.

## Commands

Prefer `--json` when parsing. Run `clickup-cli <command> --help` instead of guessing flags.

```bash
clickup-cli auth status
clickup-cli auth whoami
clickup-cli spaces list
clickup-cli lists list --space SPACE_ID
clickup-cli tasks list --list LIST_ID
clickup-cli tasks get TASK_ID
clickup-cli tasks create LIST_ID "Name" --assignee USER_ID --due UNIX_MS
clickup-cli tasks update TASK_ID --status "to do"
clickup-cli relationships add-dep TASK_ID --depends-on OTHER_ID
clickup-cli relationships add-dep TASK_ID --dependency-of OTHER_ID
clickup-cli comments add TASK_ID "..."
```

`--depends-on`: this task waits on the other. `--dependency-of`: this task blocks the other. They are the same ClickUp link. Do not create both for one pair.

Trust `relies_on` / `--depends-on` as the source of truth when a spec lists both sides and they disagree.
