---
title: 'GitHub Actions workflow: vet, lint, test, validate'
status: pending
priority: medium
created: "2026-04-25"
labels:
    - tooling
    - ci
---

## Description

Wire up a GitHub Actions workflow that runs on every push and pull request, gating merges on the same checks contributors run locally: `go vet`, `golangci-lint`, `go test ./... -race -cover`, and `tickets validate` (so a malformed ticket fails the build).

## Acceptance Criteria

- [ ] `.github/workflows/ci.yml` triggers on `push` and `pull_request`
- [ ] Job runs `go vet ./...`
- [ ] Job runs `golangci-lint run` using the committed config
- [ ] Job runs `go test ./... -race -cover`
- [ ] Job builds the `tickets` binary and runs `tickets validate` against the repo's tickets/ tree
- [ ] Workflow caches the Go module cache to keep CI under ~2 minutes

## Notes

Pin the Go version in the workflow to whatever `go.mod` declares so local and CI never drift.

## Dependencies

- `lint-and-makefile` — the workflow shells out to `make lint`, so the Makefile and golangci-lint config must land first
