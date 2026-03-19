package claudehook

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// safeArg matches strings that are safe to embed unquoted in a shell command:
// alphanumeric, hyphens, underscores, and dots only.
var safeArg = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// Install adds a SessionEnd hook to .claude/settings.json in the worktree directory.
// The hook invokes `lw session-end --repo <repo> --session <sessionID>` which reads
// the Claude Code session data from stdin and updates the session JSON file.
// It preserves any existing settings and hooks in the file.
// If the file does not exist, it creates it. If a SessionEnd hook already exists,
// it replaces it with the new command.
func Install(worktreeDir, repoName, sessionID string) error {
	if !safeArg.MatchString(repoName) {
		return fmt.Errorf("unsafe repo name for hook command: %q", repoName)
	}
	if !safeArg.MatchString(sessionID) {
		return fmt.Errorf("unsafe session ID for hook command: %q", sessionID)
	}
	claudeDir := filepath.Join(worktreeDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return fmt.Errorf("creating .claude directory: %w", err)
	}

	settingsPath := filepath.Join(claudeDir, "settings.json")

	// Load existing settings or start fresh
	settings := make(map[string]any)
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("reading settings file: %w", err)
		}
	} else {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("parsing settings file: %w", err)
		}
	}

	// Ensure hooks map exists
	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		hooks = make(map[string]any)
	}

	// Build the SessionEnd hook entry
	command := fmt.Sprintf("lw session-end --repo %s --session %s", repoName, sessionID)
	hookEntry := []any{
		map[string]any{
			"matcher": "",
			"hooks": []any{
				map[string]any{
					"type":    "command",
					"command": command,
				},
			},
		},
	}

	hooks["SessionEnd"] = hookEntry
	settings["hooks"] = hooks

	// Write back
	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, out, 0644); err != nil {
		return fmt.Errorf("writing settings file: %w", err)
	}

	return nil
}

