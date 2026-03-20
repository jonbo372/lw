# Session Handoff: LIN-8 — Add --verbose flag for detailed output

**Ticket:** LIN-8
**Phase:** execute
**Date:** 2026-03-20
**Status:** Implementation complete, pending review

## What Was Done

Added a global `--verbose` / `-v` flag to the `lw` CLI that enables detailed progress output during operations. Without the flag, `lw` remains silent (errors only). The flag is wired into the Linear API key resolver so it prints discovery progress (e.g., "Trying API keys for prefix LIN...").

1. **Global flag** — Added `--verbose` / `-v` as a `PersistentFlag` on the root Cobra command in `main.go`, with a `PersistentPreRun` that calls `setVerbose()`.
2. **`verbose()` function in `ui.go`** — New package-level `verboseMode` bool, with `setVerbose()`, `isVerbose()`, and `verbose()` functions. `verbose()` prints to stderr with "▸ " prefix only when enabled.
3. **Resolver integration** — Added `VerboseFunc` callback type and `SetVerbose()` method to `linear.Resolver`. The resolver now logs when using cached keys, trying discovery, and iterating over env vars.
4. **Wiring in `create.go`** — `fetchTicket()` now calls `resolver.SetVerbose(verbose)` when `isVerbose()` is true, connecting the UI layer to the resolver.

## Files Changed

| File | Change |
|------|--------|
| `cmd/lw/main.go` | Added `--verbose` / `-v` PersistentFlag + PersistentPreRun |
| `cmd/lw/ui.go` | Added `verboseMode` var, `setVerbose()`, `isVerbose()`, `verbose()` |
| `cmd/lw/ui_test.go` | New file: 3 tests for verbose output enabled/disabled and state functions |
| `cmd/lw/cobra_test.go` | Added 4 tests: default value, --verbose, -v, availability on subcommands |
| `cmd/lw/create.go` | Wired verbose callback into `fetchTicket()` |
| `internal/linear/resolver.go` | Added `VerboseFunc` type, `SetVerbose()` method, `logVerbose()` helper, verbose calls in `ResolveAndFetch` |
| `internal/linear/resolver_verbose_test.go` | New file: 5 tests covering discovery logging, cached hit, nil safety, stale rediscovery, method chaining |

## Test Results

- All 7 testable packages pass (`go test ./...`)
- 12 new tests (3 ui + 4 cobra + 5 resolver verbose)
- 100% coverage on new code
- `go build ./...` clean

## Decisions

- Used package-level `verboseMode` var (matches existing `die()`/`info()` pattern — CLI is single-threaded).
- Resolver uses an optional `VerboseFunc` callback rather than importing the `cmd` package, keeping the internal/cmd dependency direction clean.
- Only the resolver has verbose output wired up. Other operations (git, tmux, session) can be added in future tickets.

## Open Items

None.
