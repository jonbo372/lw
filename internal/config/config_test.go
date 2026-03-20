package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// clearLinearAPIKeys uses t.Setenv to blank all LINEAR_API_KEY* env vars,
// ensuring tests start with a clean slate and values are restored after.
func clearLinearAPIKeys(t *testing.T) {
	t.Helper()
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if strings.HasPrefix(parts[0], "LINEAR_API_KEY") {
			t.Setenv(parts[0], "")
		}
	}
}

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

func TestSessionsDir(t *testing.T) {
	home := os.Getenv("HOME")
	expected := filepath.Join(home, ".lw", "sessions")
	if got := SessionsDir(); got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestLinearAPIKeys_Single(t *testing.T) {
	clearLinearAPIKeys(t)
	t.Setenv("LINEAR_API_KEY", "key1")
	keys := LinearAPIKeys()
	if len(keys) != 1 || keys[0] != "LINEAR_API_KEY" {
		t.Errorf("expected [LINEAR_API_KEY], got %v", keys)
	}
}

func TestLinearAPIKeys_Multiple(t *testing.T) {
	clearLinearAPIKeys(t)
	t.Setenv("LINEAR_API_KEY", "key1")
	t.Setenv("LINEAR_API_KEY2", "key2")
	t.Setenv("LINEAR_API_KEY_WORK", "key3")
	keys := LinearAPIKeys()
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d: %v", len(keys), keys)
	}
	// Should be sorted for deterministic ordering
	expected := []string{"LINEAR_API_KEY", "LINEAR_API_KEY2", "LINEAR_API_KEY_WORK"}
	for i, exp := range expected {
		if keys[i] != exp {
			t.Errorf("keys[%d] = %s, want %s", i, keys[i], exp)
		}
	}
}

func TestLinearAPIKeys_None(t *testing.T) {
	clearLinearAPIKeys(t)
	keys := LinearAPIKeys()
	if len(keys) != 0 {
		t.Errorf("expected empty, got %v", keys)
	}
}

func TestLinearAPIKeys_SkipsEmpty(t *testing.T) {
	clearLinearAPIKeys(t)
	t.Setenv("LINEAR_API_KEY", "")
	t.Setenv("LINEAR_API_KEY2", "has-value")
	keys := LinearAPIKeys()
	if len(keys) != 1 || keys[0] != "LINEAR_API_KEY2" {
		t.Errorf("expected [LINEAR_API_KEY2], got %v", keys)
	}
}
