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
	AgentClaude      AgentKind = "claude"
	AgentCodex       AgentKind = "codex"
	AgentCursor      AgentKind = "cursor"
	AgentCline       AgentKind = "cline"
	AgentContinue    AgentKind = "continue"
	AgentAider       AgentKind = "aider"
	AgentAntigravity AgentKind = "antigravity"
	AgentTrae        AgentKind = "trae"
	AgentWindsurf    AgentKind = "windsurf"
	AgentGeneric     AgentKind = "generic"
)

// AllAgentKinds returns all defined agent kinds.
func AllAgentKinds() []AgentKind {
	return []AgentKind{
		AgentClaude, AgentCodex, AgentCursor, AgentCline,
		AgentContinue, AgentAider, AgentAntigravity, AgentTrae,
		AgentWindsurf, AgentGeneric,
	}
}

// InstallTarget identifies a skill installation directory within a project.
type InstallTarget string

const (
	TargetShared       InstallTarget = "shared"
	TargetClaude       InstallTarget = "claude"
	TargetCodex        InstallTarget = "codex"
	TargetCursor       InstallTarget = "cursor"
	TargetCline        InstallTarget = "cline"
	TargetContinue     InstallTarget = "continue"
	TargetAider        InstallTarget = "aider"
	TargetAntigravity  InstallTarget = "antigravity"
	TargetTrae         InstallTarget = "trae"
	TargetWindsurf     InstallTarget = "windsurf"
)

// AllTargets returns all defined install targets.
func AllTargets() []InstallTarget {
	return []InstallTarget{
		TargetShared, TargetClaude, TargetCodex, TargetCursor,
		TargetCline, TargetContinue, TargetAider, TargetAntigravity,
		TargetTrae, TargetWindsurf,
	}
}

// AgentDir returns the directory name for an agent (e.g. ".claude" for "claude").
func AgentDir(agent AgentKind) string {
	if agent == "" || agent == AgentGeneric {
		return ".agents"
	}
	return "." + string(agent)
}

// TargetDir returns the full directory path for a given target within a project.
// Shared target uses .agents/skills; agent-specific targets use .<agent>/skills.
func TargetDir(projectPath string, target InstallTarget) string {
	if target == TargetShared {
		return filepath.Join(projectPath, ".agents", "skills")
	}
	dir := "." + string(target)
	return filepath.Join(projectPath, dir, "skills")
}

// TargetAgent maps an InstallTarget to its corresponding AgentKind.
func TargetAgent(target InstallTarget) AgentKind {
	if target == TargetShared {
		return AgentGeneric
	}
	return AgentKind(target)
}

// AgentDirs scans a project directory for all agent-specific directories that exist.
// Returns the list of InstallTargets found (both shared and agent-specific).
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

			// Check for SKILL.md or README.md
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

			// Try to read version from metadata.json
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
