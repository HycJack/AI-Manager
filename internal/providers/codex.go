package providers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ai-manager/internal/agents"
)

// codexAdapter reads ~/.codex/config.toml.
// TOML is read as raw content; structured fields are extracted via string matching.
type codexAdapter struct{}

func (a *codexAdapter) Agent() agents.AgentKind { return agents.AgentCodex }

func (a *codexAdapter) Read() (*ProviderConfig, error) {
	path := expandHome("~/.codex/config.toml")
	cfg := &ProviderConfig{
		Agent:  agents.AgentCodex,
		Path:   path,
		Format: "toml",
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
	cfg.Model = extractTOMLValue(cfg.Raw, "model", "default_model")
	cfg.BaseURL = extractTOMLValue(cfg.Raw, "base_url", "baseUrl", "wire_api")
	cfg.APIKey = MaskAPIKey(extractTOMLValue(cfg.Raw, "api_key", "apiKey", "OPENAI_API_KEY"))
	return cfg, nil
}

func (a *codexAdapter) Write(cfg *ProviderConfig) error {
	if cfg == nil || !cfg.Exists {
		return nil
	}
	// For TOML, preserve raw content and modify/add key fields.
	var lines []string
	for _, line := range strings.Split(cfg.Raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if isTOMLKeyLine(trimmed, "model") || isTOMLKeyLine(trimmed, "api_key") || isTOMLKeyLine(trimmed, "base_url") {
			continue
		}
		lines = append(lines, line)
	}
	if cfg.Model != "" {
		lines = append(lines, fmt.Sprintf("model = %q", cfg.Model))
	}
	if cfg.BaseURL != "" {
		lines = append(lines, fmt.Sprintf("base_url = %q", cfg.BaseURL))
	}
	if cfg.APIKey != "" && !IsMaskedAPIKey(cfg.APIKey) {
		lines = append(lines, fmt.Sprintf("api_key = %q", cfg.APIKey))
	}
	if err := os.MkdirAll(filepath.Dir(cfg.Path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(cfg.Path, []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}

func init() { Register(&codexAdapter{}) }

// extractTOMLValue extracts a value for a TOML key from raw content.
// Supports "key = value" format.
func extractTOMLValue(raw string, keys ...string) string {
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		// Skip comments
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		for _, key := range keys {
			prefix := key + " ="
			if strings.HasPrefix(trimmed, prefix) {
				value := strings.TrimSpace(trimmed[len(prefix):])
				value = strings.Trim(value, `"'`)
				return value
			}
		}
	}
	return ""
}

// isTOMLKeyLine returns true if the line is a TOML key assignment.
func isTOMLKeyLine(line, key string) bool {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "#") {
		return false
	}
	return strings.HasPrefix(trimmed, key+" =")
}
