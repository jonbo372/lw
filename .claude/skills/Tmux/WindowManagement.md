# tmux Session Management

## Overview

The `lw` tool creates and manages tmux sessions — one per worktree. All tmux operations are gated behind `[[ -n "${TMUX:-}" ]]`, so the tool degrades gracefully when not running inside tmux.

## Session naming

| Mode | Pattern | Example |
|------|---------|---------|
| Ticket | `[<repo>] <TICKET>- <title>` | `[myrepo] ENG-123- Add user auth` |
| Scratch | `[<repo>] <silly_name>- scratch` | `[myrepo] fuzzy_cobra- scratch` |
| Review | `[<repo>] review- <branch>` | `[myrepo] review- feature/auth` |

Names are truncated to 50 characters and sanitized (colons and periods replaced with hyphens, since they are special in tmux target syntax).

## Session creation (setup)

1. Sanitize the name (replace `:` and `.` with `-`)
2. Check if a session with the exact name already exists: `tmux has-session -t "=<name>"`
3. If it exists, switch to it: `tmux switch-client -t "<name>"`
4. If not, create it: `tmux new-session -d -s "<name>" -c "<worktreeDir>"` then `tmux switch-client -t "<name>"`

The `-c` flag sets the initial working directory to the worktree path. The `-d` flag creates the session detached, then `switch-client` moves the current client to it.

## Session teardown

1. Find the session by prefix match: `tmux list-sessions -F '#{session_name}'` filtered by `strings.Contains`
2. If found, kill it: `tmux kill-session -t "=<name>"` (exact match via `=` prefix)
3. If not found, log and continue (non-fatal)

The teardown uses a prefix (e.g. `[myrepo] ENG-123`) for matching rather than exact name, since the title portion may vary.

## Session reference in hooks

The session name is exported as `LW_TMUX_SESSION` for hook scripts. Hooks can use this to:
- Send keystrokes: `tmux send-keys -t "$LW_TMUX_SESSION" "command" Enter`
- Capture output: `tmux capture-pane -t "$LW_TMUX_SESSION" -p`

Setup hooks run **after** session creation. Teardown hooks run **before** session close.

## Key tmux commands used

| Command | Purpose |
|---------|---------|
| `tmux has-session -t "=<name>"` | Check if session exists (exact match) |
| `tmux new-session -d -s <name> -c <dir>` | Create new detached session with working dir |
| `tmux switch-client -t <name>` | Switch current client to session |
| `tmux list-sessions -F '#{session_name}'` | List all session names |
| `tmux kill-session -t "=<name>"` | Close session by name (exact match) |
| `tmux send-keys -t <name> <keys> Enter` | Send keystrokes to a session (used by hooks) |
| `tmux capture-pane -t <name> -p -S -N` | Capture last N lines of pane output (used by hooks) |
