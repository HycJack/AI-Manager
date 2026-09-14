package agents

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// TestDefaultAgentConfigIncludesUniversal guards the shared .agents/skills
// target: it must exist in the defaults so the Skills page can toggle every
// library skill into the universal directory.
func TestDefaultAgentConfigIncludesUniversal(t *testing.T) {
	cfg := DefaultAgentConfig()
	if len(cfg.Agents) == 0 {
		t.Fatal("default config has no agents")
	}

	first := cfg.Agents[0]
	if first.Key != "universal" {
		t.Errorf("first default agent = %q, want universal", first.Key)
	}
	if first.Label != "Universal" {
		t.Errorf("universal label = %q, want Universal", first.Label)
	}
	if first.Path != "{home}/.agents/skills" {
		t.Errorf("universal path = %q, want {home}/.agents/skills", first.Path)
	}
	if first.IconType != "universal" {
		t.Errorf("universal iconType = %q, want universal", first.IconType)
	}
	if first.ColorClass == "" {
		t.Error("universal colorClass is empty")
	}
}

func TestDefaultAgentConfigKeysUniqueAndComplete(t *testing.T) {
	seen := map[string]bool{}
	for _, a := range DefaultAgentConfig().Agents {
		if a.Key == "" {
			t.Error("agent with empty key")
		}
		if seen[a.Key] {
			t.Errorf("duplicate agent key %q", a.Key)
		}
		seen[a.Key] = true
		if a.Path == "" {
			t.Errorf("agent %q has no path", a.Key)
		}
		if a.Label == "" {
			t.Errorf("agent %q has no label", a.Key)
		}
	}
}

func TestResolvePathUniversal(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got := ResolvePath("{home}/.agents/skills")
	want := filepath.Join(home, ".agents", "skills")
	if got != want {
		t.Errorf("ResolvePath = %q, want %q", got, want)
	}
}

// TestLoadAgentConfigSeedsDefaults verifies a missing config file falls back
// to the defaults and persists them. HOME is redirected so the real
// ~/.aimanager/agents.json is never touched.
func TestLoadAgentConfigSeedsDefaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := LoadAgentConfig()
	if err != nil {
		t.Fatalf("LoadAgentConfig: %v", err)
	}
	if cfg.Agents[0].Key != "universal" {
		t.Errorf("first loaded agent = %q, want universal", cfg.Agents[0].Key)
	}

	path := filepath.Join(home, ".aimanager", "agents.json")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("agents.json was not written: %v", err)
	}
}

func TestSaveAndLoadAgentConfigRoundTrip(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg := DefaultAgentConfig()
	cfg.Agents = append(cfg.Agents, AgentConfig{
		Key:      "my-agent",
		Label:    "My Agent",
		Path:     "{home}/.myagent/skills",
		IconType: "default",
	})
	if err := SaveAgentConfig(cfg); err != nil {
		t.Fatalf("SaveAgentConfig: %v", err)
	}

	loaded, err := LoadAgentConfig()
	if err != nil {
		t.Fatalf("LoadAgentConfig: %v", err)
	}
	if len(loaded.Agents) != len(cfg.Agents) {
		t.Fatalf("agent count = %d, want %d", len(loaded.Agents), len(cfg.Agents))
	}
	if loaded.Agents[0].Key != "universal" {
		t.Errorf("first agent = %q, want universal", loaded.Agents[0].Key)
	}
	last := loaded.Agents[len(loaded.Agents)-1]
	if last.Key != "my-agent" || last.Path != "{home}/.myagent/skills" {
		t.Errorf("roundtrip last agent = %+v", last)
	}
}

// writeAgentConfigFile writes cfg to ~/.aimanager/agents.json inside the
// redirected HOME.
func writeAgentConfigFile(t *testing.T, home string, cfg *AgentConfigFile) string {
	t.Helper()
	configPath := filepath.Join(home, ".aimanager", "agents.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return configPath
}

func agentKeys(cfg *AgentConfigFile) []string {
	keys := make([]string, 0, len(cfg.Agents))
	for _, a := range cfg.Agents {
		keys = append(keys, a.Key)
	}
	return keys
}

// TestLoadAgentConfigMergesMissingBuiltins covers the upgrade path: an install
// whose agents.json predates "universal" picks it up on read without losing its
// own order or its custom entries.
func TestLoadAgentConfigMergesMissingBuiltins(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeAgentConfigFile(t, home, &AgentConfigFile{Agents: []AgentConfig{
		{Key: "codex", Label: "Codex", Path: "{home}/.codex/skills", IconType: "codex"},
		{Key: "my-agent", Label: "My Agent", Path: "{home}/.myagent/skills", IconType: "default"},
	}})

	cfg, err := LoadAgentConfig()
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"codex", "my-agent", "universal", "claude-code", "pi", "opencode", "hermes"}
	if got := agentKeys(cfg); !reflect.DeepEqual(got, want) {
		t.Errorf("agent order = %v, want %v", got, want)
	}
}

// TestLoadAgentConfigIsReadOnly pins the invariant that merging built-ins is a
// read-time concern; the file on disk is only rewritten by SaveAgentConfig.
func TestLoadAgentConfigIsReadOnly(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	raw := []byte(`{"agents":[{"key":"codex","label":"Codex","path":"{home}/.codex/skills"}]}`)
	configPath := filepath.Join(home, ".aimanager", "agents.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadAgentConfig(); err != nil {
		t.Fatal(err)
	}

	after, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, after) {
		t.Error("LoadAgentConfig rewrote agents.json")
	}
}
