package main

import (
	"os"

	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "lw",
		Short: "Manage git worktrees for Linear tickets",
		// No RunE — root command displays help when invoked without subcommands.
		SilenceUsage: true,
	}

	// new subcommand
	newCmd := newNewCmd()

	// continue subcommand
	continueCmd := newContinueCmd()

	// done subcommand
	doneCmd := newDoneCmd()

	// review subcommand
	reviewCmd := newReviewCmd()

	// list subcommand
	listCmd := newListCmd()

	// session-end subcommand (used by Claude Code SessionEnd hook)
	sessionEndCmd := newSessionEndCmd()

	rootCmd.AddCommand(newCmd, continueCmd, doneCmd, reviewCmd, listCmd, sessionEndCmd)

	return rootCmd
}

func newNewCmd() *cobra.Command {
	var flagTicket string
	var flagName string
	var flagBranchName string
	var flagCurrentTmuxWindow bool

	cmd := &cobra.Command{
		Use:   "new",
		Short: "Create a new git worktree with a tmux window",
		Long: `Create a new git worktree with a new tmux window.

By default, generates a silly-name for the branch and tmux window.

Naming precedence (highest to lowest):
  --name              Uses value as branch, worktree, and tmux window name
  --branch_name       Uses value as branch; tmux window derived from ticket or branch
  --ticket            Uses Linear gitBranchName as branch name
  (none)              Generates a silly-name`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdNew(flagTicket, flagName, flagBranchName, flagCurrentTmuxWindow)
			return nil
		},
	}

	cmd.Flags().StringVar(&flagTicket, "ticket", "", "Linear ticket ID to fetch branch name from")
	cmd.Flags().StringVar(&flagName, "name", "", "Name for branch, worktree, and tmux window")
	cmd.Flags().StringVar(&flagBranchName, "branch_name", "", "Branch name (overrides ticket branch)")
	cmd.Flags().BoolVar(&flagCurrentTmuxWindow, "current-tmux-window", false, "Stay in current tmux window")

	return cmd
}

func newContinueCmd() *cobra.Command {
	var flagCurrentTmuxWindow bool

	cmd := &cobra.Command{
		Use:   "continue <session_identifier>",
		Short: "Continue a previous session",
		Long: `Continue a previous session by locating an existing git worktree
and opening a tmux window pointing to it.

The session identifier can be a ticket ID, name, or full worktree path.
It is matched against the output of 'git worktree list'.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdContinue(args[0], flagCurrentTmuxWindow)
			return nil
		},
	}

	cmd.Flags().BoolVar(&flagCurrentTmuxWindow, "current-tmux-window", false, "Stay in current tmux window")

	return cmd
}

func newDoneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "done <session_identifier>",
		Short: "Tear down a session",
		Long: `Tear down a session. Resolves the session identifier against
'git worktree list', warns about uncommitted/unpushed changes,
removes the worktree, and kills the associated tmux window.

Handles all worktree types including review worktrees.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdDone(args[0])
			return nil
		},
	}

	return cmd
}

func newReviewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review <branch>",
		Short: "Create a review worktree for a branch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdReview(args)
			return nil
		},
	}

	return cmd
}

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all sessions and worktrees",
		Long: `List all sessions and worktrees, cross-referencing session JSON files
with the output of 'git worktree list'.

Status labels:
  active    - Session and worktree both exist
  orphaned  - Worktree exists but no session JSON
  dead      - Session JSON exists but worktree is gone`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdList()
		},
	}

	return cmd
}

func newSessionEndCmd() *cobra.Command {
	var flagRepo string
	var flagSession string

	cmd := &cobra.Command{
		Use:    "session-end",
		Short:  "Update session with Claude Code session ID (used by hooks)",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdSessionEnd(flagRepo, flagSession)
		},
	}

	cmd.Flags().StringVar(&flagRepo, "repo", "", "Repository name")
	cmd.Flags().StringVar(&flagSession, "session", "", "Session identifier")
	cmd.MarkFlagRequired("repo")
	cmd.MarkFlagRequired("session")

	return cmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
