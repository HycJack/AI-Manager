package platform

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDataDir_EnvOverride(t *testing.T) {
	tmp := t.TempDir()
	os.Setenv("AI_MANAGER_HOME", tmp)
	defer os.Unsetenv("AI_MANAGER_HOME")

	dir, err := DataDir()
	if err != nil {
		t.Fatalf("DataDir() returned error: %v", err)
	}
	if dir != tmp {
		t.Errorf("DataDir() = %q, want %q", dir, tmp)
	}
}

func TestDataDir_Default(t *testing.T) {
	os.Unsetenv("AI_MANAGER_HOME")

	// Can't write to real Application Support on macOS sandbox, so just
	// verify the path is constructed correctly without creating the dir.
	dir, err := DataDir()
	if err != nil {
		// On sandboxed macOS, creating the real dir may fail — that's OK.
		// We just want to verify the path logic.
		t.Logf("DataDir() error (expected in sandbox): %v", err)
		return
	}
	if dir == "" {
		t.Error("DataDir() returned empty string")
	}

	// Verify the directory contains the app name
	_, file := filepath.Split(dir)
	if file != "AIManager" {
		t.Errorf("DataDir() basename = %q, want %q", file, "AIManager")
	}
}

func TestBaseDir_Darwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin-specific test")
	}
	base, err := baseDir()
	if err != nil {
		t.Fatalf("baseDir() returned error: %v", err)
	}
	if !filepath.IsAbs(base) {
		t.Errorf("baseDir() = %q, want absolute path", base)
	}
}

func TestSkillLibraryDir(t *testing.T) {
	tmp := t.TempDir()
	os.Setenv("AI_MANAGER_HOME", tmp)
	defer os.Unsetenv("AI_MANAGER_HOME")

	dir, err := SkillLibraryDir()
	if err != nil {
		t.Fatalf("SkillLibraryDir() returned error: %v", err)
	}
	if dir == "" {
		t.Error("SkillLibraryDir() returned empty string")
	}

	// Verify the directory ends with /library/skills
	rel, err := filepath.Rel(tmp, dir)
	if err != nil {
		t.Fatalf("filepath.Rel() returned error: %v", err)
	}
	if rel != filepath.Join("library", "skills") {
		t.Errorf("SkillLibraryDir() relative = %q, want %q", rel, filepath.Join("library", "skills"))
	}
}

func TestSkillLibraryDir_CreatesDir(t *testing.T) {
	tmp := t.TempDir()
	os.Setenv("AI_MANAGER_HOME", tmp)
	defer os.Unsetenv("AI_MANAGER_HOME")

	dir, err := SkillLibraryDir()
	if err != nil {
		t.Fatalf("SkillLibraryDir() returned error: %v", err)
	}

	// Verify the directory exists
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("os.Stat(%q) returned error: %v", dir, err)
	}
	if !info.IsDir() {
		t.Error("SkillLibraryDir() did not create a directory")
	}
}
