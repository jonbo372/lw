package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jonbo372/lw/internal/config"
	"github.com/jonbo372/lw/internal/git"
	"github.com/jonbo372/lw/internal/hook"
	"github.com/jonbo372/lw/internal/linear"
	"github.com/jonbo372/lw/internal/namegen"
)

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
			apiKey := config.LinearAPIKey()
			if apiKey == "" {
				die("LINEAR_API_KEY is not set.")
			}
			t, err := linear.FetchTicket(apiKey, strings.ToUpper(ticket))
			if err != nil {
				die("%v", err)
			}
			title = t.Title
		}

	case branchName != "":
		// --branch_name overrides branch; tmux derived from ticket or branch_name
		branch = branchName

		if ticket != "" {
			apiKey := config.LinearAPIKey()
			if apiKey == "" {
				die("LINEAR_API_KEY is not set.")
			}
			t, err := linear.FetchTicket(apiKey, strings.ToUpper(ticket))
			if err != nil {
				die("%v", err)
			}
			title = t.Title
			safeLabel = strings.ReplaceAll(strings.ToUpper(ticket), "/", "-")
			tmuxWindowLabel = safeLabel
		} else {
			safeLabel = strings.ReplaceAll(branchName, "/", "-")
			tmuxWindowLabel = safeLabel
			title = branchName
		}

	case ticket != "":
		// --ticket only: use Linear gitBranchName
		apiKey := config.LinearAPIKey()
		if apiKey == "" {
			die("LINEAR_API_KEY is not set. Export it or set it in your environment.")
		}

		ticket = strings.ToUpper(ticket)
		info("Fetching branch for %s from Linear…", ticket)
		t, err := linear.FetchTicket(apiKey, ticket)
		if err != nil {
			die("%v", err)
		}
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
