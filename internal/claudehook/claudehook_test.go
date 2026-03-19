package claudehook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestInstall_CreatesSettingsFile(t *testing.T) {
	worktreeDir := t.TempDir()

	err := Install(worktreeDir, "myrepo", "VOI-42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	settingsPath := filepath.Join(worktreeDir, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("failed to read settings file: %v", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("failed to parse settings JSON: %v", err)
	}

	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		t.Fatal("expected hooks object in settings")
	}

	sessionEnd, ok := hooks["SessionEnd"].([]any)
	if !ok {
		t.Fatal("expected SessionEnd array in hooks")
	}

	if len(sessionEnd) != 1 {
		t.Fatalf("expected 1 SessionEnd entry, got %d", len(sessionEnd))
	}

	entry := sessionEnd[0].(map[string]any)
	if entry["matcher"] != "" {
		t.Errorf("expected empty matcher, got %v", entry["matcher"])
	}

	innerHooks := entry["hooks"].([]any)
	if len(innerHooks) != 1 {
		t.Fatalf("expected 1 inner hook, got %d", len(innerHooks))
	}

	hook := innerHooks[0].(map[string]any)
	if hook["type"] != "command" {
		t.Errorf("expected type command, got %v", hook["type"])
	}

	expected := "lw session-end --repo myrepo --session VOI-42"
	if hook["command"] != expected {
		t.Errorf("expected command %q, got %v", expected, hook["command"])
	}
}

func TestInstall_PreservesExistingSettings(t *testing.T) {
	worktreeDir := t.TempDir()
	claudeDir := filepath.Join(worktreeDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	existing := map[string]any{
		"customSetting": "value",
		"permissions": map[string]any{
			"allow": []string{"read"},
		},
	}
	data, _ := json.MarshalIndent(existing, "", "  ")
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0644)

	err := Install(worktreeDir, "myrepo", "VOI-42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	if err != nil {
		t.Fatalf("failed to read settings: %v", err)
	}

	var settings map[string]any
	json.Unmarshal(result, &settings)

	if settings["customSetting"] != "value" {
		t.Error("existing customSetting was not preserved")
	}

	perms, ok := settings["permissions"].(map[string]any)
	if !ok {
		t.Error("existing permissions were not preserved")
	} else {
		allow := perms["allow"].([]any)
		if len(allow) != 1 || allow[0] != "read" {
			t.Error("existing permissions.allow was not preserved")
		}
	}

	hooks := settings["hooks"].(map[string]any)
	sessionEnd := hooks["SessionEnd"].([]any)
	if len(sessionEnd) != 1 {
		t.Fatalf("expected 1 SessionEnd entry, got %d", len(sessionEnd))
	}
}

func TestInstall_PreservesExistingHooks(t *testing.T) {
	worktreeDir := t.TempDir()
	claudeDir := filepath.Join(worktreeDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	existing := map[string]any{
		"hooks": map[string]any{
			"PreToolUse": []any{
				map[string]any{
					"matcher": "Bash",
					"hooks": []any{
						map[string]any{
							"type":    "command",
							"command": "echo hello",
						},
					},
				},
			},
		},
	}
	data, _ := json.MarshalIndent(existing, "", "  ")
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0644)

	err := Install(worktreeDir, "myrepo", "VOI-42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	if err != nil {
		t.Fatalf("failed to read settings: %v", err)
	}

	var settings map[string]any
	json.Unmarshal(result, &settings)

	hooks := settings["hooks"].(map[string]any)

	preToolUse, ok := hooks["PreToolUse"].([]any)
	if !ok {
		t.Fatal("PreToolUse hook was not preserved")
	}
	if len(preToolUse) != 1 {
		t.Errorf("expected 1 PreToolUse entry, got %d", len(preToolUse))
	}

	sessionEnd, ok := hooks["SessionEnd"].([]any)
	if !ok {
		t.Fatal("SessionEnd hook was not added")
	}
	if len(sessionEnd) != 1 {
		t.Errorf("expected 1 SessionEnd entry, got %d", len(sessionEnd))
	}
}

