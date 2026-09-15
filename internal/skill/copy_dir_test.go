package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyDirContents_DereferencesSymlinks(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	write := func(rel, content string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(filepath.Join(src, rel)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(src, rel), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("shared/guide.md", "# guide\n")
	write("root.md", "# root\n")

	// A skill repo reuses one references/ dir across several skills via links.
	if err := os.Symlink("shared", filepath.Join(src, "linked-dir")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("root.md", filepath.Join(src, "linked.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("does-not-exist", filepath.Join(src, "broken.md")); err != nil {
		t.Fatal(err)
	}
	// Cyclic link: must not recurse forever.
	if err := os.Symlink(".", filepath.Join(src, "cyclic")); err != nil {
		t.Fatal(err)
	}

	if err := copyDirContents(src, dst); err != nil {
		t.Fatalf("copyDirContents: %v", err)
	}

	for _, p := range []string{"root.md", "shared/guide.md", "linked-dir/guide.md", "linked.md"} {
		if _, err := os.Stat(filepath.Join(dst, p)); err != nil {
			t.Errorf("%s was not copied: %v", p, err)
		}
	}
	// Copied entries must be real files, not links: the library is self-contained.
	if info, err := os.Lstat(filepath.Join(dst, "linked-dir")); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			t.Error("linked-dir should be a real directory, not a symlink")
		}
	}
	if info, err := os.Lstat(filepath.Join(dst, "linked.md")); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			t.Error("linked.md should be a real file, not a symlink")
		}
	}
	// Broken links are skipped rather than failing the whole copy.
	if _, err := os.Stat(filepath.Join(dst, "broken.md")); err == nil {
		t.Error("broken.md should have been skipped")
	}
}
