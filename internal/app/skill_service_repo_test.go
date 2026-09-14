package app

import (
	"strings"
	"testing"
)

func TestListRepoSkills(t *testing.T) {
	s := NewSkillService(newTestState(t))

	if _, err := s.ListRepoSkills("   "); err == nil {
		t.Error("expected an error for a blank repository")
	}

	out, err := s.ListRepoSkills("demo-skill")
	if err != nil {
		t.Fatalf("ListRepoSkills: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("got %d skills, want 1", len(out))
	}
	if !strings.Contains(out[0].Slug, "demo") {
		t.Errorf("slug = %q, want it to contain %q", out[0].Slug, "demo")
	}
	if out[0].Name == "" {
		t.Error("name should not be empty")
	}
}
