# tickets

`tickets` is a lightweight, git-backed ticketing system that lives inside your repository. Tickets are plain Markdown files with YAML frontmatter — no external services or databases and no syncing. They travel with the code, get reviewed in PRs, and are readable by both humans and AI agents.

## Install

```sh
go install github.com/krzhapps/GitTickets/cmd/tickets@latest
```

Requires Go 1.21+. Make sure Go's bin directory is on your `PATH` so the `tickets` command is found:

```sh
export PATH="$PATH:$(go env GOPATH)/bin"
```

Add that line to your shell profile (e.g. `~/.bashrc` or `~/.zshrc`) to make it permanent.

## Quickstart

```sh
# 1. Initialise the .tickets/ directory in your repo
tickets init

# 2. Create a new ticket
tickets new "add rate limiting to the API"

# 3. List open tickets
tickets list

# 4. Start working on a ticket (moves it to in-progress)
tickets start add-rate-limiting-to-the-api

# 5. Mark it done (moves it to done/, uses git mv to preserve history)
tickets done add-rate-limiting-to-the-api

# 6. Validate all tickets for schema errors
tickets validate
```

## Directory layout

```
.tickets/
  to-do/          # open tickets waiting to be worked on
  in-progress/    # tickets actively being worked on
  done/           # completed tickets, kept for history
  archived/       # closed without implementation (won't fix, duplicate, superseded)
```

Each ticket is a directory named after its slug, containing a single `DESCRIPTION.md`:

```
.tickets/to-do/add-rate-limiting-to-the-api/
  DESCRIPTION.md
```

## DESCRIPTION.md frontmatter schema

| Field      | Type                                              | Required |
|------------|---------------------------------------------------|----------|
| `title`    | string                                            | yes      |
| `status`   | `pending` \| `in-progress` \| `blocked`           | yes      |
| `priority` | `low` \| `medium` \| `high`                       | yes      |
| `created`  | `YYYY-MM-DD`                                      | yes      |
| `labels`   | list of strings                                   | no       |
| `assignee` | string                                            | no       |

Full example:

```markdown
---
title: Add rate limiting to the API
status: pending
priority: high
created: "2026-04-26"
labels:
    - backend
    - security
assignee: alice
---

## Description

Protect the public API endpoints from abuse by adding per-IP rate limiting.

## Acceptance Criteria

- [ ] Requests exceeding 100 req/min per IP receive 429
- [ ] Limit is configurable via environment variable

## Notes

Consider using a token-bucket algorithm. Redis not available — use in-process state for now.

## Dependencies

- `auth-middleware-refactor` — rate limiter hooks into the middleware chain
```

## All commands

```
tickets init        Initialise the .tickets/ directory
tickets new         Create a new ticket
tickets list        List tickets (default: to-do + in-progress)
tickets show        Print a ticket's full content
tickets edit        Open a ticket in $EDITOR
tickets start       Move a ticket to in-progress and check out a ticket/<slug> branch
                    (use --worktree for a dedicated sibling working directory)
tickets done        Move a ticket to done
tickets archive     Move a ticket to archived
tickets move        Move a ticket to an arbitrary status
tickets search      Full-text search across all tickets
tickets validate    Check all tickets for schema errors
tickets deps        Show the dependency graph
```

## Parallel work with worktrees

To work on several tickets at once without switching branches in place, start a
ticket with `--worktree`:

```sh
tickets start add-rate-limiting-to-the-api --worktree
```

This checks the `ticket/<slug>` branch out in its own [git worktree](https://git-scm.com/docs/git-worktree),
a sibling directory next to the repo:

```
<repo>/
<repo>-worktrees/
  add-rate-limiting-to-the-api/   # full working tree on branch ticket/add-rate-limiting-to-the-api
```

Each worktree is an independent working directory sharing the same git history,
so multiple agents or developers can work on different tickets concurrently
without stepping on each other. `cd` into the printed path to begin.

When you `tickets done` or `tickets archive` the ticket, its worktree is removed
automatically. If the worktree has uncommitted changes, removal is skipped with
a warning so nothing is lost — clean it up by hand with `git worktree remove`.

## Conventions

- **Slug naming:** `<short-description>` in kebab-case, e.g. `auth-google-oauth-errors`, `tests-scraper-coverage`.
- **One concern per ticket.** If a ticket grows into multiple independent concerns, split it.
- **Dependencies declared explicitly.** A ticket that cannot start until another is done lists it under `Dependencies`.

