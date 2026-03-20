package linear

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolver_VerboseCallback_Discovery(t *testing.T) {
	clearLinearAPIKeys(t)
	dir := t.TempDir()
	configPath := filepath.Join(dir, "linear.json")

	t.Setenv("LINEAR_API_KEY", "secret-key-1")
	t.Setenv("LINEAR_API_KEY2", "secret-key-2")

	var messages []string
	r := NewResolver(configPath, fakeFetcher(map[string]bool{"secret-key-2": true}))
	r.SetVerbose(func(msg string, args ...any) {
		messages = append(messages, msg)
	})

	_, _, err := r.ResolveAndFetch("LIN-55")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(messages) == 0 {
		t.Fatal("expected verbose messages during discovery, got none")
	}

	// Should mention trying API keys
	found := false
	for _, m := range messages {
		if strings.Contains(m, "LIN") || strings.Contains(m, "API") || strings.Contains(m, "key") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected verbose messages about API key discovery, got: %v", messages)
	}
}

func TestResolver_VerboseCallback_CachedHit(t *testing.T) {
	clearLinearAPIKeys(t)
	dir := t.TempDir()
	configPath := filepath.Join(dir, "linear.json")

	cfg := linearConfig{APIKeys: map[string]string{"LIN": "LINEAR_API_KEY"}}
	data, _ := json.Marshal(cfg)
	os.WriteFile(configPath, data, 0644)

	t.Setenv("LINEAR_API_KEY", "cached-secret")

	var messages []string
	r := NewResolver(configPath, fakeFetcher(map[string]bool{"cached-secret": true}))
	r.SetVerbose(func(msg string, args ...any) {
		messages = append(messages, msg)
	})

	_, _, err := r.ResolveAndFetch("LIN-99")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(messages) == 0 {
		t.Fatal("expected verbose messages for cached hit, got none")
	}

	// Should mention using cached mapping
	found := false
	for _, m := range messages {
		if strings.Contains(m, "cache") || strings.Contains(m, "cached") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected verbose message about cached mapping, got: %v", messages)
	}
}

func TestResolver_VerboseCallback_Nil(t *testing.T) {
	clearLinearAPIKeys(t)
	dir := t.TempDir()
	configPath := filepath.Join(dir, "linear.json")

	t.Setenv("LINEAR_API_KEY", "secret-key-1")

	// No verbose callback set — should not panic
	r := NewResolver(configPath, fakeFetcher(map[string]bool{"secret-key-1": true}))

	_, _, err := r.ResolveAndFetch("LIN-55")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolver_VerboseCallback_StaleMapping(t *testing.T) {
	clearLinearAPIKeys(t)
	dir := t.TempDir()
	configPath := filepath.Join(dir, "linear.json")

	cfg := linearConfig{APIKeys: map[string]string{"LIN": "LINEAR_API_KEY"}}
	data, _ := json.Marshal(cfg)
	os.WriteFile(configPath, data, 0644)

	t.Setenv("LINEAR_API_KEY", "stale-secret")
	t.Setenv("LINEAR_API_KEY2", "good-secret")

	var messages []string
	r := NewResolver(configPath, fakeFetcher(map[string]bool{"good-secret": true}))
	r.SetVerbose(func(msg string, args ...any) {
		messages = append(messages, msg)
	})

	_, _, err := r.ResolveAndFetch("LIN-42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(messages) < 2 {
		t.Errorf("expected multiple verbose messages for stale+rediscovery, got %d: %v", len(messages), messages)
	}
}

func TestResolver_SetVerbose_ReturnsResolver(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "linear.json")
	r := NewResolver(configPath, fakeFetcher(map[string]bool{}))

	ret := r.SetVerbose(func(msg string, args ...any) {})
	if ret != r {
		t.Error("expected SetVerbose to return the receiver for chaining")
	}
}
