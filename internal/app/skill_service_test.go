package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSkillService_ListSkills_Empty(t *testing.T) {
	s := NewSkillService(newTestState(t))
	skills, err := s.ListSkills()
	if err != nil {
		t.Fatalf("ListSkills() returned error: %v", err)
	}
	if len(skills) != 0 {
		t.Errorf("ListSkills() returned %d skills, want 0", len(skills))
	}
}

func TestSkillService_AddSkill_UnknownKind(t *testing.T) {
	s := NewSkillService(newTestState(t))
	err := s.AddSkill("unknown", "test", "")
	if err == nil {
		t.Error("AddSkill() with unknown kind expected error, got nil")
	}
}

func TestSkillService_AddSkill_Npx(t *testing.T) {
	s := NewSkillService(newTestState(t))
	err := s.AddSkill("npx", "owner/repo", "")
	if err != nil {
		t.Fatalf("AddSkill() returned error: %v", err)
	}

	skills, err := s.ListSkills()
	if err != nil {
		t.Fatalf("ListSkills() returned error: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("ListSkills() returned %d skills, want 1", len(skills))
	}
	if skills[0].Name != "Repo" {
		t.Errorf("skill name = %q, want %q", skills[0].Name, "Repo")
	}
}

func TestSkillService_AddSkill_NpxSimple(t *testing.T) {
	s := NewSkillService(newTestState(t))
	err := s.AddSkill("npx", "superpowers", "")
	if err != nil {
		t.Fatalf("AddSkill() returned error: %v", err)
	}

	skills, _ := s.ListSkills()
	if len(skills) != 1 {
		t.Fatalf("ListSkills() returned %d skills, want 1", len(skills))
	}
	if skills[0].Name != "Superpowers" {
		t.Errorf("skill name = %q, want %q", skills[0].Name, "Superpowers")
	}
}

func TestSkillService_AddSkill_Claude(t *testing.T) {
	s := NewSkillService(newTestState(t))
	err := s.AddSkill("claude", "superpowers", "")
	if err != nil {
		t.Fatalf("AddSkill() returned error: %v", err)
	}

	skills, _ := s.ListSkills()
	if len(skills) != 1 {
		t.Fatalf("ListSkills() returned %d skills, want 1", len(skills))
	}
	if skills[0].Name != "Superpowers" {
		t.Errorf("skill name = %q, want %q", skills[0].Name, "Superpowers")
	}
}

func TestSkillService_AddSkill_Existing(t *testing.T) {
	s := NewSkillService(newTestState(t))

	// Create a test directory with a skill
	testDir := t.TempDir()
	skillDir := filepath.Join(testDir, "my-skill")
	os.MkdirAll(skillDir, 0o755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# My Skill"), 0o644)

	err := s.AddSkill("existing", testDir, "")
	if err != nil {
		t.Fatalf("AddSkill() returned error: %v", err)
	}

	skills, _ := s.ListSkills()
	if len(skills) != 1 {
		t.Fatalf("ListSkills() returned %d skills, want 1", len(skills))
	}
	if skills[0].Name != "My Skill" {
		t.Errorf("skill name = %q, want %q", skills[0].Name, "My Skill")
	}
}

func TestSkillService_AddSkill_Local(t *testing.T) {
	s := NewSkillService(newTestState(t))

	// Create a test directory with a skill
	testDir := t.TempDir()
	skillDir := filepath.Join(testDir, "local-skill")
	os.MkdirAll(skillDir, 0o755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Local Skill"), 0o644)

	err := s.AddSkill("local", testDir, "")
	if err != nil {
		t.Fatalf("AddSkill() returned error: %v", err)
	}

	skills, _ := s.ListSkills()
	if len(skills) != 1 {
		t.Fatalf("ListSkills() returned %d skills, want 1", len(skills))
	}
	if skills[0].Name != "Local Skill" {
		t.Errorf("skill name = %q, want %q", skills[0].Name, "Local Skill")
	}
}

func TestSkillService_GetSkill(t *testing.T) {
	s := NewSkillService(newTestState(t))

	// Add a skill first
	s.AddSkill("npx", "owner/repo", "")

	detail, err := s.GetSkill("Repo")
	if err != nil {
		t.Fatalf("GetSkill() returned error: %v", err)
	}
	if detail.Record.Name != "Repo" {
		t.Errorf("detail.Record.Name = %q, want %q", detail.Record.Name, "Repo")
	}
}

func TestSkillService_GetSkill_NotFound(t *testing.T) {
	s := NewSkillService(newTestState(t))
	_, err := s.GetSkill("nonexistent")
	if err == nil {
		t.Error("GetSkill() expected error for non-existent skill, got nil")
	}
}

func TestSkillService_RemoveSkill(t *testing.T) {
	s := NewSkillService(newTestState(t))

	// Add a skill
	s.AddSkill("npx", "owner/repo", "")

	// Remove it
	err := s.RemoveSkill("Repo")
	if err != nil {
		t.Fatalf("RemoveSkill() returned error: %v", err)
	}

	skills, _ := s.ListSkills()
	if len(skills) != 0 {
		t.Errorf("ListSkills() returned %d skills, want 0", len(skills))
	}
}

func TestSkillService_RemoveSkill_NotFound(t *testing.T) {
	s := NewSkillService(newTestState(t))
	err := s.RemoveSkill("nonexistent")
	if err == nil {
		t.Error("RemoveSkill() expected error for non-existent skill, got nil")
	}
}

func TestSkillService_CheckUpdates(t *testing.T) {
	s := NewSkillService(newTestState(t))

	// Empty library
	updates, err := s.CheckUpdates()
	if err != nil {
		t.Fatalf("CheckUpdates() returned error: %v", err)
	}
	if len(updates) != 0 {
		t.Errorf("CheckUpdates() returned %d updates, want 0", len(updates))
	}

	// Add a skill
	s.AddSkill("npx", "owner/repo", "")

	updates, err = s.CheckUpdates()
	if err != nil {
		t.Fatalf("CheckUpdates() returned error: %v", err)
	}
	if len(updates) != 1 {
		t.Fatalf("CheckUpdates() returned %d updates, want 1", len(updates))
	}
	if updates[0].Name != "Repo" {
		t.Errorf("update name = %q, want %q", updates[0].Name, "Repo")
	}
	if updates[0].Current != updates[0].Latest {
		t.Error("expected current == latest (no remote check)")
	}
}

func TestSkillService_UpdateSkill(t *testing.T) {
	s := NewSkillService(newTestState(t))

	// Add a skill
	s.AddSkill("npx", "owner/repo", "")

	// Update it (no-op currently)
	err := s.UpdateSkill("Repo")
	if err != nil {
		t.Fatalf("UpdateSkill() returned error: %v", err)
	}
}

func TestSkillService_UpdateSkill_NotFound(t *testing.T) {
	s := NewSkillService(newTestState(t))
	err := s.UpdateSkill("nonexistent")
	if err == nil {
		t.Error("UpdateSkill() expected error for non-existent skill, got nil")
	}
}

func TestSkillService_MultipleAddAndRemove(t *testing.T) {
	s := NewSkillService(newTestState(t))

	// Add multiple skills
	s.AddSkill("npx", "owner/repo1", "")
	s.AddSkill("npx", "owner/repo2", "")
	s.AddSkill("claude", "plugin", "")

	skills, _ := s.ListSkills()
	if len(skills) != 3 {
		t.Fatalf("ListSkills() returned %d skills, want 3", len(skills))
	}

	// Remove one
	s.RemoveSkill("Repo1")

	skills, _ = s.ListSkills()
	if len(skills) != 2 {
		t.Fatalf("after remove: ListSkills() returned %d skills, want 2", len(skills))
	}

	// Verify the remaining skills
	names := map[string]bool{}
	for _, s := range skills {
		names[s.Name] = true
	}
	if !names["Repo2"] || !names["Plugin"] {
		t.Errorf("remaining skills: %+v", names)
	}
}

func TestSkillService_DuplicateAdd(t *testing.T) {
	s := NewSkillService(newTestState(t))

	err := s.AddSkill("npx", "owner/repo", "")
	if err != nil {
		t.Fatalf("first AddSkill() returned error: %v", err)
	}

	// Adding the same skill again should fail
	err = s.AddSkill("npx", "owner/repo", "")
	if err == nil {
		t.Error("duplicate AddSkill() expected error, got nil")
	}
}
