package skill

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
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

// NewLibrary creates a Library. If dataDir is provided, uses it as the root
// (for testing). Otherwise defaults to ~/.aimanager/.
func NewLibrary(dataDir string) (*Library, error) {
	var libRoot string
	if dataDir != "" {
		libRoot = dataDir
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home dir: %w", err)
		}
		libRoot = filepath.Join(home, ".aimanager")
	}
	skillsDir := filepath.Join(libRoot, "skills")
	registryPath := filepath.Join(libRoot, "registry.json")

	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		return nil, fmt.Errorf("create skills dir: %w", err)
	}

	reg, err := config.LoadRegistry(registryPath)
	if err != nil {
		return nil, fmt.Errorf("load registry: %w", err)
	}

	return &Library{
		dataDir:      libRoot,
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
// any directory containing a SKILL.md file. The root itself may be a skill
// directory; if so, it is returned as a single record. Otherwise, the root is
// walked for subdirectories that contain SKILL.md. Returns the discovered
// SkillRecords, or an error if no skill directory is found.
func (l *Library) ScanLocal(root string) ([]Record, error) {
	root = filepath.Clean(root)
	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("scan local: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("scan local: %s is not a directory", root)
	}

	// If the root itself is a skill directory (contains SKILL.md), return it
	// as a single record. This handles the common case where the user picks
	// a single skill folder directly.
	if _, err := os.Stat(filepath.Join(root, "SKILL.md")); err == nil {
		return []Record{l.makeRecordFromDir(root)}, nil
	}

	// Root is not a skill; walk subdirectories for skill folders.
	var records []Record
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if !d.IsDir() || path == root {
			return nil
		}

		// Check if this directory has a SKILL.md
		if _, err := os.Stat(filepath.Join(path, "SKILL.md")); err != nil {
			return nil // not a skill directory
		}

		records = append(records, l.makeRecordFromDir(path))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %s: %w", root, err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("scan local: no SKILL.md found in %s", root)
	}

	return records, nil
}

// makeRecordFromDir builds a SkillRecord from a directory that contains a
// SKILL.md file. It tries to load metadata.json for additional fields.
func (l *Library) makeRecordFromDir(path string) Record {
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

	// Fallback: try to parse version from SKILL.md frontmatter.
	if rec.Version == "1.0.0" {
		if v := parseSkillFrontmatterVersion(path); v != "" {
			rec.Version = v
		}
	}

	return rec
}

// ScanNpx resolves an npx / skills.sh / GitHub package name to SkillRecords.
// The input can be:
//   - "owner/repo" — GitHub shorthand (clones and scans)
//   - "owner/repo/subdir" — GitHub with subdirectory (clones, scans subdir)
//   - "package-name" — treated as skills.sh lookup (returns metadata only)
//
// For GitHub inputs, clones the repo to a temp dir and scans for skills.
func (l *Library) ScanNpx(packageName string) ([]Record, error) {
	packageName = strings.TrimSpace(packageName)
	if packageName == "" {
		return nil, fmt.Errorf("scan npx: empty package name")
	}

	parts := strings.Split(packageName, "/")
	switch len(parts) {
	case 1:
		// Simple name — skills.sh lookup (metadata only, no download)
		return []Record{{
			ID:        hashString(packageName),
			Name:      titleCase(packageName),
			Slug:      slugify(packageName),
			Version:   "1.0.0",
			Origin:    Origin{Type: OriginSkillsSh, Path: packageName},
			Installed: false,
			UpdatedAt: time.Now().UTC(),
		}}, nil

	case 2, 3:
		// owner/repo or owner/repo/subdir — GitHub clone + scan
		repo := parts[0] + "/" + parts[1]
		subdir := ""
		if len(parts) == 3 {
			subdir = parts[2]
		}
		return l.cloneAndScanGitHub(repo, subdir)

	default:
		return nil, fmt.Errorf("scan npx: invalid package name %q (too many parts)", packageName)
	}
}

// cloneAndScanGitHub clones a GitHub repo to a temp dir and scans for skills.
func (l *Library) cloneAndScanGitHub(repo, subdir string) ([]Record, error) {
	tmpDir, err := os.MkdirTemp("", "ai-manager-clone-")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	cloneDir := filepath.Join(tmpDir, "repo")
	cmd := exec.Command("git", "clone", "--depth", "1",
		fmt.Sprintf("https://github.com/%s.git", repo), cloneDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("git clone %s: %w\n%s", repo, err, string(out))
	}

	// If a subdirectory is specified, scan that; otherwise scan the repo root
	scanPath := cloneDir
	if subdir != "" {
		scanPath = filepath.Join(cloneDir, subdir)
		if _, err := os.Stat(scanPath); err != nil {
			return nil, fmt.Errorf("subdirectory %q not found in %s", subdir, repo)
		}
	}

	records, err := l.ScanLocal(scanPath)
	if err != nil {
		return nil, fmt.Errorf("scan cloned repo %s: %w", repo, err)
	}

	// Tag each record with its GitHub origin
	for i := range records {
		records[i].Origin = Origin{Type: OriginGitHub, Repo: repo, Subdir: subdir}
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no skills found in %s", repo)
	}
	return records, nil
}

