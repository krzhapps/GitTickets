---
title: Improve ticket search and indexing
status: pending
priority: low
created: "2026-04-28"
labels:
    - search
    - performance
---

## Description

`internal/cli/search.go` runs a linear `strings.Contains` scan over every
ticket's title, description, notes, and labels. The scan itself is cheap;
the real cost is `internal/store/load.go` — every CLI invocation walks
all five bucket directories, reads each `<slug>/DESCRIPTION.md`, and runs
the YAML+Markdown parser on every ticket. At a few hundred tickets this
is invisible; somewhere between 5k and 10k tickets the per-command load
time starts to dominate, and any feature that needs richer queries
(dependency resolution, ranked search, structured filters) makes it
worse.

This ticket captures the design space rather than committing to an
approach. The framing that matters: **the filesystem is the source of
truth (git tracks it), so any index/cache must be derived state and must
handle external mutation** (`git pull`, `git checkout`, manual edits,
merge conflicts can all change tickets without going through `Store`).

### Approach 1 — Cache the parse, keep the linear search

Smallest possible change. Add an on-disk cache (gob or JSON) at
`tasks/.index.cache`, keyed by `{slug → (mtime, parsed Ticket)}`. On
`Load`, stat each `DESCRIPTION.md`, reparse only the entries whose mtime
moved, rewrite the cache. Search continues to be the existing linear
scan over the in-memory slice.

- Pros: zero new dependencies, trivial to reason about, reversible. Buys
  most of the practical win because the bottleneck is parsing not
  searching. Probably gets the tool to ~100k tickets before search
  itself becomes visible.
- Cons: doesn't help dependency resolution (still O(N) graph rebuild),
  doesn't give ranked search or structured queries, mtime invalidation
  has known edge cases (some tools rewrite mtimes on checkout).

### Approach 2 — Custom in-memory inverted index on top of (1)

Same on-disk parse cache as Approach 1, plus an in-memory inverted
index built lazily: `label → []slug`, `token → []slug` with simple
tokenization (lowercase, split on non-alphanumeric). Persist the
inverted index alongside the parse cache so we don't rebuild it on every
invocation.

- Pros: still no new deps. Faster than linear once N is large enough.
  Gives a clean place to plug in label/status/priority filters.
- Cons: you reimplement tokenization, prefix matching, and ranking
  yourself — that's the part of FTS engines that's tedious to get
  right. Persisted index format becomes a thing you have to version.
  Dependency queries are still ad-hoc map walks.

### Approach 3 — SQLite + FTS5 as a derived cache

One file at `tasks/.index.db` (gitignored). Schema roughly:

```sql
CREATE TABLE tickets (
    slug TEXT PRIMARY KEY,
    title TEXT, status TEXT, priority TEXT,
    created TEXT, assignee TEXT,
    description TEXT, notes TEXT,
    mtime INTEGER, content_hash TEXT
);
CREATE TABLE labels      (slug TEXT, label TEXT);
CREATE TABLE deps        (slug TEXT, depends_on TEXT);
CREATE VIRTUAL TABLE tickets_fts USING fts5(
    title, description, notes, content='tickets'
);
```

Refresh: stat every `DESCRIPTION.md`, reparse + upsert rows whose mtime
changed, drop rows for slugs no longer present. Search becomes an FTS5
query; `tickets list --status in-progress --label infra` becomes a real
SQL query; dependency tree becomes a recursive CTE.

- Pros: ranked full-text search, structured filters, dependency
  resolution all fall out of one engine. Mature, well-documented, no
  custom tokenizer/ranker code. Scales well beyond what this tool is
  ever likely to see.
- Cons: a real dependency (use `modernc.org/sqlite` to stay cgo-free —
  worth verifying its FTS5 support before committing). A schema you
  have to version and migrate. The `.db` file *feels* authoritative,
  which makes "it drifted from the filesystem" bugs more likely to
  confuse users than a flat cache file would.

### Recommendation (subject to revisit)

Land Approach 1 first as a small, reversible change and measure where
the cliff actually is. Graduate to Approach 3 specifically when the
driver is **dependency resolution or structured queries**, not search
speed — that's what justifies the schema. Approach 2 is a viable
middle ground but ends up being most of the engineering of (3) without
the query power.

### Cross-cutting concerns (apply to all three)

- **Invalidation policy.** Default to mtime; consider a content-hash
  fallback for environments where mtimes are unreliable. On schema/
  format version mismatch, do a full rebuild rather than trying to
  migrate.
- **Source-of-truth discipline.** Reads can hit the cache. Writes
  (`Create`, `Save`, `Move`) must update the filesystem first, then
  invalidate or update the cache. A corrupt cache should never block a
  valid filesystem operation — log and rebuild.
- **Gitignore.** Whatever the cache file is, it goes in `.gitignore`
  (and arguably in `tasks/.gitignore` so it travels with the layout).
- **`tickets validate` interaction.** Validate should still read from
  disk, not the cache, so it can flag drift.
- **CI / fresh clones.** First invocation after clone must work without
  the cache present. Build it lazily on demand; don't require an
  explicit `tickets reindex` step.

## Acceptance Criteria

- [ ] Decide which approach to pursue (or explicitly defer) and record
  the decision in this ticket's Notes.
- [ ] If proceeding: file a follow-up implementation ticket scoped to
  the chosen approach, depending on this one.
- [ ] Benchmark fixture exists (script that generates N synthetic
  tickets) so future perf claims are measured, not asserted.

## Notes

Triggered by a brainstorm on 2026-04-28 about `internal/cli/search.go`
becoming sluggish at scale. Key reframing from that conversation: the
search loop isn't the bottleneck — `Load()` is — so optimizing the
inner loop without addressing parse cost is misdirected effort.

## Dependencies

- `benchmark-ticket-store-hot-paths` — need baseline numbers before
  picking a design; otherwise we're guessing which layer to optimise.
