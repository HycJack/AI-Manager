package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewLibrary(t *testing.T) {
	tmp := t.TempDir()

	lib, err := NewLibrary(tmp)
	if err != nil {
		t.Fatalf("NewLibrary() returned error: %v", err)
	}
	if lib == nil {
		t.Fatal("NewLibrary() returned nil")
	}
	if lib.Registry() == nil {
		t.Error("Registry() is nil")
	}
}

func TestLibrary_SaveLoad(t *testing.T) {
	tmp := t.TempDir()

	// Create library
	lib, err := NewLibrary(tmp)
	if err != nil {
		t.Fatalf("NewLibrary() returned error: %v", err)
	}

	// Add a record
	rec := Record{
		ID:      "test-id",
		Name:    "Test Skill",
		Slug:    "test-skill",
		Version: "1.0.0",
	}
	if err := lib.AddRecord(rec); err != nil {
		t.Fatalf("AddRecord() returned error: %v", err)
	}

	// Save
	if err := lib.Save(); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	// Create a new library from the same directory (simulates reload)
	lib2, err := NewLibrary(tmp)
	if err != nil {
		t.Fatalf("NewLibrary() reload returned error: %v", err)
	}

	if len(lib2.Registry().Skills) != 1 {
		t.Fatalf("registry skills length = %d, want 1", len(lib2.Registry().Skills))
	}
	if lib2.Registry().Skills[0].Name != "Test Skill" {
		t.Errorf("skill name = %q, want %q", lib2.Registry().Skills[0].Name, "Test Skill")
	}
}

func TestScanLocal(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string // relative path -> content
		wantCount int
		wantErr  bool
	}{
		{
			name: "no skills",
			files: map[string]string{},
			wantCount: 0,
		},
		{
			name: "one skill with SKILL.md",
			files: map[string]string{
				"my-skill/SKILL.md": "# My Skill\n\nContent here.",
			},
			wantCount: 1,
		},
		{
			name: "one skill with README.md fallback",
			files: map[string]string{
				"my-skill/README.md": "# My Skill\n\nREADME content.",
			},
			wantCount: 1,
		},
		{
			name: "SKILL.md takes precedence",
			files: map[string]string{
				"my-skill/SKILL.md": "# My Skill",
				"my-skill/README.md": "# My Skill",
			},
			wantCount: 1,
		},
		{
			name: "non-skill directories ignored",
			files: map[string]string{
				"my-skill/SKILL.md": "# Skill",
				"other-dir/file.txt": "not a skill",
			},
			wantCount: 1,
		},
		{
			name: "multiple skills",
			files: map[string]string{
				"skill-a/SKILL.md": "# A",
				"skill-b/SKILL.md": "# B",
				"skill-c/README.md": "# C",
			},
			wantCount: 3,
		},
		{
			name: "metadata.json is loaded",
			files: map[string]string{
				"my-skill/SKILL.md": "# My Skill",
				"my-skill/metadata.json": `{"name":"My Skill","version":"2.0.0","license":"MIT"}`,
			},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()

			// Create the library
			lib, err := NewLibrary(tmp)
			if err != nil {
				t.Fatalf("NewLibrary() returned error: %v", err)
			}

			// Create the scan root with test files
			scanRoot := filepath.Join(tmp, "scan-root")
			if err := os.MkdirAll(scanRoot, 0o755); err != nil {
				t.Fatalf("MkdirAll() returned error: %v", err)
			}
			for name, content := range tt.files {
				fullPath := filepath.Join(scanRoot, name)
				if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
					t.Fatalf("MkdirAll() returned error: %v", err)
				}
				if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
					t.Fatalf("WriteFile() returned error: %v", err)
				}
			}

			records, err := lib.ScanLocal(scanRoot)
			if tt.wantErr {
				if err == nil {
					t.Error("ScanLocal() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("ScanLocal() returned error: %v", err)
			}
			if len(records) != tt.wantCount {
				t.Errorf("ScanLocal() returned %d records, want %d", len(records), tt.wantCount)
			}
		})
	}
}

func TestScanLocal_NonExistentDir(t *testing.T) {
	tmp := t.TempDir()
	lib, _ := NewLibrary(tmp)

	_, err := lib.ScanLocal(filepath.Join(tmp, "nonexistent"))
	if err == nil {
		t.Error("ScanLocal() expected error for non-existent dir, got nil")
	}
}

func TestScanLocal_NotADirectory(t *testing.T) {
	tmp := t.TempDir()
	lib, _ := NewLibrary(tmp)

	filePath := filepath.Join(tmp, "file.txt")
	os.WriteFile(filePath, []byte("not a dir"), 0o644)

	_, err := lib.ScanLocal(filePath)
	if err == nil {
		t.Error("ScanLocal() expected error for non-directory, got nil")
	}
}

