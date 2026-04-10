package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jonbo372/lw/internal/config"
	"github.com/jonbo372/lw/internal/git"
	"github.com/jonbo372/lw/internal/hook"
	"github.com/jonbo372/lw/internal/session"
	"github.com/jonbo372/lw/internal/tmux"
)

// cmdDone implements the `lw done <session_identifier>` subcommand.
// It resolves the session identifier against `git worktree list`,
// warns about uncommitted/unpushed changes, removes the worktree,
// and kills the associated tmux session.
// Handles all worktree types including review worktrees.
func cmdDone(identifier string) {
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

	// Warn about uncommitted changes
	if git.DirExists(worktreeDir) && git.IsDirty(worktreeDir) {
		fmt.Fprintln(os.Stderr, "warning: worktree has uncommitted changes:")
		fmt.Fprintln(os.Stderr, git.ShortStatus(worktreeDir))
		if !confirm("Continue with teardown anyway?") {
			info("Aborted.")
			os.Exit(0)
		}
	}

	// Warn about unpushed commits
	if branch != "" && git.DirExists(worktreeDir) {
		if git.HasUnpushedCommits(worktreeDir, branch) {
			fmt.Fprintf(os.Stderr, "warning: branch '%s' has commits not pushed to origin.\n", branch)
			if !confirm("Continue with teardown anyway?") {
				info("Aborted.")
				os.Exit(0)
			}
		}
	}

	// Resolve tmux session
	sessionPrefix := fmt.Sprintf("[%s] %s", repoName, dirName)
	tmuxSessionName := tmux.FindSession(sessionPrefix)

	// Run teardown hooks
	if err := hook.Run("teardown", gitRoot, hook.Env{
		WorktreeDir: worktreeDir,
		Branch:      branch,
		RepoName:    repoName,
		Phase:       "teardown",
		TmuxSession: tmuxSessionName,
	}); err != nil {
		die("%v", err)
	}

	// Clean up session file
	sessionID := dirName
	sessionsDir := config.SessionsDir()
	if err := session.Delete(sessionsDir, repoName, sessionID); err != nil {
		info("Warning: failed to remove session file: %v", err)
	} else {
		info("Removed session file for '%s'.", sessionID)
	}

	// Remove worktree
	if git.DirExists(worktreeDir) {
		info("Removing worktree at %s…", worktreeDir)
		git.WorktreeRemove(worktreeDir)
	} else {
		info("No worktree found at %s — skipping.", worktreeDir)
	}

	// Delete local branch
	if branch != "" && git.BranchExists(branch) {
		if !git.WorktreeBranchInUse(branch) {
			info("Deleting local branch '%s'…", branch)
			git.DeleteBranch(branch)
		} else {
			info("Branch '%s' still in use by another worktree — leaving it.", branch)
		}
	}

	// Close tmux session
	if tmux.Active() {
		if tmuxSessionName != "" {
			info("Closing tmux session '%s'…", tmuxSessionName)
			tmux.KillSession(tmuxSessionName)
		} else {
			info("No tmux session found matching '%s'.", sessionPrefix)
		}
	}

	label := branch
	if label == "" {
		label = dirName
	}
	info("Done. %s torn down.", label)
}
