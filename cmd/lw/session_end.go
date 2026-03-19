package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/jonbo372/lw/internal/config"
	"github.com/jonbo372/lw/internal/session"
)

// claudeSessionInput is the JSON structure Claude Code sends on stdin to SessionEnd hooks.
type claudeSessionInput struct {
	SessionID string `json:"session_id"`
}

// cmdSessionEnd reads a Claude Code session ID from stdin and persists it
// in the session JSON file for the given repo and session.
func cmdSessionEnd(repo, sessionID string) error {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("reading stdin: %w", err)
	}

	var input claudeSessionInput
	if err := json.Unmarshal(data, &input); err != nil {
		return fmt.Errorf("parsing stdin JSON: %w", err)
	}

	if input.SessionID == "" {
		return nil
	}

	return session.UpdateClaudeSessionID(config.SessionsDir(), repo, sessionID, input.SessionID)
}
