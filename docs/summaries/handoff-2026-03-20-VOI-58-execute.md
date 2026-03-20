# Session Handoff: VOI-58 — Add a list all sessions command to `lw`

**Ticket:** VOI-58
**Phase:** execute
**Date:** 2026-03-20
**Status:** Implementation complete, pending review

## What Was Done

Implemented `lw list` (alias `lw ls`) command that displays all sessions and worktrees in a colored table:

1. **Session-worktree cross-referencing** — Reads session JSON files from `~/.lw/sessions/<repo>/`, cross-references against `git worktree list` output, and classifies each entry as active, orphaned, or dead.
2. **Colored table output** — Renders a formatted text table with columns: Name, Branch, Worktree Path, Status, Claude Session ID. Status labels are color-coded: green (active), yellow (orphaned), red (dead).
3. **Main worktree entry** — The main repo worktree is always shown first as a separate entry with "main" status.
4. **`session.ListAll` and `session.DefaultSessionsDir`** — New functions in the session package to enumerate all session JSON files and resolve the default sessions directory.
5. **Dependency injection** — Core logic in `cmdListWithDeps` accepts injected functions for testability, following the existing codebase pattern.

## Files Changed

| File | Change |
|------|--------|
| `cmd/lw/list.go` | **NEW** — `lw list`/`lw ls` command: `buildListEntries`, `renderTable`, `statusLabel`, `statusLabelPlain`, `maxLen`, `cmdListWithDeps` (262 lines) |
| `cmd/lw/list_test.go` | **NEW** — 27 tests covering Cobra registration, alias, all status classifications, table rendering, dependency injection error paths, empty state (448 lines) |
| `cmd/lw/main.go` | Registered `newListCmd()` in the command tree (+1 line import change) |
| `internal/session/session.go` | Added `ListAll(sessionsDir, repoName)` and `DefaultSessionsDir()` functions (+44 lines) |
| `internal/session/session_test.go` | Added 7 tests for `ListAll` and `DefaultSessionsDir` (124 lines) |

## Test Results

- 34 new tests, all passing
- Full suite: `go test ./...` — 0 failures
- Coverage: cmd/lw 33.4%, session 96.5%, claudehook 96.4%, config 100%, tmux 10.0%
- All new testable functions at 100%. Only uncovered function is `cmdList` (12-line Cobra entrypoint wiring real dependencies), matching the existing pattern for all other commands.

## Decisions

- Entry ordering: main worktree first, then active sessions, then dead sessions, then orphaned worktrees. Within groups, order is non-deterministic (Go map iteration).
- `session.ListAll` depends on `internal/config` for `DefaultSessionsDir()` — one-directional dependency, no cycle risk.
- Informational only — no cleanup actions or prompts for dead/orphaned entries (deferred to future command).

## Open Items

- **OPEN:** `docs/context/git-conventions.md` still does not exist — referenced by CLAUDE.md pre-flight checks
