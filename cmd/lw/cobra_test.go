package main

import (
	"testing"

	"github.com/spf13/cobra"
)

// stubRuns replaces all RunE functions in the command tree with no-ops,
// so we only test Cobra's argument/flag validation without invoking business logic.
func stubRuns(cmd *cobra.Command) {
	if cmd.RunE != nil {
		cmd.RunE = func(cmd *cobra.Command, args []string) error { return nil }
	}
	if cmd.Run != nil {
		cmd.Run = func(cmd *cobra.Command, args []string) {}
	}
	for _, sub := range cmd.Commands() {
		stubRuns(sub)
	}
}

// executeCommand sets up the root command tree, stubs the runners, and
// executes with the given args, returning any error from arg validation.
func executeCommand(args ...string) error {
	cmd := newRootCmd()
	stubRuns(cmd)
	cmd.SetArgs(args)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	_, err := cmd.ExecuteC()
	return err
}

// --- Root command (create worktree) ---

func TestRootCommand_NoArgs(t *testing.T) {
	if err := executeCommand(); err != nil {
		t.Fatalf("expected no error for root with no args, got: %v", err)
	}
}

func TestRootCommand_OneArg_TicketID(t *testing.T) {
	if err := executeCommand("VOI-123"); err != nil {
		t.Fatalf("expected no error for root with ticket ID arg, got: %v", err)
	}
}

func TestRootCommand_OneArg_Branch(t *testing.T) {
	if err := executeCommand("feature-branch"); err != nil {
		t.Fatalf("expected no error for root with branch arg, got: %v", err)
	}
}

func TestRootCommand_CurrentFlag(t *testing.T) {
	if err := executeCommand("--current"); err != nil {
		t.Fatalf("expected no error for root with --current flag, got: %v", err)
	}
}

func TestRootCommand_TwoArgs(t *testing.T) {
	if err := executeCommand("VOI-123", "some-branch"); err != nil {
		t.Fatalf("expected no error for root with two args, got: %v", err)
	}
}

func TestRootCommand_TicketAndCurrentFlag(t *testing.T) {
	if err := executeCommand("VOI-123", "--current"); err != nil {
		t.Fatalf("expected no error for root with ticket and --current, got: %v", err)
	}
}

func TestRootCommand_TooManyArgs(t *testing.T) {
	if err := executeCommand("VOI-123", "branch", "extra"); err == nil {
		t.Fatal("expected error for root with three args, got nil")
	}
}

// --- Review subcommand ---

func TestReviewCommand_OneArg(t *testing.T) {
	if err := executeCommand("review", "my-branch"); err != nil {
		t.Fatalf("expected no error for review with one arg, got: %v", err)
	}
}

func TestReviewCommand_NoArgs(t *testing.T) {
	if err := executeCommand("review"); err == nil {
		t.Fatal("expected error for review with no args, got nil")
	}
}

func TestReviewCommand_TooManyArgs(t *testing.T) {
	if err := executeCommand("review", "a", "b"); err == nil {
		t.Fatal("expected error for review with two args, got nil")
	}
}

// --- Done subcommand ---

func TestDoneCommand_OneArg(t *testing.T) {
	if err := executeCommand("done", "VOI-123"); err != nil {
		t.Fatalf("expected no error for done with one arg, got: %v", err)
	}
}

func TestDoneCommand_NameArg(t *testing.T) {
	if err := executeCommand("done", "silly-name"); err != nil {
		t.Fatalf("expected no error for done with name arg, got: %v", err)
	}
}

func TestDoneCommand_NoArgs(t *testing.T) {
	if err := executeCommand("done"); err == nil {
		t.Fatal("expected error for done with no args, got nil")
	}
}

func TestDoneCommand_TooManyArgs(t *testing.T) {
	if err := executeCommand("done", "a", "b"); err == nil {
		t.Fatal("expected error for done with two non-review args, got nil")
	}
}

// --- Done review subcommand ---

func TestDoneReviewCommand_OneArg(t *testing.T) {
	if err := executeCommand("done", "review", "my-branch"); err != nil {
		t.Fatalf("expected no error for done review with one arg, got: %v", err)
	}
}

func TestDoneReviewCommand_NoArgs(t *testing.T) {
	if err := executeCommand("done", "review"); err == nil {
		t.Fatal("expected error for done review with no args, got nil")
	}
}

func TestDoneReviewCommand_TooManyArgs(t *testing.T) {
	if err := executeCommand("done", "review", "a", "b"); err == nil {
		t.Fatal("expected error for done review with two args, got nil")
	}
}

// --- Positional args treated as root command args, not unknown subcommands ---

func TestUnknownStringTreatedAsRootArg(t *testing.T) {
	// "foobar" is not a subcommand; it should be treated as a positional arg to root
	if err := executeCommand("foobar"); err != nil {
		t.Fatalf("expected no error (root treats unknown as positional arg), got: %v", err)
	}
}
