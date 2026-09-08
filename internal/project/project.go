// Package project manages project registrations and skill installations via
// symlinks (macOS/Linux) or directory junctions (Windows).
package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"ai-manager/internal/agents"
	"ai-manager/internal/platform"
)

// Project represents a registered project directory.
type Project struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

// Installation represents a skill installation (symlink) in a project.
type Installation struct {
	SkillName string               `json:"skillName"`
	Target    agents.InstallTarget `json:"target"`
	Path      string               `json:"path"`
	IsSymlink bool                 `json:"isSymlink"`
}

// ProjectsFile is the JSON structure for projects.json.
type ProjectsFile struct {
	Projects []Project `json:"projects"`
}

// Manager handles project CRUD and skill installations.
type Manager struct {
	mu      sync.Mutex
	dataDir string
}

// NewManager creates a project manager rooted at the app data directory.
func NewManager() *Manager {
	dataDir, _ := platform.DataDir()
	if dataDir == "" {
		dataDir = "."
	}
	return &Manager{dataDir: dataDir}
}

// projectsPath returns the on-disk location of projects.json.
func (m *Manager) projectsPath() string {
	return filepath.Join(m.dataDir, "projects.json")
}

// LoadProjects reads the projects list from disk.
func (m *Manager) LoadProjects() ([]Project, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	path := m.projectsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var pf ProjectsFile
	if err := json.Unmarshal(data, &pf); err != nil {
		return nil, err
	}
	return pf.Projects, nil
}

// SaveProjects writes the projects list to disk.
func (m *Manager) SaveProjects(projects []Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pf := ProjectsFile{Projects: projects}
	data, err := json.MarshalIndent(pf, "", "  ")
	if err != nil {
		return err
	}
	return saveAtomic(m.projectsPath(), data)
}

// AddProject registers a project. Name defaults to the directory basename.
func (m *Manager) AddProject(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	projects, err := m.loadProjectsLocked()
	if err != nil {
		return err
	}

	// Check if already exists
	for _, p := range projects {
		if p.Path == path {
			return nil // Already added
		}
	}

	name := filepath.Base(path)
	projects = append(projects, Project{Path: path, Name: name})
	return m.saveProjectsLocked(projects)
}

// RemoveProject removes a project by path.
func (m *Manager) RemoveProject(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	projects, err := m.loadProjectsLocked()
	if err != nil {
		return err
	}

	filtered := make([]Project, 0, len(projects))
	found := false
	for _, p := range projects {
		if p.Path == path {
			found = true
			continue
		}
		filtered = append(filtered, p)
	}
	if !found {
		return os.ErrNotExist
	}
	return m.saveProjectsLocked(filtered)
}

// loadProjectsLocked reads projects without holding the mutex.
func (m *Manager) loadProjectsLocked() ([]Project, error) {
	path := m.projectsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var pf ProjectsFile
	if err := json.Unmarshal(data, &pf); err != nil {
		return nil, err
	}
	return pf.Projects, nil
}

// saveProjectsLocked writes projects without holding the mutex.
func (m *Manager) saveProjectsLocked(projects []Project) error {
	pf := ProjectsFile{Projects: projects}
	data, err := json.MarshalIndent(pf, "", "  ")
	if err != nil {
		return err
	}
	return saveAtomic(m.projectsPath(), data)
}

// InstallSkill creates a symlink from the library skill to the project's target
// directory. If the skill is already installed (symlink exists), it returns nil.
//
// Cross-platform: uses os.Symlink on macOS/Linux. On Windows, os.Symlink
// works for NTFS; for non-NTFS, a copy fallback would be needed (future).
func InstallSkill(librarySkillPath, projectPath string, target agents.InstallTarget) error {
	targetDir := agents.TargetDir(projectPath, target)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return err
	}

	skillName := filepath.Base(librarySkillPath)
	linkPath := filepath.Join(targetDir, skillName)

	// Check if already installed
	if _, err := os.Lstat(linkPath); err == nil {
		return nil // Already installed
	}

	// Create symlink
	return os.Symlink(librarySkillPath, linkPath)
}

// UninstallSkill removes the symlink for a skill from a project.
func UninstallSkill(projectPath string, skillName string, target agents.InstallTarget) error {
	targetDir := agents.TargetDir(projectPath, target)
	linkPath := filepath.Join(targetDir, skillName)

	// Only remove if it's a symlink (don't delete real directories)
	info, err := os.Lstat(linkPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Already uninstalled
		}
		return err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return os.ErrInvalid // Not a symlink, refuse to delete
	}

	return os.Remove(linkPath)
}

// ScanProject returns all installations in a project (symlinks in agent dirs).
func ScanProject(projectPath string) []Installation {
	var installations []Installation

	targets := agents.AgentDirs(projectPath)
	for _, target := range targets {
		targetDir := agents.TargetDir(projectPath, target)
		entries, err := os.ReadDir(targetDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			linkPath := filepath.Join(targetDir, entry.Name())
			info, err := os.Lstat(linkPath)
			if err != nil {
				continue
			}
			// Skip non-directory entries (including symlinks to files)
			if !info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
				continue
			}

			inst := Installation{
				SkillName: entry.Name(),
				Target:    target,
				Path:      linkPath,
				IsSymlink: info.Mode()&os.ModeSymlink != 0,
			}
			installations = append(installations, inst)
		}
	}

	return installations
}

// saveAtomic writes data to a file atomically using tmp + rename.
func saveAtomic(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	success := false
	defer func() {
		if !success {
			tmp.Close()
			os.Remove(tmpName)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	success = true
	return nil
}
