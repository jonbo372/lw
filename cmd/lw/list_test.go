package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/jonbo372/lw/internal/git"
	"github.com/jonbo372/lw/internal/session"
)

// --- Cobra registration tests ---

func TestListCommand_NoArgs(t *testing.T) {
	if err := executeCommand("list"); err != nil {
		t.Fatalf("expected no error for list with no args, got: %v", err)
	}
}

func TestListCommand_LsAlias(t *testing.T) {
	if err := executeCommand("ls"); err != nil {
		t.Fatalf("expected no error for ls alias, got: %v", err)
	}
}

func TestListCommand_RejectsArgs(t *testing.T) {
	if err := executeCommand("list", "extra"); err == nil {
		t.Fatal("expected error for list with positional arg, got nil")
	}
}

// --- buildListEntries tests ---

func TestBuildListEntries_MainWorktreeFirst(t *testing.T) {
	sessions := map[string]*session.Session{}
	worktrees := []git.WorktreeEntry{
		{Path: "/home/user/repo", Branch: "main"},
	}

	entries := buildListEntries(sessions, worktrees, "/home/user/repo")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Status != statusMain {
		t.Errorf("expected statusMain, got %d", entries[0].Status)
	}
	if entries[0].Name != "repo" {
		t.Errorf("expected name 'repo', got %q", entries[0].Name)
	}
}

func TestBuildListEntries_ActiveSession(t *testing.T) {
	sessions := map[string]*session.Session{
		"VOI-42": {
			Branch:          "jonbo372/VOI-42/add-auth",
			WorktreeDir:     "/home/user/.superclaude/repo/VOI-42",
			ClaudeSessionID: "claude-abc",
		},
	}
	worktrees := []git.WorktreeEntry{
		{Path: "/home/user/repo", Branch: "main"},
		{Path: "/home/user/.superclaude/repo/VOI-42", Branch: "jonbo372/VOI-42/add-auth"},
	}

	entries := buildListEntries(sessions, worktrees, "/home/user/repo")
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	// First should be main
	if entries[0].Status != statusMain {
		t.Error("first entry should be main")
	}
	// Second should be active
	if entries[1].Status != statusActive {
		t.Errorf("expected statusActive, got %d", entries[1].Status)
	}
	if entries[1].ClaudeSessionID != "claude-abc" {
		t.Errorf("expected claude ID 'claude-abc', got %q", entries[1].ClaudeSessionID)
	}
}

func TestBuildListEntries_OrphanedWorktree(t *testing.T) {
	sessions := map[string]*session.Session{}
	worktrees := []git.WorktreeEntry{
		{Path: "/home/user/repo", Branch: "main"},
		{Path: "/home/user/.superclaude/repo/orphan", Branch: "orphan-branch"},
	}

	entries := buildListEntries(sessions, worktrees, "/home/user/repo")
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[1].Status != statusOrphaned {
		t.Errorf("expected statusOrphaned, got %d", entries[1].Status)
	}
	if entries[1].Name != "orphan" {
		t.Errorf("expected name 'orphan', got %q", entries[1].Name)
	}
}

func TestBuildListEntries_DeadSession(t *testing.T) {
	sessions := map[string]*session.Session{
		"dead-one": {
			Branch:          "dead-branch",
			WorktreeDir:     "/home/user/.superclaude/repo/dead-one",
			ClaudeSessionID: "old-session",
		},
	}
	worktrees := []git.WorktreeEntry{
		{Path: "/home/user/repo", Branch: "main"},
		// dead-one's worktree path is NOT in this list
	}

	entries := buildListEntries(sessions, worktrees, "/home/user/repo")
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[1].Status != statusDead {
		t.Errorf("expected statusDead, got %d", entries[1].Status)
	}
	if entries[1].ClaudeSessionID != "old-session" {
		t.Errorf("expected claude ID 'old-session', got %q", entries[1].ClaudeSessionID)
	}
}

