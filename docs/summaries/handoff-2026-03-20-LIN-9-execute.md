# Session Handoff: LIN-9 — Update the README to match current behavior

**Ticket:** LIN-9
**Phase:** execute
**Date:** 2026-03-20
**Status:** Implementation complete, pending review

## What Was Done

Updated README.md to reflect the current CLI behavior after VOI-56 (subcommand restructure), VOI-58 (list command), LIN-7 (multi-API-key support), and LIN-8 (verbose flag).

Key changes:
1. **Prerequisites** — updated from single `LINEAR_API_KEY` to `LINEAR_API_KEY*` pattern
2. **Configuration** — documented multi-workspace support with `~/.lw/linear.json` caching
3. **Usage** — rewrote from positional-arg style (`lw ENG-123`) to subcommand style (`lw new --ticket ENG-123`)
4. **New commands** — documented `lw new` (with all flags and naming precedence), `lw continue`, `lw list`/`lw ls`
5. **Updated commands** — `lw done` now uses session identifier resolution, `lw review` unchanged
6. **Global flags** — documented `--verbose`/`-v`
7. **Hooks example** — updated example workflow to use new command syntax

## Files Changed

| File | Change |
|------|--------|
| `README.md` | Full rewrite to match current CLI structure |
| `docs/summaries/handoff-2026-03-20-LIN-9-execute.md` | This handoff |

## Open Items

None.
