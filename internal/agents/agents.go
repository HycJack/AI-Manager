// Package agents defines agent kinds, install targets, and skill discovery
// for cross-agent project scanning.
package agents

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// AgentKind identifies a specific agent runtime.
type AgentKind string

const (
	AgentUniversal   AgentKind = "universal"
	AgentClaudeCode  AgentKind = "claude-code"
	AgentCodex       AgentKind = "codex"
	AgentCursor      AgentKind = "cursor"
	AgentOpenCode    AgentKind = "opencode"
	AgentPi          AgentKind = "pi"
	AgentGrok        AgentKind = "grok"
	AgentAntigravity AgentKind = "antigravity"
	AgentDroid       AgentKind = "droid"
	AgentCopilot     AgentKind = "copilot"
	AgentHermes      AgentKind = "hermes"
	AgentCcSwitch    AgentKind = "cc-switch"

	// Agents added for provider/session/memory integration (T12+).
	// Source: docs/agents/memory-session-paths.md (verified 2026-09-11).
	AgentGemini    AgentKind = "gemini"
	AgentContinue  AgentKind = "continue"
	AgentCline     AgentKind = "cline"
	AgentWindsurf  AgentKind = "windsurf"
	AgentOpenClaw  AgentKind = "openclaw"
	AgentKiro      AgentKind = "kiro"
	AgentAmp       AgentKind = "amp"
	AgentGoose     AgentKind = "goose"
	AgentRooCode   AgentKind = "roo-code"
)

// AllAgentKinds returns all defined agent kinds in display order.
func AllAgentKinds() []AgentKind {
	return []AgentKind{
		AgentUniversal, AgentClaudeCode, AgentCodex, AgentCursor,
		AgentOpenCode, AgentPi, AgentGrok, AgentAntigravity,
		AgentDroid, AgentCopilot, AgentHermes, AgentCcSwitch,
		AgentGemini, AgentContinue, AgentCline, AgentWindsurf,
		AgentOpenClaw, AgentKiro, AgentAmp, AgentGoose, AgentRooCode,
	}
}

// InstallTarget identifies a skill installation directory within a project.
type InstallTarget string

const (
	TargetUniversal   InstallTarget = "universal"
	TargetClaudeCode  InstallTarget = "claude-code"
	TargetCodex       InstallTarget = "codex"
	TargetCursor      InstallTarget = "cursor"
	TargetOpenCode    InstallTarget = "opencode"
	TargetPi          InstallTarget = "pi"
	TargetGrok        InstallTarget = "grok"
	TargetAntigravity InstallTarget = "antigravity"
	TargetDroid       InstallTarget = "droid"
	TargetCopilot     InstallTarget = "copilot"
	TargetHermes      InstallTarget = "hermes"
	TargetCcSwitch    InstallTarget = "cc-switch"
)

// AllTargets returns all defined install targets.
func AllTargets() []InstallTarget {
	return []InstallTarget{
		TargetUniversal, TargetClaudeCode, TargetCodex, TargetCursor,
		TargetOpenCode, TargetPi, TargetGrok, TargetAntigravity,
		TargetDroid, TargetCopilot, TargetHermes, TargetCcSwitch,
	}
}

// AgentDir returns the directory name for an agent.
// Universal uses ".agents"; Claude Code uses ".claude" (not ".claude-code");
// all others use ".<agent-name>".
func AgentDir(agent AgentKind) string {
	switch agent {
	case AgentUniversal:
		return ".agents"
	case AgentClaudeCode:
		return ".claude"
	default:
		return "." + string(agent)
	}
}

// TargetDir returns the full directory path for a given target within a project.
// Universal target uses .agents/skills; Claude Code uses .claude/skills;
// agent-specific targets use .<agent>/skills.
func TargetDir(projectPath string, target InstallTarget) string {
	switch target {
	case TargetUniversal:
		return filepath.Join(projectPath, ".agents", "skills")
	case TargetClaudeCode:
		return filepath.Join(projectPath, ".claude", "skills")
	default:
		dir := "." + string(target)
		return filepath.Join(projectPath, dir, "skills")
	}
}

// TargetAgent maps an InstallTarget to its corresponding AgentKind.
func TargetAgent(target InstallTarget) AgentKind {
	return AgentKind(target)
}

// UserTargetDir returns the user-level (home-directory) skill directory for a
// given target. Some tools use XDG-style paths (~/.config/<tool>/skills)
// instead of the dot-directory convention (~/.<tool>/skills). This function
// handles those exceptions.
func UserTargetDir(target InstallTarget) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	switch target {
	case TargetUniversal:
		return filepath.Join(home, ".agents", "skills")
	case TargetClaudeCode:
		return filepath.Join(home, ".claude", "skills")
	case TargetOpenCode:
		return filepath.Join(home, ".config", "opencode", "skills")
	case TargetCursor:
		return filepath.Join(home, ".cursor", "skills")
	default:
		return filepath.Join(home, "."+string(target), "skills")
	}
}

// AgentDirs scans a project directory for all agent-specific directories that exist.
// Returns the list of InstallTargets found (both universal and agent-specific).
func AgentDirs(projectPath string) []InstallTarget {
	found := make([]InstallTarget, 0, len(AllTargets()))

	for _, target := range AllTargets() {
		dir := TargetDir(projectPath, target)
		if _, err := os.Stat(dir); err == nil {
			found = append(found, target)
		}
	}

	return found
}

// DiscoveredSkill represents a skill found during project scanning.
type DiscoveredSkill struct {
	Name    string        `json:"name"`
	Agent   AgentKind     `json:"agent"`
	Target  InstallTarget `json:"target"`
	Path    string        `json:"path"`
	Version string        `json:"version,omitempty"`
}

// DiscoverSkills scans all agent directories in a project for skill folders.
// A skill folder is identified by containing SKILL.md or README.md.
func DiscoverSkills(projectPath string) []DiscoveredSkill {
	var skills []DiscoveredSkill

	targets := AgentDirs(projectPath)
	for _, target := range targets {
		dir := TargetDir(projectPath, target)
		agent := TargetAgent(target)

		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			skillPath := filepath.Join(dir, entry.Name())

			hasManifest := false
			for _, name := range []string{"SKILL.md", "README.md"} {
				if _, err := os.Stat(filepath.Join(skillPath, name)); err == nil {
					hasManifest = true
					break
				}
			}
			if !hasManifest {
				continue
			}

			ds := DiscoveredSkill{
				Name:   entry.Name(),
				Agent:  agent,
				Target: target,
				Path:   skillPath,
			}

			metaPath := filepath.Join(skillPath, "metadata.json")
			if data, err := os.ReadFile(metaPath); err == nil {
				var meta struct {
					Version string `json:"version"`
				}
				if err := json.Unmarshal(data, &meta); err == nil && meta.Version != "" {
					ds.Version = meta.Version
				}
			}

			skills = append(skills, ds)
		}
	}

	return skills
}
