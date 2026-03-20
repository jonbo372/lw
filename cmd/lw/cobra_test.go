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

// --- Root command (no args = help) ---

func TestRootCommand_NoArgs_ShowsHelp(t *testing.T) {
	// Root with no args should succeed (shows help)
	if err := executeCommand(); err != nil {
		t.Fatalf("expected no error for root with no args, got: %v", err)
	}
}

func TestRootCommand_UnknownSubcommand(t *testing.T) {
	// Unknown subcommands should error since root no longer accepts args
	if err := executeCommand("foobar"); err == nil {
		t.Fatal("expected error for unknown subcommand, got nil")
	}
}

// --- New subcommand ---

func TestNewCommand_NoFlags(t *testing.T) {
	if err := executeCommand("new"); err != nil {
		t.Fatalf("expected no error for new with no flags, got: %v", err)
	}
}

func TestNewCommand_TicketFlag(t *testing.T) {
	if err := executeCommand("new", "--ticket", "VOI-123"); err != nil {
		t.Fatalf("expected no error for new with --ticket, got: %v", err)
	}
}

func TestNewCommand_NameFlag(t *testing.T) {
	if err := executeCommand("new", "--name", "my-feature"); err != nil {
		t.Fatalf("expected no error for new with --name, got: %v", err)
	}
}

func TestNewCommand_BranchNameFlag(t *testing.T) {
	if err := executeCommand("new", "--branch_name", "custom-branch"); err != nil {
		t.Fatalf("expected no error for new with --branch_name, got: %v", err)
	}
}

func TestNewCommand_CurrentTmuxWindowFlag(t *testing.T) {
	if err := executeCommand("new", "--current-tmux-window"); err != nil {
		t.Fatalf("expected no error for new with --current-tmux-window, got: %v", err)
	}
}

func TestNewCommand_AllFlags(t *testing.T) {
	if err := executeCommand("new", "--ticket", "VOI-123", "--name", "my-feature", "--branch_name", "custom", "--current-tmux-window"); err != nil {
		t.Fatalf("expected no error for new with all flags, got: %v", err)
	}
}

func TestNewCommand_NoPositionalArgs(t *testing.T) {
	// new should not accept positional args
	if err := executeCommand("new", "some-arg"); err == nil {
		t.Fatal("expected error for new with positional arg, got nil")
	}
}

// --- Continue subcommand ---

func TestContinueCommand_OneArg(t *testing.T) {
	if err := executeCommand("continue", "fuzzy_cobra"); err != nil {
		t.Fatalf("expected no error for continue with one arg, got: %v", err)
	}
}

func TestContinueCommand_NoArgs(t *testing.T) {
	if err := executeCommand("continue"); err == nil {
		t.Fatal("expected error for continue with no args, got nil")
	}
}

func TestContinueCommand_TooManyArgs(t *testing.T) {
	if err := executeCommand("continue", "a", "b"); err == nil {
		t.Fatal("expected error for continue with two args, got nil")
	}
}

func TestContinueCommand_CurrentTmuxWindowFlag(t *testing.T) {
	if err := executeCommand("continue", "--current-tmux-window", "fuzzy_cobra"); err != nil {
		t.Fatalf("expected no error for continue with --current-tmux-window, got: %v", err)
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
		t.Fatal("expected error for done with two args, got nil")
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

// --- validateName ---

func TestValidateName_Valid(t *testing.T) {
	for _, name := range []string{"my-feature", "cool_thing", "fix-123"} {
		if err := validateName(name); err != nil {
			t.Errorf("expected %q to be valid, got: %v", name, err)
		}
	}
}

func TestValidateName_RejectsSlash(t *testing.T) {
	if err := validateName("foo/bar"); err == nil {
		t.Fatal("expected error for name containing '/', got nil")
	}
}

func TestValidateName_RejectsDotDot(t *testing.T) {
	if err := validateName("foo..bar"); err == nil {
		t.Fatal("expected error for name containing '..', got nil")
	}
}

func TestValidateName_RejectsPathTraversal(t *testing.T) {
	if err := validateName("../etc"); err == nil {
		t.Fatal("expected error for path traversal name, got nil")
	}
}

// --- Session-end subcommand ---

func TestSessionEndCommand_RequiresFlags(t *testing.T) {
	if err := executeCommand("session-end"); err == nil {
		t.Fatal("expected error for session-end without flags, got nil")
	}
}

func TestSessionEndCommand_RequiresSessionFlag(t *testing.T) {
	if err := executeCommand("session-end", "--repo", "myrepo"); err == nil {
		t.Fatal("expected error for session-end without --session, got nil")
	}
}

func TestSessionEndCommand_RequiresRepoFlag(t *testing.T) {
	if err := executeCommand("session-end", "--session", "VOI-42"); err == nil {
		t.Fatal("expected error for session-end without --repo, got nil")
	}
}

func TestSessionEndCommand_BothFlags(t *testing.T) {
	if err := executeCommand("session-end", "--repo", "myrepo", "--session", "VOI-42"); err != nil {
		t.Fatalf("expected no error for session-end with both flags, got: %v", err)
	}
}

func TestSessionEndCommand_NoPositionalArgs(t *testing.T) {
	if err := executeCommand("session-end", "--repo", "myrepo", "--session", "VOI-42", "extra"); err == nil {
		t.Fatal("expected error for session-end with positional arg, got nil")
	}
}

// --- Verbose flag ---

func TestVerboseFlag_Default(t *testing.T) {
	cmd := newRootCmd()
	stubRuns(cmd)
	cmd.SetArgs([]string{"new"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.ExecuteC()

	v, err := cmd.PersistentFlags().GetBool("verbose")
	if err != nil {
		t.Fatalf("expected --verbose flag to exist, got: %v", err)
	}
	if v {
		t.Error("expected --verbose to default to false")
	}
}

func TestVerboseFlag_Set(t *testing.T) {
	cmd := newRootCmd()
	stubRuns(cmd)
	cmd.SetArgs([]string{"--verbose", "new"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.ExecuteC()

	v, _ := cmd.PersistentFlags().GetBool("verbose")
	if !v {
		t.Error("expected --verbose to be true when set")
	}
}

func TestVerboseFlag_ShortFlag(t *testing.T) {
	cmd := newRootCmd()
	stubRuns(cmd)
	cmd.SetArgs([]string{"-v", "new"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.ExecuteC()

	v, _ := cmd.PersistentFlags().GetBool("verbose")
	if !v {
		t.Error("expected -v to set verbose to true")
	}
}

func TestVerboseFlag_AvailableOnSubcommands(t *testing.T) {
	subcommands := [][]string{
		{"--verbose", "new"},
		{"--verbose", "continue", "foo"},
		{"--verbose", "done", "foo"},
		{"--verbose", "review", "foo"},
		{"--verbose", "list"},
		{"--verbose", "session-end", "--repo", "r", "--session", "s"},
	}
	for _, args := range subcommands {
		if err := executeCommand(args...); err != nil {
			t.Errorf("expected no error for %v, got: %v", args, err)
		}
	}
}

// --- done review is no longer a subcommand ---

func TestDoneReview_IsNoLongerSubcommand(t *testing.T) {
	// "done review" should now treat "review" as the session identifier argument
	// to the done command, not as a subcommand
	if err := executeCommand("done", "review"); err != nil {
		t.Fatalf("expected 'done review' to treat review as an arg, got: %v", err)
	}
}