func TestBuildListEntries_MixedStatuses(t *testing.T) {
	sessions := map[string]*session.Session{
		"active-one": {
			Branch:      "active-branch",
			WorktreeDir: "/home/user/.superclaude/repo/active-one",
		},
		"dead-one": {
			Branch:      "dead-branch",
			WorktreeDir: "/home/user/.superclaude/repo/dead-one",
		},
	}
	worktrees := []git.WorktreeEntry{
		{Path: "/home/user/repo", Branch: "main"},
		{Path: "/home/user/.superclaude/repo/active-one", Branch: "active-branch"},
		{Path: "/home/user/.superclaude/repo/orphan-wt", Branch: "orphan-branch"},
	}

	entries := buildListEntries(sessions, worktrees, "/home/user/repo")

	// Should have: main, active-one, dead-one, orphan-wt = 4 entries
	if len(entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(entries))
	}

	// First entry must be main
	if entries[0].Status != statusMain {
		t.Error("first entry should be main")
	}

	// Check we have one of each non-main status
	statusCounts := map[sessionStatus]int{}
	for _, e := range entries {
		statusCounts[e.Status]++
	}
	if statusCounts[statusActive] != 1 {
		t.Errorf("expected 1 active, got %d", statusCounts[statusActive])
	}
	if statusCounts[statusDead] != 1 {
		t.Errorf("expected 1 dead, got %d", statusCounts[statusDead])
	}
	if statusCounts[statusOrphaned] != 1 {
		t.Errorf("expected 1 orphaned, got %d", statusCounts[statusOrphaned])
	}
}

func TestBuildListEntries_SkipsBareWorktrees(t *testing.T) {
	sessions := map[string]*session.Session{}
	worktrees := []git.WorktreeEntry{
		{Path: "/home/user/repo", Branch: "main"},
		{Path: "/home/user/bare-repo", Branch: "", Bare: true},
	}

	entries := buildListEntries(sessions, worktrees, "/home/user/repo")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry (bare should be skipped), got %d", len(entries))
	}
}

func TestBuildListEntries_NoMainWorktree(t *testing.T) {
	sessions := map[string]*session.Session{
		"s1": {Branch: "b1", WorktreeDir: "/wt/s1"},
	}
	worktrees := []git.WorktreeEntry{
		{Path: "/wt/s1", Branch: "b1"},
	}

	entries := buildListEntries(sessions, worktrees, "/nonexistent/main")
	// No main entry, just the active session
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Status != statusActive {
		t.Errorf("expected active, got %d", entries[0].Status)
	}
}

func TestBuildListEntries_EmptyInputs(t *testing.T) {
	entries := buildListEntries(map[string]*session.Session{}, nil, "")
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

// --- statusLabel tests ---

func TestStatusLabel_AllStatuses(t *testing.T) {
	tests := []struct {
		status sessionStatus
		want   string
	}{
		{statusMain, "main"},
		{statusActive, "active"},
		{statusOrphaned, "orphaned"},
		{statusDead, "dead"},
	}
	for _, tc := range tests {
		plain := statusLabelPlain(tc.status)
		if plain != tc.want {
			t.Errorf("statusLabelPlain(%d) = %q, want %q", tc.status, plain, tc.want)
		}
		colored := statusLabel(tc.status)
		if !strings.Contains(colored, tc.want) {
			t.Errorf("statusLabel(%d) = %q, should contain %q", tc.status, colored, tc.want)
		}
	}
}

func TestStatusLabel_Unknown(t *testing.T) {
	plain := statusLabelPlain(sessionStatus(99))
	if plain != "unknown" {
		t.Errorf("expected 'unknown', got %q", plain)
	}
	colored := statusLabel(sessionStatus(99))
	if colored != "unknown" {
		t.Errorf("expected 'unknown', got %q", colored)
	}
}

// --- renderTable tests ---

func TestRenderTable_EmptyEntries(t *testing.T) {
	var buf bytes.Buffer
	renderTable(&buf, nil)
	output := buf.String()
	if !strings.Contains(output, "No sessions or worktrees found") {
		t.Errorf("expected 'no sessions' message, got: %q", output)
	}
}

func TestRenderTable_HasHeader(t *testing.T) {
	var buf bytes.Buffer
	entries := []listEntry{
		{Name: "repo", Branch: "main", WorktreePath: "/repo", Status: statusMain},
	}
	renderTable(&buf, entries)
	output := buf.String()
	if !strings.Contains(output, "NAME") {
		t.Error("expected NAME header")
	}
	if !strings.Contains(output, "BRANCH") {
		t.Error("expected BRANCH header")
	}
	if !strings.Contains(output, "WORKTREE") {
		t.Error("expected WORKTREE header")
	}
	if !strings.Contains(output, "STATUS") {
		t.Error("expected STATUS header")
	}
	if !strings.Contains(output, "CLAUDE ID") {
		t.Error("expected CLAUDE ID header")
	}
}

func TestRenderTable_ShowsAllEntries(t *testing.T) {
	var buf bytes.Buffer
	entries := []listEntry{
		{Name: "repo", Branch: "main", WorktreePath: "/repo", Status: statusMain},
		{Name: "VOI-42", Branch: "feature", WorktreePath: "/wt/42", Status: statusActive, ClaudeSessionID: "abc"},
		{Name: "orphan", Branch: "orph", WorktreePath: "/wt/orphan", Status: statusOrphaned},
		{Name: "dead", Branch: "old", WorktreePath: "/wt/dead", Status: statusDead, ClaudeSessionID: "old-id"},
	}
	renderTable(&buf, entries)
	output := buf.String()

	// Check entries are present
	if !strings.Contains(output, "VOI-42") {
		t.Error("expected VOI-42 in output")
	}
	if !strings.Contains(output, "abc") {
		t.Error("expected Claude session ID 'abc' in output")
	}
	if !strings.Contains(output, "orphan") {
		t.Error("expected orphan in output")
	}
	if !strings.Contains(output, "dead") {
		t.Error("expected dead in output")
	}
	if !strings.Contains(output, "old-id") {
		t.Error("expected Claude session ID 'old-id' in output")
	}
}

// --- cmdListWithDeps tests ---

func TestCmdListWithDeps_Success(t *testing.T) {
	var buf bytes.Buffer
	deps := listDeps{
		sessionsDir: "/fake/sessions",
		repoName:    "myrepo",
		listSessions: func(sessionsDir, repoName string) (map[string]*session.Session, error) {
			return map[string]*session.Session{
				"VOI-42": {
					Branch:      "feature",
					WorktreeDir: "/wt/VOI-42",
				},
			}, nil
		},
		listWorktrees: func() ([]git.WorktreeEntry, error) {
			return []git.WorktreeEntry{
				{Path: "/home/user/repo", Branch: "main"},
				{Path: "/wt/VOI-42", Branch: "feature"},
			}, nil
		},
		mainWorktree: func() (string, error) {
			return "/home/user/repo", nil
		},
		output: &buf,
	}

	err := cmdListWithDeps(deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "VOI-42") {
		t.Error("expected VOI-42 in output")
	}
	if !strings.Contains(output, "repo") {
		t.Error("expected repo in output")
	}
}

func TestCmdListWithDeps_SessionsError(t *testing.T) {
	deps := listDeps{
		listSessions: func(_, _ string) (map[string]*session.Session, error) {
			return nil, fmt.Errorf("disk failure")
		},
		listWorktrees: func() ([]git.WorktreeEntry, error) { return nil, nil },
		mainWorktree:  func() (string, error) { return "", nil },
		output:        &bytes.Buffer{},
	}

	err := cmdListWithDeps(deps)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "reading sessions") {
		t.Errorf("expected 'reading sessions' in error, got: %v", err)
	}
}

