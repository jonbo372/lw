package main

import (
	"os"

	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	var flagCurrent bool

	rootCmd := &cobra.Command{
		Use:   "lw [<TICKET-ID>] [<branch>|--current]",
		Short: "Manage git worktrees for Linear tickets",
		Args:  cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Build the effective args list for cmdCreate.
			// If --current was passed as a flag, inject it into the args slice
			// so cmdCreate sees the same interface it always did.
			effective := make([]string, 0, len(args)+1)
			effective = append(effective, args...)
			if flagCurrent {
				effective = append(effective, "--current")
			}
			cmdCreate(effective)
			return nil
		},
		SilenceUsage: true,
	}

	rootCmd.Flags().BoolVar(&flagCurrent, "current", false, "Use the current branch as base")

	// review subcommand
	reviewCmd := &cobra.Command{
		Use:   "review <branch>",
		Short: "Create a review worktree",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdReview(args)
			return nil
		},
	}

	// done subcommand
	doneCmd := &cobra.Command{
		Use:   "done <TICKET-ID|name>",
		Short: "Tear down a worktree",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdDone(args)
			return nil
		},
	}

	// done review subcommand
	doneReviewCmd := &cobra.Command{
		Use:   "review <branch>",
		Short: "Tear down a review worktree",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdDone(append([]string{"review"}, args...))
			return nil
		},
	}

	doneCmd.AddCommand(doneReviewCmd)
	rootCmd.AddCommand(reviewCmd, doneCmd)

	return rootCmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
