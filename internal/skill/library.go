package skill

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ai-manager/internal/config"
)

// Library manages the skill library: the collection of skills on this machine,
// persisted as a registry.json file with skill metadata and group definitions.
type Library struct {
	dataDir    string
	skillsDir  string
	registryPath string
	registry   *config.Registry
}

// NewLibrary creates a Library rooted at the given data directory. The registry
// is loaded from disk if it exists; otherwise an empty registry is created.
func NewLibrary(dataDir string) (*Library, error) {
	skillsDir := filepath.Join(dataDir, "library", "skills")
	registryPath := filepath.Join(dataDir, "library", "registry.json")

	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		return nil, fmt.Errorf("create skills dir: %w", err)
	}

	reg, err := config.LoadRegistry(registryPath)
	if err != nil {
		return nil, fmt.Errorf("load registry: %w", err)
	}

	return &Library{
		dataDir:      dataDir,
		skillsDir:    skillsDir,
		registryPath: registryPath,
		registry:     reg,
	}, nil
}

// Save persists the current registry to disk atomically.
func (l *Library) Save() error {
	return l.registry.SaveRegistry(l.registryPath)
}

// Registry returns the current registry (for read access by callers).
func (l *Library) Registry() *config.Registry {
	return l.registry
}

// SkillDir returns the base directory where skill folders live.
func (l *Library) SkillDir() string {
	return l.skillsDir
}

// ScanLocal scans a local folder for skill directories. A skill directory is
// any subdirectory containing a SKILL.md or README.md file. Returns the
// discovered SkillRecords.
func (l *Library) ScanLocal(root string) ([]Record, error) {
	root = filepath.Clean(root)
	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("scan local: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("scan local: %s is not a directory", root)
	}

	var records []Record
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if !d.IsDir() || path == root {
			return nil
		}

		// Check if this directory has a SKILL.md or README.md
		hasReadme := false
		for _, name := range []string{"SKILL.md", "README.md"} {
			if _, err := os.Stat(filepath.Join(path, name)); err == nil {
				hasReadme = true
				break
			}
		}
		if !hasReadme {
			return nil // not a skill directory
		}

		name := filepath.Base(path)
		slug := slugify(name)

		rec := Record{
			ID:        hashString(name),
			Name:      titleCase(name),
			Slug:      slug,
			Version:   "1.0.0",
			Origin:    Origin{Type: OriginLocal, Path: path},
			Installed: true,
			UpdatedAt: time.Now().UTC(),
		}

		// Try to load metadata.json if it exists
		if meta, err := LoadMetadata(path); err == nil {
			rec.Authors = meta.Authors
			rec.License = meta.License
			rec.Keywords = meta.Keywords
			rec.Tags = meta.Tags
			if meta.Version != "" {
				rec.Version = meta.Version
			}
			if meta.ID != "" {
				rec.ID = meta.ID
			}
		}

		records = append(records, rec)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %s: %w", root, err)
	}
	return records, nil
}

// ScanNpx resolves an npx / skills.sh / GitHub package name to a SkillRecord.
// The input can be:
//   - "owner/repo" — GitHub shorthand (uses raw GitHub content)
//   - "owner/repo/subdir" — GitHub with subdirectory
//   - "package-name" — treated as skills.sh lookup
//
// Returns a SkillRecord with the resolved origin. The actual content download
// is handled by the caller (AddSkill copies the files into the library).
func (l *Library) ScanNpx(packageName string) (*Record, error) {
	packageName = strings.TrimSpace(packageName)
	if packageName == "" {
		return nil, fmt.Errorf("scan npx: empty package name")
	}

	parts := strings.Split(packageName, "/")
	switch len(parts) {
	case 1:
		// Simple name — treat as skills.sh lookup
		return &Record{
			ID:    hashString(packageName),
			Name:  titleCase(packageName),
			Slug:  slugify(packageName),
			Version: "1.0.0",
			Origin: Origin{Type: OriginSkillsSh, Path: packageName},
			Installed: false,
			UpdatedAt: time.Now().UTC(),
		}, nil
	case 2:
		// owner/repo — GitHub shorthand
		return &Record{
			ID:    hashString(packageName),
			Name:  titleCase(parts[1]),
			Slug:  slugify(parts[1]),
			Version: "1.0.0",
			Origin: Origin{Type: OriginGitHub, Repo: packageName},
			Installed: false,
			UpdatedAt: time.Now().UTC(),
		}, nil
	case 3:
		// owner/repo/subdir — GitHub with subdirectory
		repo := parts[0] + "/" + parts[1]
		return &Record{
			ID:    hashString(packageName),
			Name:  titleCase(parts[2]),
			Slug:  slugify(parts[2]),
			Version: "1.0.0",
			Origin: Origin{Type: OriginGitHub, Repo: repo, Subdir: parts[2]},
			Installed: false,
			UpdatedAt: time.Now().UTC(),
		}, nil
	default:
		return nil, fmt.Errorf("scan npx: invalid package name %q (too many parts)", packageName)
	}
}

// ScanClaude resolves a Claude plugin name to a SkillRecord.
// The input is a plugin identifier (e.g. "superpowers" or "owner/plugin").
func (l *Library) ScanClaude(plugin string) (*Record, error) {
	plugin = strings.TrimSpace(plugin)
	if plugin == "" {
		return nil, fmt.Errorf("scan claude: empty plugin name")
	}

	name := plugin
	parts := strings.Split(plugin, "/")
	if len(parts) == 2 {
		name = parts[1]
	}

	return &Record{
		ID:    hashString(plugin),
		Name:  titleCase(name),
		Slug:  slugify(name),
		Version: "1.0.0",
		Origin: Origin{Type: OriginClaude, Path: plugin},
		Installed: false,
		UpdatedAt: time.Now().UTC(),
	}, nil
}

