# Linear Worktrees

A CLI tool that creates isolated git worktrees for Linear tickets, each in its own tmux window.

## Prerequisites

- `git`, `curl`, `jq`
- `tmux` (optional — for automatic window management)
- A [Linear API key](https://linear.app/settings/api) exported as `LINEAR_API_KEY`

## Installation

```bash
git clone <repo-url> && cd lw
./install.sh
```

This copies the `lw` script to `~/.local/bin` (Linux) or `/usr/local/bin` (macOS).

## Configuration

| Variable | Default | Description |
|---|---|---|
| `LINEAR_API_KEY` | *(required)* | Your Linear personal API key |
| `WORKTREE_HOME` | `~/.superclaude` | Directory where worktrees are created |

## Usage

Run all commands from within an existing git repository.

### Start work on a Linear ticket

```bash
lw ENG-123
```

Fetches the branch name and title from Linear, creates (or fetches) the branch, sets up a worktree at `$WORKTREE_HOME/<repo>/<TICKET>`, and opens a tmux window for it.

### Review a branch

```bash
lw review feature/auth-redesign
```

Checks out an existing branch into a review worktree without needing a Linear ticket.

### Tear down a ticket worktree

```bash
lw done ENG-123
```

Removes the worktree, deletes the local branch, and closes the tmux window. Warns if there are uncommitted or unpushed changes.

### Tear down a review worktree

```bash
lw done review feature/auth-redesign
```

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

The result: `lw ENG-123` opens a worktree with Claude already running. `lw done ENG-123` saves the session. Next time you `lw ENG-123`, Claude resumes where it left off.

To install, copy into your global or repo-local hooks directory:

```bash
mkdir -p ~/.lw/setup ~/.lw/teardown
cp examples/hooks/setup/s50claude.sh ~/.lw/setup/
cp examples/hooks/teardown/s50claude.sh ~/.lw/teardown/
```
