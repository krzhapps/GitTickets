---
title: tickets new should auto-track description file
status: done
priority: medium
created: "2026-06-01"
labels:
    - enhancement
---

## Description

When `tickets new` creates a ticket, the generated `DESCRIPTION.md` file is not staged in git. This means that when `tickets start` is subsequently called, the git mv operation fails because the file is untracked.

Modify the `tickets new` command to automatically run `git add` on the newly created `DESCRIPTION.md` file after writing it.

## Acceptance Criteria

- [ ] After `tickets new`, the created `DESCRIPTION.md` is automatically staged (tracked) in git
- [ ] `tickets start` succeeds immediately after `tickets new` without requiring a manual `git add`
- [ ] Existing behavior of `tickets new` is otherwise unchanged
