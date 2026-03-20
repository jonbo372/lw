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
	"github.com/jonbo372/lw/internal/linear"
	"github.com/jonbo372/lw/internal/namegen"
	"github.com/jonbo372/lw/internal/session"
)

// fetchTicket resolves the correct API key and fetches the ticket from Linear.
func fetchTicket(ticketID string) *linear.Ticket {
	resolver := linear.NewResolver(linear.DefaultConfigPath(), linear.FetchTicket)
	if isVerbose() {
		resolver.SetVerbose(verbose)
	}
	_, t, err := resolver.ResolveAndFetch(ticketID)
	if err != nil {
		die("%v", err)
	}
	return t
}

// validateName checks that a --name value is safe for use as a directory component.
func validateName(name string) error {
	if strings.Contains(name, "/") || strings.Contains(name, "..") {
		return fmt.Errorf("--name must not contain '/' or '..'")
	}
	return nil
}

// cmdNew implements the `lw new` subcommand.
//
// Naming precedence (highest to lowest):
//
//	--name         → branch=name, tmuxWindow=name
//	--branch_name  → branch=branch_name, tmuxWindow derived from ticket or branch_name
//	--ticket       → branch=Linear gitBranchName, tmuxWindow derived from gitBranchName
//	(none)         → branch=silly-name, tmuxWindow=silly-name
func cmdNew(ticket, name, branchName string, currentTmuxWindow bool) {
	gitRoot, err := git.MainRoot()
	if err != nil {
		die("Not inside a git repository.")
	}
	repoName := filepath.Base(gitRoot)

	ticket = strings.ToUpper(ticket)

	var branch, title, safeLabel, tmuxWindowLabel string

	switch {
	case name != "":
		// --name takes precedence over everything
		if err := validateName(name); err != nil {
			die("%v", err)
		}
		branch = name
		safeLabel = name
		tmuxWindowLabel = name
		title = name

		if ticket != "" {
			t := fetchTicket(ticket)
			title = t.Title
		}

	case branchName != "":
		// --branch_name overrides branch; tmux derived from ticket or branch_name
		branch = branchName

		if ticket != "" {
			t := fetchTicket(ticket)
			title = t.Title
			safeLabel = strings.ReplaceAll(ticket, "/", "-")
			tmuxWindowLabel = safeLabel
		} else {
			safeLabel = strings.ReplaceAll(branchName, "/", "-")
			tmuxWindowLabel = safeLabel
			title = branchName
		}

	case ticket != "":
		// --ticket only: use Linear gitBranchName
		info("Fetching branch for %s from Linear…", ticket)
		t := fetchTicket(ticket)
		if t.Branch == "" {
			die("No gitBranchName found for ticket %s.", ticket)
		}
		branch = t.Branch
		title = t.Title
		safeLabel = strings.ReplaceAll(ticket, "/", "-")
		tmuxWindowLabel = safeLabel
		info("Found: \"%s\"", title)
		info("Branch: %s", branch)

	default:
		// No flags: generate silly-name
		sillyName := namegen.Generate()
		branch = sillyName
		safeLabel = sillyName
		tmuxWindowLabel = sillyName
		title = "scratch"
		info("Creating scratch worktree '%s'…", sillyName)
	}

	// Ensure branch exists
	if !git.BranchExists(branch) {
		defaultBranch, err := git.DetectDefaultBranch()
		if err != nil {
			die("%v", err)
		}
		if err := git.EnsureLocalBranch(defaultBranch); err != nil {
			die("Could not find default branch '%s'.", defaultBranch)
		}
		info("Creating '%s' off '%s'…", branch, defaultBranch)
		if err := git.CreateBranch(branch, defaultBranch); err != nil {
			die("Failed to create branch '%s': %v", branch, err)
		}
	}

	if !git.BranchExists(branch) {
		die("Branch '%s' could not be created.", branch)
	}

	// Create worktree
	worktreeDir := filepath.Join(config.WorktreeHome(), repoName, safeLabel)
	if err := os.MkdirAll(filepath.Dir(worktreeDir), 0755); err != nil {
		die("Failed to create worktree parent directory: %v", err)
	}

	if git.DirExists(worktreeDir) {
		info("Worktree already exists at %s", worktreeDir)
	} else {
		info("Creating worktree at %s…", worktreeDir)
		if err := git.WorktreeAdd(worktreeDir, branch); err != nil {
			die("Failed to create worktree: %v", err)
		}
		info("Done.")
	}

	// tmux window
	var tmuxWindow string
	if !currentTmuxWindow {
		windowName := fmt.Sprintf("[%s] %s: %s", repoName, tmuxWindowLabel, title)
		tmuxWindow = tmuxCreateOrSwitchInfo(windowName, worktreeDir)
	}

	// Persist session metadata
	sessionsDir := config.SessionsDir()
	sessionID := safeLabel
	sess := &session.Session{
		Branch:      branch,
		Ticket:      ticket,
		WorktreeDir: worktreeDir,
		TmuxWindow:  tmuxWindow,
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

	// hooks
	if err := hook.Run("setup", gitRoot, hook.Env{
		WorktreeDir: worktreeDir,
		Branch:      branch,
		Ticket:      ticket,
		RepoName:    repoName,
		Phase:       "setup",
		TmuxWindow:  tmuxWindow,
	}); err != nil {
		die("%v", err)
	}

	fmt.Println(worktreeDir)
}
