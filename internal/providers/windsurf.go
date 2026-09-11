package providers

import (
	"ai-manager/internal/agents"
)

// windsurfAdapter reads ~/.codeium/windsurf/settings.json.
type windsurfAdapter struct{}

func (a *windsurfAdapter) Agent() agents.AgentKind { return agents.AgentWindsurf }

func (a *windsurfAdapter) Read() (*ProviderConfig, error) {
	path := expandHome("~/.codeium/windsurf/settings.json")
	data, exists, errs := readJSONConfig(path)

	cfg := &ProviderConfig{
		Agent:  agents.AgentWindsurf,
		Path:   path,
		Format: "json",
		Exists: exists,
		Errors: errs,
	}
	if !exists || len(errs) > 0 {
		return cfg, nil
	}

	cfg.Extra = data
	cfg.Provider = inferProviderFromJSON(data)
	cfg.Model = extractString(data, "model", "defaultModel")
	cfg.BaseURL = extractString(data, "baseUrl", "base_url", "apiBase")
	cfg.APIKey = MaskAPIKey(extractString(data, "apiKey", "api_key"))
	return cfg, nil
}

func (a *windsurfAdapter) Write(cfg *ProviderConfig) error {
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

func init() { Register(&windsurfAdapter{}) }
