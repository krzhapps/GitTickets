---
title: Write README with install and usage
status: pending
priority: medium
created: "2026-04-25"
labels:
    - docs
---

## Description

Write the project README — what `tickets` is, how to install it, the on-disk file layout, and a short usage walkthrough covering the common lifecycle (init → new → list → start → done → validate). The README is the first thing someone sees on the GitHub page; it should be enough to evaluate the tool in under a minute.

## Acceptance Criteria

- [ ] One-paragraph elevator pitch at the top
- [ ] Install section: `go install github.com/krzhapps/GithubTickets/cmd/tickets@latest`
- [ ] Quickstart: 5–7 commands that walk through the full lifecycle
- [ ] Reference of the directory layout (to-do / in-progress / done / archived)
- [ ] Reference of the DESCRIPTION.md frontmatter schema
- [ ] Link to TicketingSystemIdea.md for the design rationale

## Notes

Keep examples copy-pasteable — no placeholder `<your-slug>` style; use real-looking slugs.
