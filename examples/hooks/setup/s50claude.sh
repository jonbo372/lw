#!/usr/bin/env bash
# s50claude.sh — Start Claude Code in the tmux window
#
# Place this in .lw/setup/ (repo-local) or ~/.lw/setup/ (global).
# Requires: claude CLI, tmux

[[ -n "$LW_TMUX_WINDOW" ]] || exit 0

SESSION_DIR="$HOME/.lw/sessions/$LW_REPO_NAME"
SESSION_FILE="$SESSION_DIR/${LW_TICKET:-${LW_BRANCH//\//-}}"

CMD="claude --dangerously-skip-permissions"
if [[ -f "$SESSION_FILE" ]]; then
  CMD="$CMD --resume $(cat "$SESSION_FILE")"
fi

tmux send-keys -t "$LW_TMUX_WINDOW" "$CMD" Enter
