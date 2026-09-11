package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"ai-manager/internal/agents"
	"ai-manager/internal/effective"
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
	skillUpdates map[string]bool // slug -> hasUpdate
}

// NewSkillService creates the skill service. The library is lazily
// initialized on first use.
func NewSkillService(s *State) *SkillService {
	return &SkillService{state: s, skillUpdates: make(map[string]bool)}
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

	if len(records) == 0 {
		return fmt.Errorf("no skills to add from %s", input)
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

// GetEffectiveSkills returns all effective skills visible to agents at the
// user level. It calls the effective package's GetEffectiveSkills with the
// user's home directory as the project path, scanning user-level agent
// directories (e.g. ~/.claude/skills, ~/.codex/skills, ~/.agents/skills).
// The result is deduplicated by canonical (resolved symlink) path.
func (s *SkillService) GetEffectiveSkills() ([]effective.EffectiveSkill, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home dir: %w", err)
	}
	return effective.GetEffectiveSkills(home), nil
}

// UninstallEffectiveSkill removes a skill from a specific location.
// For managed skills (symlinks), it removes the symlink.
// For unmanaged skills (real directories), it refuses to delete.
// The location is the directory containing the skill (e.g. ~/.claude/skills).
func (s *SkillService) UninstallEffectiveSkill(location, name string) error {
	path := filepath.Join(location, name)

	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Already removed
		}
		return err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("skill %q at %s is not a managed symlink; cannot uninstall", name, path)
	}
	return os.Remove(path)
}

// OpenSkillDirectory opens the source directory of a library skill in the
// system file manager (Finder on macOS, Explorer on Windows, etc.).
// The skill is identified by name or slug.
func (s *SkillService) OpenSkillDirectory(name string) error {
	lib, err := s.lib()
	if err != nil {
		return err
	}

	entry, found := lib.FindSkill(name)
	if !found {
		return fmt.Errorf("skill %q not found", name)
	}

	// The source directory is the skill's install location in the library
	// (e.g. ~/.ai-manager/library/skills/<slug>/).
	sourceDir := filepath.Join(lib.SkillDir(), entry.Slug)
	return openInFileManager(sourceDir)
}

// OpenEffectiveSkillDirectory opens the canonical source directory of an
// effective skill in the system file manager.
func (s *SkillService) OpenEffectiveSkillDirectory(canonicalPath string) error {
	return openInFileManager(canonicalPath)
}

