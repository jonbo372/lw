package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jonbo372/lw/internal/config"
)

// Session represents the persisted metadata for an lw worktree session.
type Session struct {
	Branch          string    `json:"branch"`
	Ticket          string    `json:"ticket"`
	WorktreeDir     string    `json:"worktreeDir"`
	TmuxWindow      string    `json:"tmuxWindow"`
	CreatedAt       time.Time `json:"createdAt"`
	ClaudeSessionID string    `json:"claudeSessionId"`
}

// Path returns the file path for a session JSON file.
func Path(sessionsDir, repoName, sessionID string) string {
	return filepath.Join(sessionsDir, repoName, sessionID+".json")
}

// Create writes a new session JSON file. It sets CreatedAt to the current time
// and ClaudeSessionID to empty string. Returns the path of the created file.
func Create(sessionsDir, repoName, sessionID string, s *Session) (string, error) {
	dir := filepath.Join(sessionsDir, repoName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("creating session directory: %w", err)
	}

	s.CreatedAt = time.Now().UTC()
	s.ClaudeSessionID = ""

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling session: %w", err)
	}

	path := Path(sessionsDir, repoName, sessionID)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("writing session file: %w", err)
	}

	return path, nil
}

// Load reads a session JSON file. Returns nil, nil if the file does not exist.
func Load(sessionsDir, repoName, sessionID string) (*Session, error) {
	path := Path(sessionsDir, repoName, sessionID)

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading session file: %w", err)
	}

	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing session file: %w", err)
	}

	return &s, nil
}

// UpdateClaudeSessionID reads the session file, updates the Claude session ID,
// and writes it back.
func UpdateClaudeSessionID(sessionsDir, repoName, sessionID, claudeSessionID string) error {
	s, err := Load(sessionsDir, repoName, sessionID)
	if err != nil {
		return err
	}
	if s == nil {
		return fmt.Errorf("session file not found: %s", Path(sessionsDir, repoName, sessionID))
	}

	s.ClaudeSessionID = claudeSessionID

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling session: %w", err)
	}

	path := Path(sessionsDir, repoName, sessionID)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing session file: %w", err)
	}

	return nil
}

// DefaultSessionsDir returns the default sessions directory from config.
func DefaultSessionsDir() string {
	return config.SessionsDir()
}

// ListAll reads all session JSON files in sessionsDir/repoName/ and returns
// them as a map keyed by session ID (the filename without .json extension).
func ListAll(sessionsDir, repoName string) (map[string]*Session, error) {
	dir := filepath.Join(sessionsDir, repoName)

	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]*Session{}, nil
		}
		return nil, fmt.Errorf("reading sessions directory: %w", err)
	}

	result := make(map[string]*Session)
	for _, entry := range dirEntries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}

		sessionID := strings.TrimSuffix(name, ".json")
		sess, err := Load(sessionsDir, repoName, sessionID)
		if err != nil {
			return nil, fmt.Errorf("loading session %s: %w", sessionID, err)
		}
		if sess != nil {
			result[sessionID] = sess
		}
	}

	return result, nil
}
