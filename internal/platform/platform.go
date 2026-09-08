// Package platform provides cross-platform path resolution for the
// AI-Manager application data directory.
//
// The base path is resolved as follows:
//  1. If the AI_MANAGER_HOME environment variable is set, use it.
//  2. Otherwise, use the OS-specific default:
//     - macOS: ~/Library/Application Support/AIManager
//     - Windows: %APPDATA%/AIManager
//     - Linux: $XDG_DATA_HOME/AIManager or ~/.local/share/AIManager
package platform

import (
	"os"
	"path/filepath"
	"runtime"
)

const appName = "AIManager"

// DataDir returns the application data directory, creating it if needed.
// The order of precedence is:
//  1. AI_MANAGER_HOME environment variable
//  2. OS-specific default location
func DataDir() (string, error) {
	if home := os.Getenv("AI_MANAGER_HOME"); home != "" {
		if err := os.MkdirAll(home, 0o755); err != nil {
			return "", err
		}
		return home, nil
	}

	base, err := baseDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(base, appName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// baseDir returns the OS-specific base directory for application data.
func baseDir() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Application Support"), nil
	case "windows":
		appdata := os.Getenv("APPDATA")
		if appdata == "" {
			return "", os.ErrNotExist
		}
		return appdata, nil
	case "linux":
		xdg := os.Getenv("XDG_DATA_HOME")
		if xdg != "" {
			return xdg, nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".local", "share"), nil
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".ai-manager"), nil
	}
}

// SkillLibraryDir returns the directory where the skill library is stored.
func SkillLibraryDir() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	skillDir := filepath.Join(dir, "library", "skills")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		return "", err
	}
	return skillDir, nil
}
