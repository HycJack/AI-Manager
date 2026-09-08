// Package effective computes the set of skills an agent would actually see
// when scanning a project, including token budget estimation.
//
// The approach mirrors Kitter: scan ALL agent directories in the project,
// resolve symlinks to their real source (canonical path), deduplicate by
// canonical path, and group the results across agents.
package effective

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"ai-manager/internal/agents"
)

// Token thresholds for warning and danger levels.
const (
	TokenWarning = 2000
	TokenDanger  = 5000
	SkillWarning = 20
	SkillDanger  = 50
)

// EffectiveSkill represents one deduplicated skill across all agents.
type EffectiveSkill struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Managed       bool     `json:"managed"`       // symlink back to library
	Builtin       bool     `json:"builtin"`
	Agents        []string `json:"agents"`        // agent kinds that see this skill
	Locations     []string `json:"locations"`     // symlink paths (agent-specific dirs)
	CanonicalPath string   `json:"canonicalPath"` // resolved real source path
	Tokens        int      `json:"tokens"`
	Warning       bool     `json:"warning"`
	Danger        bool     `json:"danger"`
	SkillCount    int      `json:"skillCount"` // number of agents that see this skill
}

// GetEffectiveSkills returns all skills visible in a project, deduplicated
// across agents by canonical (resolved symlink) path.
//
// Scans:
//  1. All agent-specific directories (.claude/skills, .codex/skills, ...)
//  2. The shared directory (.agents/skills)
//  3. User-level directories (~/.claude/skills, etc.)
//
// For each skill found:
//  - Resolves symlinks to the real source path
//  - Deduplicates by canonical path (same source = one skill)
//  - Records which agents see it
//  - Estimates token count
func GetEffectiveSkills(projectPath string) []EffectiveSkill {
	// Discover all agent directories that exist
	targets := agents.AgentDirs(projectPath)

	// Also add shared directory
	targets = append(targets, agents.TargetShared)

	// Scan each agent directory
	type discoveredSkill struct {
		name      string
		path      string
		canonical string
		agent     string
		location  string // the symlink path (agent dir)
	}

	var allDiscovered []discoveredSkill

	for _, target := range targets {
		dir := agents.TargetDir(projectPath, target)
		agentName := string(agents.TargetAgent(target))

		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			skillPath := filepath.Join(dir, entry.Name())

			// Check if it's a symlink or directory
			info, err := os.Lstat(skillPath)
			if err != nil {
				continue
			}
			isSymlink := info.Mode()&os.ModeSymlink != 0
			if !info.IsDir() && !isSymlink {
				continue
			}

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

			// Resolve to canonical path (resolves symlinks)
			canonical, err := filepath.EvalSymlinks(skillPath)
			if err != nil {
				canonical = skillPath
			}

			allDiscovered = append(allDiscovered, discoveredSkill{
				name:      entry.Name(),
				path:      skillPath,
				canonical: canonical,
				agent:     agentName,
				location:  dir,
			})
		}
	}

	// Also scan user-level directories
	home, _ := os.UserHomeDir()
	if home != "" {
		for _, target := range targets {
			userDir := agents.TargetDir(home, target)
			if userDir == "" {
				continue
			}
			agentName := string(agents.TargetAgent(target))

			entries, err := os.ReadDir(userDir)
			if err != nil {
				continue
			}

			for _, entry := range entries {
				skillPath := filepath.Join(userDir, entry.Name())

				info, err := os.Lstat(skillPath)
				if err != nil {
					continue
				}
				isSymlink := info.Mode()&os.ModeSymlink != 0
				if !info.IsDir() && !isSymlink {
					continue
				}

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

				canonical, err := filepath.EvalSymlinks(skillPath)
				if err != nil {
					canonical = skillPath
				}

				allDiscovered = append(allDiscovered, discoveredSkill{
					name:      entry.Name(),
					path:      skillPath,
					canonical: canonical,
					agent:     agentName,
					location:  userDir,
				})
			}
		}
	}

	// Group by canonical path
	type skillGroup struct {
		name        string
		canonical   string
		agents      []string
		locations   []string
		managed     bool
		builtin     bool
		description string
	}

	groups := make(map[string]*skillGroup)
	for _, ds := range allDiscovered {
		g, ok := groups[ds.canonical]
		if !ok {
			g = &skillGroup{
				name:      ds.name,
				canonical: ds.canonical,
			}
			groups[ds.canonical] = g
		}

		// Check if it's a symlink (managed)
		info, err := os.Lstat(ds.path)
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			g.managed = true
		}

		// Check if builtin (in the library source)
		if isBuiltin(ds.canonical) {
			g.builtin = true
		}

		// Track agents
		if !contains(g.agents, ds.agent) {
			g.agents = append(g.agents, ds.agent)
		}

		// Track locations
		if !contains(g.locations, ds.location) {
			g.locations = append(g.locations, ds.location)
		}

		// Read description from metadata
		if g.description == "" {
			g.description = readDescription(ds.canonical)
		}
	}

	// Convert groups to EffectiveSkill
	result := make([]EffectiveSkill, 0, len(groups))
	for _, g := range groups {
		tokens := EstimateTokenCount(g.canonical)
		skillCount := len(g.agents)

		result = append(result, EffectiveSkill{
			Name:          g.name,
			Description:   g.description,
			Managed:       g.managed,
			Builtin:       g.builtin,
			Agents:        g.agents,
			Locations:     g.locations,
			CanonicalPath: g.canonical,
			Tokens:        tokens,
			Warning:       tokens > TokenWarning || skillCount > SkillWarning,
			Danger:        tokens > TokenDanger || skillCount > SkillDanger,
			SkillCount:    skillCount,
		})
	}

	return result
}

