---
title: Benchmark ticket store hot paths
status: pending
priority: medium
created: "2026-04-28"
labels:
    - performance
    - testing
---

## Description

We've identified `internal/store.Load()` and `internal/cli/search.go`'s
linear scan as the likely bottlenecks once the ticket count grows, and
filed `improve-ticket-search-and-indexing` to explore index/cache
designs. Before picking a design we need numbers, not intuitions: this
ticket is to land Go benchmarks plus a synthetic-fixture generator so
every later perf claim is measured.

The benchmarks should isolate each layer so we can tell *which* part of
a slow command is actually slow — a single end-to-end "search is fast"
number hides whether the win came from caching parses, indexing tokens,
or reducing allocations.

### Layers to cover

1. **`store.Load()`** — full directory walk, file reads, YAML+Markdown
   parse for every ticket. This is the prime suspect.
2. **`ticket.Parse()`** in isolation — YAML+Markdown parse of a single
   `DESCRIPTION.md` byte slice, no filesystem. Lets us see whether the
   parse itself or the I/O around it dominates `Load`.
3. **`search.matchesQuery` over a loaded slice** — the linear scan
   itself, given an already-loaded `[]ticket.Ticket`. Should be very
   fast; if it isn't, that's the surprise we want to find.
4. **`store.Find(slug)`** — hot path for `show`, `start`, `done`,
   `move`. Today it iterates buckets and calls `readTicket`; worth
   knowing how it scales.
5. **Dependency walk** — whatever the current `tickets deps` codepath
   is. Even if it's fine today, the indexing ticket may want to
   replace it, so we need a baseline.

### Fixture generator

A test helper (e.g. `internal/store/testfixture.go` behind a build tag,
or a small `cmd/genfixture` tool) that materialises N synthetic tickets
into a tempdir with realistic-ish content: variable-length descriptions,
2–4 labels each, occasional dependency edges, distributed across the
buckets. Deterministic seed so benchmark runs are comparable.

Suggested scales: **100, 1k, 10k, 100k**. 100k is past anything we
expect in practice but exposes asymptotic behaviour cheaply.

### What to report

- `go test -bench` output committed as a baseline somewhere durable —
  either a `BENCHMARKS.md` or a comment on the indexing ticket — so
  later changes can claim concrete deltas.
- For each benchmark, both ns/op and allocs/op. Allocations matter for
  `Load` because that's where a parse cache buys the most.
- A short README note on how to regenerate the fixture and rerun the
  benchmarks, so this stays reproducible after the original author
  forgets the invocation.

### Out of scope

- Optimising anything. This ticket is purely measurement; the indexing
  ticket consumes the numbers.
- CI integration / regression gates. Tempting but premature — we don't
  yet know which numbers are stable enough to gate on.

## Acceptance Criteria

- [ ] Fixture generator that produces N tickets with deterministic
  content into a given root directory.
- [ ] `Benchmark*` functions covering at minimum: `store.Load`,
  `ticket.Parse`, `matchesQuery` end-to-end search, `store.Find`, and
  the dependency walk.
- [ ] Benchmark results captured at 100 / 1k / 10k / 100k tickets,
  recorded somewhere the indexing ticket can reference.
- [ ] Brief instructions for rerunning the benchmark suite locally.

## Notes

Pairs with `improve-ticket-search-and-indexing`; that ticket's first
acceptance criterion ("benchmark fixture exists") is satisfied by this
one.

## Dependencies
