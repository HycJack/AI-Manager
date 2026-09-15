package skill

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const frontmatterSkill = "---\nname: swiftui-pro\ndescription: Reviews SwiftUI code\n---\n# swiftui-pro\n"

// newGitRepo initialises a local git repository with the given files so the
// clone/install path can be exercised without network access.
func newGitRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-q")
	run("add", "-A")
	run("commit", "-q", "-m", "init")
	return root
}

func TestMakeRecordFromDir_UsesFrontmatterName(t *testing.T) {
	lib, err := NewLibrary(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	dir := filepath.Join(t.TempDir(), "swiftui-pro")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(frontmatterSkill), 0o644); err != nil {
		t.Fatal(err)
	}

	rec := lib.makeRecordFromDir(dir)

	if rec.Name != "swiftui-pro" {
		t.Errorf("Name = %q, want %q (frontmatter name, not titleCase of the directory)", rec.Name, "swiftui-pro")
	}
	if rec.Description != "Reviews SwiftUI code" {
		t.Errorf("Description = %q, want the frontmatter description", rec.Description)
	}
	if rec.Slug != "swiftui-pro" {
		t.Errorf("Slug = %q, want %q", rec.Slug, "swiftui-pro")
	}
}

func TestMakeRecordFromDir_FallsBackToDirectoryName(t *testing.T) {
	lib, err := NewLibrary(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	dir := filepath.Join(t.TempDir(), "plain-skill")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// No frontmatter at all.
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# Plain Skill\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	rec := lib.makeRecordFromDir(dir)

	if rec.Name != "Plain Skill" {
		t.Errorf("Name = %q, want %q (titleCase fallback)", rec.Name, "Plain Skill")
	}
	if rec.Description != "" {
		t.Errorf("Description = %q, want empty", rec.Description)
	}
}

// TestInstallFromURL_CopiesSkillFiles is the regression test for the
// metadata.json-only install: the checkout has to outlive the scan so AddRecord
// can copy the skill files out of it, and tagging the GitHub origin must not
// wipe Origin.Path.
func TestInstallFromURL_CopiesSkillFiles(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not available")
	}

	lib, err := NewLibrary(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	src := newGitRepo(t, map[string]string{
		"SKILL.md":            frontmatterSkill,
		"references/guide.md": "# guide\n",
	})

	installed, err := lib.installFromURL(src, "swiftui-pro", "", "my-group", nil)
	if err != nil {
		t.Fatalf("installFromURL: %v", err)
	}
	if len(installed) != 1 {
		t.Fatalf("installed %d skills, want 1", len(installed))
	}

	skillPath := filepath.Join(lib.SkillDir(), "swiftui-pro")
	for _, want := range []string{"SKILL.md", "references/guide.md", "metadata.json"} {
		if _, err := os.Stat(filepath.Join(skillPath, want)); err != nil {
			t.Errorf("%s missing from the library copy: %v", want, err)
		}
	}

	entry, ok := lib.FindSkill("swiftui-pro")
	if !ok {
		t.Fatal("skill was not registered")
	}
	if entry.Name != "swiftui-pro" {
		t.Errorf("registered Name = %q, want %q", entry.Name, "swiftui-pro")
	}
	if entry.Path != skillPath {
		t.Errorf("registry Path = %q, want %q", entry.Path, skillPath)
	}
}

func TestInstallFromURL_SelectedSlugs(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not available")
	}

	lib, err := NewLibrary(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	src := newGitRepo(t, map[string]string{
		"a-skill/SKILL.md": "---\nname: a-skill\n---\n# A\n",
		"b-skill/SKILL.md": "---\nname: b-skill\n---\n# B\n",
	})

	installed, err := lib.installFromURL(src, "two-skills", "", "", []string{"b-skill"})
	if err != nil {
		t.Fatalf("installFromURL: %v", err)
	}
	if len(installed) != 1 {
		t.Fatalf("installed %d skills, want 1", len(installed))
	}
	if installed[0].Slug != "b-skill" {
		t.Errorf("Slug = %q, want %q", installed[0].Slug, "b-skill")
	}
	if _, err := os.Stat(filepath.Join(lib.SkillDir(), "a-skill")); err == nil {
		t.Error("a-skill should not have been installed")
	}

	if _, err := lib.installFromURL(src, "two-skills", "", "", []string{"nope"}); err == nil {
		t.Error("expected an error when no selected slug matches")
	}
}

func TestInstallFromRepo_InvalidRef(t *testing.T) {
	lib, err := NewLibrary(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, ref := range []string{"", "onlyowner", "a/b/c/d"} {
		if _, err := lib.InstallFromRepo(ref, "", nil); err == nil {
			t.Errorf("InstallFromRepo(%q): expected an error", ref)
		}
	}
}
