---
title: Interactive live-search mode for tickets search
status: pending
priority: medium
created: "2026-04-28"
labels:
    - ux
    - cli
---

## Description

Improve the UX of `tickets search` by adding an interactive (TUI) mode when no query argument is provided. Instead of requiring the user to pass a query string upfront, the command should launch an interactive prompt where results update in real time as the user types, showing the top 10 matches.

Current behaviour: `tickets search "<query>"` — runs a one-shot search and exits.

Desired behaviour: `tickets search` (no args) — opens an interactive search prompt; the top 10 matching tickets re-render on every keystroke.

## Acceptance Criteria

- [ ] Running `tickets search` with no arguments enters interactive mode
- [ ] Results update on every keystroke, showing at most 10 matches ranked by relevance
- [ ] Searches across title, description, notes, and labels (same fields as today)
- [ ] User can press Enter to select a ticket (or Esc/Ctrl-C to exit without selection)
- [ ] Running `tickets search "<query>"` with an explicit argument still works as before (non-interactive, backward-compatible)
- [ ] Works in terminals that support ANSI escape sequences; degrades gracefully (falls back to non-interactive) when stdin is not a TTY

## Notes

Candidate libraries for the interactive TUI layer: `github.com/charmbracelet/bubbletea` + `github.com/charmbracelet/bubbles` (textinput + list components). Alternatively a lightweight readline-style approach using `github.com/manifoldco/promptui` if a full Bubble Tea dep is too heavy.

## Dependencies
