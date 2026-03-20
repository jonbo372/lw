# Session Handoff: LIN-6 — Clean up session file on `lw done`

**Ticket:** LIN-6
**Phase:** execute
**Date:** 2026-03-20
**Status:** Implementation complete, pending review

## What Was Done

Fixed bug where `lw done` removed the worktree, branch, and tmux window but did not delete the session JSON file from `~/.lw/sessions/<repo>/`, causing stale "dead" entries in `lw ls`.

1. **`session.Delete`** — Added a `Delete(sessionsDir, repoName, sessionID)` function to the session package. Returns nil if the file doesn't exist (idempotent).
2. **Session cleanup in `cmdDone`** — After teardown hooks run but before worktree removal, `cmdDone` now calls `session.Delete` to remove the session file. Failure is non-fatal (logs a warning and continues).

## Files Changed

| File | Change |
|------|--------|
| `internal/session/session.go` | Added `Delete` function (+9 lines) |
| `internal/session/session_test.go` | Added 4 tests for `Delete`: existing session, non-existent file, permission denied, integration with `ListAll` (+96 lines) |
| `cmd/lw/done.go` | Added session cleanup call + imports for `config` and `session` packages (+11 lines) |

## Test Results

- 4 new tests, all passing
- Session package coverage: 96.7%
- `go vet ./...` clean

## Decisions

- Session cleanup is non-fatal: if the file can't be deleted, `lw done` logs a warning but continues with worktree/branch/tmux teardown. Rationale: the session file is metadata — failing to clean it up should not block the primary teardown workflow.
- Session cleanup runs after teardown hooks but before worktree removal, so hooks can still read session data if needed.

## Open Items

None.
