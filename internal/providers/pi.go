// Pi provider adapter: reads ~/.pi/agent/settings.json + models.json
package providers

import (
	"encoding/json"
	"os"
	"path/filepath"

	"ai-manager/internal/agents"
)

func init() {
	Register(&piAdapter{})
}

type piAdapter struct{}

func (a *piAdapter) Agent() agents.AgentKind { return agents.AgentPi }

func (a *piAdapter) Read() (*ProviderConfig, error) {
	configPath := expandHome("~/.pi/agent/settings.json")

	data, exists, errs := readJSONConfig(configPath)
	if !exists {
		return &ProviderConfig{
			Agent:  agents.AgentPi,
			Path:   configPath,
			Format: "json",
			Exists: false,
		}, nil
	}
	if len(errs) > 0 {
		return &ProviderConfig{
			Agent:  agents.AgentPi,
			Path:   configPath,
			Format: "json",
			Exists: true,
			Errors: errs,
		}, nil
	}

	cfg := &ProviderConfig{
		Agent:  agents.AgentPi,
		Path:   configPath,
		Format: "json",
		Exists: true,
		Extra:  data,
		Provider: "pi",
	}

	// Extract default provider/model from settings.json
	defaultProvider := ""
	if dm, ok := data["defaultModel"].(string); ok {
		cfg.Model = dm
	}
	if dp, ok := data["defaultProvider"].(string); ok {
		defaultProvider = dp
		cfg.Provider = dp
	}

	// Try to read models.json for API key and base URL of all providers
	modelsPath := filepath.Join(filepath.Dir(configPath), "models.json")
	modelsData, merr := os.ReadFile(modelsPath)
	if merr == nil {
		var models struct {
			Providers map[string]struct {
				APIKey  string `json:"apiKey"`
				BaseURL string `json:"baseUrl"`
				Models  []struct {
					ID string `json:"id"`
				} `json:"models"`
			} `json:"providers"`
		}
		if err := json.Unmarshal(modelsData, &models); err == nil {
			// Populate all providers
			for name, p := range models.Providers {
				model := ""
				if len(p.Models) > 0 {
					model = p.Models[0].ID
				}
				cfg.Providers = append(cfg.Providers, ProviderEntry{
					Name:   name,
					Model:  model,
					BaseURL: p.BaseURL,
					APIKey: MaskAPIKey(p.APIKey),
					Active: name == defaultProvider,
				})
			}
			// Set active provider details
			if p, ok := models.Providers[defaultProvider]; ok {
				cfg.APIKey = MaskAPIKey(p.APIKey)
				cfg.BaseURL = p.BaseURL
				if len(p.Models) > 0 {
					cfg.Model = p.Models[0].ID
				}
			}
		}
	}

	return cfg, nil
}

func (a *piAdapter) Write(cfg *ProviderConfig) error {
	return nil // Pi config is not editable yet
}
