package agents

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// AgentConfig defines a single agent entry in the configuration file.
type AgentConfig struct {
	Key   string `json:"key"`   // unique identifier (e.g. "claude-code")
	Label string `json:"label"` // display name (e.g. "Claude Code")
	Path  string `json:"path"`  // skill directory path template, {home} = user home
	// IconType is a string key for the frontend to pick the right icon.
	// Frontend maps this to a Lucide icon or inline SVG.
	IconType   string `json:"iconType"`
	ColorClass string `json:"colorClass"` // CSS class for active state
}

// AgentConfigFile is the root of the agents.json config file.
type AgentConfigFile struct {
	Agents []AgentConfig `json:"agents"`
}

// DefaultAgentConfig returns the default agent configuration.
// Order matters — it determines the display order in the UI.
//
// "universal" is listed first: it targets the shared .agents/skills
// directory that many agents read directly, so a skill linked there is
// available to all of them at once.
func DefaultAgentConfig() *AgentConfigFile {
	return &AgentConfigFile{
		Agents: []AgentConfig{
			{
				Key:        "universal",
				Label:      "Universal",
				Path:       "{home}/.agents/skills",
				IconType:   "universal",
				ColorClass: "bg-sky-500/10 ring-sky-500/20 text-sky-600 dark:text-sky-400 hover:bg-sky-500/20",
			},
			{
				Key:        "claude-code",
				Label:      "Claude Code",
				Path:       "{home}/.claude/skills",
				IconType:   "claude",
				ColorClass: "bg-orange-500/10 ring-orange-500/20 text-orange-600 dark:text-orange-400 hover:bg-orange-500/20",
			},
			{
				Key:        "codex",
				Label:      "Codex",
				Path:       "{home}/.codex/skills",
				IconType:   "codex",
				ColorClass: "bg-green-500/10 ring-green-500/20 text-green-600 dark:text-green-400 hover:bg-green-500/20",
			},
			{
				Key:        "pi",
				Label:      "Pi",
				Path:       "{home}/.pi/agent/skills",
				IconType:   "pi",
				ColorClass: "bg-fuchsia-500/10 ring-fuchsia-500/20 text-fuchsia-600 dark:text-fuchsia-400 hover:bg-fuchsia-500/20",
			},
			{
				Key:        "opencode",
				Label:      "OpenCode",
				Path:       "{home}/.config/opencode/skills",
				IconType:   "opencode",
				ColorClass: "bg-indigo-500/10 ring-indigo-500/20 text-indigo-600 dark:text-indigo-400 hover:bg-indigo-500/20",
			},
			{
				Key:        "hermes",
				Label:      "Hermes",
				Path:       "{home}/.hermes/skills",
				IconType:   "hermes",
				ColorClass: "bg-violet-500/10 ring-violet-500/20 text-violet-600 dark:text-violet-400 hover:bg-violet-500/20",
			},
		},
	}
}

// ResolvePath replaces {home} placeholder with the user's home directory.
func ResolvePath(template string) string {
	home, _ := os.UserHomeDir()
	return replaceAll(template, "{home}", home)
}

func replaceAll(s, old, new string) string {
	result := ""
	for {
		idx := indexOf(s, old)
		if idx < 0 {
			result += s
			break
		}
		result += s[:idx] + new
		s = s[idx+len(old):]
	}
	return result
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// LoadAgentConfig loads the agent config from ~/.aimanager/agents.json.
// If the file doesn't exist, creates it with the default config.
//
// Always returns a config that contains every built-in agent: missing ones are
// merged in on read (see mergeBuiltins), so adding a new default agent reaches
// already-installed copies without a manual migration.
func LoadAgentConfig() (*AgentConfigFile, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return mergeBuiltins(DefaultAgentConfig()), nil
	}
	configPath := filepath.Join(home, ".aimanager", "agents.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultAgentConfig()
			_ = SaveAgentConfig(cfg) // best-effort; return defaults regardless
			return mergeBuiltins(cfg), nil
		}
		return mergeBuiltins(DefaultAgentConfig()), nil
	}

	var cfg AgentConfigFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return mergeBuiltins(DefaultAgentConfig()), nil
	}
	if len(cfg.Agents) == 0 {
		return mergeBuiltins(DefaultAgentConfig()), nil
	}
	return mergeBuiltins(&cfg), nil
}

// mergeBuiltins appends every DefaultAgentConfig() entry that is absent from
// cfg, preserving the user's own order and any agents they added. It never
// removes or reorders existing entries and never writes the file: callers
// persist explicitly through SaveAgentConfig.
//
// Trade-off: a built-in agent deleted in Settings reappears on the next load,
// because the default list is treated as authoritative for built-ins. Agents
// the user added themselves are never touched.
func mergeBuiltins(cfg *AgentConfigFile) *AgentConfigFile {
	present := make(map[string]bool, len(cfg.Agents))
	for _, a := range cfg.Agents {
		present[a.Key] = true
	}
	for _, def := range DefaultAgentConfig().Agents {
		if present[def.Key] {
			continue
		}
		cfg.Agents = append(cfg.Agents, def)
		present[def.Key] = true
	}
	return cfg
}

// SaveAgentConfig writes the agent config to ~/.aimanager/agents.json.
func SaveAgentConfig(cfg *AgentConfigFile) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(home, ".aimanager", "agents.json")

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0o644)
}