func TestInstall_ReplacesExistingSessionEndHook(t *testing.T) {
	worktreeDir := t.TempDir()
	claudeDir := filepath.Join(worktreeDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	existing := map[string]any{
		"hooks": map[string]any{
			"SessionEnd": []any{
				map[string]any{
					"matcher": "",
					"hooks": []any{
						map[string]any{
							"type":    "command",
							"command": "/old/script.sh",
						},
					},
				},
			},
		},
	}
	data, _ := json.MarshalIndent(existing, "", "  ")
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0644)

	err := Install(worktreeDir, "myrepo", "VOI-99")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	if err != nil {
		t.Fatalf("failed to read settings: %v", err)
	}

	var settings map[string]any
	json.Unmarshal(result, &settings)

	hooks := settings["hooks"].(map[string]any)
	sessionEnd := hooks["SessionEnd"].([]any)
	if len(sessionEnd) != 1 {
		t.Fatalf("expected 1 SessionEnd entry, got %d", len(sessionEnd))
	}

	entry := sessionEnd[0].(map[string]any)
	innerHooks := entry["hooks"].([]any)
	hook := innerHooks[0].(map[string]any)

	expected := "lw session-end --repo myrepo --session VOI-99"
	if hook["command"] != expected {
		t.Errorf("expected command %q, got %v", expected, hook["command"])
	}
}

func TestInstall_CreatesDotClaudeDirectory(t *testing.T) {
	worktreeDir := t.TempDir()

	err := Install(worktreeDir, "myrepo", "VOI-42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	claudeDir := filepath.Join(worktreeDir, ".claude")
	info, err := os.Stat(claudeDir)
	if err != nil {
		t.Fatalf(".claude directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected .claude to be a directory")
	}
}

func TestInstall_MalformedExistingJSON(t *testing.T) {
	worktreeDir := t.TempDir()
	claudeDir := filepath.Join(worktreeDir, ".claude")
	os.MkdirAll(claudeDir, 0755)
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte("{bad json"), 0644)

	err := Install(worktreeDir, "myrepo", "VOI-42")
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestInstall_UnwritableClaudeDir(t *testing.T) {
	err := Install("/dev/null/impossible", "myrepo", "VOI-42")
	if err == nil {
		t.Fatal("expected error for impossible directory path")
	}
}

func TestInstall_UnwritableSettingsFile(t *testing.T) {
	worktreeDir := t.TempDir()
	claudeDir := filepath.Join(worktreeDir, ".claude")
	os.MkdirAll(claudeDir, 0755)
	settingsPath := filepath.Join(claudeDir, "settings.json")
	os.WriteFile(settingsPath, []byte(`{}`), 0444)
	defer os.Chmod(settingsPath, 0644)

	err := Install(worktreeDir, "myrepo", "VOI-42")
	if err == nil {
		t.Fatal("expected error for read-only settings file")
	}
}

func TestInstall_UnreadableSettingsFile(t *testing.T) {
	worktreeDir := t.TempDir()
	claudeDir := filepath.Join(worktreeDir, ".claude")
	os.MkdirAll(claudeDir, 0755)
	settingsPath := filepath.Join(claudeDir, "settings.json")
	os.WriteFile(settingsPath, []byte(`{}`), 0000)
	defer os.Chmod(settingsPath, 0644)

	err := Install(worktreeDir, "myrepo", "VOI-42")
	if err == nil {
		t.Fatal("expected error for unreadable settings file")
	}
}

func TestInstall_RejectsUnsafeRepoName(t *testing.T) {
	for _, name := range []string{"my repo", "repo;rm -rf", "repo\nname", "$(evil)"} {
		err := Install(t.TempDir(), name, "VOI-42")
		if err == nil {
			t.Errorf("expected error for unsafe repo name %q", name)
		}
	}
}

func TestInstall_RejectsUnsafeSessionID(t *testing.T) {
	for _, id := range []string{"id with space", "id;cmd", "$(evil)", "id\nnewline"} {
		err := Install(t.TempDir(), "myrepo", id)
		if err == nil {
			t.Errorf("expected error for unsafe session ID %q", id)
		}
	}
}

func TestInstall_AcceptsSafeNames(t *testing.T) {
	for _, tc := range []struct{ repo, session string }{
		{"myrepo", "VOI-42"},
		{"my.repo", "fuzzy_cobra"},
		{"repo-name", "some-session.123"},
	} {
		err := Install(t.TempDir(), tc.repo, tc.session)
		if err != nil {
			t.Errorf("expected no error for repo=%q session=%q, got: %v", tc.repo, tc.session, err)
		}
	}
}