// ScanClaude resolves a Claude plugin path to SkillRecords.
// The input is a local directory path containing skill folders.
func (l *Library) ScanClaude(pluginPath string) ([]Record, error) {
	pluginPath = strings.TrimSpace(pluginPath)
	if pluginPath == "" {
		return nil, fmt.Errorf("scan claude: empty plugin path")
	}

	records, err := l.ScanLocal(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("scan claude plugin %s: %w", pluginPath, err)
	}

	for i := range records {
		records[i].Origin = Origin{Type: OriginClaude, Path: pluginPath}
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no skills found in Claude plugin path %s", pluginPath)
	}
	return records, nil
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

	// Copy skill files from source to library so the library holds the
	// complete skill (not just metadata.json). This makes the library the
	// single source of truth as described in CONTEXT.md.
	if rec.Origin.Path != "" {
		if srcInfo, err := os.Stat(rec.Origin.Path); err == nil && srcInfo.IsDir() {
			if err := copyDirContents(rec.Origin.Path, skillPath); err != nil {
				return fmt.Errorf("copy skill files: %w", err)
			}
		}
	}

	return l.Save()
}

// copyDirContents copies all files and subdirectories from src to dst,
// preserving the directory structure. It skips metadata.json at the root
// level (the library writes its own). Returns an error if any file cannot
// be copied.
func copyDirContents(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil // skip the root itself
		}
		// Skip metadata.json at the root level (library writes its own)
		if d.Type().IsRegular() && rel == "metadata.json" {
			return nil
		}
		dstPath := filepath.Join(dst, rel)
		if d.IsDir() {
			if err := os.MkdirAll(dstPath, 0o755); err != nil {
				return fmt.Errorf("create dir %s: %w", dstPath, err)
			}
			return nil
		}
		// Copy regular files
		info, err := d.Info()
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
			return fmt.Errorf("create parent dir: %w", err)
		}
		if err := os.WriteFile(dstPath, data, info.Mode()); err != nil {
			return fmt.Errorf("write %s: %w", dstPath, err)
		}
		return nil
	})
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
		// Try to load description from SKILL.md
		var desc string
		if entry.Path != "" {
			desc = ParseSkillFrontmatterDescription(entry.Path)
		}
		summaries = append(summaries, Summary{
			ID:        entry.ID,
			Name:      entry.Name,
			Slug:      entry.Slug,
			Version:   entry.Version,
			Description: desc,
			Installed: entry.Installed,
			UpdatedAt: time.Now().UTC(),
		})
	}
	return summaries
}

// RescanAndRegister scans the skills directory for unregistered skills and
// adds them to the registry. Also cleans up empty entries from previous runs.
func (l *Library) RescanAndRegister() (int, error) {
	// First, clean up empty entries (no slug or name)
	cleaned := l.registry.Skills[:0]
	for _, entry := range l.registry.Skills {
		if entry.Slug != "" && entry.Name != "" {
			cleaned = append(cleaned, entry)
		}
	}
	l.registry.Skills = cleaned

	entries, err := os.ReadDir(l.skillsDir)
	if err != nil {
		return 0, nil
	}

	added := 0
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "." || entry.Name() == ".." {
			continue
		}
		skillDir := filepath.Join(l.skillsDir, entry.Name())

		hasManifest := false
		for _, name := range []string{"SKILL.md", "README.md"} {
			if _, err := os.Stat(filepath.Join(skillDir, name)); err == nil {
				hasManifest = true
				break
			}
		}
		if !hasManifest {
			continue
		}

		// Check if already registered (by slug or name matching dir name)
		dirName := entry.Name()
		found := false
		for _, reg := range l.registry.Skills {
			if strings.EqualFold(reg.Slug, dirName) || strings.EqualFold(reg.Name, dirName) {
				found = true
				break
			}
		}
		if found {
			continue
		}

		// Load metadata if available, otherwise create a minimal record
		rec, _ := LoadMetadata(skillDir)
		if rec == nil || rec.Name == "" || rec.Slug == "" {
			rec = &Record{
				ID:      entry.Name(),
				Slug:    entry.Name(),
				Name:    entry.Name(),
				Version: "0.0.0",
			}
		}

		// Directly add to registry without writing metadata.json
		l.registry.Skills = append(l.registry.Skills, config.SkillEntry{
			ID:        rec.ID,
			Name:      rec.Name,
			Slug:      rec.Slug,
			Version:   rec.Version,
			Installed: true,
			Path:      skillDir,
		})
		added++
	}

	if added > 0 {
		_ = l.Save()
	}
	return added, nil
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
