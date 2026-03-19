package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWorktreeHome_Default(t *testing.T) {
	t.Setenv("WORKTREE_HOME", "")
	home := os.Getenv("HOME")
	expected := filepath.Join(home, ".superclaude")
	if got := WorktreeHome(); got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestWorktreeHome_Override(t *testing.T) {
	t.Setenv("WORKTREE_HOME", "/custom/path")
	if got := WorktreeHome(); got != "/custom/path" {
		t.Errorf("expected /custom/path, got %s", got)
	}
}

func TestLinearAPIKey(t *testing.T) {
	t.Setenv("LINEAR_API_KEY", "test-key")
	if got := LinearAPIKey(); got != "test-key" {
		t.Errorf("expected test-key, got %s", got)
	}
}

func TestLinearAPIKey_Empty(t *testing.T) {
	t.Setenv("LINEAR_API_KEY", "")
	if got := LinearAPIKey(); got != "" {
		t.Errorf("expected empty, got %s", got)
	}
}

func TestSessionsDir(t *testing.T) {
	home := os.Getenv("HOME")
	expected := filepath.Join(home, ".lw", "sessions")
	if got := SessionsDir(); got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}
