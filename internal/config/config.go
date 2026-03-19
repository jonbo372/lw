package config

import (
	"os"
	"path/filepath"
)

func WorktreeHome() string {
	if v := os.Getenv("WORKTREE_HOME"); v != "" {
		return v
	}
	return filepath.Join(os.Getenv("HOME"), ".superclaude")
}

func LinearAPIKey() string {
	return os.Getenv("LINEAR_API_KEY")
}

// SessionsDir returns the directory where session JSON files are stored.
func SessionsDir() string {
	return filepath.Join(os.Getenv("HOME"), ".lw", "sessions")
}
