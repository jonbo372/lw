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
	"github.com/jonbo372/lw/internal/linear"
	"github.com/jonbo372/lw/internal/namegen"
)

func cmdCreate(args []string) {
	if len(args) > 2 {
		die("Usage: lw [<TICKET-ID>] [<branch>|--current]")
	}

	gitRoot, err := git.MainRoot()
	if err != nil {
		die("Not inside a git repository.")
	}
	repoName := filepath.Base(gitRoot)

	var ticket, baseBranch, branch, title, safeLabel string
	scratch := false
	ticketRe := regexp.MustCompile(`(?i)^[A-Za-z]+-[0-9]+`)

	switch len(args) {
	case 0:
		scratch = true
	case 1:
		if args[0] == "--current" {
			scratch = true
			baseBranch, err = git.CurrentBranch()
			if err != nil {
				die("%v", err)
			}
			info("Using current branch '%s' as base.", baseBranch)
		} else if ticketRe.MatchString(args[0]) {
			ticket = strings.ToUpper(args[0])
		} else {
			scratch = true
			baseBranch = args[0]
			info("Using '%s' as base branch.", baseBranch)
		}
	case 2:
		ticket = strings.ToUpper(args[0])
		if args[1] == "--current" {
			baseBranch, err = git.CurrentBranch()
			if err != nil {
				die("%v", err)
			}
			info("Using current branch '%s' as base.", baseBranch)
		} else {
			baseBranch = args[1]
			info("Using '%s' as base branch.", baseBranch)
		}
	}

	if scratch {
		sillyName := namegen.Generate()
		branch = sillyName
		title = "scratch"
		safeLabel = sillyName
		info("Creating scratch worktree '%s'…", sillyName)

		if baseBranch == "" {
			baseBranch, err = git.DetectDefaultBranch()
			if err != nil {
				die("%v", err)
			}
		}

		if err := git.EnsureLocalBranch(baseBranch); err != nil {
			die("Base branch '%s' not found locally or on origin.", baseBranch)
		}

		info("Creating '%s' off '%s'…", branch, baseBranch)
		if err := git.CreateBranch(branch, baseBranch); err != nil {
			die("Failed to create branch '%s': %v", branch, err)
		}
	} else {
		apiKey := config.LinearAPIKey()
		if apiKey == "" {
			die("LINEAR_API_KEY is not set. Export it or set it in your environment.")
		}

		info("Fetching branch for %s from Linear…", ticket)
		t, err := linear.FetchTicket(apiKey, ticket)
		if err != nil {
			die("%v", err)
		}
		branch = t.Branch
		title = t.Title
		info("Found: \"%s\"", title)
		info("Branch: %s", branch)
		safeLabel = strings.ReplaceAll(ticket, "/", "-")

		if !git.BranchExists(branch) {
			if baseBranch == "" {
				if err := git.FetchBranch(branch); err == nil {
					info("Fetched '%s' from origin.", branch)
				} else {
					info("Branch not found on origin — detecting default branch…")
					defaultBranch, err := git.DetectDefaultBranch()
					if err != nil {
						die("%v", err)
					}
					git.FetchBranch(defaultBranch)
					info("Creating '%s' off '%s'…", branch, defaultBranch)
					git.CreateBranch(branch, defaultBranch)
				}
			} else {
				if err := git.EnsureLocalBranch(baseBranch); err != nil {
					die("Base branch '%s' not found locally or on origin.", baseBranch)
				}
				info("Creating '%s' off '%s'…", branch, baseBranch)
				git.CreateBranch(branch, baseBranch)
			}
		}
	}

	if !git.BranchExists(branch) {
		die("Branch '%s' could not be created.", branch)
	}

	// Create worktree
	worktreeDir := filepath.Join(config.WorktreeHome(), repoName, safeLabel)
	os.MkdirAll(filepath.Dir(worktreeDir), 0755)

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
	windowName := fmt.Sprintf("[%s] %s: %s", repoName, safeLabel, title)
	tmuxWindow := tmuxCreateOrSwitchInfo(windowName, worktreeDir)

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
