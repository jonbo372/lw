package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/jonbo372/lw/internal/session"
)

func TestCmdSessionEnd_UpdatesSessionFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	sessionsDir := dir + "/.lw/sessions"

	// Create a session file
	sess := &session.Session{
		Branch:      "test-branch",
		Ticket:      "VOI-42",
		WorktreeDir: "/tmp/wt",
	}
	_, err := session.Create(sessionsDir, "myrepo", "VOI-42", sess)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Simulate stdin with Claude Code session data
	input := map[string]string{"session_id": "claude-abc-123"}
	data, _ := json.Marshal(input)

	r, w, _ := os.Pipe()
	w.Write(data)
	w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	err = cmdSessionEnd("myrepo", "VOI-42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify session was updated
	loaded, err := session.Load(sessionsDir, "myrepo", "VOI-42")
	if err != nil {
		t.Fatalf("failed to load session: %v", err)
	}
	if loaded.ClaudeSessionID != "claude-abc-123" {
		t.Errorf("expected claudeSessionId 'claude-abc-123', got %q", loaded.ClaudeSessionID)
	}
}

func TestCmdSessionEnd_EmptySessionID(t *testing.T) {
	input := map[string]string{"session_id": ""}
	data, _ := json.Marshal(input)

	r, w, _ := os.Pipe()
	w.Write(data)
	w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	// Should return nil (no-op) when session_id is empty
	err := cmdSessionEnd("myrepo", "VOI-42")
	if err != nil {
		t.Fatalf("expected no error for empty session_id, got: %v", err)
	}
}

func TestCmdSessionEnd_InvalidJSON(t *testing.T) {
	r, w, _ := os.Pipe()
	w.Write([]byte("{bad json"))
	w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	err := cmdSessionEnd("myrepo", "VOI-42")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
