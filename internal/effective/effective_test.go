package effective

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEstimateTokenCount_EmptyDir(t *testing.T) {
	tmp := t.TempDir()
	got := EstimateTokenCount(tmp)
	if got != 0 {
		t.Errorf("EstimateTokenCount(empty) = %d, want 0", got)
	}
}

func TestEstimateTokenCount_WithSKILLmd(t *testing.T) {
	tmp := t.TempDir()
	content := "# My Skill\n\n" + repeat("A", 400)
	if err := os.WriteFile(filepath.Join(tmp, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	got := EstimateTokenCount(tmp)
	if got < 90 || got > 110 {
		t.Errorf("EstimateTokenCount() = %d, want ~100", got)
	}
}

func TestEstimateTokenCount_WithREADME(t *testing.T) {
	tmp := t.TempDir()
	content := "# README\n" + repeat("B", 800)
	if err := os.WriteFile(filepath.Join(tmp, "README.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	got := EstimateTokenCount(tmp)
	if got < 190 || got > 210 {
		t.Errorf("EstimateTokenCount() = %d, want ~200", got)
	}
}

func TestEstimateTokenCount_WithScript(t *testing.T) {
	tmp := t.TempDir()
	content := "#!/bin/bash\n" + repeat("echo hello\n", 100)
	if err := os.WriteFile(filepath.Join(tmp, "run.sh"), []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	got := EstimateTokenCount(tmp)
	if got < 250 || got > 300 {
		t.Errorf("EstimateTokenCount() = %d, want ~275", got)
	}
}

func TestEstimateTokenCount_WarningThreshold(t *testing.T) {
	tmp := t.TempDir()
	content := repeat("C", 5000)
	if err := os.WriteFile(filepath.Join(tmp, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	got := EstimateTokenCount(tmp)
	if got < 1200 || got > 1300 {
		t.Errorf("EstimateTokenCount() = %d, want ~1250", got)
	}
}

func TestGetEffectiveSkills_Empty(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	projDir := filepath.Join(tmp, "proj")
	if err := os.MkdirAll(projDir, 0o755); err != nil {
		t.Fatal(err)
	}

	got := GetEffectiveSkills(projDir)
	if len(got) != 0 {
		t.Errorf("GetEffectiveSkills(empty) = %v, want empty", got)
	}
}

func TestGetEffectiveSkills_Unmanaged(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	projDir := filepath.Join(tmp, "proj")

	skillDir := filepath.Join(projDir, ".claude", "skills", "my-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Skill"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := GetEffectiveSkills(projDir)
	if len(got) != 1 {
		t.Fatalf("GetEffectiveSkills() len = %d, want 1", len(got))
	}
	if got[0].Name != "my-skill" {
		t.Errorf("Name = %q, want %q", got[0].Name, "my-skill")
	}
	if got[0].Managed {
		t.Error("Managed = true, want false (unmanaged)")
	}
	if len(got[0].Agents) != 1 || got[0].Agents[0] != "claude" {
		t.Errorf("Agents = %v, want [claude]", got[0].Agents)
	}
}

func TestGetEffectiveSkills_Managed(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	projDir := filepath.Join(tmp, "proj")
	libSkillDir := filepath.Join(tmp, "library", "test-skill")

	if err := os.MkdirAll(libSkillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(libSkillDir, "SKILL.md"), []byte("# Library Skill"), 0o644); err != nil {
		t.Fatal(err)
	}

	targetDir := filepath.Join(projDir, ".claude", "skills")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	linkPath := filepath.Join(targetDir, "test-skill")
	if err := os.Symlink(libSkillDir, linkPath); err != nil {
		t.Fatal(err)
	}

	got := GetEffectiveSkills(projDir)
	if len(got) != 1 {
		t.Fatalf("GetEffectiveSkills() len = %d, want 1", len(got))
	}
	if !got[0].Managed {
		t.Error("Managed = false, want true")
	}
	// Canonical path should be the resolved library path (may include /private on macOS)
	expectedCanonical, _ := filepath.EvalSymlinks(libSkillDir)
	if got[0].CanonicalPath != expectedCanonical {
		t.Errorf("CanonicalPath = %q, want %q", got[0].CanonicalPath, expectedCanonical)
	}
}

func TestGetEffectiveSkills_SymlinkDeduplication(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	projDir := filepath.Join(tmp, "proj")
	libSkillDir := filepath.Join(tmp, "library", "shared-skill")

	// Create one library skill
	if err := os.MkdirAll(libSkillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(libSkillDir, "SKILL.md"), []byte("# Shared"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create symlinks in both .claude and .codex pointing to same source
	for _, dir := range []string{".claude", ".codex"} {
		targetDir := filepath.Join(projDir, dir, "skills")
		if err := os.MkdirAll(targetDir, 0o755); err != nil {
			t.Fatal(err)
		}
		linkPath := filepath.Join(targetDir, "shared-skill")
		if err := os.Symlink(libSkillDir, linkPath); err != nil {
			t.Fatal(err)
		}
	}

	got := GetEffectiveSkills(projDir)
	if len(got) != 1 {
		t.Fatalf("GetEffectiveSkills() len = %d, want 1 (deduplicated by symlink target)", len(got))
	}
	// Should show both agents
	if len(got[0].Agents) != 2 {
		t.Errorf("Agents = %v, want 2 agents", got[0].Agents)
	}
	// Should have both locations
	if len(got[0].Locations) != 2 {
		t.Errorf("Locations = %v, want 2 locations", got[0].Locations)
	}
}

func TestGetEffectiveSkills_WarningThreshold(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	projDir := filepath.Join(tmp, "proj")

	skillDir := filepath.Join(projDir, ".claude", "skills", "big-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := repeat("D", 9000) // 2250 tokens > 2000 warning
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	got := GetEffectiveSkills(projDir)
	if len(got) != 1 {
		t.Fatalf("GetEffectiveSkills() len = %d, want 1", len(got))
	}
	if !got[0].Warning {
		t.Error("Warning = false, want true")
	}
	if got[0].Danger {
		t.Error("Danger = true, want false")
	}
}

func TestGetEffectiveSkills_DangerThreshold(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	projDir := filepath.Join(tmp, "proj")

	skillDir := filepath.Join(projDir, ".claude", "skills", "huge-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := repeat("E", 21000)
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	got := GetEffectiveSkills(projDir)
	if len(got) != 1 {
		t.Fatalf("GetEffectiveSkills() len = %d, want 1", len(got))
	}
	if !got[0].Warning {
		t.Error("Warning = false, want true")
	}
	if !got[0].Danger {
		t.Error("Danger = false, want true")
	}
}

func TestGetEffectiveSkills_BothSharedAndAgent(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	projDir := filepath.Join(tmp, "proj")

	claudeDir := filepath.Join(projDir, ".claude", "skills", "agent-skill")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "SKILL.md"), []byte("# Agent"), 0o644); err != nil {
		t.Fatal(err)
	}

	sharedDir := filepath.Join(projDir, ".agents", "skills", "shared-skill")
	if err := os.MkdirAll(sharedDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sharedDir, "SKILL.md"), []byte("# Shared"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := GetEffectiveSkills(projDir)
	if len(got) != 2 {
		t.Fatalf("GetEffectiveSkills() len = %d, want 2", len(got))
	}
}

func TestIsScriptExt(t *testing.T) {
	tests := []struct {
		ext  string
		want bool
	}{
		{".sh", true}, {".js", true}, {".ts", true}, {".py", true},
		{".go", true}, {".rb", true}, {".php", true}, {".pl", true},
		{".bash", true}, {".zsh", true}, {".fish", true}, {".ps1", true},
		{".bat", true}, {".cmd", true},
		{".md", false}, {".json", false}, {".txt", false}, {"", false},
	}
	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			got := isScriptExt(tt.ext)
			if got != tt.want {
				t.Errorf("isScriptExt(%q) = %v, want %v", tt.ext, got, tt.want)
			}
		})
	}
}

func repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
