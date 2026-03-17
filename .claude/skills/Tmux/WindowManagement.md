# tmux Window Management

## Overview

The `lw` script optionally creates and manages tmux windows — one per worktree. All tmux operations are gated behind `[[ -n "${TMUX:-}" ]]`, so the script degrades gracefully when not running inside tmux.

## Window naming

| Mode | Pattern | Example |
|------|---------|---------|
| Ticket | `[<repo>] <TICKET>: <title>` | `[myrepo] ENG-123: Add user auth` |
| Scratch | `[<repo>] <silly_name>: scratch` | `[myrepo] fuzzy_cobra: scratch` |
| Review | `[<repo>] review: <branch>` | `[myrepo] review: feature/auth` |

Names are truncated to 50 characters (`${TMUX_WINDOW_NAME:0:50}`).

## Window creation (setup)

1. Check if a window with the exact name already exists: `tmux list-windows -F '#{window_name}' | grep -Fx`
2. If it exists, switch to it: `tmux select-window -t "$TMUX_WINDOW_NAME"`
3. If not, create it: `tmux new-window -c "$WORKTREE_DIR" -n "$TMUX_WINDOW_NAME"`

The `-c` flag sets the initial working directory to the worktree path.

## Window teardown

1. Find the window by prefix match: `tmux list-windows -F '#{window_index} #{window_name}'` piped through `awk` matching against `$WINDOW_PREFIX`
2. If found, kill it: `tmux kill-window -t "$TMUX_WINDOW_INDEX"`
3. If not found, log and continue (non-fatal)

The teardown uses `WINDOW_PREFIX` (e.g. `[myrepo] ENG-123`) for matching rather than exact name, since the title portion may vary.

## Window reference in hooks

The window name is exported as `LW_TMUX_WINDOW` for hook scripts. Hooks can use this to:
- Send keystrokes: `tmux send-keys -t "$LW_TMUX_WINDOW" "command" Enter`
- Capture output: `tmux capture-pane -t "$LW_TMUX_WINDOW" -p`

Setup hooks run **after** window creation. Teardown hooks run **before** window close.

## Key tmux commands used

| Command | Purpose |
|---------|---------|
| `tmux list-windows -F '#{window_name}'` | List all window names in current session |
| `tmux select-window -t <name>` | Switch to existing window |
| `tmux new-window -c <dir> -n <name>` | Create new window with working dir and name |
| `tmux kill-window -t <index>` | Close window by index |
| `tmux send-keys -t <name> <keys> Enter` | Send keystrokes to a window (used by hooks) |
| `tmux capture-pane -t <name> -p -S -N` | Capture last N lines of pane output (used by hooks) |
