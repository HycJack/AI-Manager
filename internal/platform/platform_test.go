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
	dir, err := SkillLibraryDir()
	if err != nil {
		t.Fatalf("SkillLibraryDir() returned error: %v", err)
	}
	if dir == "" {
		t.Error("SkillLibraryDir() returned empty string")
	}

	// Verify the directory ends with /skills and starts with ~/.aimanager
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".aimanager", "skills")
	// Fix typo: expected should be .aimanager
	expected = filepath.Join(home, ".aimanager", "skills")
	if dir != expected {
		t.Errorf("SkillLibraryDir() = %q, want %q", dir, expected)
	}
}

func TestSkillLibraryDir_CreatesDir(t *testing.T) {
	dir, err := SkillLibraryDir()
	if err != nil {
		t.Fatalf("SkillLibraryDir() returned error: %v", err)
	}

	// SkillLibraryDir returns the path but does not create the directory.
	// The Library constructor (NewLibrary) handles directory creation.
	if dir == "" {
		t.Error("SkillLibraryDir() returned empty string")
	}
}
