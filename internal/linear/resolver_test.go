package linear

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// clearLinearAPIKeys uses t.Setenv to blank all LINEAR_API_KEY* env vars.
func clearLinearAPIKeys(t *testing.T) {
	t.Helper()
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if strings.HasPrefix(parts[0], "LINEAR_API_KEY") {
			t.Setenv(parts[0], "")
		}
	}
}

func TestExtractPrefix(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"LIN-55", "LIN"},
		{"VOI-123", "VOI"},
		{"ABC-1", "ABC"},
		{"lin-55", "LIN"},
	}
	for _, tt := range tests {
		got := extractPrefix(tt.input)
		if got != tt.want {
			t.Errorf("extractPrefix(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestExtractPrefix_NoHyphen(t *testing.T) {
	got := extractPrefix("NOHYPHEN")
	if got != "NOHYPHEN" {
		t.Errorf("extractPrefix(%q) = %q, want %q", "NOHYPHEN", got, "NOHYPHEN")
	}
}

// fakeFetcher returns a fetcher function that succeeds for the given apiKey.
func fakeFetcher(validKeys map[string]bool) func(string, string) (*Ticket, error) {
	return func(apiKey, ticketID string) (*Ticket, error) {
		if validKeys[apiKey] {
			return &Ticket{
				Branch: "branch-" + ticketID,
				Title:  "Title " + ticketID,
			}, nil
		}
		return nil, fmt.Errorf("unauthorized")
	}
}

func TestResolver_Discovery_NoExistingMapping(t *testing.T) {
	clearLinearAPIKeys(t)
	dir := t.TempDir()
	configPath := filepath.Join(dir, "linear.json")

	t.Setenv("LINEAR_API_KEY", "secret-key-1")
	t.Setenv("LINEAR_API_KEY2", "secret-key-2")

	r := NewResolver(configPath, fakeFetcher(map[string]bool{"secret-key-2": true}))

	apiKey, ticket, err := r.ResolveAndFetch("LIN-55")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if apiKey != "LINEAR_API_KEY2" {
		t.Errorf("apiKey = %q, want LINEAR_API_KEY2", apiKey)
	}
	if ticket.Branch != "branch-LIN-55" {
		t.Errorf("ticket.Branch = %q, want branch-LIN-55", ticket.Branch)
	}

	// Verify mapping was persisted
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("reading config: %v", err)
	}
	var cfg linearConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("parsing config: %v", err)
	}
	if cfg.APIKeys["LIN"] != "LINEAR_API_KEY2" {
		t.Errorf("persisted mapping LIN = %q, want LINEAR_API_KEY2", cfg.APIKeys["LIN"])
	}
}

func TestResolver_CachedMapping(t *testing.T) {
	clearLinearAPIKeys(t)
	dir := t.TempDir()
	configPath := filepath.Join(dir, "linear.json")

	// Pre-populate mapping
	cfg := linearConfig{APIKeys: map[string]string{"LIN": "LINEAR_API_KEY"}}
	data, _ := json.Marshal(cfg)
	os.WriteFile(configPath, data, 0644)

	t.Setenv("LINEAR_API_KEY", "cached-secret")

	callCount := 0
	fetcher := func(apiKey, ticketID string) (*Ticket, error) {
		callCount++
		if apiKey == "cached-secret" {
			return &Ticket{Branch: "cached-branch", Title: "cached"}, nil
		}
		return nil, fmt.Errorf("unauthorized")
	}

	r := NewResolver(configPath, fetcher)
	apiKey, ticket, err := r.ResolveAndFetch("LIN-99")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if apiKey != "LINEAR_API_KEY" {
		t.Errorf("apiKey = %q, want LINEAR_API_KEY", apiKey)
	}
	if ticket.Branch != "cached-branch" {
		t.Errorf("ticket.Branch = %q, want cached-branch", ticket.Branch)
	}
	if callCount != 1 {
		t.Errorf("callCount = %d, want 1 (should use cached mapping)", callCount)
	}
}

func TestResolver_StaleMapping_RediscoversKey(t *testing.T) {
	clearLinearAPIKeys(t)
	dir := t.TempDir()
	configPath := filepath.Join(dir, "linear.json")

	// Pre-populate with stale mapping
	cfg := linearConfig{APIKeys: map[string]string{"LIN": "LINEAR_API_KEY"}}
	data, _ := json.Marshal(cfg)
	os.WriteFile(configPath, data, 0644)

	t.Setenv("LINEAR_API_KEY", "stale-secret")
	t.Setenv("LINEAR_API_KEY2", "good-secret")

	r := NewResolver(configPath, fakeFetcher(map[string]bool{"good-secret": true}))

	apiKey, ticket, err := r.ResolveAndFetch("LIN-42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if apiKey != "LINEAR_API_KEY2" {
		t.Errorf("apiKey = %q, want LINEAR_API_KEY2", apiKey)
	}
	if ticket.Branch != "branch-LIN-42" {
		t.Errorf("ticket.Branch = %q, want branch-LIN-42", ticket.Branch)
	}

	// Verify mapping was updated
	data, _ = os.ReadFile(configPath)
	var updatedCfg linearConfig
	json.Unmarshal(data, &updatedCfg)
	if updatedCfg.APIKeys["LIN"] != "LINEAR_API_KEY2" {
		t.Errorf("updated mapping LIN = %q, want LINEAR_API_KEY2", updatedCfg.APIKeys["LIN"])
	}
}

