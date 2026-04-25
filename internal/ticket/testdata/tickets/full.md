---
title: Better Google OAuth errors
status: pending
priority: high
created: "2026-04-25"
labels:
    - auth
    - oauth
assignee: alice
---

## Description

Surface friendlier error messages when Google OAuth fails so users can self-correct instead of contacting support.

## Acceptance Criteria

- [ ] Map common Google error codes to user-facing strings
- [x] Add structured logging for OAuth failures

## Notes

Reference: https://developers.google.com/identity/protocols/oauth2

## Dependencies

- `auth-session-store` — needs new error types
