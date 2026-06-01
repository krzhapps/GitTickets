---
title: Add configuration file support to init command
status: pending
priority: medium
created: "2026-04-28"
labels:
    - enhancement
    - cli
---

## Description

Extend the `tickets init` command to accept an optional configuration file that bootstraps project-level defaults. On init, the file is written into the tickets directory so it is easy to find and modify later.

The configuration file (e.g. `tickets/config.yml`) should support:

1. **Ticket path** — override the directory where tickets are stored (default: `tickets/`).
2. **Possible assignees** — a list of valid assignee names; used for validation and autocomplete.
3. **Possible labels** — a list of valid labels; used for validation and autocomplete.

## Acceptance Criteria

- [ ] `tickets init --config <file>` accepts a path to a YAML configuration file.
- [ ] The configuration file is parsed and written to `<tickets-root>/config.yml` during init.
- [ ] If a `config.yml` already exists in the tickets root, the command warns and does not overwrite unless `--force` is passed.
- [ ] `tickets validate` reads `config.yml` and rejects tickets whose `assignee` or `labels` are not in the allowed lists (when those lists are non-empty).
- [ ] Ticket path setting in `config.yml` is respected by all subcommands.
- [ ] Documented config schema with an example file in the README or `tickets init --help` output.

## Notes

Example `config.yml` shape:

```yaml
ticket_path: tasks          # optional; defaults to "tickets"
assignees:
  - alice
  - bob
labels:
  - backend
  - frontend
  - enhancement
  - bug
```

## Dependencies
