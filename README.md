# Linear Worktrees

A CLI tool that creates isolated git worktrees for Linear tickets, each in its own tmux window.

## Prerequisites

- `go`, `make`, `git`
- `tmux` (optional — for automatic window management)
- One or more [Linear API keys](https://linear.app/settings/api) exported as `LINEAR_API_KEY*` environment variables

## Installation

```bash
git clone <repo-url> && cd lw
./install.sh
```

This builds the `lw` binary and installs it to `~/.local/bin` (Linux) or `/usr/local/bin` (macOS).

## Configuration

| Variable | Default | Description |
|---|---|---|
| `LINEAR_API_KEY*` | *(at least one required)* | Linear API keys — any env var starting with `LINEAR_API_KEY` is discovered (e.g. `LINEAR_API_KEY`, `LINEAR_API_KEY2`, `LINEAR_API_KEY_WORK`) |
| `WORKTREE_HOME` | `~/.superclaude` | Directory where worktrees are created |

### Multiple Linear workspaces

If you work across multiple Linear workspaces, export one `LINEAR_API_KEY*` env var per workspace. When `lw` encounters a ticket prefix (e.g. `LIN`, `VOI`) for the first time, it tries each key until it finds a match, then caches the association in `~/.lw/linear.json`. Subsequent lookups for the same prefix use the cached key automatically. If a cached key becomes stale, `lw` re-discovers automatically.

The cache stores only env var names (e.g. `"LINEAR_API_KEY2"`), never the actual secrets.

## Usage

Run all commands from within an existing git repository.

### Global flags

| Flag | Description |
|---|---|
| `--verbose`, `-v` | Enable detailed progress output (API key resolution, git operations) |

### Create a new worktree

```bash
lw new --ticket ENG-123
```

Fetches the branch name and title from Linear, creates (or fetches) the branch, sets up a worktree at `$WORKTREE_HOME/<repo>/<TICKET>`, and opens a tmux window for it.

#### Flags

| Flag | Description |
|---|---|
| `--ticket` | Linear ticket ID to fetch branch name from |
| `--name` | Name for branch, worktree, and tmux window |
| `--branch_name` | Branch name (overrides ticket branch) |
| `--current-tmux-window` | Stay in current tmux window instead of creating a new one |

**Naming precedence** (highest to lowest):

1. `--name` — uses value as branch, worktree directory, and tmux window name
2. `--branch_name` — uses value as branch; tmux window derived from ticket or branch
3. `--ticket` — uses Linear `gitBranchName` as branch name
4. *(none)* — generates a random silly-name for a scratch worktree

### Continue a previous session

```bash
lw continue ENG-123
```

Locates an existing worktree matching the session identifier and opens a tmux window for it. The identifier can be a ticket ID, name, or worktree path.

| Flag | Description |
|---|---|
| `--current-tmux-window` | Stay in current tmux window |

### List all sessions

```bash
lw list    # or: lw ls
```

Displays all sessions and worktrees in a colored table, cross-referencing session JSON files with `git worktree list` output.

Status labels:
- **active** — session and worktree both exist
- **orphaned** — worktree exists but no session JSON
- **dead** — session JSON exists but worktree is gone

### Review a branch

```bash
lw review feature/auth-redesign
```

Checks out an existing branch into a review worktree without needing a Linear ticket.

### Tear down a session

```bash
lw done ENG-123
```

Resolves the session identifier against `git worktree list`, removes the worktree, deletes the local branch (unless still in use by another worktree), and closes the tmux window. Warns if there are uncommitted or unpushed changes. Handles all worktree types including review worktrees.

## Hooks

You can run custom scripts automatically during setup (worktree creation) and teardown (worktree removal). Hooks can be defined in two locations:

- **Global** (`~/.lw/`) — apply to all repositories
- **Repo-local** (`.lw/` in the repo root) — apply to a single repository

Global hooks run first, then repo-local hooks. Within each location, scripts are sorted lexicographically (so `s10` runs before `s50`).

```
~/.lw/                        # global hooks
├── setup/
│   └── s10default-env.sh
└── teardown/
    └── s90cleanup.sh

<repo>/.lw/                   # repo-local hooks
├── setup/
│   ├── s10install-deps.sh
│   └── s50start-services.sh
└── teardown/
    └── s50stop-services.sh
```

Scripts are executed from within the worktree directory and must be executable (`chmod +x`).

The following environment variables are available to hook scripts:

| Variable | Description |
|---|---|
| `LW_WORKTREE_DIR` | Path to the worktree |
| `LW_BRANCH` | Branch name |
| `LW_TICKET` | Ticket ID (empty for review worktrees) |
| `LW_REPO_NAME` | Repository directory name |
| `LW_ACTION` | `setup` or `teardown` |
| `LW_TMUX_WINDOW` | tmux window name (empty if not running in tmux) |

Setup hooks run **after** the tmux window is created, so they can send commands to it via `tmux send-keys`. Teardown hooks run **before** the window is closed.

If any hook exits non-zero, `lw` aborts immediately.

### Example: auto-start Claude Code

The `examples/hooks/` directory contains a pair of hooks that automatically manage a Claude Code session per worktree.

**`setup/s50claude.sh`** — runs after the tmux window is created:
1. Checks for a saved session ID in `~/.lw/sessions/<repo>/<ticket>`
2. Sends `claude --dangerously-skip-permissions` (with `--resume <session-id>` if a previous session exists) to the tmux window via `tmux send-keys`

**`teardown/s50claude.sh`** — runs before the tmux window is closed:
1. Sends `/exit` to the tmux window to gracefully stop Claude
2. Waits for Claude to exit, then captures the pane content
3. Extracts the session ID and saves it to `~/.lw/sessions/<repo>/<ticket>`

The result: `lw new --ticket ENG-123` opens a worktree with Claude already running. `lw done ENG-123` saves the session. Next time you `lw continue ENG-123`, Claude resumes where it left off.

To install, copy into your global or repo-local hooks directory:

```bash
mkdir -p ~/.lw/setup ~/.lw/teardown
cp examples/hooks/setup/s50claude.sh ~/.lw/setup/
cp examples/hooks/teardown/s50claude.sh ~/.lw/teardown/
```
