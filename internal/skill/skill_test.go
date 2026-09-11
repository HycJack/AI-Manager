package skill

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRecord_JSONRoundTrip(t *testing.T) {
	original := Record{
		ID:        "skill-001",
		Name:      "Test Skill",
		Slug:      "test-skill",
		Version:   "1.0.0",
		Authors:   []string{"Author A", "Author B"},
		License:   "MIT",
		Keywords:  []string{"test", "skill"},
		Tags:      []string{"frontend", "react"},
		Group:     "frontend-group",
		Origin: Origin{
			Type:   OriginGitHub,
			Repo:   "owner/repo",
			Subdir: "skills/test-skill",
		},
		Installed:   true,
		UpdatedAt:   time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		InstalledAt: time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(&original)
	if err != nil {
		t.Fatalf("Marshal() returned error: %v", err)
	}

	var decoded Record
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() returned error: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, original.ID)
	}
	if decoded.Name != original.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, original.Name)
	}
	if decoded.Version != original.Version {
		t.Errorf("Version = %q, want %q", decoded.Version, original.Version)
	}
	if len(decoded.Authors) != 2 {
		t.Errorf("Authors length = %d, want 2", len(decoded.Authors))
	}
	if decoded.License != original.License {
		t.Errorf("License = %q, want %q", decoded.License, original.License)
	}
	if decoded.Origin.Type != OriginGitHub {
		t.Errorf("Origin.Type = %q, want %q", decoded.Origin.Type, OriginGitHub)
	}
	if decoded.Origin.Repo != original.Origin.Repo {
		t.Errorf("Origin.Repo = %q, want %q", decoded.Origin.Repo, original.Origin.Repo)
	}
	if !decoded.Installed {
		t.Error("Installed = false, want true")
	}
}

func TestRecord_NewSummary(t *testing.T) {
	rec := Record{
		ID:        "skill-002",
		Name:      "Summary Test",
		Slug:      "summary-test",
		Version:   "2.0.0",
		Installed: true,
		UpdatedAt: time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC),
	}

	summary := rec.NewSummary()
	if summary.ID != rec.ID {
		t.Errorf("Summary.ID = %q, want %q", summary.ID, rec.ID)
	}
	if summary.Name != rec.Name {
		t.Errorf("Summary.Name = %q, want %q", summary.Name, rec.Name)
	}
	if summary.Installed != rec.Installed {
		t.Errorf("Summary.Installed = %v, want %v", summary.Installed, rec.Installed)
	}
}

func TestLoadReadme(t *testing.T) {
	tests := []struct {
		name    string
		files   map[string]string
		wantErr bool
		wantVal string
	}{
		{
			name: "SKILL.md exists",
			files: map[string]string{
				"SKILL.md": "# My Skill\n\nThis is a skill.",
			},
			wantVal: "# My Skill\n\nThis is a skill.",
		},
		{
			name: "README.md fallback",
			files: map[string]string{
				"README.md": "# My Skill\n\nREADME content.",
			},
			wantVal: "# My Skill\n\nREADME content.",
		},
		{
			name:    "no files",
			files:   map[string]string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()
			for name, content := range tt.files {
				if err := os.WriteFile(filepath.Join(tmp, name), []byte(content), 0o644); err != nil {
					t.Fatalf("WriteFile() returned error: %v", err)
				}
			}

			got, err := LoadReadme(tmp)
			if tt.wantErr {
				if err == nil {
					t.Error("LoadReadme() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("LoadReadme() returned error: %v", err)
			}
			if got != tt.wantVal {
				t.Errorf("LoadReadme() = %q, want %q", got, tt.wantVal)
			}
		})
	}
}

func TestLoadMetadata(t *testing.T) {
	tmp := t.TempDir()
	rec := Record{
		ID:      "skill-003",
		Name:    "Meta Test",
		Slug:    "meta-test",
		Version: "3.0.0",
	}
	data, _ := json.Marshal(&rec)
	if err := os.WriteFile(filepath.Join(tmp, "metadata.json"), data, 0o644); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	got, err := LoadMetadata(tmp)
	if err != nil {
		t.Fatalf("LoadMetadata() returned error: %v", err)
	}
	if got.ID != rec.ID {
		t.Errorf("ID = %q, want %q", got.ID, rec.ID)
	}
	if got.Name != rec.Name {
		t.Errorf("Name = %q, want %q", got.Name, rec.Name)
	}
}

func TestParseSkillFrontmatterVersion(t *testing.T) {
	tests := []struct {
		name  string
		content string
		want  string
	}{
		{
			name:  "version in frontmatter",
			content: "---\nname: My Skill\nversion: 2.1.0\ndescription: A test skill\n---\n# My Skill\n",
			want:  "2.1.0",
		},
		{
			name:  "quoted version",
			content: "---\nversion: \"3.0.0\"\nname: Quoted\n---\n# Quoted\n",
			want:  "3.0.0",
		},
		{
			name:  "no frontmatter",
			content: "# My Skill\n\nNo frontmatter here.\n",
			want:  "",
		},
		{
			name:  "frontmatter without version",
			content: "---\nname: No Version\ndescription: Just a name\n---\n# No Version\n",
			want:  "",
		},
		{
			name:  "empty frontmatter",
			content: "---\n---\n# Empty\n",
			want:  "",
		},
		{
			name:  "comments ignored",
			content: "---\n# this is a comment\nversion: 1.5.0\n# another comment\n---\n# Comments\n",
			want:  "1.5.0",
		},
		{
			name:  "extra fields ignored",
			content: "---\nname: Extra\nauthor: Someone\nversion: 4.2.0\ntags: [a, b]\n---\n# Extra\n",
			want:  "4.2.0",
		},
		{
			name:  "no closing delimiter",
			content: "---\nversion: 1.0.0\nname: Unclosed\n",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()
			if err := os.WriteFile(filepath.Join(tmp, "SKILL.md"), []byte(tt.content), 0o644); err != nil {
				t.Fatalf("WriteFile() returned error: %v", err)
			}
			got := parseSkillFrontmatterVersion(tmp)
			if got != tt.want {
				t.Errorf("parseSkillFrontmatterVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseSkillFrontmatterVersion_NoFile(t *testing.T) {
	tmp := t.TempDir()
	// No SKILL.md in the directory
	got := parseSkillFrontmatterVersion(tmp)
	if got != "" {
		t.Errorf("parseSkillFrontmatterVersion() = %q, want empty string for missing SKILL.md", got)
	}
}

func TestOriginType_Values(t *testing.T) {
	types := []OriginType{
		OriginLocal, OriginGitHub, OriginSkillsSh,
		OriginClaude, OriginBuiltin, OriginUnknown,
	}
	for _, ot := range types {
		if ot == "" {
			t.Error("OriginType is empty string")
		}
	}
}
