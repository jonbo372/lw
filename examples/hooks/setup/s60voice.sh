#!/usr/bin/env bash
# s60voice.sh — Set up a voice project in a split tmux pane
#
# Place this in .lw/setup/ (repo-local) or ~/.lw/setup/ (global).
# Requires: claude CLI, tmux, ~/development/aicodeflow/install.sh

[[ -n "$LW_TMUX_WINDOW" ]] || exit 0
[[ -n "$LW_WORKTREE_DIR" ]] || exit 0

if [[ "$LW_REPO_NAME" == "voicekit" ]]; then
  WORK_DIR="$LW_WORKTREE_DIR/voice"
else
  WORK_DIR="$LW_WORKTREE_DIR"
fi

# Install the project
cd "$WORK_DIR" || exit 1
~/development/aicodeflow/install.sh "$WORK_DIR"

# Resolve the window index by name (avoids special-char issues in tmux targets)
WIN_ID=$(tmux list-windows -F '#{window_id} #{window_name}' \
  | grep -F "$LW_TMUX_WINDOW" | head -1 | awk '{print $1}')
[[ -n "$WIN_ID" ]] || exit 1

# Start Claude in the current pane
tmux send-keys -t "$WIN_ID" "cd '$WORK_DIR' && claude --dangerously-skip-permissions" Enter

# Split the window and cd into the work directory
tmux split-window -t "$WIN_ID" -v
tmux send-keys -t "$WIN_ID" "cd '$WORK_DIR'" Enter
