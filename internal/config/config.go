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
