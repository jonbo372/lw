package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jonbo372/lw/internal/config"
	"github.com/jonbo372/lw/internal/git"
	"github.com/jonbo372/lw/internal/hook"
	"github.com/jonbo372/lw/internal/session"
	"github.com/jonbo372/lw/internal/tmux"
)

// cmdContinue implements the `lw continue <session_identifier>` subcommand.
// It locates an existing git worktree and opens a tmux window pointing to it.
func cmdContinue(identifier string, currentTmuxWindow bool) {
	gitRoot, err := git.MainRoot()
	if err != nil {
		die("Not inside a git repository.")
	}
	repoName := filepath.Base(gitRoot)

	entries, err := git.WorktreeList()
	if err != nil {
		die("Failed to list worktrees: %v", err)
	}

	matches := git.MatchWorktrees(entries, identifier)

	switch len(matches) {
	case 0:
		die("No worktree found matching '%s'.", identifier)
	case 1:
		// Exactly one match — proceed
	default:
		die("Multiple worktrees match '%s':\n%s\nPlease be more specific.",
			identifier, formatWorktreeList(matches))
	}

	entry := matches[0]
	worktreeDir := entry.Path
	branch := entry.Branch
	dirName := filepath.Base(worktreeDir)

	info("Continuing session in %s (branch: %s)", worktreeDir, branch)

	// tmux window
	var tmuxWindow string
	if !currentTmuxWindow {
		windowName := fmt.Sprintf("[%s] %s", repoName, dirName)
		tmuxWindow = tmuxCreateOrSwitchInfo(windowName, worktreeDir)
	}

	// Print saved Claude session ID to tmux window if available
	if tmuxWindow != "" {
		sess, err := session.Load(config.SessionsDir(), repoName, dirName)
		if err == nil && sess != nil && sess.ClaudeSessionID != "" {
			// Strip newlines to prevent command injection via tmux send-keys
			safeID := strings.ReplaceAll(strings.ReplaceAll(sess.ClaudeSessionID, "\n", ""), "\r", "")
			info("Found Claude session: %s", safeID)
			tmux.SendKeys(tmuxWindow, fmt.Sprintf("# Previous Claude session: %s", safeID))
		}
	}

	// hooks
	if err := hook.Run("setup", gitRoot, hook.Env{
		WorktreeDir: worktreeDir,
		Branch:      branch,
		RepoName:    repoName,
		Phase:       "setup",
		TmuxWindow:  tmuxWindow,
	}); err != nil {
		die("%v", err)
	}

	fmt.Println(worktreeDir)
}

func formatWorktreeList(entries []git.WorktreeEntry) string {
	var result string
	for _, e := range entries {
		if e.Branch != "" {
			result += fmt.Sprintf("  %s [%s]\n", e.Path, e.Branch)
		} else {
			result += fmt.Sprintf("  %s\n", e.Path)
		}
	}
	return result
}
