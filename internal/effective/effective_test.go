package effective

import (
	"os"
	"path/filepath"
	"testing"
	"ai-manager/internal/agents"
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
	// ~400 chars = ~100 tokens
	content := "# My Skill\n\n" + repeat("A", 400)
	if err := os.WriteFile(filepath.Join(tmp, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	got := EstimateTokenCount(tmp)
	// Should be roughly 100 tokens (400/4)
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
	content := "#!/bin/bash\n" + repeat("echo hello\n", 100) // ~1100 chars = ~275 tokens
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
	// 5000 chars = 1250 tokens (above warning threshold)
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
	// Prevent user-level directory scan from finding real skills
	t.Setenv("HOME", tmp)

	projDir := filepath.Join(tmp, "proj")
	if err := os.MkdirAll(projDir, 0o755); err != nil {
		t.Fatal(err)
	}

	got := GetEffectiveSkills(projDir, agents.AgentClaude)
	if len(got) != 0 {
		t.Errorf("GetEffectiveSkills(empty) = %v, want empty", got)
	}
}

func TestGetEffectiveSkills_Unmanaged(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	projDir := filepath.Join(tmp, "proj")

	// Create .claude/skills/my-skill/SKILL.md
	skillDir := filepath.Join(projDir, ".claude", "skills", "my-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Skill"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := GetEffectiveSkills(projDir, agents.AgentClaude)
	if len(got) != 1 {
		t.Fatalf("GetEffectiveSkills() len = %d, want 1", len(got))
	}
	if got[0].Name != "my-skill" {
		t.Errorf("Name = %q, want %q", got[0].Name, "my-skill")
	}
	if got[0].Managed {
		t.Error("Managed = true, want false (unmanaged)")
	}
	if got[0].Source != "unmanaged" {
		t.Errorf("Source = %q, want %q", got[0].Source, "unmanaged")
	}
}

func TestGetEffectiveSkills_Managed(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	projDir := filepath.Join(tmp, "proj")
	libSkillDir := filepath.Join(tmp, "library", "test-skill")

	// Create library skill
	if err := os.MkdirAll(libSkillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(libSkillDir, "SKILL.md"), []byte("# Library Skill"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create symlink in project
	targetDir := filepath.Join(projDir, ".claude", "skills")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	linkPath := filepath.Join(targetDir, "test-skill")
	if err := os.Symlink(libSkillDir, linkPath); err != nil {
		t.Fatal(err)
	}

	got := GetEffectiveSkills(projDir, agents.AgentClaude)
	if len(got) != 1 {
		t.Fatalf("GetEffectiveSkills() len = %d, want 1", len(got))
	}
	if !got[0].Managed {
		t.Error("Managed = false, want true")
	}
	if got[0].Source != "managed" {
		t.Errorf("Source = %q, want %q", got[0].Source, "managed")
	}
}

func TestGetEffectiveSkills_WarningThreshold(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	projDir := filepath.Join(tmp, "proj")

	// Create a large skill (>1000 tokens = warning)
	skillDir := filepath.Join(projDir, ".claude", "skills", "big-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := repeat("D", 5000) // 1250 tokens
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	got := GetEffectiveSkills(projDir, agents.AgentClaude)
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

	// Create a very large skill (>5000 tokens = danger)
	skillDir := filepath.Join(projDir, ".claude", "skills", "huge-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := repeat("E", 21000) // 5250 tokens
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	got := GetEffectiveSkills(projDir, agents.AgentClaude)
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

	// Create .claude/skills/agent-skill/SKILL.md
	claudeDir := filepath.Join(projDir, ".claude", "skills", "agent-skill")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "SKILL.md"), []byte("# Agent"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create .agents/skills/shared-skill/SKILL.md
	sharedDir := filepath.Join(projDir, ".agents", "skills", "shared-skill")
	if err := os.MkdirAll(sharedDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sharedDir, "SKILL.md"), []byte("# Shared"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := GetEffectiveSkills(projDir, agents.AgentClaude)
	if len(got) != 2 {
		t.Fatalf("GetEffectiveSkills() len = %d, want 2", len(got))
	}
}

func TestGetEffectiveSkills_Deduplication(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	projDir := filepath.Join(tmp, "proj")

	// Create same skill in both .claude and .agents
	for _, dir := range []string{".claude", ".agents"} {
		skillDir := filepath.Join(projDir, dir, "skills", "dup-skill")
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Dup"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	got := GetEffectiveSkills(projDir, agents.AgentClaude)
	if len(got) != 1 {
		t.Fatalf("GetEffectiveSkills() len = %d, want 1 (deduplicated)", len(got))
	}
}

func TestIsScriptExt(t *testing.T) {
	tests := []struct {
		ext  string
		want bool
	}{
		{".sh", true},
		{".js", true},
		{".ts", true},
		{".py", true},
		{".go", true},
		{".rb", true},
		{".php", true},
		{".pl", true},
		{".bash", true},
		{".zsh", true},
		{".fish", true},
		{".ps1", true},
		{".bat", true},
		{".cmd", true},
		{".md", false},
		{".json", false},
		{".txt", false},
		{"", false},
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