func TestScanNpx(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		wantType  OriginType
		wantCount int
	}{
		{
			name:      "simple name (skills.sh)",
			input:     "superpowers",
			wantType:  OriginSkillsSh,
			wantCount: 1,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "too many parts",
			input:   "a/b/c/d",
			wantErr: true,
		},
	}

	tmp := t.TempDir()
	lib, _ := NewLibrary(tmp)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			records, err := lib.ScanNpx(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("ScanNpx() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("ScanNpx() returned error: %v", err)
			}
			if len(records) != tt.wantCount {
				t.Errorf("ScanNpx() returned %d records, want %d", len(records), tt.wantCount)
			}
			if tt.wantCount > 0 && records[0].Origin.Type != tt.wantType {
				t.Errorf("Origin.Type = %q, want %q", records[0].Origin.Type, tt.wantType)
			}
		})
	}
}

func TestScanNpx_GitHubClone(t *testing.T) {
	// GitHub clone tests require network — skipped in unit tests.
	// Integration tests should test with actual repos.
	t.Skip("requires network access to github.com")
}

func TestScanClaude(t *testing.T) {
	tmp := t.TempDir()
	lib, _ := NewLibrary(tmp)

	// Create a fake plugin directory with a skill
	pluginDir := filepath.Join(tmp, "my-plugin")
	skillDir := filepath.Join(pluginDir, "my-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# My Skill"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Run("valid plugin path", func(t *testing.T) {
		records, err := lib.ScanClaude(pluginDir)
		if err != nil {
			t.Fatalf("ScanClaude() returned error: %v", err)
		}
		if len(records) != 1 {
			t.Fatalf("ScanClaude() returned %d records, want 1", len(records))
		}
		if records[0].Origin.Type != OriginClaude {
			t.Errorf("Origin.Type = %q, want %q", records[0].Origin.Type, OriginClaude)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		_, err := lib.ScanClaude("")
		if err == nil {
			t.Error("ScanClaude() expected error, got nil")
		}
	})

	t.Run("nonexistent path", func(t *testing.T) {
		_, err := lib.ScanClaude(filepath.Join(tmp, "nope"))
		if err == nil {
			t.Error("ScanClaude() expected error for nonexistent path, got nil")
		}
	})
}

func TestScanExisting(t *testing.T) {
	tests := []struct {
		name      string
		files     map[string]string // path -> content (paths are relative to tmp)
		scanPaths []string          // absolute paths to scan
		wantErr   bool
		wantCount int
	}{
		{
			name: "one skill in one path",
			files: map[string]string{
				"path1/skill-a/SKILL.md": "# A",
			},
			scanPaths: []string{"path1"},
			wantCount: 1,
		},
		{
			name: "skills across multiple paths",
			files: map[string]string{
				"path1/skill-a/SKILL.md": "# A",
				"path2/skill-b/SKILL.md": "# B",
			},
			scanPaths: []string{"path1", "path2"},
			wantCount: 2,
		},
		{
			name:    "no skills found",
			files:   map[string]string{},
			scanPaths: []string{"path1"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()

			// Create the library
			lib, _ := NewLibrary(tmp)

			// Create test files
			scanRoots := make([]string, len(tt.scanPaths))
			for i, relPath := range tt.scanPaths {
				scanRoots[i] = filepath.Join(tmp, relPath)
			}

			for name, content := range tt.files {
				fullPath := filepath.Join(tmp, name)
				if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			records, err := lib.ScanExisting(scanRoots)
			if tt.wantErr {
				if err == nil {
					t.Error("ScanExisting() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("ScanExisting() returned error: %v", err)
			}
			if len(records) != tt.wantCount {
				t.Errorf("ScanExisting() returned %d records, want %d", len(records), tt.wantCount)
			}
		})
	}
}

func TestAddRecord_Duplicate(t *testing.T) {
	tmp := t.TempDir()
	lib, _ := NewLibrary(tmp)

	rec := Record{ID: "dup-id", Name: "Dup Skill", Slug: "dup-skill", Version: "1.0.0"}
	if err := lib.AddRecord(rec); err != nil {
		t.Fatalf("first AddRecord() returned error: %v", err)
	}

	err := lib.AddRecord(rec)
	if err == nil {
		t.Error("duplicate AddRecord() expected error, got nil")
	}
}

func TestRemoveRecord(t *testing.T) {
	tmp := t.TempDir()
	lib, _ := NewLibrary(tmp)

	rec := Record{ID: "rem-id", Name: "Remove Me", Slug: "remove-me", Version: "1.0.0"}
	if err := lib.AddRecord(rec); err != nil {
		t.Fatalf("AddRecord() returned error: %v", err)
	}

	if err := lib.RemoveRecord("Remove Me"); err != nil {
		t.Fatalf("RemoveRecord() returned error: %v", err)
	}

	if len(lib.Registry().Skills) != 0 {
		t.Errorf("registry skills length = %d, want 0", len(lib.Registry().Skills))
	}
}

func TestRemoveRecord_NotFound(t *testing.T) {
	tmp := t.TempDir()
	lib, _ := NewLibrary(tmp)

	err := lib.RemoveRecord("nonexistent")
	if err == nil {
		t.Error("RemoveRecord() expected error for non-existent skill, got nil")
	}
}

func TestFindSkill(t *testing.T) {
	tmp := t.TempDir()
	lib, _ := NewLibrary(tmp)

	rec := Record{ID: "find-id", Name: "Find Me", Slug: "find-me", Version: "1.0.0"}
	if err := lib.AddRecord(rec); err != nil {
		t.Fatal(err)
	}

	// Find by name
	entry, found := lib.FindSkill("Find Me")
	if !found {
		t.Error("FindSkill() by name not found")
	} else if entry.ID != "find-id" {
		t.Errorf("FindSkill() ID = %q, want %q", entry.ID, "find-id")
	}

	// Find by slug
	entry, found = lib.FindSkill("find-me")
	if !found {
		t.Error("FindSkill() by slug not found")
	} else if entry.Name != "Find Me" {
		t.Errorf("FindSkill() Name = %q, want %q", entry.Name, "Find Me")
	}

	// Case insensitive
	entry, found = lib.FindSkill("find me")
	if !found {
		t.Error("FindSkill() case insensitive not found")
	}

	// Not found
	_, found = lib.FindSkill("nonexistent")
	if found {
		t.Error("FindSkill() expected not found")
	}
}

func TestListSkills(t *testing.T) {
	tmp := t.TempDir()
	lib, _ := NewLibrary(tmp)

	lib.AddRecord(Record{ID: "l1", Name: "Skill 1", Slug: "skill-1", Version: "1.0.0"})
	lib.AddRecord(Record{ID: "l2", Name: "Skill 2", Slug: "skill-2", Version: "2.0.0"})

	summaries := lib.ListSkills()
	if len(summaries) != 2 {
		t.Fatalf("ListSkills() length = %d, want 2", len(summaries))
	}
	if summaries[0].Name != "Skill 1" {
		t.Errorf("summaries[0].Name = %q, want %q", summaries[0].Name, "Skill 1")
	}
}

func TestListSkillFiles(t *testing.T) {
	tmp := t.TempDir()
	lib, _ := NewLibrary(tmp)

	// Create a skill with files
	skillPath := filepath.Join(lib.skillsDir, "test-skill")
	os.MkdirAll(filepath.Join(skillPath, "subdir"), 0o755)
	os.WriteFile(filepath.Join(skillPath, "SKILL.md"), []byte("# Test"), 0o644)
	os.WriteFile(filepath.Join(skillPath, "script.js"), []byte("console.log()"), 0o644)
	os.WriteFile(filepath.Join(skillPath, "subdir", "helper.js"), []byte("export {}"), 0o644)

	files, err := lib.ListSkillFiles("test-skill")
	if err != nil {
		t.Fatalf("ListSkillFiles() returned error: %v", err)
	}
	if len(files) != 3 {
		t.Errorf("ListSkillFiles() returned %d files, want 3", len(files))
	}
}

func TestLoadSkillReadme(t *testing.T) {
	tmp := t.TempDir()
	lib, _ := NewLibrary(tmp)

	skillPath := filepath.Join(lib.skillsDir, "readme-skill")
	os.MkdirAll(skillPath, 0o755)
	os.WriteFile(filepath.Join(skillPath, "SKILL.md"), []byte("# My Skill\n\nContent."), 0o644)

	readme, err := lib.LoadSkillReadme("readme-skill")
	if err != nil {
		t.Fatalf("LoadSkillReadme() returned error: %v", err)
	}
	if readme == "" {
		t.Error("LoadSkillReadme() returned empty string")
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"my-skill", "my-skill"},
		{"My_Skill", "my-skill"},
		{"My Skill", "my-skill"},
		{"my/skill", "my-skill"},
		{"my\\skill", "my-skill"},
		{"  my-skill  ", "my-skill"},
		{"MySkill", "myskill"},
		{"my--skill", "my--skill"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := slugify(tt.input)
			if got != tt.want {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestTitleCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"my-skill", "My Skill"},
		{"my_skill", "My Skill"},
		{"my skill", "My Skill"},
		{"myskill", "Myskill"},
		{"my/skill.name", "My Skill Name"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := titleCase(tt.input)
			if got != tt.want {
				t.Errorf("titleCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestHashString(t *testing.T) {
	h := hashString("test")
	if h == "" {
		t.Error("hashString() returned empty string")
	}
	if len(h) != 12 {
		t.Errorf("hashString() length = %d, want 12", len(h))
	}

	// Deterministic
	h2 := hashString("test")
	if h != h2 {
		t.Error("hashString() is not deterministic")
	}

	// Different inputs produce different hashes
	h3 := hashString("other")
	if h == h3 {
		t.Error("hashString() produced same hash for different inputs")
	}
}