// openInFileManager opens a directory in the system file manager.
func openInFileManager(path string) error {
	// Verify the path exists
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("path does not exist: %s", path)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", "-R", path)
	case "windows":
		cmd = exec.Command("explorer", "/select,", path)
	case "linux":
		cmd = exec.Command("xdg-open", filepath.Dir(path))
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// SkillAgentStatus represents a Library skill with its agent installation status.
type SkillAgentStatus struct {
	Name        string            `json:"name"`
	Slug        string            `json:"slug"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	SourceDir   string            `json:"sourceDir"`
	Origin      string            `json:"origin"`       // "local", "github", "skills.sh"
	HasUpdate   bool              `json:"hasUpdate"`    // true if a newer version is available
	Agents      map[string]bool   `json:"agents"`       // agent key -> installed (symlink exists)
}

// ExternalSkill represents a skill installed in an agent directory but not
// managed by the AI-Manager library. It shows which agents have it and
// allows the user to register it or associate it with other agents.
type ExternalSkill struct {
	Name        string            `json:"name"`
	Slug        string            `json:"slug"`
	Description string            `json:"description"`
	SourceDir   string            `json:"sourceDir"`
	Canonical   string            `json:"canonical"`
	Agents      map[string]bool   `json:"agents"` // agent key -> true if skill exists in that agent's dir
}

// GetExternalSkills scans all agent skill directories for skills that are
// not registered in the AI-Manager library. Returns them with their agent
// associations so the user can see and manage them.
func (s *SkillService) GetExternalSkills() ([]ExternalSkill, error) {
	lib, err := s.lib()
	if err != nil {
		return nil, err
	}

	cfg, _ := agents.LoadAgentConfig()

	// Collect all library skill canonical paths for deduplication
	libCanonical := make(map[string]bool)
	for _, entry := range lib.Registry().Skills {
		path := entry.Path
		if path == "" {
			path = filepath.Join(lib.SkillDir(), entry.Slug)
		}
		if canon, err := filepath.EvalSymlinks(path); err == nil {
			libCanonical[canon] = true
		}
	}

	// Track external skills by canonical path to avoid duplicates
	seen := make(map[string]*ExternalSkill)
	var results []ExternalSkill

	for _, agent := range cfg.Agents {
		userDir := agents.ResolvePath(agent.Path)
		if userDir == "" {
			continue
		}

		entries, err := os.ReadDir(userDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			skillPath := filepath.Join(userDir, entry.Name())
			linkPath := skillPath

			// Resolve symlinks to get canonical path
			canonical := skillPath
			if info, err := os.Lstat(skillPath); err == nil && info.Mode()&os.ModeSymlink != 0 {
				if resolved, err := filepath.EvalSymlinks(skillPath); err == nil {
					canonical = resolved
				}
			}

			// Skip if already in the library
			if libCanonical[canonical] {
				continue
			}

			// Check if we've already seen this canonical path
			if existing, ok := seen[canonical]; ok {
				// Just add the agent association
				existing.Agents[agent.Key] = true
				continue
			}

			// Load skill metadata if available
			name := entry.Name()
			slug := entry.Name()
			description := ""

			// Try to load metadata.json
			metaPath := filepath.Join(linkPath, "metadata.json")
			if _, err := os.Stat(metaPath); err == nil {
				if rec, err := skill.LoadMetadata(linkPath); err == nil {
					if rec.Name != "" {
						name = rec.Name
					}
					if rec.Slug != "" {
						slug = rec.Slug
					}
					if rec.Description != "" {
						description = rec.Description
					}
				}
			}

			// Try to load description from SKILL.md frontmatter
			if description == "" {
				desc := skill.ParseSkillFrontmatterDescription(linkPath)
				if desc != "" {
					description = desc
				}
			}

			// Check which other agents have this skill
			agentAssociations := make(map[string]bool)
			agentAssociations[agent.Key] = true

			// Check other agents for the same skill
			for _, otherAgent := range cfg.Agents {
				if otherAgent.Key == agent.Key {
					continue
				}
				otherDir := agents.ResolvePath(otherAgent.Path)
				if otherDir == "" {
					continue
				}
				otherPath := filepath.Join(otherDir, entry.Name())
				if _, err := os.Lstat(otherPath); err == nil {
					agentAssociations[otherAgent.Key] = true
				}
			}

			ext := &ExternalSkill{
				Name:        name,
				Slug:        slug,
				Description: description,
				SourceDir:   linkPath,
				Canonical:   canonical,
				Agents:      agentAssociations,
			}
			seen[canonical] = ext
			results = append(results, *ext)
		}
	}

	return results, nil
}

// RegisterExternalSkill adds an external skill (found in an agent directory)
// to the AI-Manager library by copying it or creating a symlink.
func (s *SkillService) RegisterExternalSkill(canonicalPath string) error {
	lib, err := s.lib()
	if err != nil {
		return err
	}

	// Verify the path exists
	if _, err := os.Stat(canonicalPath); err != nil {
		return fmt.Errorf("skill path not found: %w", err)
	}

	// Load metadata if available
	rec, err := skill.LoadMetadata(canonicalPath)
	if err != nil {
		// Create a minimal record
		name := filepath.Base(canonicalPath)
		rec = &skill.Record{
			ID:    name,
			Name:  name,
			Slug:  name,
		}
	}

	// Copy the skill to the library
	skillDir := filepath.Join(lib.SkillDir(), rec.Slug)
	if err := copyDir(canonicalPath, skillDir); err != nil {
		return fmt.Errorf("copy skill: %w", err)
	}

	// Update the record with the new path
	rec.Installed = true
	rec.UpdatedAt = time.Now().UTC()

	return lib.AddRecord(*rec)
}

// copyDir copies a directory recursively.
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else if entry.Type()&os.ModeSymlink != 0 {
			// Copy symlinks as-is
			target, err := os.Readlink(srcPath)
			if err != nil {
				return err
			}
			if err := os.Symlink(target, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// GetAgentConfig returns the dynamic agent configuration.
func (s *SkillService) GetAgentConfig() (*agents.AgentConfigFile, error) {
	return agents.LoadAgentConfig()
}

// SaveAgentConfig saves the agent configuration.
func (s *SkillService) SaveAgentConfig(cfg *agents.AgentConfigFile) error {
	return agents.SaveAgentConfig(cfg)
}

// GetSkillAgentStatus returns all Library skills with per-agent installation
// status. For each skill, checks whether a symlink exists in each agent's
// user-level directory pointing to the skill's source in the Library.
func (s *SkillService) GetSkillAgentStatus() ([]SkillAgentStatus, error) {
	lib, err := s.lib()
	if err != nil {
		return nil, err
	}

	// Auto-register any unregistered skills found in the Library directory
	_, _ = lib.RescanAndRegister()

	cfg, _ := agents.LoadAgentConfig()
	summaries := lib.ListSkills()
	results := make([]SkillAgentStatus, 0, len(summaries))

	for _, sum := range summaries {
		// Try to find by name first, then by slug
		entry, found := lib.FindSkill(sum.Name)
		if !found {
			entry, found = lib.FindSkill(sum.Slug)
		}
		if !found {
			continue
		}

		// Use entry.Path if available, otherwise fall back to library skills dir
		sourceDir := entry.Path
		if sourceDir == "" {
			sourceDir = filepath.Join(lib.SkillDir(), entry.Slug)
		}

		agentStatus := make(map[string]bool)
		for _, agent := range cfg.Agents {
			userDir := agents.ResolvePath(agent.Path)
			if userDir == "" {
				continue
			}
			linkPath := filepath.Join(userDir, entry.Slug)

			// Check if symlink exists (installed) or if it's a real directory
			info, err := os.Lstat(linkPath)
			installed := false
			if err == nil {
				// Symlink exists → installed
				if info.Mode()&os.ModeSymlink != 0 {
					installed = true
				} else if info.IsDir() {
					// Real directory at this path — also counts as installed
					installed = true
				}
			}
			agentStatus[agent.Key] = installed
		}

		// Fallback to slug if name is empty
		name := sum.Name
		if name == "" {
			name = sum.Slug
		}

		// Get origin from metadata if available
		origin := ""
		if lib, err := s.lib(); err == nil {
			for _, entry := range lib.Registry().Skills {
				if entry.Slug == sum.Slug && entry.Path != "" {
					if rec, err := skill.LoadMetadata(entry.Path); err == nil {
						origin = string(rec.Origin.Type)
					}
					break
				}
			}
		}

		results = append(results, SkillAgentStatus{
			Name:        name,
			Slug:        sum.Slug,
			Version:     sum.Version,
			Description: sum.Description,
			SourceDir:   sourceDir,
			Origin:      origin,
			HasUpdate:   s.skillUpdates[sum.Slug],
			Agents:      agentStatus,
		})
	}

	return results, nil
}

// ToggleSkillAgent installs or uninstalls a skill for a specific agent.
// Returns the new installed state.
func (s *SkillService) ToggleSkillAgent(skillName string, agentKey string) (bool, error) {
	lib, err := s.lib()
	if err != nil {
		return false, err
	}

	entry, found := lib.FindSkill(skillName)
	if !found {
		return false, fmt.Errorf("skill %q not found", skillName)
	}

	// Use entry.Path if available, otherwise fall back to library skills dir
	sourceDir := entry.Path
	if sourceDir == "" {
		sourceDir = filepath.Join(lib.SkillDir(), entry.Slug)
	}
	if _, err := os.Stat(sourceDir); err != nil {
		return false, fmt.Errorf("skill source not found: %s", sourceDir)
	}

	cfg, _ := agents.LoadAgentConfig()

	// Find the agent config
	var agentCfg agents.AgentConfig
	found = false
	for _, a := range cfg.Agents {
		if a.Key == agentKey {
			agentCfg = a
			found = true
			break
		}
	}
	if !found {
		return false, fmt.Errorf("unknown agent %q", agentKey)
	}

	userDir := agents.ResolvePath(agentCfg.Path)
	if userDir == "" {
		return false, fmt.Errorf("no directory for agent %q", agentKey)
	}

	if err := os.MkdirAll(userDir, 0o755); err != nil {
		return false, fmt.Errorf("create agent dir: %w", err)
	}

	linkPath := filepath.Join(userDir, entry.Slug)

	info, err := os.Lstat(linkPath)
	if err == nil && info.Mode()&os.ModeSymlink != 0 {
		if err := os.Remove(linkPath); err != nil {
			return false, fmt.Errorf("remove symlink: %w", err)
		}
		return false, nil
	}

	if err := os.Symlink(sourceDir, linkPath); err != nil {
		return false, fmt.Errorf("create symlink: %w", err)
	}
	return true, nil
}

// DiscoverableSkill represents a skill found via search that can be installed.
type DiscoverableSkill struct {
	Key       string `json:"key"`
	Name      string `json:"name"`
	Source    string `json:"source"`
	RepoOwner string `json:"repoOwner"`
	RepoName  string `json:"repoName"`
	RepoBranch string `json:"repoBranch"`
	Installs  int    `json:"installs"`
	ReadmeURL string `json:"readmeUrl"`
}

// SkillsShSearchResult is the result of searching skills.sh.
type SkillsShSearchResult struct {
	Skills    []DiscoverableSkill `json:"skills"`
	TotalCount int               `json:"totalCount"`
	Query     string              `json:"query"`
}

// skillsShAPIResponse is the raw response from the skills.sh API.
type skillsShAPIResponse struct {
	Query  string `json:"query"`
	Skills []struct {
		ID      string `json:"id"`
		SkillID string `json:"skillId"`
		Name    string `json:"name"`
		Installs int   `json:"installs"`
		Source  string `json:"source"`
	} `json:"skills"`
	Count int `json:"count"`
}

// SearchSkillsSh searches the skills.sh registry for skills matching the query.
func (s *SkillService) SearchSkillsSh(query string, limit, offset int) (*SkillsShSearchResult, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	params := url.Values{}
	params.Set("q", query)
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("offset", fmt.Sprintf("%d", offset))

	u, err := url.Parse("https://skills.sh/api/search")
	if err != nil {
		return nil, err
	}
	u.RawQuery = params.Encode()

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("search skills.sh: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("skills.sh API returned %d", resp.StatusCode)
	}

	var apiResp skillsShAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode skills.sh response: %w", err)
	}

	results := &SkillsShSearchResult{
		Skills:     make([]DiscoverableSkill, 0, len(apiResp.Skills)),
		TotalCount: apiResp.Count,
		Query:      apiResp.Query,
	}

	for _, s := range apiResp.Skills {
		parts := strings.SplitN(s.Source, "/", 2)
		if len(parts) != 2 {
			continue
		}
		owner, repo := parts[0], parts[1]
		results.Skills = append(results.Skills, DiscoverableSkill{
			Key:        s.ID,
			Name:       s.Name,
			Source:     s.Source,
			RepoOwner:  owner,
			RepoName:   repo,
			RepoBranch: "main",
			Installs:   s.Installs,
			ReadmeURL:  fmt.Sprintf("https://github.com/%s/%s", owner, repo),
		})
	}

	return results, nil
}

// GitHubSearchResult is the result of searching GitHub for skill repositories.
type GitHubSearchResult struct {
	Items []struct {
		FullName string `json:"full_name"`
		HTMLURL  string `json:"html_url"`
		Desc     string `json:"description"`
		Stars    int    `json:"stargazers_count"`
	} `json:"items"`
	TotalCount int `json:"total_count"`
}

// SearchGitHubSkills searches GitHub for repositories containing skills.
func (s *SkillService) SearchGitHubSkills(query string) (*GitHubSearchResult, error) {
	params := url.Values{}
	params.Set("q", query+" skill in:name,description")
	params.Set("per_page", "20")
	params.Set("sort", "stars")
	params.Set("order", "desc")

	u, err := url.Parse("https://api.github.com/search/repositories")
	if err != nil {
		return nil, err
	}
	u.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "AI-Manager/1.0")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var result GitHubSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode GitHub response: %w", err)
	}

	return &result, nil
}

// CheckSkillUpdates checks if any installed skills have newer versions available.
// Updates the HasUpdate field in the returned statuses.
func (s *SkillService) CheckSkillUpdates() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	lib, err := s.lib()
	if err != nil {
		return err
	}

	for _, entry := range lib.Registry().Skills {
		if entry.Path == "" {
			continue
		}

		// Load metadata to get origin
		rec, err := skill.LoadMetadata(entry.Path)
		if err != nil {
			continue
		}

		// Only check GitHub and skills.sh origins
		if rec.Origin.Type != skill.OriginGitHub && rec.Origin.Type != skill.OriginSkillsSh {
			continue
		}

		// For GitHub, check the latest release/tag
		if rec.Origin.Type == skill.OriginGitHub && rec.Origin.Repo != "" {
			latestVersion, err := s.checkGitHubLatestVersion(rec.Origin.Repo)
			if err != nil {
				continue
			}
			// Compare versions (simple string comparison for now)
			if latestVersion != "" && latestVersion != rec.Version {
				// Mark as having update
				s.skillUpdates[entry.Slug] = true
			}
		}
	}

	return nil
}

// TriggerUpdateCheck runs the update check and returns the updated statuses.
func (s *SkillService) TriggerUpdateCheck() ([]SkillAgentStatus, error) {
	if err := s.CheckSkillUpdates(); err != nil {
		return nil, err
	}
	return s.GetSkillAgentStatus()
}

// checkGitHubLatestVersion returns the latest release tag or version from a GitHub repo.
func (s *SkillService) checkGitHubLatestVersion(repo string) (string, error) {
	params := url.Values{}
	params.Set("per_page", "1")

	u, err := url.Parse(fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo))
	if err != nil {
		return "", err
	}
	u.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "AI-Manager/1.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	// Remove leading 'v' from tag name
	version := strings.TrimPrefix(release.TagName, "v")
	return version, nil
}
