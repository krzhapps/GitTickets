---
name: tickets
description: Use when the user wants to create, list, search, edit, validate, or change the lifecycle (start/done/archive/move) of tickets in a repo that uses the `tickets` CLI — a git-backed ticketing system where each ticket is a Markdown file with YAML frontmatter under `tasks/{to-do,in-progress,done,archived}/<slug>/DESCRIPTION.md`. Also use when the user asks about tickets, ticket dependencies, , creating a branch from a ticket, or initializing the `tasks/` layout in a new repo.
---

# tickets

`tickets` is a git-backed ticketing CLI. Tickets are plain Markdown files with YAML frontmatter living under `tasks/<status>/<slug>/DESCRIPTION.md`. Lifecycle moves use `git mv` to preserve history, so always prefer the CLI over editing or moving files by hand.

## When this skill applies

Trigger on requests that mention tickets, issues-in-repo, the `tasks/` directory, or any of the verbs: open / create / file / list / show / search / start / finish / close / archive / move / validate / depend on. If the repo has no `tasks/` directory yet, the user likely wants `tickets init` first.

## Core rules (do not skip)

1. **Always go through the CLI for lifecycle changes.** Use `tickets start|done|archive|move` — never `mv` files yourself, never edit the `status:` frontmatter field directly. The CLI runs `git mv` and keeps history attached.
2. **Slugs are kebab-case** and act as the ticket's identity. Generate them from the title: lowercase, ASCII, words joined by `-`. Example: `"Add rate limiting to the API"` → `add-rate-limiting-to-the-api`.
3. **Run from inside the repo.** The CLI auto-discovers the `tasks/` root from CWD; pass `--root <path>` only if running from outside.
4. **One concern per ticket.** If a request bundles multiple independent pieces of work, file them as separate tickets and link them via `Dependencies`.
5. **After any create/move, run `tickets validate`** to catch frontmatter or dependency-graph errors before handing back to the user.

## Invocation

If the binary is on `$PATH`, call `tickets …`. If not you need to install it via `go install github.com/krzhapps/GithubTickets/cmd/tickets@latest`

## Common workflows

### Create a new ticket
```sh
tickets new <slug> \
  --title "<human title>" \
  --priority low|medium|high \
  --label <label> --label <label> \
  --assignee <name> \
  --no-edit
```
Pass `--no-edit` when scripting; omit it if the user wants to drop straight into `$EDITOR`. Then open the generated `DESCRIPTION.md` and fill in `## Description`, `## Acceptance Criteria`, `## Notes`, and `## Dependencies` from the user's request before reporting back.

### List / find tickets
- Default open work: `tickets list`
- Include closed: `tickets list --all`
- Filter: `tickets list --status in-progress --priority high --label backend --assignee alice`
- Machine-readable: `tickets list --json` (use this when piping into further processing)
- Full-text: `tickets search "<query>"` — searches title, description, notes, and labels

### Inspect one ticket
- Human: `tickets show <slug>`
- Structured: `tickets show <slug> --json`

### Lifecycle moves
| Verb the user uses                      | Command                            |
|-----------------------------------------|------------------------------------|
| "start working on", "pick up", "begin"  | `tickets start <slug>`             |
| "done", "finished", "ship", "close"     | `tickets done <slug>`              |
| "won't fix", "duplicate", "supersede"   | `tickets archive <slug>`           |
| arbitrary (e.g. set `blocked`)          | `tickets move <slug> blocked`      |

Valid statuses for `move`: `pending`, `in-progress`, `blocked`, `done`, `archived`.

### Branch off a ticket
`tickets branch <slug> --checkout` — creates and switches to a git branch named after the slug. Use this when the user says "let's start working on X" and they want a branch in one shot (combine with `tickets start <slug>`).

### Dependencies
- Tree for one ticket: `tickets deps <slug>`
- Whole edge list: `tickets deps`
- Graphviz: `tickets deps --graph` (pipe into `dot -Tpng`)

To declare a dependency, list the slug under `## Dependencies` in the dependent ticket's `DESCRIPTION.md`:
```markdown
## Dependencies

- `auth-middleware-refactor` — short reason
```

### Validate
`tickets validate` checks all tickets; `tickets validate <slug>` checks one. Run after creating or editing tickets.

## Frontmatter schema

```yaml
---
title: <string, required>
status: pending | in-progress | blocked        # required; done/archived live in their dirs
priority: low | medium | high                  # required
created: "YYYY-MM-DD"                          # required, quoted
labels:                                        # optional
    - <label>
assignee: <string>                             # optional
---
```

The body should follow this shape (the `new` command scaffolds it):
```markdown
## Description
…

## Acceptance Criteria
- [ ] …

## Notes
…

## Dependencies
- `<other-slug>` — why
```

Do not invent extra frontmatter fields — `tickets validate` will reject them. If the user wants to track something not covered by the schema, put it under `## Notes`.

## Initialising a new repo

If `tasks/` does not exist, run `tickets init` first — it scaffolds `tasks/{to-do,in-progress,done,archived}/`. Don't create those directories by hand.

## Reporting back to the user

After running tickets commands, summarise the resulting state succinctly: the slug, new status, and (for `list`/`search`) a short rendered table rather than the raw CLI output. For destructive-feeling actions (`archive`, `move … archived`), confirm with the user first unless they were explicit.
