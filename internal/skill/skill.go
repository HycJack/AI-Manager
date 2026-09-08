// Package skill defines the core data model for agent skills.
package skill

import (
	"encoding/json"
	"os"
	"path/filepath"
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
