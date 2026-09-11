// Package skill defines the core data model for agent skills.
package skill

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// OriginType identifies where a skill was sourced from.
type OriginType string

const (
	OriginLocal     OriginType = "local"
	OriginGitHub    OriginType = "github"
	OriginSkillsSh  OriginType = "skills.sh"
	OriginClaude    OriginType = "claude"
	OriginBuiltin   OriginType = "builtin"
	OriginUnknown   OriginType = "unknown"
)

// Origin describes where a skill came from.
type Origin struct {
	Type    OriginType `json:"type"`
	Path    string     `json:"path,omitempty"`
	Repo    string     `json:"repo,omitempty"`
	Subdir  string     `json:"subdir,omitempty"`
}

// Source describes a skill's source metadata.
type Source struct {
	Origin  Origin     `json:"origin"`
	License string     `json:"license,omitempty"`
	Authors []string   `json:"authors,omitempty"`
	Keywords []string  `json:"keywords,omitempty"`
}

// Group is a named collection of skills for batch installation.
type Group struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Skills []string `json:"skills"`
}

// Record is the full metadata for a single skill in the library.
type Record struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Slug      string   `json:"slug"`
	Version   string   `json:"version"`
	Description string `json:"description,omitempty"`
	Authors   []string `json:"authors,omitempty"`
	License   string   `json:"license,omitempty"`
	Keywords  []string `json:"keywords,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	Group     string   `json:"group,omitempty"`
	Origin    Origin   `json:"origin"`
	Installed bool     `json:"installed"`
	UpdatedAt time.Time `json:"updatedAt"`
	InstalledAt time.Time `json:"installedAt,omitempty"`
}

// Summary is a lightweight representation of a skill for list views.
type Summary struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Version   string `json:"version"`
	Description string `json:"description,omitempty"`
	Installed bool   `json:"installed"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// NewSummary converts a full Record to a Summary.
func (r *Record) NewSummary() Summary {
	return Summary{
		ID:        r.ID,
		Name:      r.Name,
		Slug:      r.Slug,
		Version:   r.Version,
		Description: r.Description,
		Installed: r.Installed,
		UpdatedAt: r.UpdatedAt,
	}
}

// LoadReadme reads and returns the SKILL.md or README.md content from the skill directory.
func LoadReadme(skillDir string) (string, error) {
	for _, name := range []string{"SKILL.md", "README.md"} {
		path := filepath.Join(skillDir, name)
		data, err := os.ReadFile(path)
		if err == nil {
			return string(data), nil
		}
	}
	return "", os.ErrNotExist
}

// LoadMetadata reads and parses a metadata.json file from the skill directory.
func LoadMetadata(skillDir string) (*Record, error) {
	path := filepath.Join(skillDir, "metadata.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rec Record
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

// parseSkillFrontmatterVersion reads SKILL.md from the given directory and
// returns the version declared in its YAML frontmatter (the block between
// the first two --- lines). Returns an empty string if no frontmatter or no
// version field is found.
func parseSkillFrontmatterVersion(skillDir string) string {
	skillPath := filepath.Join(skillDir, "SKILL.md")
	data, err := os.ReadFile(skillPath)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(data), "\n")

	// Frontmatter must start on the first line.
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return ""
	}

	// Find the closing ---.
	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			endIdx = i
			break
		}
	}
	if endIdx == -1 {
		return ""
	}

	// Parse key: value pairs between the markers.
	for i := 1; i < endIdx; i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)
		if key == "version" && value != "" {
			return value
		}
	}

	return ""
}

// ParseSkillFrontmatterDescription reads SKILL.md and returns the description
// from its YAML frontmatter. Returns empty string if no frontmatter or no description.
func ParseSkillFrontmatterDescription(skillDir string) string {
	skillPath := filepath.Join(skillDir, "SKILL.md")
	data, err := os.ReadFile(skillPath)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(data), "\n")

	// Frontmatter must start on the first line.
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return ""
	}

	// Find the closing ---.
	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			endIdx = i
			break
		}
	}
	if endIdx == -1 {
		return ""
	}

	// Parse key: value pairs between the markers.
	for i := 1; i < endIdx; i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)
		if key == "description" && value != "" {
			return value
		}
	}

	return ""
}
