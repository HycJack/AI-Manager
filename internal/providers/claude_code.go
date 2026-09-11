package providers

import (
	"os"

	"ai-manager/internal/agents"
)

// claudeCodeAdapter reads ~/.claude/settings.json (and ~/.claude.json as fallback).
type claudeCodeAdapter struct{}

func (a *claudeCodeAdapter) Agent() agents.AgentKind { return agents.AgentClaudeCode }

func (a *claudeCodeAdapter) Read() (*ProviderConfig, error) {
	// Try settings.json first, then .claude.json
	for _, path := range []string{
		expandHome("~/.claude/settings.json"),
		expandHome("~/.claude.json"),
	} {
		data, exists, errs := readJSONConfig(path)
		if !exists {
			continue
		}
		if len(errs) > 0 {
			return &ProviderConfig{
				Agent:    agents.AgentClaudeCode,
				Path:     path,
				Format:   "json",
				Exists:   true,
				Errors:   errs,
			}, nil
		}

		cfg := &ProviderConfig{
			Agent:    agents.AgentClaudeCode,
			Path:     path,
			Format:   "json",
			Exists:   true,
			Extra:    data,
			Provider: inferProviderFromJSON(data),
			Model:    extractString(data, "model", "defaultModel", "preferred_model"),
			BaseURL:  extractString(data, "baseUrl", "base_url", "apiBase", "baseURL"),
			APIKey:   MaskAPIKey(extractString(data, "apiKey", "api_key", "anthropic_api_key")),
		}
		return cfg, nil
	}

	return &ProviderConfig{
		Agent:  agents.AgentClaudeCode,
		Format: "json",
		Exists: false,
	}, nil
}

func (a *claudeCodeAdapter) Write(cfg *ProviderConfig) error {
	if cfg == nil || !cfg.Exists {
		return nil
	}
	data := make(map[string]any)
	if cfg.Extra != nil {
		for k, v := range cfg.Extra {
			data[k] = v
		}
	}
	if cfg.Model != "" {
		data["model"] = cfg.Model
	}
	if cfg.BaseURL != "" {
		data["baseUrl"] = cfg.BaseURL
	}
	if cfg.APIKey != "" && !IsMaskedAPIKey(cfg.APIKey) {
		data["apiKey"] = cfg.APIKey
	}
	return writeJSONConfig(cfg.Path, data)
}

func init() { Register(&claudeCodeAdapter{}) }

// expandHome replaces a leading ~ with the user's home directory.
func expandHome(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		home, _ := os.UserHomeDir()
		return home + path[1:]
	}
	return path
}
