package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCreate_WritesJSONFile(t *testing.T) {
	dir := t.TempDir()

	s := &Session{
		Branch:      "jonbo372/VOI-42/add-auth",
		Ticket:      "VOI-42",
		WorktreeDir: "/home/user/.superclaude/repo/VOI-42",
		TmuxWindow:  "[repo] VOI-42: Add auth",
	}

	path, err := Create(dir, "myrepo", "VOI-42", s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedPath := filepath.Join(dir, "myrepo", "VOI-42.json")
	if path != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read session file: %v", err)
	}

	var loaded Session
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("failed to parse session JSON: %v", err)
	}

	if loaded.Branch != "jonbo372/VOI-42/add-auth" {
		t.Errorf("expected branch jonbo372/VOI-42/add-auth, got %s", loaded.Branch)
	}
	if loaded.Ticket != "VOI-42" {
		t.Errorf("expected ticket VOI-42, got %s", loaded.Ticket)
	}
	if loaded.WorktreeDir != "/home/user/.superclaude/repo/VOI-42" {
		t.Errorf("expected worktreeDir, got %s", loaded.WorktreeDir)
	}
	if loaded.TmuxWindow != "[repo] VOI-42: Add auth" {
		t.Errorf("expected tmuxWindow, got %s", loaded.TmuxWindow)
	}
	if loaded.ClaudeSessionID != "" {
		t.Errorf("expected empty claudeSessionId, got %s", loaded.ClaudeSessionID)
	}
	if loaded.CreatedAt.IsZero() {
		t.Error("expected createdAt to be set")
	}
	// CreatedAt should be recent (within last 5 seconds)
	if time.Since(loaded.CreatedAt) > 5*time.Second {
		t.Errorf("createdAt too old: %v", loaded.CreatedAt)
	}
}

func TestCreate_ScratchSession(t *testing.T) {
	dir := t.TempDir()

	s := &Session{
		Branch:      "fuzzy_cobra",
		Ticket:      "",
		WorktreeDir: "/home/user/.superclaude/repo/fuzzy_cobra",
		TmuxWindow:  "[repo] fuzzy_cobra: scratch",
	}

	path, err := Create(dir, "myrepo", "fuzzy_cobra", s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedPath := filepath.Join(dir, "myrepo", "fuzzy_cobra.json")
	if path != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read session file: %v", err)
	}

	var loaded Session
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("failed to parse session JSON: %v", err)
	}

	if loaded.Ticket != "" {
		t.Errorf("expected empty ticket for scratch, got %s", loaded.Ticket)
	}
}

func TestCreate_CreatesDirectories(t *testing.T) {
	dir := t.TempDir()
	sessionsDir := filepath.Join(dir, "nested", "deep")

	s := &Session{
		Branch:      "test-branch",
		WorktreeDir: "/tmp/wt",
	}

	_, err := Create(sessionsDir, "myrepo", "test-id", s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify directory was created
	info, err := os.Stat(filepath.Join(sessionsDir, "myrepo"))
	if err != nil {
		t.Fatalf("directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory")
	}
}

func TestLoad_ExistingSession(t *testing.T) {
	dir := t.TempDir()

	// Create a session file manually
	repoDir := filepath.Join(dir, "myrepo")
	os.MkdirAll(repoDir, 0755)

	s := Session{
		Branch:           "test-branch",
		Ticket:           "VOI-99",
		WorktreeDir:      "/tmp/wt",
		TmuxWindow:       "win",
		CreatedAt:        time.Now().UTC(),
		ClaudeSessionID:  "abc-123-def",
	}
	data, _ := json.Marshal(s)
	os.WriteFile(filepath.Join(repoDir, "VOI-99.json"), data, 0644)

	loaded, err := Load(dir, "myrepo", "VOI-99")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loaded.ClaudeSessionID != "abc-123-def" {
		t.Errorf("expected claudeSessionId abc-123-def, got %s", loaded.ClaudeSessionID)
	}
	if loaded.Ticket != "VOI-99" {
		t.Errorf("expected ticket VOI-99, got %s", loaded.Ticket)
	}
}

func TestLoad_NonExistentFile(t *testing.T) {
	dir := t.TempDir()

	loaded, err := Load(dir, "myrepo", "nonexistent")
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if loaded != nil {
		t.Error("expected nil session for missing file")
	}
}

func TestLoad_NonExistentDirectory(t *testing.T) {
	loaded, err := Load("/nonexistent/path", "myrepo", "test")
	if err != nil {
		t.Fatalf("expected no error for missing directory, got: %v", err)
	}
	if loaded != nil {
		t.Error("expected nil session for missing directory")
	}
}

func TestLoad_MalformedJSON(t *testing.T) {
	dir := t.TempDir()
	repoDir := filepath.Join(dir, "myrepo")
	os.MkdirAll(repoDir, 0755)
	os.WriteFile(filepath.Join(repoDir, "bad.json"), []byte("{invalid"), 0644)

	_, err := Load(dir, "myrepo", "bad")
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestUpdateClaudeSessionID(t *testing.T) {
	dir := t.TempDir()

	s := &Session{
		Branch:      "test-branch",
		Ticket:      "VOI-99",
		WorktreeDir: "/tmp/wt",
		TmuxWindow:  "win",
	}

	_, err := Create(dir, "myrepo", "VOI-99", s)
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}

	err = UpdateClaudeSessionID(dir, "myrepo", "VOI-99", "session-id-456")
	if err != nil {
		t.Fatalf("failed to update: %v", err)
	}

	loaded, err := Load(dir, "myrepo", "VOI-99")
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}
	if loaded.ClaudeSessionID != "session-id-456" {
		t.Errorf("expected session-id-456, got %s", loaded.ClaudeSessionID)
	}
	// Other fields should be preserved
	if loaded.Branch != "test-branch" {
		t.Errorf("expected branch preserved, got %s", loaded.Branch)
	}
}

