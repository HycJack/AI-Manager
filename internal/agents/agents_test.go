package agents

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTargetDir_Universal(t *testing.T) {
	got := TargetDir("/projects/myproj", TargetUniversal)
	want := "/projects/myproj/.agents/skills"
	if got != want {
		t.Errorf("TargetDir(universal) = %q, want %q", got, want)
	}
}

func TestTargetDir_ClaudeCode(t *testing.T) {
	got := TargetDir("/projects/myproj", TargetClaudeCode)
	want := "/projects/myproj/.claude-code/skills"
	if got != want {
		t.Errorf("TargetDir(claude-code) = %q, want %q", got, want)
	}
}

func TestAgentDir(t *testing.T) {
	tests := []struct {
		agent AgentKind
		want  string
	}{
		{AgentClaudeCode, ".claude-code"},
		{AgentCodex, ".codex"},
		{AgentCursor, ".cursor"},
		{AgentOpenCode, ".opencode"},
		{AgentUniversal, ".agents"},
	}
	for _, tt := range tests {
		t.Run(string(tt.agent), func(t *testing.T) {
			got := AgentDir(tt.agent)
			if got != tt.want {
				t.Errorf("AgentDir(%q) = %q, want %q", tt.agent, got, tt.want)
			}
		})
	}
}

func TestTargetAgent(t *testing.T) {
	tests := []struct {
		target InstallTarget
		want   AgentKind
	}{
		{TargetUniversal, AgentUniversal},
		{TargetClaudeCode, AgentClaudeCode},
		{TargetCodex, AgentCodex},
		{TargetCursor, AgentCursor},
	}
	for _, tt := range tests {
		t.Run(string(tt.target), func(t *testing.T) {
			got := TargetAgent(tt.target)
			if got != tt.want {
				t.Errorf("TargetAgent(%q) = %q, want %q", tt.target, got, tt.want)
			}
		})
	}
}

func TestAgentDirs_Empty(t *testing.T) {
	tmp := t.TempDir()
	got := AgentDirs(tmp)
	if len(got) != 0 {
		t.Errorf("AgentDirs(empty) = %v, want empty", got)
	}
}

func TestAgentDirs_WithDirs(t *testing.T) {
	tmp := t.TempDir()

	claudeDir := filepath.Join(tmp, ".claude-code", "skills")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	codexDir := filepath.Join(tmp, ".codex", "skills")
	if err := os.MkdirAll(codexDir, 0o755); err != nil {
		t.Fatal(err)
	}

	got := AgentDirs(tmp)
	if len(got) != 2 {
		t.Fatalf("AgentDirs() = %v (len=%d), want 2", got, len(got))
	}

	found := map[InstallTarget]bool{}
	for _, tg := range got {
		found[tg] = true
	}
	if !found[TargetClaudeCode] || !found[TargetCodex] {
		t.Errorf("AgentDirs() = %v, want [claude-code codex]", got)
	}
}

func TestDiscoverSkills_Empty(t *testing.T) {
	tmp := t.TempDir()
	got := DiscoverSkills(tmp)
	if len(got) != 0 {
		t.Errorf("DiscoverSkills(empty) = %v, want empty", got)
	}
}

func TestDiscoverSkills_FindsSkill(t *testing.T) {
	tmp := t.TempDir()

	skillDir := filepath.Join(tmp, ".claude-code", "skills", "my-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# My Skill"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := DiscoverSkills(tmp)
	if len(got) != 1 {
		t.Fatalf("DiscoverSkills() = %v (len=%d), want 1", got, len(got))
	}
	if got[0].Name != "my-skill" {
		t.Errorf("Name = %q, want %q", got[0].Name, "my-skill")
	}
	if got[0].Agent != AgentClaudeCode {
		t.Errorf("Agent = %q, want %q", got[0].Agent, AgentClaudeCode)
	}
	if got[0].Target != TargetClaudeCode {
		t.Errorf("Target = %q, want %q", got[0].Target, TargetClaudeCode)
	}
}

func TestDiscoverSkills_MultipleAgents(t *testing.T) {
	tmp := t.TempDir()

	_, _, err := createSkillDir(tmp, ".claude-code", "skills", "skill-a", "SKILL.md", "# Skill A")
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = createSkillDir(tmp, ".codex", "skills", "skill-b", "README.md", "# Skill B")
	if err != nil {
		t.Fatal(err)
	}

	got := DiscoverSkills(tmp)
	if len(got) != 2 {
		t.Fatalf("DiscoverSkills() len = %d, want 2", len(got))
	}
}

func TestDiscoverSkills_IgnoresNonSkillDirs(t *testing.T) {
	tmp := t.TempDir()

	dir := filepath.Join(tmp, ".claude-code", "skills", "not-a-skill")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	got := DiscoverSkills(tmp)
	if len(got) != 0 {
		t.Errorf("DiscoverSkills() = %v, want empty (no manifest)", got)
	}
}

func TestDiscoverSkills_WithVersion(t *testing.T) {
	tmp := t.TempDir()

	skillDir := filepath.Join(tmp, ".claude-code", "skills", "versioned")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Skill"), 0o644); err != nil {
		t.Fatal(err)
	}
	meta := `{"version":"1.2.3","name":"versioned"}`
	if err := os.WriteFile(filepath.Join(skillDir, "metadata.json"), []byte(meta), 0o644); err != nil {
		t.Fatal(err)
	}

	got := DiscoverSkills(tmp)
	if len(got) != 1 {
		t.Fatalf("DiscoverSkills() len = %d, want 1", len(got))
	}
	if got[0].Version != "1.2.3" {
		t.Errorf("Version = %q, want %q", got[0].Version, "1.2.3")
	}
}

func TestAllAgentKinds(t *testing.T) {
	kinds := AllAgentKinds()
	if len(kinds) != 10 {
		t.Errorf("AllAgentKinds() len = %d, want 10", len(kinds))
	}
}

func TestAllTargets(t *testing.T) {
	targets := AllTargets()
	if len(targets) != 10 {
		t.Errorf("AllTargets() len = %d, want 10", len(targets))
	}
}

func createSkillDir(base string, parts ...string) (string, string, error) {
	dir := filepath.Join(append([]string{base}, parts[:len(parts)-1]...)...)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", err
	}
	manifest := filepath.Join(dir, parts[len(parts)-1])
	if err := os.WriteFile(manifest, []byte("# Test"), 0o644); err != nil {
		return "", "", err
	}
	return dir, manifest, nil
}
