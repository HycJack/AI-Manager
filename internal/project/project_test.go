package project

import (
	"os"
	"path/filepath"
	"testing"
	"ai-manager/internal/agents"
)

func TestInstallSkill_CreatesSymlink(t *testing.T) {
	// Create a library skill directory
	tmp := t.TempDir()
	libSkillDir := filepath.Join(tmp, "library", "skills", "test-skill")
	if err := os.MkdirAll(libSkillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(libSkillDir, "SKILL.md"), []byte("# Test"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a project directory
	projDir := filepath.Join(tmp, "projects", "myproj")
	if err := os.MkdirAll(projDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Install the skill
	err := InstallSkill(libSkillDir, projDir, agents.TargetClaude)
	if err != nil {
		t.Fatalf("InstallSkill() returned error: %v", err)
	}

	// Verify symlink exists
	linkPath := filepath.Join(projDir, ".claude", "skills", "test-skill")
	info, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatalf("Lstat() returned error: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("Installed skill is not a symlink")
	}
}

func TestInstallSkill_AlreadyInstalled(t *testing.T) {
	tmp := t.TempDir()
	libSkillDir := filepath.Join(tmp, "library", "test-skill")
	if err := os.MkdirAll(libSkillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	projDir := filepath.Join(tmp, "proj")
	if err := os.MkdirAll(projDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Install twice
	err := InstallSkill(libSkillDir, projDir, agents.TargetClaude)
	if err != nil {
		t.Fatalf("First install: %v", err)
	}
	err = InstallSkill(libSkillDir, projDir, agents.TargetClaude)
	if err != nil {
		t.Fatalf("Second install (should be no-op): %v", err)
	}
}

func TestInstallSkill_SharesTarget(t *testing.T) {
	tmp := t.TempDir()
	libSkillDir := filepath.Join(tmp, "library", "test-skill")
	if err := os.MkdirAll(libSkillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	projDir := filepath.Join(tmp, "proj")
	if err := os.MkdirAll(projDir, 0o755); err != nil {
		t.Fatal(err)
	}

	err := InstallSkill(libSkillDir, projDir, agents.TargetShared)
	if err != nil {
		t.Fatalf("InstallSkill() returned error: %v", err)
	}

	linkPath := filepath.Join(projDir, ".agents", "skills", "test-skill")
	if _, err := os.Lstat(linkPath); err != nil {
		t.Errorf("Symlink not found at %s: %v", linkPath, err)
	}
}

func TestUninstallSkill_RemovesSymlink(t *testing.T) {
	tmp := t.TempDir()
	libSkillDir := filepath.Join(tmp, "library", "test-skill")
	if err := os.MkdirAll(libSkillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	projDir := filepath.Join(tmp, "proj")
	if err := os.MkdirAll(projDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Install
	if err := InstallSkill(libSkillDir, projDir, agents.TargetClaude); err != nil {
		t.Fatal(err)
	}

	// Uninstall
	err := UninstallSkill(projDir, "test-skill", agents.TargetClaude)
	if err != nil {
		t.Fatalf("UninstallSkill() returned error: %v", err)
	}

	// Verify symlink removed
	linkPath := filepath.Join(projDir, ".claude", "skills", "test-skill")
	if _, err := os.Lstat(linkPath); !os.IsNotExist(err) {
		t.Errorf("Symlink still exists after uninstall")
	}
}

func TestUninstallSkill_AlreadyUninstalled(t *testing.T) {
	tmp := t.TempDir()
	projDir := filepath.Join(tmp, "proj")
	if err := os.MkdirAll(projDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Should be no-op
	err := UninstallSkill(projDir, "nonexistent", agents.TargetClaude)
	if err != nil {
		t.Fatalf("UninstallSkill() on missing: %v", err)
	}
}

func TestUninstallSkill_RefusesRealDir(t *testing.T) {
	tmp := t.TempDir()
	projDir := filepath.Join(tmp, "proj")
	// Create a real directory (not a symlink)
	realDir := filepath.Join(projDir, ".claude", "skills", "real-dir")
	if err := os.MkdirAll(realDir, 0o755); err != nil {
		t.Fatal(err)
	}

	err := UninstallSkill(projDir, "real-dir", agents.TargetClaude)
	if err == nil {
		t.Error("UninstallSkill() should refuse to delete non-symlink directory")
	}
}

func TestScanProject_Empty(t *testing.T) {
	tmp := t.TempDir()
	projDir := filepath.Join(tmp, "proj")
	if err := os.MkdirAll(projDir, 0o755); err != nil {
		t.Fatal(err)
	}

	got := ScanProject(projDir)
	if len(got) != 0 {
		t.Errorf("ScanProject(empty) = %v, want empty", got)
	}
}

func TestScanProject_FindsInstallations(t *testing.T) {
	tmp := t.TempDir()
	libSkillDir := filepath.Join(tmp, "library", "test-skill")
	if err := os.MkdirAll(libSkillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	projDir := filepath.Join(tmp, "proj")
	if err := os.MkdirAll(projDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Install two skills
	if err := InstallSkill(libSkillDir, projDir, agents.TargetClaude); err != nil {
		t.Fatal(err)
	}
	if err := InstallSkill(libSkillDir, projDir, agents.TargetCodex); err != nil {
		t.Fatal(err)
	}

	got := ScanProject(projDir)
	if len(got) != 2 {
		t.Fatalf("ScanProject() len = %d, want 2", len(got))
	}

	// Check first
	found := map[agents.InstallTarget]bool{}
	for _, inst := range got {
		if inst.SkillName != "test-skill" {
			t.Errorf("SkillName = %q, want %q", inst.SkillName, "test-skill")
		}
		if !inst.IsSymlink {
			t.Error("IsSymlink = false, want true")
		}
		found[inst.Target] = true
	}
	if !found[agents.TargetClaude] || !found[agents.TargetCodex] {
		t.Errorf("ScanProject() targets = %v, want claude + codex", found)
	}
}

func TestManager_AddRemoveProject(t *testing.T) {
	tmp := t.TempDir()
	os.Setenv("AI_MANAGER_HOME", tmp)
	defer os.Unsetenv("AI_MANAGER_HOME")

	m := NewManager()

	// Add
	err := m.AddProject("/projects/myproj")
	if err != nil {
		t.Fatalf("AddProject() returned error: %v", err)
	}

	projects, err := m.LoadProjects()
	if err != nil {
		t.Fatalf("LoadProjects() returned error: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("Projects len = %d, want 1", len(projects))
	}
	if projects[0].Path != "/projects/myproj" {
		t.Errorf("Path = %q, want %q", projects[0].Path, "/projects/myproj")
	}
	if projects[0].Name != "myproj" {
		t.Errorf("Name = %q, want %q", projects[0].Name, "myproj")
	}

	// Remove
	err = m.RemoveProject("/projects/myproj")
	if err != nil {
		t.Fatalf("RemoveProject() returned error: %v", err)
	}

	projects, _ = m.LoadProjects()
	if len(projects) != 0 {
		t.Errorf("Projects after remove = %d, want 0", len(projects))
	}
}

func TestManager_AddProject_Duplicate(t *testing.T) {
	tmp := t.TempDir()
	os.Setenv("AI_MANAGER_HOME", tmp)
	defer os.Unsetenv("AI_MANAGER_HOME")

	m := NewManager()

	// Add same project twice
	err := m.AddProject("/projects/dup")
	if err != nil {
		t.Fatal(err)
	}
	err = m.AddProject("/projects/dup")
	if err != nil {
		t.Fatal(err)
	}

	projects, _ := m.LoadProjects()
	if len(projects) != 1 {
		t.Errorf("Projects len = %d, want 1 (no duplicates)", len(projects))
	}
}

func TestManager_RemoveProject_NotFound(t *testing.T) {
	tmp := t.TempDir()
	os.Setenv("AI_MANAGER_HOME", tmp)
	defer os.Unsetenv("AI_MANAGER_HOME")

	m := NewManager()
	err := m.RemoveProject("/nonexistent")
	if err == nil {
		t.Error("RemoveProject() should return error for non-existent project")
	}
}
