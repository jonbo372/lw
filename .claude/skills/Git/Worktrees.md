# Git Worktree Lifecycle and Branch Management

## Overview

The `lw` script manages git worktrees as isolated working directories, one per ticket, scratch session, or review. All worktree and branch operations live in `scripts/lw`.

## Directory convention

Worktrees are created at `$WORKTREE_HOME/<repo-name>/<label>`:

| Mode | Label | Example path |
|------|-------|-------------|
| Ticket | `SAFE_TICKET` (ticket ID, `/` → `-`) | `~/.superclaude/myrepo/ENG-123` |
| Scratch | silly name (e.g. `fuzzy_cobra`) | `~/.superclaude/myrepo/fuzzy_cobra` |
| Review | `review-<safe-branch>` (`/` → `-`) | `~/.superclaude/myrepo/review-feature-auth` |

Both creation and teardown paths must agree on this convention. If you change how `SAFE_LABEL` / `SAFE_DIR` is derived, update both.

## Branch resolution order

When creating a ticket worktree, the branch is resolved in this priority:

1. If the branch already exists locally (`refs/heads/<branch>`), use it as-is.
2. If a `BASE_BRANCH` was provided (explicit branch or `--current`), create the branch off that base.
3. Try to fetch the branch from origin (`git fetch origin <branch>:<branch>`).
4. Fall back to detecting the default branch: `origin/HEAD` → probe `main`, `master`, `trunk`, `develop`.

For scratch worktrees, the branch is always created new (random name), so step 1 never applies.

## Key git commands used

- `git worktree add <path> <branch>` — create worktree
- `git worktree remove --force <path>` — remove worktree
- `git worktree prune` — clean up stale worktree references
- `git worktree list --porcelain` — check which branches are in use (teardown safety)
- `git show-ref --verify --quiet refs/heads/<branch>` — check if local branch exists
- `git symbolic-ref refs/remotes/origin/HEAD` — detect default branch
- `git fetch origin <branch>:<branch>` — fetch a specific branch
- `git branch <new> <base>` — create branch off base
- `git branch -D <branch>` — force-delete local branch on teardown

## Teardown safety checks

Before removing a worktree, the script checks:
1. **Uncommitted changes** — `git -C <worktree> status --porcelain`. Prompts user to confirm.
2. **Unpushed commits** — compares `HEAD` vs `origin/<branch>`. Skipped for review worktrees. Prompts user to confirm.
3. **Branch in use** — for reviews, checks `git worktree list --porcelain` to avoid deleting a branch used by another worktree.

## Ticket ID detection

First positional argument is classified as a ticket ID if it matches `^[A-Za-z]+-[0-9]+` (e.g. `ENG-123`, `PROJ-42`). Anything else is treated as a branch name, triggering scratch mode.
