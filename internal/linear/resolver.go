package linear

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jonbo372/lw/internal/config"
)

// linearConfig is the on-disk format for ~/.lw/linear.json.
type linearConfig struct {
	APIKeys map[string]string `json:"API_KEYS"`
}

// VerboseFunc is a callback for emitting verbose progress messages.
type VerboseFunc func(msg string, args ...any)

// Resolver resolves which LINEAR_API_KEY* env var to use for a given ticket.
type Resolver struct {
	configPath string
	fetcher    func(apiKey, ticketID string) (*Ticket, error)
	verboseFn  VerboseFunc
}

// NewResolver creates a Resolver with the given config path and fetcher function.
func NewResolver(configPath string, fetcher func(apiKey, ticketID string) (*Ticket, error)) *Resolver {
	return &Resolver{
		configPath: configPath,
		fetcher:    fetcher,
	}
}

// SetVerbose sets a callback for verbose progress output.
// Returns the Resolver for method chaining.
func (r *Resolver) SetVerbose(fn VerboseFunc) *Resolver {
	r.verboseFn = fn
	return r
}

// logVerbose emits a verbose message if a callback is set.
func (r *Resolver) logVerbose(msg string, args ...any) {
	if r.verboseFn != nil {
		r.verboseFn(msg, args...)
	}
}

// DefaultConfigPath returns the default path for linear.json.
func DefaultConfigPath() string {
	return filepath.Join(os.Getenv("HOME"), ".lw", "linear.json")
}

// ResolveAndFetch determines the correct API key for a ticket and fetches it.
// Returns the env var name (not the secret), the ticket, and any error.
func (r *Resolver) ResolveAndFetch(ticketID string) (string, *Ticket, error) {
	prefix := extractPrefix(ticketID)
	r.logVerbose("Resolving API key for prefix %s (ticket %s)", prefix, ticketID)

	// Try cached mapping first
	cfg := r.loadConfig()
	if envVarName, ok := cfg.APIKeys[prefix]; ok {
		r.logVerbose("Found cached mapping: %s -> %s", prefix, envVarName)
		apiKeyValue := os.Getenv(envVarName)
		if apiKeyValue != "" {
			ticket, err := r.fetcher(apiKeyValue, ticketID)
			if err == nil {
				r.logVerbose("Cached key %s succeeded", envVarName)
				return envVarName, ticket, nil
			}
			r.logVerbose("Cached key %s failed, falling through to discovery", envVarName)
			// Stale mapping, fall through to discovery
		}
	}

	// Discovery: try each LINEAR_API_KEY* env var
	keyNames := config.LinearAPIKeys()
	r.logVerbose("Trying %d API key(s) for prefix %s...", len(keyNames), prefix)
	for _, envVarName := range keyNames {
		r.logVerbose("Trying %s...", envVarName)
		apiKeyValue := os.Getenv(envVarName)
		ticket, err := r.fetcher(apiKeyValue, ticketID)
		if err == nil {
			r.logVerbose("Key %s succeeded, saving mapping", envVarName)
			// Save the mapping
			cfg.APIKeys[prefix] = envVarName
			if err := r.saveConfig(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not save API key mapping: %v\n", err)
			}
			return envVarName, ticket, nil
		}
	}

	return "", nil, fmt.Errorf("no LINEAR_API_KEY* env var could resolve ticket %s", ticketID)
}

// extractPrefix extracts the project prefix from a ticket ID (e.g., "LIN" from "LIN-55").
func extractPrefix(ticketID string) string {
	idx := strings.Index(ticketID, "-")
	if idx < 0 {
		return strings.ToUpper(ticketID)
	}
	return strings.ToUpper(ticketID[:idx])
}

// loadConfig reads the linear.json file, returning an empty config on any error.
func (r *Resolver) loadConfig() linearConfig {
	data, err := os.ReadFile(r.configPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "warning: could not read %s: %v\n", r.configPath, err)
		}
		return linearConfig{APIKeys: make(map[string]string)}
	}

	var cfg linearConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return linearConfig{APIKeys: make(map[string]string)}
	}
	if cfg.APIKeys == nil {
		cfg.APIKeys = make(map[string]string)
	}
	return cfg
}

// saveConfig writes the linear.json file, creating parent directories as needed.
func (r *Resolver) saveConfig(cfg linearConfig) error {
	if err := os.MkdirAll(filepath.Dir(r.configPath), 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	if err := os.WriteFile(r.configPath, data, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}
