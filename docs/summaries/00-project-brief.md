# Project Brief: Linear Worktrees (lw)

## What it is

A CLI tool (`lw`) that creates isolated git worktrees for Linear tickets, each in its own tmux window. It bridges Linear's ticket system with local git workflow by automatically fetching branch metadata from the Linear API and managing the full lifecycle of worktrees.

## Core capabilities

1. **Ticket worktrees** — `lw <TICKET-ID>` fetches branch name and title from Linear, creates/fetches the branch, sets up a worktree at `$WORKTREE_HOME/<repo>/<TICKET>`, and opens a tmux window.
2. **Scratch worktrees** — `lw` (no ticket) generates a random branch name (`adjective_noun`) for quick, untracked work. Accepts an optional base branch or `--current`.
3. **Review worktrees** — `lw review <branch>` checks out an existing branch into a review worktree without needing a Linear ticket.
4. **Teardown** — `lw done <TICKET-ID|name>` removes the worktree, deletes the local branch, and closes the tmux window. Warns on uncommitted/unpushed changes.
5. **Hooks** — extensible setup/teardown hooks in `~/.lw/` (global) and `.lw/` (repo-local), executed in sort order. Hooks receive env vars (`LW_WORKTREE_DIR`, `LW_BRANCH`, `LW_TICKET`, `LW_REPO_NAME`, `LW_ACTION`, `LW_TMUX_WINDOW`).

## Tech stack

- Pure Bash (single script: `scripts/lw`)
- Dependencies: `git`, `curl`, `jq`, optionally `tmux`
- Linear GraphQL API for ticket metadata
- Git worktrees for branch isolation

## Configuration

| Variable | Default | Description |
|---|---|---|
| `LINEAR_API_KEY` | *(required for ticket mode)* | Linear personal API key |
| `WORKTREE_HOME` | `~/.superclaude` | Directory where worktrees are created |

## Installation

`install.sh` copies the `lw` script to `~/.local/bin` (Linux) or `/usr/local/bin` (macOS).

## Repository layout

```
scripts/lw          — the main CLI script
install.sh          — installer
examples/hooks/     — example hook scripts (e.g. auto-start Claude Code)
.lw/                — repo-local hooks (setup/, teardown/)
docs/               — project documentation
templates/          — summary and handoff templates
```

## Current state

- Core functionality is stable and in use.
- Recent additions: scratch worktree mode (no ticket required), optional base branch selection (`--current` / explicit branch).
- No automated tests.
- No CI/CD.
