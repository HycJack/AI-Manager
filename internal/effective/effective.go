// Package effective computes the set of skills an agent would actually see
// when scanning a project, including token budget estimation.
package effective

import (
	"os"
	"path/filepath"

	"ai-manager/internal/agents"
)

// Token thresholds for warning and danger levels.
const (
	TokenWarning = 1000
	TokenDanger  = 5000
)

// EffectiveSkill represents a skill as an agent would see it.
type EffectiveSkill struct {
	Name    string `json:"name"`
	Source  string `json:"source"` // "managed", "unmanaged", "builtin"
	Managed bool   `json:"managed"`
	Tokens  int    `json:"tokens"`
	Warning bool   `json:"warning"`
	Danger  bool   `json:"danger"`
}

// EstimateTokenCount estimates the token count for a skill based on its files.
// Rough estimate: character count / 4.
func EstimateTokenCount(skillDir string) int {
	total := 0

	// Count chars in SKILL.md or README.md
	for _, name := range []string{"SKILL.md", "README.md"} {
		data, err := os.ReadFile(filepath.Join(skillDir, name))
		if err == nil {
			total += len(data)
		}
	}

	// Count chars in metadata.json
	data, err := os.ReadFile(filepath.Join(skillDir, "metadata.json"))
	if err == nil {
		total += len(data)
	}

	// Count chars in config files (config.json, package.json, etc.)
	configFiles := []string{"config.json", "package.json", "settings.json"}
	for _, name := range configFiles {
		data, err := os.ReadFile(filepath.Join(skillDir, name))
		if err == nil {
			total += len(data)
		}
	}

	// Count script files (.sh, .js, .ts, .py, .go)
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

// GetEffectiveSkills returns all skills an agent would see in a project.
// It scans:
//  1. The agent-specific directory (e.g. .claude/skills) — "unmanaged"
//  2. The shared directory (.agents/skills) — "managed" if symlink, "unmanaged" otherwise
//  3. User-level directories (e.g. ~/.claude/skills) — "managed" if symlink
//
// Managed skills are those that are symlinks back to the library.
func GetEffectiveSkills(projectPath string, agent agents.AgentKind) []EffectiveSkill {
	var skills []EffectiveSkill

	// Determine which directories to scan
	dirsToScan := []struct {
		dir     string
		managed bool
	}{
		{agents.TargetDir(projectPath, agents.InstallTarget(agent)), false},
		{agents.TargetDir(projectPath, agents.TargetShared), false},
	}

	// Also scan user-level directories
	home, _ := os.UserHomeDir()
	if home != "" {
		userTarget := agents.TargetDir(home, agents.InstallTarget(agent))
		dirsToScan = append(dirsToScan, struct {
			dir     string
			managed bool
		}{userTarget, false})
	}

	seen := make(map[string]bool)

	for _, ds := range dirsToScan {
		dir := ds.dir
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			skillDir := filepath.Join(dir, entry.Name())

			// Check if it's a symlink or directory using Lstat
			info, err := os.Lstat(skillDir)
			if err != nil {
				continue
			}
			isSymlink := info.Mode()&os.ModeSymlink != 0
			if !info.IsDir() && !isSymlink {
				continue
			}

			// Skip if already seen (prefer managed over unmanaged)
			if seen[entry.Name()] {
				continue
			}

			// Only include if it has a manifest
			hasManifest := false
			for _, name := range []string{"SKILL.md", "README.md"} {
				if _, err := os.Stat(filepath.Join(skillDir, name)); err == nil {
					hasManifest = true
					break
				}
			}
			if !hasManifest {
				continue
			}

			tokens := EstimateTokenCount(skillDir)

			// Resolve symlink target to check if it's from the library
			source := "unmanaged"
			managed := false
			if isSymlink {
				source = "managed"
				managed = true
			}

			es := EffectiveSkill{
				Name:    entry.Name(),
				Source:  source,
				Managed: managed,
				Tokens:  tokens,
				Warning: tokens > TokenWarning,
				Danger:  tokens > TokenDanger,
			}

			skills = append(skills, es)
			seen[entry.Name()] = true
		}
	}

	return skills
}