// EstimateTokenCount estimates the token count for a skill based on its files.
// Rough estimate: character count / 4.
func EstimateTokenCount(skillDir string) int {
	total := 0

	for _, name := range []string{"SKILL.md", "README.md"} {
		data, err := os.ReadFile(filepath.Join(skillDir, name))
		if err == nil {
			total += len(data)
		}
	}

	data, err := os.ReadFile(filepath.Join(skillDir, "metadata.json"))
	if err == nil {
		total += len(data)
	}

	configFiles := []string{"config.json", "package.json", "settings.json"}
	for _, name := range configFiles {
		data, err := os.ReadFile(filepath.Join(skillDir, name))
		if err == nil {
			total += len(data)
		}
	}

	entries, err := os.ReadDir(skillDir)
	if err != nil {
		return total / 4
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		if isScriptExt(ext) {
			data, err := os.ReadFile(filepath.Join(skillDir, entry.Name()))
			if err == nil {
				total += len(data)
			}
		}
	}

	return total / 4
}

func isScriptExt(ext string) bool {
	switch ext {
	case ".sh", ".js", ".ts", ".py", ".go", ".rb", ".php", ".pl",
		".bash", ".zsh", ".fish", ".ps1", ".bat", ".cmd":
		return true
	}
	return false
}

func isBuiltin(skillDir string) bool {
	dataDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	libraryDir := filepath.Join(dataDir, ".ai-manager", "library")
	if strings.HasPrefix(skillDir, libraryDir) {
		return true
	}
	libraryDir2, _ := filepath.Abs(filepath.Join(dataDir, "Library", "Application Support", "AIManager", "library"))
	if strings.HasPrefix(skillDir, libraryDir2) {
		return true
	}
	return false
}

func readDescription(skillDir string) string {
	// Try to read description from metadata.json
	metaPath := filepath.Join(skillDir, "metadata.json")
	data, err := os.ReadFile(metaPath)
	if err == nil {
		var meta struct {
			Description string `json:"description"`
		}
		if json.Unmarshal(data, &meta) == nil {
			return meta.Description
		}
	}

	// Fallback: read first paragraph of SKILL.md or README.md
	for _, name := range []string{"SKILL.md", "README.md"} {
		data, err := os.ReadFile(filepath.Join(skillDir, name))
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "#") && trimmed != "---" {
				return trimmed
			}
		}
	}
	return ""
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