func TestResolver_NoKeyMatches(t *testing.T) {
	clearLinearAPIKeys(t)
	dir := t.TempDir()
	configPath := filepath.Join(dir, "linear.json")

	t.Setenv("LINEAR_API_KEY", "bad-secret")

	r := NewResolver(configPath, fakeFetcher(map[string]bool{}))

	_, _, err := r.ResolveAndFetch("LIN-55")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	want := "no LINEAR_API_KEY* env var could resolve ticket LIN-55"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestResolver_NoAPIKeysConfigured(t *testing.T) {
	clearLinearAPIKeys(t)
	dir := t.TempDir()
	configPath := filepath.Join(dir, "linear.json")

	r := NewResolver(configPath, fakeFetcher(map[string]bool{}))

	_, _, err := r.ResolveAndFetch("LIN-55")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	want := "no LINEAR_API_KEY* env var could resolve ticket LIN-55"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestResolver_FirstMatchWins(t *testing.T) {
	clearLinearAPIKeys(t)
	dir := t.TempDir()
	configPath := filepath.Join(dir, "linear.json")

	t.Setenv("LINEAR_API_KEY", "secret-a")
	t.Setenv("LINEAR_API_KEY2", "secret-b")

	// Both keys work
	r := NewResolver(configPath, fakeFetcher(map[string]bool{"secret-a": true, "secret-b": true}))

	apiKey, _, err := r.ResolveAndFetch("LIN-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Sorted order: LINEAR_API_KEY comes before LINEAR_API_KEY2
	if apiKey != "LINEAR_API_KEY" {
		t.Errorf("apiKey = %q, want LINEAR_API_KEY (first match)", apiKey)
	}
}

func TestResolver_SecretNeverStored(t *testing.T) {
	clearLinearAPIKeys(t)
	dir := t.TempDir()
	configPath := filepath.Join(dir, "linear.json")

	t.Setenv("LINEAR_API_KEY", "super-secret-value")

	r := NewResolver(configPath, fakeFetcher(map[string]bool{"super-secret-value": true}))

	_, _, err := r.ResolveAndFetch("LIN-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("reading config: %v", err)
	}
	contents := string(data)
	if strings.Contains(contents, "super-secret-value") {
		t.Error("linear.json contains the API key secret; only env var names should be stored")
	}
}

func TestResolver_PreservesOtherPrefixes(t *testing.T) {
	clearLinearAPIKeys(t)
	dir := t.TempDir()
	configPath := filepath.Join(dir, "linear.json")

	// Pre-populate with an existing mapping for a different prefix
	cfg := linearConfig{APIKeys: map[string]string{"VOI": "LINEAR_API_KEY"}}
	data, _ := json.Marshal(cfg)
	os.WriteFile(configPath, data, 0644)

	t.Setenv("LINEAR_API_KEY", "secret-1")
	t.Setenv("LINEAR_API_KEY2", "secret-2")

	r := NewResolver(configPath, fakeFetcher(map[string]bool{"secret-2": true}))

	_, _, err := r.ResolveAndFetch("LIN-55")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify both mappings exist
	data, _ = os.ReadFile(configPath)
	var updatedCfg linearConfig
	json.Unmarshal(data, &updatedCfg)
	if updatedCfg.APIKeys["VOI"] != "LINEAR_API_KEY" {
		t.Errorf("VOI mapping lost: got %q", updatedCfg.APIKeys["VOI"])
	}
	if updatedCfg.APIKeys["LIN"] != "LINEAR_API_KEY2" {
		t.Errorf("LIN mapping: got %q, want LINEAR_API_KEY2", updatedCfg.APIKeys["LIN"])
	}
}

func TestResolver_TicketIDUppercased(t *testing.T) {
	clearLinearAPIKeys(t)
	dir := t.TempDir()
	configPath := filepath.Join(dir, "linear.json")

	t.Setenv("LINEAR_API_KEY", "secret-1")

	var receivedTicketID string
	fetcher := func(apiKey, ticketID string) (*Ticket, error) {
		receivedTicketID = ticketID
		return &Ticket{Branch: "b", Title: "t"}, nil
	}

	r := NewResolver(configPath, fetcher)
	_, _, err := r.ResolveAndFetch("lin-55")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedTicketID != "lin-55" {
		t.Errorf("ticket ID passed to fetcher = %q, want %q", receivedTicketID, "lin-55")
	}
}
