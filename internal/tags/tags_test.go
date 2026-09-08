package tags

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNewDocument(t *testing.T) {
	doc := NewDocument()
	if len(doc.Skills.Tags) != 0 {
		t.Errorf("Skills.Tags length = %d, want 0", len(doc.Skills.Tags))
	}
	if doc.Skills.Assigned == nil {
		t.Error("Skills.Assigned is nil, want empty map")
	}
	if len(doc.Projects.Tags) != 0 {
		t.Errorf("Projects.Tags length = %d, want 0", len(doc.Projects.Tags))
	}
}

func TestAddTag(t *testing.T) {
	state := NewState()

	tag1 := state.AddTag("frontend")
	if tag1.ID == "" {
		t.Error("AddTag() returned empty ID")
	}
	if tag1.Name != "frontend" {
		t.Errorf("AddTag() name = %q, want %q", tag1.Name, "frontend")
	}

	// Adding the same tag again should return the existing tag
	tag2 := state.AddTag("frontend")
	if tag2.ID != tag1.ID {
		t.Errorf("AddTag() duplicate ID = %q, want %q", tag2.ID, tag1.ID)
	}

	if len(state.Tags) != 1 {
		t.Errorf("Tags length = %d, want 1", len(state.Tags))
	}
}

func TestAddTag_Multiple(t *testing.T) {
	state := NewState()
	state.AddTag("frontend")
	state.AddTag("backend")
	state.AddTag("database")

	if len(state.Tags) != 3 {
		t.Errorf("Tags length = %d, want 3", len(state.Tags))
	}
}

func TestRemoveTag(t *testing.T) {
	state := NewState()
	tag := state.AddTag("frontend")
	state.Assign("skill-001", tag.ID)
	state.Assign("skill-002", tag.ID)

	state.RemoveTag(tag.ID)

	if len(state.Tags) != 0 {
		t.Errorf("Tags length = %d, want 0", len(state.Tags))
	}
	if len(state.Assigned) != 0 {
		t.Errorf("Assigned length = %d, want 0", len(state.Assigned))
	}
}

func TestAssignUnassign(t *testing.T) {
	state := NewState()
	tag := state.AddTag("frontend")

	state.Assign("skill-001", tag.ID)

	tags := state.GetTagsBySkill("skill-001")
	if len(tags) != 1 {
		t.Fatalf("GetTagsBySkill() length = %d, want 1", len(tags))
	}
	if tags[0].ID != tag.ID {
		t.Errorf("GetTagsBySkill() ID = %q, want %q", tags[0].ID, tag.ID)
	}

	// Unassign
	state.Unassign("skill-001", tag.ID)

	tags = state.GetTagsBySkill("skill-001")
	if len(tags) != 0 {
		t.Errorf("GetTagsBySkill() after unassign = %d, want 0", len(tags))
	}
}

func TestGetSkillsByTag(t *testing.T) {
	state := NewState()
	tag := state.AddTag("frontend")
	state.Assign("skill-001", tag.ID)
	state.Assign("skill-002", tag.ID)
	state.Assign("skill-003", tag.ID)

	skills := state.GetSkillsByTag(tag.ID)
	if len(skills) != 3 {
		t.Fatalf("GetSkillsByTag() length = %d, want 3", len(skills))
	}
}

func TestDocument_JSONRoundTrip(t *testing.T) {
	doc := NewDocument()
	tag := doc.Skills.AddTag("frontend")
	doc.Skills.Assign("skill-001", tag.ID)

	data, err := json.Marshal(&doc)
	if err != nil {
		t.Fatalf("Marshal() returned error: %v", err)
	}

	var decoded Document
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() returned error: %v", err)
	}

	if len(decoded.Skills.Tags) != 1 {
		t.Errorf("Tags length = %d, want 1", len(decoded.Skills.Tags))
	}
	if decoded.Skills.Tags[0].Name != "frontend" {
		t.Errorf("Tag name = %q, want %q", decoded.Skills.Tags[0].Name, "frontend")
	}
}

func TestDocument_SaveLoad(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "tags.json")

	doc := NewDocument()
	tag := doc.Skills.AddTag("test")
	doc.Skills.Assign("skill-001", tag.ID)

	if err := doc.Save(path); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("os.Stat() returned error: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if len(loaded.Skills.Tags) != 1 {
		t.Errorf("Tags length = %d, want 1", len(loaded.Skills.Tags))
	}
}

func TestGenerateID(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"frontend", "frontend"},
		{"Backend", "backend"},
		{"Database", "database"},
		{"My Tag", "my-tag"},
		{"", "tag"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateID(tt.name)
			if got != tt.want {
				t.Errorf("generateID(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}
