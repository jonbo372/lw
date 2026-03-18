package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jonbo372/lw/internal/config"
	"github.com/jonbo372/lw/internal/git"
	"github.com/jonbo372/lw/internal/hook"
	"github.com/jonbo372/lw/internal/tmux"
)

func cmdDone(args []string) {
	gitRoot, err := git.MainRoot()
	if err != nil {
		die("Not inside a git repository.")
	}
	repoName := filepath.Base(gitRoot)

	isReview := false
	var branch, safeDir, windowPrefix string

	if len(args) >= 1 && args[0] == "review" {
		isReview = true
		if len(args) != 2 {
			die("Usage: lw done review <branch>")
		}
		branch = args[1]
		safeDir = "review-" + strings.ReplaceAll(branch, "/", "-")
		windowPrefix = fmt.Sprintf("[%s] review:", repoName)
	} else {
		if len(args) != 1 {
			die("Usage: lw done <TICKET-ID|name>")
		}
		arg := args[0]
		ticketRe := regexp.MustCompile(`(?i)^[A-Za-z]+-[0-9]+`)
		if ticketRe.MatchString(arg) {
			ticket := strings.ToUpper(arg)
			safeDir = strings.ReplaceAll(ticket, "/", "-")
		} else {
			safeDir = arg
		}
		windowPrefix = fmt.Sprintf("[%s] %s", repoName, safeDir)
	}

	worktreeDir := filepath.Join(config.WorktreeHome(), repoName, safeDir)

	// Resolve branch from worktree
	if git.DirExists(worktreeDir) {
		if b, err := git.CurrentBranchInDir(worktreeDir); err == nil {
			branch = b
		}
	}

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
	if !isReview && branch != "" && git.DirExists(worktreeDir) {
		if git.HasUnpushedCommits(worktreeDir, branch) {
			fmt.Fprintf(os.Stderr, "warning: branch '%s' has commits not pushed to origin.\n", branch)
			if !confirm("Continue with teardown anyway?") {
				info("Aborted.")
				os.Exit(0)
			}
		}
	}

	// Resolve tmux window
	tmuxWindowIndex, tmuxWindowName := tmux.FindWindow(windowPrefix)

	// Run teardown hooks
	if err := hook.Run("teardown", gitRoot, hook.Env{
		WorktreeDir: worktreeDir,
		Branch:      branch,
		RepoName:    repoName,
		Phase:       "teardown",
		TmuxWindow:  tmuxWindowName,
	}); err != nil {
		die("%v", err)
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
		if isReview {
			if !git.WorktreeBranchInUse(branch) {
				info("Deleting local branch '%s'…", branch)
				git.DeleteBranch(branch)
			} else {
				info("Branch '%s' still in use by another worktree — leaving it.", branch)
			}
		} else {
			info("Deleting local branch '%s'…", branch)
			git.DeleteBranch(branch)
		}
	}

	// Close tmux window
	if tmux.Active() {
		if tmuxWindowIndex != "" {
			info("Closing tmux window %s…", tmuxWindowIndex)
			tmux.KillWindow(tmuxWindowIndex)
		} else {
			info("No tmux window found matching '%s'.", windowPrefix)
		}
	}

	label := branch
	if label == "" {
		label = safeDir
	}
	info("Done. %s torn down.", label)
}
