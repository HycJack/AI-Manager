package app

import (
	"fmt"
	"strings"
	"sync"

	"ai-manager/internal/skill"
)

// SkillDetail is the full detail returned by GetSkill, including the
// underlying record, file list, and readme content.
type SkillDetail struct {
	Record skill.Record `json:"record"`
	Files  []string     `json:"files"`
	Readme string       `json:"readme"`
}

// UpdateInfo describes a skill update check result.
type UpdateInfo struct {
	Name    string `json:"name"`
	Current string `json:"current"`
	Latest  string `json:"latest"`
}

// SkillService manages the skill library through the Wails3 binding layer.
// It exposes 6 methods that the frontend calls to list, add, remove, and
// check updates for skills.
type SkillService struct {
	state    *State
	mu       sync.Mutex
	library  *skill.Library
}

// NewSkillService creates the skill service. The library is lazily
// initialized on first use.
func NewSkillService(s *State) *SkillService {
	return &SkillService{state: s}
}

// lib initializes and returns the library, creating it on first call.
func (s *SkillService) lib() (*skill.Library, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.library != nil {
		return s.library, nil
	}
	lib, err := skill.NewLibrary(s.state.cfg.DataDir)
	if err != nil {
		return nil, err
	}
	s.library = lib
	return lib, nil
}

// ListSkills returns all skills in the library as Summaries.
func (s *SkillService) ListSkills() ([]skill.Summary, error) {
	lib, err := s.lib()
	if err != nil {
		return nil, err
	}
	return lib.ListSkills(), nil
}

// GetSkill returns the full detail (record + files + readme) for a skill
// identified by name or slug.
func (s *SkillService) GetSkill(name string) (SkillDetail, error) {
	lib, err := s.lib()
	if err != nil {
		return SkillDetail{}, err
	}

	entry, found := lib.FindSkill(name)
	if !found {
		return SkillDetail{}, fmt.Errorf("skill %q not found", name)
	}

	rec, err := lib.LoadSkillRecord(entry.Slug)
	if err != nil {
		return SkillDetail{}, fmt.Errorf("load metadata: %w", err)
	}

	files, err := lib.ListSkillFiles(entry.Slug)
	if err != nil {
		// Non-fatal: return empty file list
		files = []string{}
	}

	readme, err := lib.LoadSkillReadme(entry.Slug)
	if err != nil {
		// Non-fatal: return empty readme
		readme = ""
	}

	return SkillDetail{
		Record: *rec,
		Files:  files,
		Readme: readme,
	}, nil
}

// SkillSummary is a lightweight view of a skill for scan results.
type SkillSummary struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Version   string `json:"version"`
	Origin    string `json:"origin"`
	Installed bool   `json:"installed"`
}

// ScanSkill scans a source without adding anything. Returns the list of
// discovered skills for preview. The frontend shows these in a result list
// with multi-select, then calls AddSkill with the selected slugs.
func (s *SkillService) ScanSkill(kind, input string) ([]SkillSummary, error) {
	lib, err := s.lib()
	if err != nil {
		return nil, err
	}

	var records []skill.Record

	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "local":
		records, err = lib.ScanLocal(input)
	case "npx":
		records, err = lib.ScanNpx(input)
	case "claude":
		records, err = lib.ScanClaude(input)
	case "existing":
		paths := strings.Split(input, ",")
		for i, p := range paths {
			paths[i] = strings.TrimSpace(p)
		}
		records, err = lib.ScanExisting(paths)
	default:
		return nil, fmt.Errorf("unknown skill source kind %q (expected: local, npx, claude, existing)", kind)
	}

	if err != nil {
		return nil, err
	}

	summaries := make([]SkillSummary, 0, len(records))
	for _, rec := range records {
		summaries = append(summaries, SkillSummary{
			ID:        rec.ID,
			Name:      rec.Name,
			Slug:      rec.Slug,
			Version:   rec.Version,
			Origin:    string(rec.Origin.Type),
			Installed: rec.Installed,
		})
	}
	return summaries, nil
}

// AddSkill adds selected skills to the library. The selectedSlugs parameter
// contains the slugs of skills to add (from a prior ScanSkill call). If empty,
// all scanned skills are added.
func (s *SkillService) AddSkill(kind, input, groupName string, selectedSlugs []string) error {
	lib, err := s.lib()
	if err != nil {
		return err
	}

	var records []skill.Record

	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "local":
		records, err = lib.ScanLocal(input)
	case "npx":
		records, err = lib.ScanNpx(input)
	case "claude":
		records, err = lib.ScanClaude(input)
	case "existing":
		paths := strings.Split(input, ",")
		for i, p := range paths {
			paths[i] = strings.TrimSpace(p)
		}
		records, err = lib.ScanExisting(paths)
	default:
		return fmt.Errorf("unknown skill source kind %q (expected: local, npx, claude, existing)", kind)
	}

	if err != nil {
		return err
	}

	// Filter by selected slugs if provided
	if len(selectedSlugs) > 0 {
		selected := make(map[string]bool, len(selectedSlugs))
		for _, slug := range selectedSlugs {
			selected[slug] = true
		}
		filtered := make([]skill.Record, 0, len(records))
		for _, rec := range records {
			if selected[rec.Slug] {
				filtered = append(filtered, rec)
			}
		}
		records = filtered
	}

	// Add each record to the library
	for i := range records {
		records[i].Group = groupName
		if err := lib.AddRecord(records[i]); err != nil {
			return err
		}
	}

	return nil
}

// RemoveSkill removes a skill from the library by name or slug.
func (s *SkillService) RemoveSkill(name string) error {
	lib, err := s.lib()
	if err != nil {
		return err
	}
	return lib.RemoveRecord(name)
}

// CheckUpdates checks all skills in the library for available updates.
// For local skills, the version from metadata.json is compared against
// itself (no remote check). For remote skills, this would check the
// source for newer versions.
// Currently returns empty list since remote update checking is not
// implemented yet — each skill's current version is reported as latest.
func (s *SkillService) CheckUpdates() ([]UpdateInfo, error) {
	lib, err := s.lib()
	if err != nil {
		return nil, err
	}

	var updates []UpdateInfo
	for _, entry := range lib.Registry().Skills {
		// Placeholder: report current version as latest
		updates = append(updates, UpdateInfo{
			Name:    entry.Name,
			Current: entry.Version,
			Latest:  entry.Version,
		})
	}
	return updates, nil
}

// UpdateSkill updates a skill to its latest version.
// Currently a no-op since remote update resolution is not implemented.
func (s *SkillService) UpdateSkill(name string) error {
	lib, err := s.lib()
	if err != nil {
		return err
	}

	entry, found := lib.FindSkill(name)
	if !found {
		return fmt.Errorf("skill %q not found", name)
	}

	// Mark as updated (no actual download yet)
	for i := range lib.Registry().Skills {
		if lib.Registry().Skills[i].ID == entry.ID {
			// Version stays the same for now
			break
		}
	}

	return lib.Save()
}