func TestUpdateClaudeSessionID_NonExistentFile(t *testing.T) {
	dir := t.TempDir()

	err := UpdateClaudeSessionID(dir, "myrepo", "nonexistent", "id")
	if err == nil {
		t.Fatal("expected error for non-existent session file")
	}
}

func TestSessionPath(t *testing.T) {
	p := Path("/home/user/.lw/sessions", "myrepo", "VOI-42")
	expected := "/home/user/.lw/sessions/myrepo/VOI-42.json"
	if p != expected {
		t.Errorf("expected %s, got %s", expected, p)
	}
}

func TestCreate_FailsOnReadOnlyDir(t *testing.T) {
	// Use a path that cannot be created
	_, err := Create("/dev/null/impossible", "myrepo", "test", &Session{Branch: "b"})
	if err == nil {
		t.Fatal("expected error for impossible directory path")
	}
}

func TestCreate_FailsOnUnwritableDir(t *testing.T) {
	dir := t.TempDir()
	repoDir := filepath.Join(dir, "myrepo")
	os.MkdirAll(repoDir, 0755)
	// Make directory read-only
	os.Chmod(repoDir, 0444)
	defer os.Chmod(repoDir, 0755)

	_, err := Create(dir, "myrepo", "test", &Session{Branch: "b"})
	if err == nil {
		t.Fatal("expected error writing to read-only directory")
	}
}

func TestLoad_PermissionDenied(t *testing.T) {
	dir := t.TempDir()
	repoDir := filepath.Join(dir, "myrepo")
	os.MkdirAll(repoDir, 0755)
	filePath := filepath.Join(repoDir, "test.json")
	os.WriteFile(filePath, []byte(`{"branch":"b"}`), 0644)
	// Make file unreadable
	os.Chmod(filePath, 0000)
	defer os.Chmod(filePath, 0644)

	_, err := Load(dir, "myrepo", "test")
	if err == nil {
		t.Fatal("expected error for unreadable file")
	}
}

func TestUpdateClaudeSessionID_WriteError(t *testing.T) {
	dir := t.TempDir()

	s := &Session{Branch: "b", WorktreeDir: "/tmp/wt"}
	_, err := Create(dir, "myrepo", "test", s)
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}

	// Make the file read-only and the directory read-only so it cannot be overwritten
	filePath := Path(dir, "myrepo", "test")
	os.Chmod(filePath, 0444)
	repoDir := filepath.Join(dir, "myrepo")
	os.Chmod(repoDir, 0555)
	defer func() {
		os.Chmod(repoDir, 0755)
		os.Chmod(filePath, 0644)
	}()

	err = UpdateClaudeSessionID(dir, "myrepo", "test", "sid")
	if err == nil {
		t.Fatal("expected error when writing to read-only file")
	}
}

func TestUpdateClaudeSessionID_LoadError(t *testing.T) {
	dir := t.TempDir()
	repoDir := filepath.Join(dir, "myrepo")
	os.MkdirAll(repoDir, 0755)

	// Write a file that can't be read
	filePath := filepath.Join(repoDir, "test.json")
	os.WriteFile(filePath, []byte(`{"branch":"b"}`), 0000)
	defer os.Chmod(filePath, 0644)

	err := UpdateClaudeSessionID(dir, "myrepo", "test", "sid")
	if err == nil {
		t.Fatal("expected error for unreadable session file")
	}
}
