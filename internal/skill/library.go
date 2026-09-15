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

// SkillSourceDir resolves the on-disk directory for a registered skill.
//
// Registry entries record an absolute path, which goes stale whenever the
// library root moves (for example from ~/Library/Application Support/AIManager
// /library to ~/.aimanager). When the recorded path no longer exists, fall back
// to this library's own skills directory so a moved library keeps working
// instead of failing every toggle.
func (l *Library) SkillSourceDir(entryPath, slug string) string {
	if entryPath != "" {
		if _, err := os.Stat(entryPath); err == nil {
			return entryPath
		}
	}
	return filepath.Join(l.skillsDir, slug)
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
// SKILL.md file.
func (l *Library) makeRecordFromDir(path string) Record {
	dirName := filepath.Base(path)
	slug := slugify(dirName)

	rec := Record{
		ID:        hashString(dirName),
		Name:      titleCase(dirName),
		Slug:      slug,
		Version:   "1.0.0",
		Origin:    Origin{Type: OriginLocal, Path: path},
		Installed: true,
		UpdatedAt: time.Now().UTC(),
	}

	// SKILL.md frontmatter holds the canonical name and description. The
	// directory name is only a fallback: deriving it with titleCase mangles
	// real names, turning "swiftui-pro" into "Swiftui Pro".
	fm := parseSkillFrontmatter(path)
	if n := fm["name"]; n != "" {
		rec.Name = n
	}
	if d := fm["description"]; d != "" {
		rec.Description = d
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

	// Version falls back to SKILL.md frontmatter when metadata.json has none.
	if rec.Version == "1.0.0" {
		if v := fm["version"]; v != "" {
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

// cloneRef clones cloneURL into a fresh temp dir named after repoName and
// returns that dir together with the path to scan. When subdir is set the
// checkout is validated and the subdir is returned instead.
//
// The caller owns the temp dir and must remove it with os.RemoveAll. Ownership
// is pushed to the caller because AddRecord has to copy the skill files out of
// the checkout: deleting it here would leave the install with metadata.json
// and nothing else.
func cloneRef(cloneURL, repoName, subdir string) (string, string, error) {
	repoName = strings.TrimSpace(repoName)
	if repoName == "" {
		return "", "", fmt.Errorf("empty repository")
	}

	tmpDir, err := os.MkdirTemp("", "ai-manager-clone-")
	if err != nil {
		return "", "", fmt.Errorf("create temp dir: %w", err)
	}
	// Name the checkout after the repository. A repo whose root is itself a
	// skill directory would otherwise be registered as a skill named "repo".
	cloneDir := filepath.Join(tmpDir, filepath.Base(repoName))

	cmd := exec.Command("git", "clone", "--depth", "1", cloneURL, cloneDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		os.RemoveAll(tmpDir)
		return "", "", fmt.Errorf("git clone %s: %w\n%s", repoName, err, string(out))
	}

	scanPath := cloneDir
	if subdir != "" {
		scanPath = filepath.Join(cloneDir, subdir)
		if _, err := os.Stat(scanPath); err != nil {
			os.RemoveAll(tmpDir)
			return "", "", fmt.Errorf("subdirectory %q not found in %s", subdir, repoName)
		}
	}
	return tmpDir, scanPath, nil
}

// cloneGitHubRepo clones a GitHub repository into a temp dir.
func cloneGitHubRepo(repo, subdir string) (string, string, error) {
	repo = strings.TrimSpace(repo)
	if repo == "" {
		return "", "", fmt.Errorf("empty repository")
	}
	return cloneRef("https://github.com/"+repo+".git", repo, subdir)
}

// cloneAndScanGitHub reports the skills a GitHub repo ships. This is a
// read-only probe: the checkout is discarded before returning, so the records
// must not be relied on for on-disk files. Use InstallFromRepo to install.
func (l *Library) cloneAndScanGitHub(repo, subdir string) ([]Record, error) {
	tmpDir, scanPath, err := cloneGitHubRepo(repo, subdir)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	records, err := l.ScanLocal(scanPath)
	if err != nil {
		return nil, fmt.Errorf("scan cloned repo %s: %w", repo, err)
	}

	// Tag each record with its GitHub origin, keeping Origin.Path so the
	// skill's real location is not lost.
	for i := range records {
		records[i].Origin.Type = OriginGitHub
		records[i].Origin.Repo = repo
		records[i].Origin.Subdir = subdir
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no skills found in %s", repo)
	}
	return records, nil
}

// InstallFromRepo clones a GitHub repository and installs its skills into the
// library, copying the skill files over. selectedSlugs is matched against
// record slugs; an empty slice installs everything. The checkout is deleted
// only after every AddRecord has copied its files out of it.
func (l *Library) InstallFromRepo(repo, groupName string, selectedSlugs []string) ([]Summary, error) {
	repo = strings.TrimSpace(repo)
	parts := strings.Split(repo, "/")
	if len(parts) < 2 || len(parts) > 3 {
		return nil, fmt.Errorf("expected owner/repo[/subdir], got %q", repo)
	}
	name := parts[0] + "/" + parts[1]
	subdir := ""
	if len(parts) == 3 {
		subdir = parts[2]
	}
	return l.installFromURL("https://github.com/"+name+".git", name, subdir, groupName, selectedSlugs)
}

// installFromURL does the clone-scan-copy work for an arbitrary git URL. It is
// split out of InstallFromRepo so the install path can be exercised with a
// local repository in tests.
func (l *Library) installFromURL(cloneURL, repo, subdir, groupName string, selectedSlugs []string) ([]Summary, error) {
	tmpDir, scanPath, err := cloneRef(cloneURL, repo, subdir)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	records, err := l.ScanLocal(scanPath)
	if err != nil {
		return nil, fmt.Errorf("scan cloned repo %s: %w", repo, err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("no skills found in %s", repo)
	}

	if len(selectedSlugs) > 0 {
		selected := make(map[string]bool, len(selectedSlugs))
		for _, slug := range selectedSlugs {
			selected[strings.TrimSpace(slug)] = true
		}
		filtered := make([]Record, 0, len(records))
		for _, rec := range records {
			if selected[rec.Slug] {
				filtered = append(filtered, rec)
			}
		}
		records = filtered
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("no matching skills to install from %s", repo)
	}

	installed := make([]Summary, 0, len(records))
	for i := range records {
		rec := records[i]
		rec.Origin.Type = OriginGitHub
		rec.Origin.Repo = repo
		rec.Origin.Subdir = subdir
		rec.Group = groupName
		if err := l.AddRecord(rec); err != nil {
			if strings.Contains(err.Error(), "already exists in library") {
				continue // keep going: one duplicate should not fail the batch
			}
			return nil, err
		}
		sum := rec.NewSummary()
		sum.Installed = true
		installed = append(installed, sum)
	}
	return installed, nil
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
	return copyDirContentsDepth(src, dst, 0)
}

// maxSymlinkDepth bounds how deep symlinked directories are followed, so a
// cyclic link cannot recurse forever.
const maxSymlinkDepth = 16

func copyDirContentsDepth(src, dst string, depth int) error {
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

		// Dereference symlinks so the library stays self-contained: a symlinked
		// directory is copied as a real directory, a symlinked file as a real
		// file. Copying the link itself would leave the installed skill pointing
		// outside the library. Skill repos use symlinks a lot (a shared
		// references/ reused by several skills), so failing on one would break
		// the whole install.
		if d.Type()&os.ModeSymlink != 0 {
			if depth >= maxSymlinkDepth {
				return nil // cyclic link: stop rather than recurse forever
			}
			target, err := os.Stat(path) // follows symlinks
			if err != nil {
				return nil // broken link: skip it
			}
			if target.IsDir() {
				if err := os.MkdirAll(dstPath, 0o755); err != nil {
					return fmt.Errorf("create dir %s: %w", dstPath, err)
				}
				// Resolve the link before walking: WalkDir reports paths
				// relative to the resolved root, so walking the link itself
				// would make filepath.Rel escape the destination.
				real, err := filepath.EvalSymlinks(path)
				if err != nil {
					return nil
				}
				// WalkDir does not descend into symlinked directories.
				return copyDirContentsDepth(real, dstPath, depth+1)
			}
			return copyFileContents(path, dstPath, target.Mode())
		}

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
		return copyFileContents(path, dstPath, info.Mode())
	})
}

func copyFileContents(srcPath, dstPath string, mode os.FileMode) error {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", srcPath, err)
	}
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}
	if err := os.WriteFile(dstPath, data, mode); err != nil {
		return fmt.Errorf("write %s: %w", dstPath, err)
	}
	return nil
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
	repaired := 0
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
		for i := range l.registry.Skills {
			reg := &l.registry.Skills[i]
			if !strings.EqualFold(reg.Slug, dirName) && !strings.EqualFold(reg.Name, dirName) {
				continue
			}
			found = true
			// Entries keep an absolute path, which goes stale when the library
			// root moves. Repair any that no longer exist (an empty path
			// counts too: os.Stat("") fails) so the stale data heals itself
			// instead of accumulating.
			if reg.Path != skillDir {
				if _, err := os.Stat(reg.Path); err != nil {
					reg.Path = skillDir
					repaired++
				}
			}
			break
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

	if added > 0 || repaired > 0 {
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
