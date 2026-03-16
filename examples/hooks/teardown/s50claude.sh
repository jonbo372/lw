#!/usr/bin/env bash
# s50claude.sh — Stop Claude Code and save session ID for resumption
#
# Place this in .lw/teardown/ (repo-local) or ~/.lw/teardown/ (global).
# Requires: claude CLI, tmux

[[ -n "$LW_TMUX_WINDOW" ]] || exit 0

# Ask claude to exit gracefully
tmux send-keys -t "$LW_TMUX_WINDOW" "/exit" Enter
sleep 2

# Capture the pane content and extract the session ID
SESSION_ID=$(tmux capture-pane -t "$LW_TMUX_WINDOW" -p -S -50 \
  | grep -oP '(?<=Session: )\S+' | tail -1)

if [[ -n "$SESSION_ID" ]]; then
  SESSION_DIR="$HOME/.lw/sessions/$LW_REPO_NAME"
  mkdir -p "$SESSION_DIR"
  echo "$SESSION_ID" > "$SESSION_DIR/${LW_TICKET:-${LW_BRANCH//\//-}}"
  echo "▸ Saved Claude session: $SESSION_ID"
else
  echo "▸ No Claude session ID found — skipping save." >&2
fi
