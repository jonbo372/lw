# Project Brief: Linear Worktrees (lw)

## What it is

A CLI tool (`lw`) that creates isolated git worktrees for Linear tickets, each in its own tmux window. It bridges Linear's ticket system with local git workflow by automatically fetching branch metadata from the Linear API and managing the full lifecycle of worktrees.

## Core capabilities

1. **Ticket worktrees** — `lw new --ticket ENG-123` fetches branch name and title from Linear, creates/fetches the branch, sets up a worktree at `$WORKTREE_HOME/<repo>/<TICKET>`, and opens a tmux window.
2. **Scratch worktrees** — `lw new` (no flags) generates a random branch name (`adjective_noun`) for quick, untracked work. Supports `--name` and `--branch_name` overrides.
3. **Review worktrees** — `lw review <branch>` checks out an existing branch into a review worktree without needing a Linear ticket.
4. **Continue sessions** — `lw continue <identifier>` locates an existing worktree and opens a tmux window for it.
5. **Teardown** — `lw done <identifier>` removes the worktree, deletes the local branch, cleans up the session file, and closes the tmux window. Warns on uncommitted/unpushed changes.
6. **List sessions** — `lw list` (or `lw ls`) displays all sessions and worktrees in a colored table with status labels (active, orphaned, dead).
7. **Session storage** — Persists session state as JSON files, enabling continue/list workflows and Claude Code session ID tracking.
8. **Hooks** — extensible setup/teardown hooks in `~/.lw/` (global) and `.lw/` (repo-local), executed in sort order. Hooks receive env vars (`LW_WORKTREE_DIR`, `LW_BRANCH`, `LW_TICKET`, `LW_REPO_NAME`, `LW_ACTION`, `LW_TMUX_WINDOW`).
9. **Multi-workspace support** — Discovers all `LINEAR_API_KEY*` env vars, tries each until it finds a match for a ticket prefix, and caches the association in `~/.lw/linear.json`.

## Tech stack

- Go (CLI built with [Cobra](https://github.com/spf13/cobra))
- Internal packages: `config`, `git`, `hook`, `linear`, `namegen`, `session`, `tmux`, `claudehook`
- Linear GraphQL API for ticket metadata
- Git worktrees for branch isolation
- Build via `make` (see `Makefile`)

## Configuration

| Variable | Default | Description |
|---|---|---|
| `LINEAR_API_KEY*` | *(at least one required)* | Linear API keys — any env var starting with `LINEAR_API_KEY` is discovered |
| `WORKTREE_HOME` | `~/.superclaude` | Directory where worktrees are created |

## Installation

`install.sh` builds the Go binary and installs it to `~/.local/bin` (Linux) or `/usr/local/bin` (macOS).

## Repository layout

```
cmd/lw/             — CLI entry point and subcommand handlers
internal/           — internal Go packages
  config/           — configuration loading
  git/              — git and worktree operations
  hook/             — setup/teardown hook execution
  linear/           — Linear API client and key resolver
  namegen/          — random name generation for scratch worktrees
  session/          — session JSON storage
  tmux/             — tmux window management
  claudehook/       — Claude Code hook integration
install.sh          — installer
Makefile            — build targets
examples/hooks/     — example hook scripts (e.g. auto-start Claude Code)
.lw/                — repo-local hooks (setup/, teardown/)
docs/               — project documentation
templates/          — summary and handoff templates
```

## Current state

- Core functionality is stable and in use.
- Go rewrite complete with Cobra-based subcommand structure (`new`, `continue`, `done`, `review`, `list`, `session-end`).
- Session storage and Claude Code hook integration implemented.
- Multi-workspace Linear API key support with auto-discovery and caching.
- `--verbose` flag for detailed progress output.
- Has unit tests for key packages.
- No CI/CD.
