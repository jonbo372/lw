package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jonbo372/lw/internal/config"
	"github.com/jonbo372/lw/internal/git"
	"github.com/jonbo372/lw/internal/hook"
	"github.com/jonbo372/lw/internal/tmux"
)

func cmdReview(args []string) {
	if len(args) != 1 {
		die("Usage: lw review <branch>")
	}
	branch := args[0]

	gitRoot, err := git.MainRoot()
	if err != nil {
		die("Not inside a git repository.")
	}
	repoName := filepath.Base(gitRoot)

	localExists := git.BranchExists(branch)
	fetchErr := git.FetchOriginBranch(branch)
	remoteExists := fetchErr == nil || git.RefExists("refs/remotes/origin/"+branch)

	if !localExists && !remoteExists {
		die("Branch '%s' not found locally or on origin.", branch)
	}

	if !localExists {
		info("Creating local branch '%s' from origin…", branch)
		git.CreateBranch(branch, "origin/"+branch)
	}

	safeDir := "review-" + strings.ReplaceAll(branch, "/", "-")
	worktreeDir := filepath.Join(config.WorktreeHome(), repoName, safeDir)

	os.MkdirAll(filepath.Dir(worktreeDir), 0755)

	if git.DirExists(worktreeDir) {
		info("Worktree already exists at %s", worktreeDir)
	} else {
		info("Creating review worktree at %s…", worktreeDir)
		if err := git.WorktreeAdd(worktreeDir, branch); err != nil {
			die("Failed to create worktree: %v", err)
		}
		info("Done.")
	}

	windowName := fmt.Sprintf("[%s] review: %s", repoName, branch)
	tmuxWindow := tmuxCreateOrSwitchInfo(windowName, worktreeDir)

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

func tmuxCreateOrSwitchInfo(name, dir string) string {
	if !tmux.Active() {
		return ""
	}
	if len(name) > 50 {
		name = name[:50]
	}
	windowName := tmux.CreateOrSwitch(name, dir)
	if windowName != "" {
		info("tmux window '%s' ready.", windowName)
	}
	return windowName
}
