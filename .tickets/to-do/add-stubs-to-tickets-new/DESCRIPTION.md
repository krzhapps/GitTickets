---
title: Scaffold body stubs when running tickets new
status: pending
priority: medium
created: "2026-04-28"
labels:
    - enhancement
---

## Description

When `tickets new` creates a `DESCRIPTION.md`, the file body is empty. Users are expected to fill in the standard sections manually. The CLI should scaffold the four standard stubs automatically so that a newly created ticket is immediately ready to edit without boilerplate hunting.

## Acceptance Criteria

- [ ] `tickets new` writes `## Description`, `## Acceptance Criteria`, `## Notes`, and `## Dependencies` sections into the generated `DESCRIPTION.md`
- [ ] Each section contains a minimal placeholder (empty bullet or ellipsis) so editors don't collapse it
- [ ] Existing behaviour of `--no-edit` (skip opening `$EDITOR`) is unchanged
- [ ] `tickets validate` passes on the scaffolded file without modifications

## Notes

The stubs should match the shape documented in the frontmatter schema section of the skill README, i.e.:
```
## Description
…

## Acceptance Criteria
- [ ] …

## Notes
…

## Dependencies
- `<other-slug>` — why
```

## Dependencies
