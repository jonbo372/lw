package main

import (
	"fmt"
	"path/filepath"

	"github.com/jonbo372/lw/internal/git"
	"github.com/jonbo372/lw/internal/hook"
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
