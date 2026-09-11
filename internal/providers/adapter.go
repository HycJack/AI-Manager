// Package providers implements per-agent adapters for reading and writing
// provider configuration files (model, base URL, API key, etc.).
//
// Each agent stores its provider config in a different format (JSON, TOML,
// YAML, VS Code settings). This package abstracts that heterogeneity behind a
// ProviderAdapter interface, exposing a common ProviderConfig view to the
// frontend.
//
// API keys are always masked in Read() output (only the last 4 characters
// are shown). Write() accepts masked values ("****abcd") and preserves the
// original key unchanged.
package providers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ai-manager/internal/agents"
)

// ProviderConfig is the common view of an agent's provider configuration.
// Agent-specific fields that don't fit the common schema go into Extra.
type ProviderConfig struct {
	Agent    agents.AgentKind `json:"agent"`
	Path     string           `json:"path"`     // resolved config file path
	Format   string           `json:"format"`   // "json", "toml", "yaml", "vscode-settings"
	Exists   bool             `json:"exists"`   // true if the config file exists on disk
	Provider string           `json:"provider"` // "openai", "anthropic", "deepseek", "custom", "unknown"
	Model    string           `json:"model"`    // model identifier (e.g. "claude-3.5-sonnet")
	BaseURL  string           `json:"base_url"` // API base URL (empty if using agent default)
	APIKey   string           `json:"api_key"`  // masked: "****abcd" (only last 4 chars shown)
	Extra    map[string]any   `json:"extra"`    // agent-specific fields
	Errors   []string         `json:"errors"`   // field-level read errors
	Raw      string           `json:"raw"`      // raw file content (for non-JSON formats)
}

// ProviderAdapter reads and writes a single agent's provider configuration.
type ProviderAdapter interface {
	// Agent returns the AgentKind this adapter handles.
	Agent() agents.AgentKind

	// Read loads the provider config from disk.
	// If the file doesn't exist, returns ProviderConfig{Exists: false}.
	Read() (*ProviderConfig, error)

	// Write persists the provider config to disk.
	// Masked API keys ("****abcd") are preserved as-is (original key kept).
	Write(cfg *ProviderConfig) error
}

// registry is the global adapter registry, keyed by AgentKind.
var registry = map[agents.AgentKind]ProviderAdapter{}

// Register adds an adapter to the registry. Panics on duplicate registration.
func Register(a ProviderAdapter) {
	if _, exists := registry[a.Agent()]; exists {
		panic(fmt.Sprintf("providers: adapter already registered for %q", a.Agent()))
	}
	registry[a.Agent()] = a
}

// For returns the adapter for the given agent, or nil if not registered.
func For(agent agents.AgentKind) ProviderAdapter {
	return registry[agent]
}

// All returns all registered adapters in insertion order.
func All() []ProviderAdapter {
	result := make([]ProviderAdapter, 0, len(registry))
	for _, a := range registry {
		result = append(result, a)
	}
	return result
}

// --- shared helpers ---

// MaskAPIKey returns a masked version of an API key, showing only the last
// 4 characters. Empty keys return empty string.
func MaskAPIKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 4 {
		return "****"
	}
	return "****" + key[len(key)-4:]
}

// IsMaskedAPIKey returns true if the key looks like a masked value ("****abcd").
func IsMaskedAPIKey(key string) bool {
	return strings.HasPrefix(key, "****")
}

// resolveConfigPath returns the resolved config file path for an agent.
// Falls back to DefaultAgentPaths if not provided.
func resolveConfigPath(agent agents.AgentKind) string {
	paths := agents.DefaultAgentPaths()
	p, ok := paths[agent]
	if !ok {
		return ""
	}
	resolved := agents.ResolveAgentPaths(p, agent, "")
	return resolved.ProviderConfigPath
}

// readJSONConfig reads a JSON config file and returns the raw map.
func readJSONConfig(path string) (map[string]any, bool, []string) {
	exists := fileExists(path)
	var errs []string
	var data map[string]any

	if !exists {
		return nil, false, errs
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, true, []string{fmt.Sprintf("read error: %v", err)}
	}

	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, true, []string{fmt.Sprintf("parse error: %v", err)}
	}

	return data, true, errs
}

// fileExists returns true if the file exists and is a regular file.
func fileExists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

// extractString extracts a string value from a nested map.
func extractString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

// writeJSONConfig writes a JSON config file with indentation.
func writeJSONConfig(path string, data map[string]any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

// inferProviderFromJSON infers the provider name from a JSON config map.
func inferProviderFromJSON(data map[string]any) string {
	// Check for explicit provider field
	if p := extractString(data, "provider", "model_provider", "modelProvider"); p != "" {
		return normalizeProvider(p)
	}
	// Infer from base URL (even without API key)
	base := strings.ToLower(extractString(data, "baseUrl", "base_url", "apiBase", "baseURL"))
	if base != "" {
		switch {
		case strings.Contains(base, "anthropic"):
			return "anthropic"
		case strings.Contains(base, "openai"):
			return "openai"
		case strings.Contains(base, "deepseek"):
			return "deepseek"
		case strings.Contains(base, "google") || strings.Contains(base, "gemini"):
			return "google"
		default:
			return "custom"
		}
	}
	// Infer from API key presence (default to openai)
	if extractString(data, "apiKey", "api_key", "anthropic_api_key") != "" {
		return "openai"
	}
	return "unknown"
}

// inferProviderFromRaw infers the provider name from raw config content.
func inferProviderFromRaw(raw string) string {
	lower := strings.ToLower(raw)
	switch {
	case strings.Contains(lower, "anthropic"):
		return "anthropic"
	case strings.Contains(lower, "openai"):
		return "openai"
	case strings.Contains(lower, "deepseek"):
		return "deepseek"
	case strings.Contains(lower, "gemini") || strings.Contains(lower, "google"):
		return "google"
	default:
		return "unknown"
	}
}

// normalizeProvider normalizes a provider name to lowercase.
func normalizeProvider(p string) string {
	return strings.ToLower(strings.TrimSpace(p))
}
