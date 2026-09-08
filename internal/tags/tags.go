// Package tags defines the tag data model for organizing and filtering skills.
package tags

import (
	"encoding/json"
	"os"
	"sort"
	"time"
)

// Tag represents a single user-assigned label.
type Tag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// State represents a two-level tag tree (parent -> children).
// Parents cannot have children; only two levels are supported.
type State struct {
	Tags      []Tag      `json:"tags"`
	Assigned  map[string][]string `json:"assigned"` // skillID -> []tagID
	UpdatedAt time.Time  `json:"updatedAt"`
}

// Document is the full tag data persisted to tags.json.
type Document struct {
	Skills   State `json:"skills"`
	Projects State `json:"projects"`
}

// NewState returns an initialized State with empty tags and assignments.
func NewState() State {
	return State{
		Tags:     []Tag{},
		Assigned: map[string][]string{},
		UpdatedAt: time.Now().UTC(),
	}
}

// NewDocument returns an initialized Document.
func NewDocument() Document {
	return Document{
		Skills:   NewState(),
		Projects: NewState(),
	}
}

// AddTag adds a tag to the state, returning the new tag.
// If a tag with the same name already exists, it returns the existing tag.
func (s *State) AddTag(name string) Tag {
	for _, t := range s.Tags {
		if t.Name == name {
			return t
		}
	}
	id := generateID(name)
	tag := Tag{ID: id, Name: name}
	s.Tags = append(s.Tags, tag)
	s.UpdatedAt = time.Now().UTC()
	return tag
}

// RemoveTag removes a tag by ID and unassigns it from all skills.
func (s *State) RemoveTag(tagID string) {
	newTags := make([]Tag, 0, len(s.Tags))
	for _, t := range s.Tags {
		if t.ID != tagID {
			newTags = append(newTags, t)
		}
	}
	s.Tags = newTags

	for skillID, tagIDs := range s.Assigned {
		newTagIDs := make([]string, 0, len(tagIDs))
		for _, id := range tagIDs {
			if id != tagID {
				newTagIDs = append(newTagIDs, id)
			}
		}
		s.Assigned[skillID] = newTagIDs
		if len(newTagIDs) == 0 {
			delete(s.Assigned, skillID)
		}
	}
	s.UpdatedAt = time.Now().UTC()
}

// Assign adds a tag to a skill. If already assigned, no-op.
func (s *State) Assign(skillID, tagID string) {
	for _, id := range s.Assigned[skillID] {
		if id == tagID {
			return
		}
	}
	s.Assigned[skillID] = append(s.Assigned[skillID], tagID)
	s.UpdatedAt = time.Now().UTC()
}

// Unassign removes a tag from a skill.
func (s *State) Unassign(skillID, tagID string) {
	tagIDs := s.Assigned[skillID]
	newTagIDs := make([]string, 0, len(tagIDs))
	for _, id := range tagIDs {
		if id != tagID {
			newTagIDs = append(newTagIDs, id)
		}
	}
	if len(newTagIDs) == 0 {
		delete(s.Assigned, skillID)
	} else {
		s.Assigned[skillID] = newTagIDs
	}
	s.UpdatedAt = time.Now().UTC()
}

// GetTagsBySkill returns all tags assigned to a skill.
func (s *State) GetTagsBySkill(skillID string) []Tag {
	tagIDs := s.Assigned[skillID]
	if len(tagIDs) == 0 {
		return nil
	}
	tags := make([]Tag, 0, len(tagIDs))
	for _, id := range tagIDs {
		for _, t := range s.Tags {
			if t.ID == id {
				tags = append(tags, t)
				break
			}
		}
	}
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Name < tags[j].Name
	})
	return tags
}

// GetSkillsByTag returns all skill IDs assigned to a tag.
func (s *State) GetSkillsByTag(tagID string) []string {
	var skillIDs []string
	for skillID, tagIDs := range s.Assigned {
		for _, id := range tagIDs {
			if id == tagID {
				skillIDs = append(skillIDs, skillID)
				break
			}
		}
	}
	sort.Strings(skillIDs)
	return skillIDs
}

// generateID creates a simple unique ID from a tag name.
func generateID(name string) string {
	// Simple slug-based ID generation
	result := ""
	for i, c := range name {
		if c >= 'a' && c <= 'z' {
			result += string(c)
		} else if c >= 'A' && c <= 'Z' {
			result += string(c + 32) // to lowercase
		} else if c >= '0' && c <= '9' {
			result += string(c)
		} else if i > 0 {
			result += "-"
		}
	}
	if result == "" {
		result = "tag"
	}
	return result
}

// Save writes the document to a JSON file.
func (d *Document) Save(path string) error {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Load reads a document from a JSON file.
func Load(path string) (*Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}
