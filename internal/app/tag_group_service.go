package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"ai-manager/internal/config"
	"ai-manager/internal/tags"
)

// TagGroupService manages tags and groups, persisted to tags.json.
// Deep module: 9 methods at the interface, full tag tree + persistence behind it.
type TagGroupService struct {
	state *State
	mu    sync.Mutex
}

// NewTagGroupService creates the tag/group service.
func NewTagGroupService(s *State) *TagGroupService {
	return &TagGroupService{state: s}
}

// tagsFilePath returns the path to tags.json in the data directory.
func (s *TagGroupService) tagsFilePath() string {
	return filepath.Join(s.state.cfg.DataDir, "tags.json")
}

// loadTags reads and returns the current tags document from disk.
func (s *TagGroupService) loadTags() (*tags.Document, error) {
	path := s.tagsFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			doc := tags.NewDocument()
			return &doc, nil
		}
		return nil, err
	}
	var doc tags.Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

// saveTags writes the tags document to disk atomically.
func (s *TagGroupService) saveTags(doc *tags.Document) error {
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	return config.SaveAtomic(s.tagsFilePath(), data)
}

// ListTags returns the full tags document (skills tags + projects tags).
func (s *TagGroupService) ListTags() (*tags.Document, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadTags()
}

// AddTag adds a tag to the document. If parentID is non-empty, it's a child tag.
// Returns the new tag.
func (s *TagGroupService) AddTag(name, parentID string) (*tags.Tag, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	doc, err := s.loadTags()
	if err != nil {
		return nil, err
	}

	// Add to skills state
	tag := doc.Skills.AddTag(name)
	if parentID != "" {
		// Mark as child by adding parent reference
		tag.ParentID = parentID
	}

	if err := s.saveTags(doc); err != nil {
		return nil, err
	}
	return &tag, nil
}

// RenameTag renames a tag by ID.
func (s *TagGroupService) RenameTag(tagID, newName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	doc, err := s.loadTags()
	if err != nil {
		return err
	}

	for i, t := range doc.Skills.Tags {
		if t.ID == tagID {
			doc.Skills.Tags[i].Name = newName
			break
		}
	}

	return s.saveTags(doc)
}

// DeleteTag removes a tag by ID and unassigns it from all skills.
func (s *TagGroupService) DeleteTag(tagID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	doc, err := s.loadTags()
	if err != nil {
		return err
	}

	doc.Skills.RemoveTag(tagID)
	return s.saveTags(doc)
}

// AssignTag assigns a tag to a skill.
func (s *TagGroupService) AssignTag(skillID, tagID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	doc, err := s.loadTags()
	if err != nil {
		return err
	}

	doc.Skills.Assign(skillID, tagID)
	return s.saveTags(doc)
}

// UnassignTag removes a tag assignment from a skill.
func (s *TagGroupService) UnassignTag(skillID, tagID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	doc, err := s.loadTags()
	if err != nil {
		return err
	}

	doc.Skills.Unassign(skillID, tagID)
	return s.saveTags(doc)
}

// ListGroups returns all groups.
func (s *TagGroupService) ListGroups() ([]Group, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	doc, err := s.loadTags()
	if err != nil {
		return nil, err
	}

	// Groups are stored in the document's Groups field
	groups := make([]Group, 0, len(doc.Groups))
	for _, g := range doc.Groups {
		groups = append(groups, Group{
			ID:     g.ID,
			Name:   g.Name,
			Skills: g.Skills,
		})
	}
	return groups, nil
}

// AddGroup creates a new group.
func (s *TagGroupService) AddGroup(name string) (*Group, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	doc, err := s.loadTags()
	if err != nil {
		return nil, err
	}

	id := generateGroupID(name)
	g := tags.GroupEntry{
		ID:     id,
		Name:   name,
		Skills: []string{},
	}
	doc.Groups = append(doc.Groups, g)

	if err := s.saveTags(doc); err != nil {
		return nil, err
	}
	return &Group{ID: g.ID, Name: g.Name, Skills: g.Skills}, nil
}

// DeleteGroup removes a group by ID. Skills are moved to Ungrouped.
func (s *TagGroupService) DeleteGroup(groupID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	doc, err := s.loadTags()
	if err != nil {
		return err
	}

	newGroups := make([]tags.GroupEntry, 0, len(doc.Groups))
	for _, g := range doc.Groups {
		if g.ID != groupID {
			newGroups = append(newGroups, g)
		}
	}
	doc.Groups = newGroups

	return s.saveTags(doc)
}

// MoveSkillToGroup adds a skill to a group (removes from other groups first).
func (s *TagGroupService) MoveSkillToGroup(skillID, groupID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	doc, err := s.loadTags()
	if err != nil {
		return err
	}

	// Remove skill from all other groups
	for i := range doc.Groups {
		skills := make([]string, 0, len(doc.Groups[i].Skills))
		for _, s := range doc.Groups[i].Skills {
			if s != skillID {
				skills = append(skills, s)
			}
		}
		doc.Groups[i].Skills = skills
	}

	// Add to target group
	for i := range doc.Groups {
		if doc.Groups[i].ID == groupID {
			doc.Groups[i].Skills = append(doc.Groups[i].Skills, skillID)
			break
		}
	}

	return s.saveTags(doc)
}

// Group is the API-level group representation.
type Group struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Skills []string `json:"skills"`
}

// generateGroupID creates a simple unique ID from a group name.
func generateGroupID(name string) string {
	result := ""
	for i, c := range name {
		if c >= 'a' && c <= 'z' {
			result += string(c)
		} else if c >= 'A' && c <= 'Z' {
			result += string(c + 32)
		} else if c >= '0' && c <= '9' {
			result += string(c)
		} else if i > 0 {
			result += "-"
		}
	}
	if result == "" {
		result = "group"
	}
	// Add timestamp suffix for uniqueness
	return result + "-" + time.Now().UTC().Format("20060102150405")
}