func TestCmdListWithDeps_WorktreeError(t *testing.T) {
	deps := listDeps{
		listSessions: func(_, _ string) (map[string]*session.Session, error) {
			return map[string]*session.Session{}, nil
		},
		listWorktrees: func() ([]git.WorktreeEntry, error) {
			return nil, fmt.Errorf("git failure")
		},
		mainWorktree: func() (string, error) { return "", nil },
		output:       &bytes.Buffer{},
	}

	err := cmdListWithDeps(deps)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "listing worktrees") {
		t.Errorf("expected 'listing worktrees' in error, got: %v", err)
	}
}

func TestCmdListWithDeps_MainWorktreeError(t *testing.T) {
	deps := listDeps{
		listSessions: func(_, _ string) (map[string]*session.Session, error) {
			return map[string]*session.Session{}, nil
		},
		listWorktrees: func() ([]git.WorktreeEntry, error) {
			return []git.WorktreeEntry{}, nil
		},
		mainWorktree: func() (string, error) {
			return "", fmt.Errorf("not a git repo")
		},
		output: &bytes.Buffer{},
	}

	err := cmdListWithDeps(deps)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "finding main worktree") {
		t.Errorf("expected 'finding main worktree' in error, got: %v", err)
	}
}

func TestCmdListWithDeps_EmptyState(t *testing.T) {
	var buf bytes.Buffer
	deps := listDeps{
		listSessions: func(_, _ string) (map[string]*session.Session, error) {
			return map[string]*session.Session{}, nil
		},
		listWorktrees: func() ([]git.WorktreeEntry, error) {
			return []git.WorktreeEntry{}, nil
		},
		mainWorktree: func() (string, error) {
			return "/home/user/repo", nil
		},
		output: &buf,
	}

	err := cmdListWithDeps(deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "No sessions or worktrees found") {
		t.Error("expected empty state message")
	}
}

// --- maxLen test ---

func TestMaxLen(t *testing.T) {
	if maxLen(3, 5) != 5 {
		t.Error("expected 5")
	}
	if maxLen(7, 2) != 7 {
		t.Error("expected 7")
	}
	if maxLen(4, 4) != 4 {
		t.Error("expected 4")
	}
}
