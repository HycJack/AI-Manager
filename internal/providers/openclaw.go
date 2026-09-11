package providers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ai-manager/internal/agents"
)

// openClawAdapter reads <state_dir>/config.toml (i.e. ~/.openclaw/config.toml).
type openClawAdapter struct{}

func (a *openClawAdapter) Agent() agents.AgentKind { return agents.AgentOpenClaw }

func (a *openClawAdapter) Read() (*ProviderConfig, error) {
	path := expandHome("~/.openclaw/config.toml")
	cfg := &ProviderConfig{
		Agent:  agents.AgentOpenClaw,
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

func (a *openClawAdapter) Write(cfg *ProviderConfig) error {
	if cfg == nil || !cfg.Exists {
		return nil
	}
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

func init() { Register(&openClawAdapter{}) }
