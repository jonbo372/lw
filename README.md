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
