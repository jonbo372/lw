# Domain Skills

Domain-specific skill files that agents should load based on ticket keywords.
Each entry maps a set of keywords to a skill file path (relative to `.claude/skills/`).
Used by `codebase-explorer`, `architect`, and `backend-engineer` to load relevant context.

| Keywords | Skill File | Description |
|----------|-----------|-------------|
| worktree, branch, git, checkout, fetch, merge | Git/Worktrees.md | Git worktree lifecycle and branch management |
| hook, setup, teardown, LW_*, s10, s50 | Hooks/HookSystem.md | Setup/teardown hook execution and env vars |
| Linear, API, GraphQL, ticket, branchName | Integration/LinearAPI.md | Linear GraphQL API integration |
| tmux, window, session, pane | Tmux/WindowManagement.md | tmux session lifecycle management |
