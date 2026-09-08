package app

import (
	"os"
	"path/filepath"
	"sync"

	"ai-manager/internal/agents"
	"ai-manager/internal/effective"
	"ai-manager/internal/platform"
	"ai-manager/internal/project"
)

// ProjectService manages projects, skill installations, and effective skill
// calculation. It is exposed to the frontend via Wails3 bindings.
type ProjectService struct {
	state   *State
	manager *project.Manager
	mu      sync.Mutex
}

// NewProjectService creates the project service.
func NewProjectService(s *State) *ProjectService {
	return &ProjectService{
		state:   s,
		manager: project.NewManager(),
	}
}

// InstallSkills installs multiple skills into a project's target directories.
// If global is true, installs into shared target only; otherwise installs into
// all specified targets.
func (s *ProjectService) InstallSkills(names []string, projectPath string, targets []agents.InstallTarget, global bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(names) == 0 {
		return nil
	}
	if projectPath == "" {
		return os.ErrInvalid
	}

	// Determine the library skill directory
	libDir, err := s.libraryDir()
	if err != nil {
		return err
	}

	if global {
		targets = []agents.InstallTarget{agents.TargetUniversal}
	}

	for _, name := range names {
		skillPath := filepath.Join(libDir, name)
		// Verify the skill exists in the library
		if _, err := os.Stat(skillPath); err != nil {
			if os.IsNotExist(err) {
				continue // Skip missing skills
			}
			return err
		}

		for _, target := range targets {
			if err := project.InstallSkill(skillPath, projectPath, target); err != nil {
				return err
			}
		}
	}

	return nil
}

// UninstallSkill removes a skill from a specific target in a project.
func (s *ProjectService) UninstallSkill(projectPath, name string, target agents.InstallTarget) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if projectPath == "" || name == "" {
		return os.ErrInvalid
	}
	return project.UninstallSkill(projectPath, name, target)
}

// ListInstallations returns all skill installations in a project.
func (s *ProjectService) ListInstallations(projectPath string) ([]project.Installation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if projectPath == "" {
		return nil, os.ErrInvalid
	}
	return project.ScanProject(projectPath), nil
}

// ListProjects returns all registered projects.
func (s *ProjectService) ListProjects() ([]project.Project, error) {
	return s.manager.LoadProjects()
}

// AddProject registers a project directory.
func (s *ProjectService) AddProject(path string) error {
	if path == "" {
		return os.ErrInvalid
	}
	return s.manager.AddProject(path)
}

// RemoveProject removes a project by path.
func (s *ProjectService) RemoveProject(path string) error {
	if path == "" {
		return os.ErrInvalid
	}
	return s.manager.RemoveProject(path)
}

// BrowseProject opens a folder picker dialog and returns the selected path.
// Uses the Wails v3 Dialog API to let the user choose a directory.
func (s *ProjectService) BrowseProject() (string, error) {
	if s.state == nil || s.state.app == nil {
		return "", os.ErrInvalid
	}

	// Use Wails v3 Dialog API for folder selection
	dialog := s.state.app.Dialog.OpenFile().
		SetTitle("Select Project Directory").
		CanChooseDirectories(true).
		CanChooseFiles(false).
		CanCreateDirectories(true)

	path, err := dialog.PromptForSingleSelection()
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", os.ErrNotExist
	}
	return path, nil
}

// GetEffectiveSkills returns all skills visible in a project, deduplicated
// across agents by canonical (resolved symlink) path. Mirrors Kitter:
// scans all agent directories, resolves symlinks, deduplicates by source.
func (s *ProjectService) GetEffectiveSkills(projectPath string) ([]effective.EffectiveSkill, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if projectPath == "" {
		return nil, os.ErrInvalid
	}
	return effective.GetEffectiveSkills(projectPath), nil
}

// libraryDir returns the skill library directory from the config.
func (s *ProjectService) libraryDir() (string, error) {
	dataDir, err := platform.DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "library", "skills"), nil
}

// EnsureProjectsFileExists creates an empty projects.json if it doesn't exist.
// This ensures the frontend always has a valid file to read.
func (s *ProjectService) EnsureProjectsFileExists() {
	path := filepath.Join(s.state.cfg.DataDir, "projects.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		data := `{"projects":[]}`
		_ = os.MkdirAll(filepath.Dir(path), 0o755)
		_ = os.WriteFile(path, []byte(data), 0o644)
	}
}
