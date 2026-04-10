#!/usr/bin/env bash
# — Set up a project in a split tmux pane
#
# Place this in .lw/setup/ (repo-local) or ~/.lw/setup/ (global).
# Requires: claude CLI, tmux, ~/development/aicodeflow/install.sh

[[ -n "$LW_TMUX_SESSION" ]] || exit 0
[[ -n "$LW_WORKTREE_DIR" ]] || exit 0

if [[ "$LW_REPO_NAME" == "courier-ingress" ]]; then
  WORK_DIR="$LW_WORKTREE_DIR/api-gateway"
else
  WORK_DIR="$LW_WORKTREE_DIR"
fi

# Install the project
cd "$WORK_DIR" || exit 1
~/development2/aicodeflow/install.sh "$WORK_DIR"

# Build the Claude command, resuming a previous session if one exists
SESSION_DIR="$HOME/.lw/sessions/$LW_REPO_NAME"
SESSION_FILE="$SESSION_DIR/${LW_TICKET:-${LW_BRANCH//\//-}}"

CMD="cd '$WORK_DIR' && claude --dangerously-skip-permissions"
if [[ -f "$SESSION_FILE" ]]; then
  CMD="$CMD --resume $(cat "$SESSION_FILE")"
fi

# Start Claude in the current pane
tmux send-keys -t "$LW_TMUX_SESSION" "$CMD" Enter

# Split the session window and cd into the work directory
tmux split-window -t "$LW_TMUX_SESSION" -v
tmux send-keys -t "$LW_TMUX_SESSION" "cd '$WORK_DIR'" Enter
