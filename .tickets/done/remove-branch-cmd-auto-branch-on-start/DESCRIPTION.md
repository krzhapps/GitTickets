---
title: Remove branch command; auto-create branch on tickets start
status: done
priority: medium
created: "2026-04-27"
labels:
    - enhancement
---

## Description

Remove the standalone `branch` command. Instead, `tickets start <slug>` should automatically create and check out a `ticket/<slug>` git branch as part of moving the ticket to in-progress.

## Acceptance Criteria

- [ ] `tickets branch` command is removed
- [ ] `tickets start <slug>` creates and checks out a `ticket/<slug>` git branch
- [ ] If the branch already exists, start still succeeds (no error)
- [ ] Docs/help text updated to reflect the change
