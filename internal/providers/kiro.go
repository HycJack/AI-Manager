package providers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ai-manager/internal/agents"
)

// kiroAdapter reads ~/.kiro/config.yaml.
type kiroAdapter struct{}

func (a *kiroAdapter) Agent() agents.AgentKind { return agents.AgentKiro }

func (a *kiroAdapter) Read() (*ProviderConfig, error) {
	path := expandHome("~/.kiro/config.yaml")
	cfg := &ProviderConfig{
		Agent:  agents.AgentKiro,
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

func (a *kiroAdapter) Write(cfg *ProviderConfig) error {
	if cfg == nil || !cfg.Exists {
		return nil
	}
	var lines []string
	for _, line := range strings.Split(cfg.Raw, "\n") {
		trimmed := strings.TrimSpace(line)
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

func init() { Register(&kiroAdapter{}) }
