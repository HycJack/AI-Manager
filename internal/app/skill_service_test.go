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
	err := s.AddSkill("unknown", "test", "", nil)
	if err == nil {
		t.Error("AddSkill() with unknown kind expected error, got nil")
	}
}

func TestSkillService_AddSkill_Npx(t *testing.T) {
	s := NewSkillService(newTestState(t))
	err := s.AddSkill("npx", "superpowers", "", nil)
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
	if skills[0].Name != "Superpowers" {
		t.Errorf("skill name = %q, want %q", skills[0].Name, "Superpowers")
	}
}

func TestSkillService_AddSkill_NpxSimple(t *testing.T) {
	s := NewSkillService(newTestState(t))
	err := s.AddSkill("npx", "superpowers", "", nil)
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

	// Create a local directory with a skill (Claude plugins are local dirs)
	testDir := t.TempDir()
	skillDir := filepath.Join(testDir, "my-plugin-skill")
	os.MkdirAll(skillDir, 0o755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Plugin Skill"), 0o644)

	err := s.AddSkill("local", testDir, "", nil)
	if err != nil {
		t.Fatalf("AddSkill() returned error: %v", err)
	}

	skills, _ := s.ListSkills()
	if len(skills) != 1 {
		t.Fatalf("ListSkills() returned %d skills, want 1", len(skills))
	}
	if skills[0].Name != "My Plugin Skill" {
		t.Errorf("skill name = %q, want %q", skills[0].Name, "My Plugin Skill")
	}
}

func TestSkillService_AddSkill_Existing(t *testing.T) {
	s := NewSkillService(newTestState(t))

	// Create a test directory with a skill
	testDir := t.TempDir()
	skillDir := filepath.Join(testDir, "my-skill")
	os.MkdirAll(skillDir, 0o755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# My Skill"), 0o644)

	err := s.AddSkill("existing", testDir, "", nil)
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

	err := s.AddSkill("local", testDir, "", nil)
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
	s.AddSkill("npx", "superpowers", "", nil)

	detail, err := s.GetSkill("Superpowers")
	if err != nil {
		t.Fatalf("GetSkill() returned error: %v", err)
	}
	if detail.Record.Name != "Superpowers" {
		t.Errorf("detail.Record.Name = %q, want %q", detail.Record.Name, "Superpowers")
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
	s.AddSkill("npx", "superpowers", "", nil)

	// Remove it
	err := s.RemoveSkill("Superpowers")
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
	s.AddSkill("npx", "superpowers", "", nil)

	updates, err = s.CheckUpdates()
	if err != nil {
		t.Fatalf("CheckUpdates() returned error: %v", err)
	}
	if len(updates) != 1 {
		t.Fatalf("CheckUpdates() returned %d updates, want 1", len(updates))
	}
	if updates[0].Name != "Superpowers" {
		t.Errorf("update name = %q, want %q", updates[0].Name, "Superpowers")
	}
	if updates[0].Current != updates[0].Latest {
		t.Error("expected current == latest (no remote check)")
	}
}

func TestSkillService_UpdateSkill(t *testing.T) {
	s := NewSkillService(newTestState(t))

	// Add a skill
	s.AddSkill("npx", "superpowers", "", nil)

	// Update it (no-op currently)
	err := s.UpdateSkill("Superpowers")
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

	// Create test dirs with skills
	dir1 := t.TempDir()
	os.MkdirAll(filepath.Join(dir1, "skill-a"), 0o755)
	os.WriteFile(filepath.Join(dir1, "skill-a", "SKILL.md"), []byte("# A"), 0o644)

	dir2 := t.TempDir()
	os.MkdirAll(filepath.Join(dir2, "skill-b"), 0o755)
	os.WriteFile(filepath.Join(dir2, "skill-b", "SKILL.md"), []byte("# B"), 0o644)

	dir3 := t.TempDir()
	os.MkdirAll(filepath.Join(dir3, "skill-c"), 0o755)
	os.WriteFile(filepath.Join(dir3, "skill-c", "SKILL.md"), []byte("# C"), 0o644)

	// Add multiple skills
	s.AddSkill("local", dir1, "", nil)
	s.AddSkill("local", dir2, "", nil)
	s.AddSkill("local", dir3, "", nil)

	skills, _ := s.ListSkills()
	if len(skills) != 3 {
		t.Fatalf("ListSkills() returned %d skills, want 3", len(skills))
	}

	// Remove one
	s.RemoveSkill("Skill A")

	skills, _ = s.ListSkills()
	if len(skills) != 2 {
		t.Fatalf("after remove: ListSkills() returned %d skills, want 2", len(skills))
	}

	// Verify the remaining skills
	names := map[string]bool{}
	for _, s := range skills {
		names[s.Name] = true
	}
	if !names["Skill B"] || !names["Skill C"] {
		t.Errorf("remaining skills: %+v", names)
	}
}

func TestSkillService_DuplicateAdd(t *testing.T) {
	s := NewSkillService(newTestState(t))

	err := s.AddSkill("npx", "superpowers", "", nil)
	if err != nil {
		t.Fatalf("first AddSkill() returned error: %v", err)
	}

	// Adding the same skill again should fail
	err = s.AddSkill("npx", "superpowers", "", nil)
	if err == nil {
		t.Error("duplicate AddSkill() expected error, got nil")
	}
}
