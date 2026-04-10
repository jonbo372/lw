# Domain Review Checklist

Domain-specific review items that `code-reviewer` should check in addition to the
standard checklist. Each item should be a concrete, checkable concern.

- Shell: does the script handle unset variables safely (`set -euo pipefail` respected)?
- Git: are worktree creation and teardown paths using consistent directory conventions?
- Hooks: are new LW_* env vars exported in run_hooks() and documented in README?
- API: does the Linear GraphQL query handle error/empty responses before using values?
- tmux: does the feature degrade gracefully when $TMUX is unset?
