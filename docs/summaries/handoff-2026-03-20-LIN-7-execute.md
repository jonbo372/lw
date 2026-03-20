# Session Handoff: LIN-7 — Add support for many LINEAR_API_KEY

**Ticket:** LIN-7
**Phase:** execute
**Date:** 2026-03-20
**Status:** Implementation complete, pending review

## What Was Done

Added multi-API-key support so `lw` can work across multiple Linear workspaces. Previously, only a single `LINEAR_API_KEY` env var was supported. Now, all env vars matching `LINEAR_API_KEY*` are discovered, and the correct key is resolved per ticket prefix (e.g., `LIN`, `VOI`) with automatic discovery and caching in `~/.lw/linear.json`.

1. **`config.LinearAPIKeys()`** — New function that discovers all `LINEAR_API_KEY*` env vars with non-empty values, returning sorted env var names for deterministic ordering.
2. **`linear.Resolver`** — New struct with injectable fetcher function. `ResolveAndFetch(ticketID)` checks cached prefix→env-var-name mappings in `~/.lw/linear.json`, falls back to discovery across all keys, and handles stale mappings by re-discovering.
3. **Updated `create.go`** — Replaced all 3 call sites that used `config.LinearAPIKey()` + `linear.FetchTicket()` with a unified `fetchTicket()` helper that uses the Resolver.

## Files Changed

| File | Change |
|------|--------|
| `internal/config/config.go` | Added `LinearAPIKeys()` function (+16 lines) |
| `internal/config/config_test.go` | Added 4 tests for `LinearAPIKeys` + `clearLinearAPIKeys` helper (+58 lines) |
| `internal/linear/resolver.go` | New file: `Resolver` struct, `ResolveAndFetch`, `extractPrefix`, `loadConfig`, `saveConfig`, `DefaultConfigPath` (111 lines) |
| `internal/linear/resolver_test.go` | New file: 11 tests covering discovery, caching, stale re-discovery, error cases, first-match-wins, secret-never-stored (317 lines) |
| `cmd/lw/create.go` | Replaced 3 direct API call sites with `fetchTicket()` helper using Resolver (-25/+13 lines) |

## Test Results

- 11 new resolver tests + 4 new config tests, all passing
- `internal/config` coverage: 93.3%
- `internal/linear` coverage: 58.6% (low due to pre-existing untested `client.go` HTTP code, not part of this ticket)
- New resolver code: ~100% coverage
- `go vet ./...` clean, `go build ./...` clean

## Decisions

- `linear.json` stores only env var names (e.g., `"LINEAR_API_KEY2"`), never secrets.
- `loadConfig`/`saveConfig` silently handle errors (corrupt JSON → re-discover). Rationale: the mapping file is a cache; losing it just triggers re-discovery.
- Env var names are sorted for deterministic "first match wins" behavior.
- The existing `config.LinearAPIKey()` function is kept for backward compatibility but is no longer used in any call site.

## Open Items

None.
