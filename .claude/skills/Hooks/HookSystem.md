# Setup and Teardown Hook System

## Overview

The `lw` script runs user-defined hook scripts at two lifecycle points: **setup** (after worktree + tmux window creation) and **teardown** (before worktree removal and window close).

## Hook locations

Hooks are discovered from two directories, in this order:

1. **Global:** `~/.lw/<phase>/` — applies to all repositories
2. **Repo-local:** `<git-root>/.lw/<phase>/` — applies to a single repository

Global hooks run first, then repo-local hooks. Within each directory, scripts are sorted lexicographically by filename.

## Naming convention

Scripts are named `s<NN><description>.sh` where `<NN>` controls execution order:
- `s10install-deps.sh` — runs early
- `s50claude.sh` — runs mid-lifecycle
- `s90cleanup.sh` — runs late

Scripts must be executable (`chmod +x`). Non-executable files are ignored.

## Hook execution (`run_hooks()`)

Defined in `scripts/lw` (function `run_hooks`). Key behaviors:

- Discovery: `find <dir> -maxdepth 1 -type f -executable -print0 | sort -z`
- Execution: each script runs in a subshell with `cd` to the worktree directory
- Failure: any non-zero exit aborts `lw` immediately via `die()`
- No-op: if no hooks are found, returns silently

## Environment variables

All hooks receive these exported variables:

| Variable | Content | Notes |
|----------|---------|-------|
| `LW_WORKTREE_DIR` | Absolute path to the worktree | |
| `LW_BRANCH` | Branch name | |
| `LW_TICKET` | Ticket ID (e.g. `ENG-123`) | Empty for review and scratch worktrees |
| `LW_REPO_NAME` | Repository directory name | e.g. `myrepo` |
| `LW_ACTION` | `setup` or `teardown` | |
| `LW_TMUX_WINDOW` | tmux window name | Empty if not running inside tmux |

When adding new environment variables:
1. Export them in the `run_hooks()` function body
2. Document them in `README.md` (hook variable table)
3. Use `${VAR:-}` pattern to avoid `set -u` failures for optional values

## Hook timing

- **Setup hooks** run after the tmux window is created, so they can use `tmux send-keys -t "$LW_TMUX_WINDOW"` to send commands to the new window.
- **Teardown hooks** run before the tmux window is closed, so they can capture pane content or gracefully stop processes.

## Example: Claude Code hooks

`examples/hooks/setup/s50claude.sh`:
- Checks for saved session in `~/.lw/sessions/<repo>/<ticket>`
- Sends `claude --dangerously-skip-permissions [--resume <id>]` to tmux window

`examples/hooks/teardown/s50claude.sh`:
- Sends `/exit` to Claude, waits 2 seconds
- Captures pane content, extracts session ID via regex
- Saves to `~/.lw/sessions/<repo>/<ticket>` for next resumption

## Session storage

The Claude hooks use `~/.lw/sessions/<repo>/<ticket-or-branch>` for persistence. The key is `$LW_TICKET` if set, otherwise `$LW_BRANCH` with `/` replaced by `-`.
