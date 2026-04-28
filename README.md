# tickets

`tickets` is a lightweight, git-backed ticketing system that lives inside your repository. Tickets are plain Markdown files with YAML frontmatter — no external services or databases and no syncing. They travel with the code, get reviewed in PRs, and are readable by both humans and AI agents.

## Install

```sh
go install github.com/krzhapps/GithubTickets/cmd/tickets@latest
```

Requires Go 1.21+.

## Quickstart

```sh
# 1. Initialise the tasks/ directory in your repo
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
tasks/
  to-do/          # open tickets waiting to be worked on
  in-progress/    # tickets actively being worked on
  done/           # completed tickets, kept for history
  archived/       # closed without implementation (won't fix, duplicate, superseded)
```

Each ticket is a directory named after its slug, containing a single `DESCRIPTION.md`:

```
tasks/to-do/add-rate-limiting-to-the-api/
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
tickets init        Initialise the tasks/ directory
tickets new         Create a new ticket
tickets list        List tickets (default: to-do + in-progress)
tickets show        Print a ticket's full content
tickets edit        Open a ticket in $EDITOR
tickets start       Move a ticket to in-progress and check out a ticket/<slug> branch
tickets done        Move a ticket to done
tickets archive     Move a ticket to archived
tickets move        Move a ticket to an arbitrary status
tickets search      Full-text search across all tickets
tickets validate    Check all tickets for schema errors
tickets deps        Show the dependency graph
```

## Conventions

- **Slug naming:** `<short-description>` in kebab-case, e.g. `auth-google-oauth-errors`, `tests-scraper-coverage`.
- **One concern per ticket.** If a ticket grows into multiple independent concerns, split it.
- **Dependencies declared explicitly.** A ticket that cannot start until another is done lists it under `Dependencies`.