// ScanExisting adopts existing skill installations by scanning the given paths
// for SKILL.md/README.md files. Returns the discovered SkillRecords.
func (l *Library) ScanExisting(paths []string) ([]Record, error) {
	var all []Record
	for _, p := range paths {
		records, err := l.ScanLocal(p)
		if err != nil {
			// Skip paths that can't be scanned; collect the error for the caller.
			continue
		}
		all = append(all, records...)
	}
	if len(all) == 0 {
		return nil, fmt.Errorf("scan existing: no skills found in provided paths")
	}
	return all, nil
}

// AddRecord adds a skill record to the library and saves the registry.
func (l *Library) AddRecord(rec Record) error {
	// Check for duplicates
	for _, entry := range l.registry.Skills {
		if entry.ID == rec.ID {
			return fmt.Errorf("skill %q already exists in library", rec.Name)
		}
	}

	l.registry.Skills = append(l.registry.Skills, config.SkillEntry{
		ID:        rec.ID,
		Name:      rec.Name,
		Slug:      rec.Slug,
		Version:   rec.Version,
		Installed: rec.Installed,
		Path:      filepath.Join(l.skillsDir, rec.Slug),
	})

	// Write metadata.json into the skill directory
	skillPath := filepath.Join(l.skillsDir, rec.Slug)
	if err := os.MkdirAll(skillPath, 0o755); err != nil {
		return fmt.Errorf("create skill dir: %w", err)
	}
	metaPath := filepath.Join(skillPath, "metadata.json")
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}
	if err := config.SaveAtomic(metaPath, data); err != nil {
		return fmt.Errorf("write metadata: %w", err)
	}

	return l.Save()
}

// RemoveRecord removes a skill record from the library by name.
func (l *Library) RemoveRecord(name string) error {
	found := false
	newSkills := make([]config.SkillEntry, 0, len(l.registry.Skills))
	for _, entry := range l.registry.Skills {
		if strings.EqualFold(entry.Name, name) || strings.EqualFold(entry.Slug, name) {
			found = true
			// Remove the skill directory
			skillPath := filepath.Join(l.skillsDir, entry.Slug)
			os.RemoveAll(skillPath)
			continue
		}
		newSkills = append(newSkills, entry)
	}

	if !found {
		return fmt.Errorf("skill %q not found in library", name)
	}

	l.registry.Skills = newSkills
	return l.Save()
}

// FindSkill looks up a skill by name or slug.
func (l *Library) FindSkill(name string) (config.SkillEntry, bool) {
	for _, entry := range l.registry.Skills {
		if strings.EqualFold(entry.Name, name) || strings.EqualFold(entry.Slug, name) {
			return entry, true
		}
	}
	return config.SkillEntry{}, false
}

// ListSkills returns all skills in the library as Summaries.
func (l *Library) ListSkills() []Summary {
	summaries := make([]Summary, 0, len(l.registry.Skills))
	for _, entry := range l.registry.Skills {
		summaries = append(summaries, Summary{
			ID:        entry.ID,
			Name:      entry.Name,
			Slug:      entry.Slug,
			Version:   entry.Version,
			Installed: entry.Installed,
			UpdatedAt: time.Now().UTC(),
		})
	}
	return summaries
}

// LoadSkillRecord loads the full Record from the metadata.json of a skill.
func (l *Library) LoadSkillRecord(slug string) (*Record, error) {
	skillPath := filepath.Join(l.skillsDir, slug)
	return LoadMetadata(skillPath)
}

// ListSkillFiles returns the file paths (relative) of a skill directory.
func (l *Library) ListSkillFiles(slug string) ([]string, error) {
	skillPath := filepath.Join(l.skillsDir, slug)
	var files []string
	err := filepath.WalkDir(skillPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(skillPath, path)
		if err != nil {
			rel = path
		}
		files = append(files, rel)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// LoadSkillReadme returns the readme content of a skill.
func (l *Library) LoadSkillReadme(slug string) (string, error) {
	skillPath := filepath.Join(l.skillsDir, slug)
	return LoadReadme(skillPath)
}

// slugify converts a name to a URL-safe slug.
func slugify(name string) string {
	name = strings.ToLower(name)
	var result strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			result.WriteRune(r)
		case r >= '0' && r <= '9':
			result.WriteRune(r)
		case r == '-':
			result.WriteRune(r)
		case r == '_' || r == ' ' || r == '/' || r == '\\':
			result.WriteRune('-')
		default:
			// Skip other characters
		}
	}
	return strings.Trim(result.String(), "-")
}

// titleCase converts a slug or name to Title Case.
func titleCase(name string) string {
	words := strings.FieldsFunc(name, func(r rune) bool {
		return r == '-' || r == '_' || r == ' ' || r == '/' || r == '.'
	})
	var result strings.Builder
	for _, w := range words {
		if result.Len() > 0 {
			result.WriteByte(' ')
		}
		if len(w) > 0 {
			result.WriteString(strings.ToUpper(w[:1]) + w[1:])
		}
	}
	if result.Len() == 0 {
		return name
	}
	return result.String()
}

// hashString creates a deterministic hash-based ID from a string.
func hashString(s string) string {
	h := sha1.Sum([]byte(s))
	return hex.EncodeToString(h[:])[:12]
}
