package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jonbo372/lw/internal/claudehook"
	"github.com/jonbo372/lw/internal/config"
	"github.com/jonbo372/lw/internal/git"
	"github.com/jonbo372/lw/internal/hook"
	"github.com/jonbo372/lw/internal/session"
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

	if err := os.MkdirAll(filepath.Dir(worktreeDir), 0755); err != nil {
		die("Failed to create worktree parent directory: %v", err)
	}

	if git.DirExists(worktreeDir) {
		info("Worktree already exists at %s", worktreeDir)
	} else {
		info("Creating review worktree at %s…", worktreeDir)
		if err := git.WorktreeAdd(worktreeDir, branch); err != nil {
			die("Failed to create worktree: %v", err)
		}
		info("Done.")
	}

	sessionName := fmt.Sprintf("[%s] review: %s", repoName, branch)
	tmuxSession := tmuxCreateOrSwitchInfo(sessionName, worktreeDir)

	// Persist session metadata
	sessionsDir := config.SessionsDir()
	sessionID := safeDir
	sess := &session.Session{
		Branch:      branch,
		Ticket:      "",
		WorktreeDir: worktreeDir,
		TmuxSession: tmuxSession,
	}
	sessionPath, err := session.Create(sessionsDir, repoName, sessionID, sess)
	if err != nil {
		die("Failed to create session file: %v", err)
	}
	info("Session saved to %s", sessionPath)

	// Install Claude Code SessionEnd hook
	if err := claudehook.Install(worktreeDir, repoName, sessionID); err != nil {
		die("Failed to install Claude hook: %v", err)
	}

	if err := hook.Run("setup", gitRoot, hook.Env{
		WorktreeDir: worktreeDir,
		Branch:      branch,
		RepoName:    repoName,
		Phase:       "setup",
		TmuxSession: tmuxSession,
	}); err != nil {
		die("%v", err)
	}

	fmt.Println(worktreeDir)
}

func tmuxCreateOrSwitchInfo(name, dir string) string {
	if !tmux.Active() {
		return ""
	}
	if len(name) > tmux.MaxNameLen {
		name = name[:tmux.MaxNameLen]
	}
	sessionName := tmux.CreateOrSwitch(name, dir)
	if sessionName != "" {
		info("tmux session '%s' ready.", sessionName)
	}
	return sessionName
}
