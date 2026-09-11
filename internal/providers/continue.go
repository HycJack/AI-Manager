package providers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ai-manager/internal/agents"
)

// continueAdapter reads ~/.continue/config.yaml.
// YAML is read as raw content; structured fields are extracted via string matching.
type continueAdapter struct{}

func (a *continueAdapter) Agent() agents.AgentKind { return agents.AgentContinue }

func (a *continueAdapter) Read() (*ProviderConfig, error) {
	path := expandHome("~/.continue/config.yaml")
	cfg := &ProviderConfig{
		Agent:  agents.AgentContinue,
		Path:   path,
		Format: "yaml",
		Exists: fileExists(path),
	}
	if !cfg.Exists {
		return cfg, nil
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		cfg.Errors = []string{fmt.Sprintf("read error: %v", err)}
		return cfg, nil
	}

	cfg.Raw = string(raw)
	cfg.Provider = inferProviderFromRaw(cfg.Raw)
	cfg.Model = extractYAMLValue(cfg.Raw, "model", "defaultModel")
	cfg.BaseURL = extractYAMLValue(cfg.Raw, "baseUrl", "base_url", "apiBase")
	cfg.APIKey = MaskAPIKey(extractYAMLValue(cfg.Raw, "apiKey", "api_key"))
	return cfg, nil
}

func (a *continueAdapter) Write(cfg *ProviderConfig) error {
	if cfg == nil || !cfg.Exists {
		return nil
	}
	// For YAML, we preserve the raw content and append/modify key fields.
	// This is a simplified approach; full YAML parsing would require a library.
	var lines []string
	for _, line := range strings.Split(cfg.Raw, "\n") {
		trimmed := strings.TrimSpace(line)
		// Skip existing model/apiKey/baseUrl lines (we'll add them fresh)
		if isYAMLKeyLine(trimmed, "model") || isYAMLKeyLine(trimmed, "apiKey") || isYAMLKeyLine(trimmed, "baseUrl") || isYAMLKeyLine(trimmed, "base_url") {
			continue
		}
		lines = append(lines, line)
	}
	if cfg.Model != "" {
		lines = append(lines, fmt.Sprintf("model: %q", cfg.Model))
	}
	if cfg.BaseURL != "" {
		lines = append(lines, fmt.Sprintf("baseUrl: %q", cfg.BaseURL))
	}
	if cfg.APIKey != "" && !IsMaskedAPIKey(cfg.APIKey) {
		lines = append(lines, fmt.Sprintf("apiKey: %q", cfg.APIKey))
	}
	if err := os.MkdirAll(filepath.Dir(cfg.Path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(cfg.Path, []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}

func init() { Register(&continueAdapter{}) }

// extractYAMLValue extracts a value for a YAML key from raw content.
// Supports "key: value" and "key: 'value'" formats.
func extractYAMLValue(raw string, keys ...string) string {
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		for _, key := range keys {
			prefix := key + ":"
			if strings.HasPrefix(trimmed, prefix) {
				value := strings.TrimSpace(trimmed[len(prefix):])
				// Strip quotes
				value = strings.Trim(value, `"'`)
				return value
			}
		}
	}
	return ""
}

// isYAMLKeyLine returns true if the line is a YAML key assignment for the given key.
func isYAMLKeyLine(line, key string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, key+":")
}
