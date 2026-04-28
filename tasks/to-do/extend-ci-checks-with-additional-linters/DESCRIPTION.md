---
title: Extend CI checks with additional linters
status: pending
priority: medium
created: "2026-04-28"
labels:
    - ci
    - tooling
---

## Description
The CI pipeline currently runs a limited set of checks. Extend it with additional linters to catch a broader class of issues (style, bugs, security, dead code) before review.

## Acceptance Criteria
- [ ] Survey candidate linters and decide which to adopt (e.g. golangci-lint preset expansion, gosec, staticcheck, govulncheck, misspell, gofumpt)
- [ ] Wire the chosen linters into the existing CI workflow
- [ ] Ensure the linters run via the Makefile so they are reproducible locally
- [ ] Fix or baseline existing violations so CI is green
- [ ] Document how to run the linters locally

## Notes

## Dependencies
