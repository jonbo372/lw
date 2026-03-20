package config

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func WorktreeHome() string {
	if v := os.Getenv("WORKTREE_HOME"); v != "" {
		return v
	}
	return filepath.Join(os.Getenv("HOME"), ".superclaude")
}

// LinearAPIKeys returns the names of all env vars starting with LINEAR_API_KEY
// that have non-empty values. The names are sorted for deterministic ordering.
func LinearAPIKeys() []string {
	var keys []string
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		name, value := parts[0], parts[1]
		if strings.HasPrefix(name, "LINEAR_API_KEY") && value != "" {
			keys = append(keys, name)
		}
	}
	sort.Strings(keys)
	return keys
}

// SessionsDir returns the directory where session JSON files are stored.
func SessionsDir() string {
	return filepath.Join(os.Getenv("HOME"), ".lw", "sessions")
}
