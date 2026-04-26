---
title: Add Makefile and golangci-lint config
status: pending
priority: medium
created: "2026-04-25"
labels:
    - tooling
---

## Description

Add a top-level Makefile that wraps the common Go workflows (build, test, lint, install, tidy) so contributors don't have to remember the exact `go` invocations, and a `.golangci.yml` config that pins the lint surface for both local runs and CI.

## Acceptance Criteria

- [ ] `Makefile` with targets: `build`, `test`, `lint`, `install`, `tidy`
- [ ] `make test` runs `go test ./... -race -cover`
- [ ] `.golangci.yml` enables errcheck, govet, ineffassign, staticcheck, gosimple, gofmt, goimports, misspell, unused
- [ ] `make lint` runs cleanly on the current tree
- [ ] `make build` produces a `tickets` binary at the repo root

## Notes

Keep the Makefile minimal — no fancy phony tricks, no hidden code generation. It should be skim-readable in 30 seconds.
