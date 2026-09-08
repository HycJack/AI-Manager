// Package config provides a small JSON-file-backed application configuration.
//
// It is intentionally generic: add new fields to Config as your app grows, and
// they will be persisted automatically by Save/Load. This is the same pattern
// the services use to store user-adjustable settings.
package config

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"ai-manager/internal/platform"
)

// Config holds all user-adjustable settings for the application.
//
// JSON tags keep the on-disk file stable across renames/refactors.
type Config struct {
	// AppName is shown in the window title / tray.
	AppName string `json:"appName"`
	// DataDir is where the application stores its runtime data. Defaults to
	// ~/.<app slug>/ on first run.
	DataDir string `json:"dataDir"`
	// Proxy is an optional HTTP(S) proxy URL (e.g. "http://127.0.0.1:7890")
	// used for all outbound HTTP requests.
	Proxy string `json:"proxy"`
	// LogLevel controls verbosity of the file logger: debug/info/warn/error.
	LogLevel string `json:"logLevel"`
	// UpdateRepo is the GitHub "owner/repo" used by the auto-updater.
	// Leave empty to disable update checks.
	UpdateRepo string `json:"updateRepo"`
}

// Default returns the default configuration rooted in the user's home
// directory, using the platform-appropriate data directory.
func Default() *Config {
	dataDir, _ := platform.DataDir()
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = filepath.Join(home, ".ai-manager")
	}
	return &Config{
		AppName:  "AI-Manager",
		DataDir:  dataDir,
		LogLevel: "info",
		// Empty by default: auto-update is opt-in until the user sets a real
		// "owner/repo" in config.json. Prevents the app from hitting GitHub
		// for a placeholder repo that doesn't exist.
		UpdateRepo: "",
	}
}

// Load reads the config file, falling back to defaults if it does not exist.
func Load(path string) (*Config, error) {
	c := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Save writes the config to disk using atomic write (tmp + rename).
func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return SaveAtomic(path, data)
}

// SaveAtomic writes data to a file atomically using tmp + rename.
// This ensures the file is never in a partially-written state.
func SaveAtomic(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	// Write to a temporary file in the same directory
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	// Ensure cleanup on error
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

	// Atomic rename
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	success = true
	return nil
}

// EnsureDirs creates all configured directories.
func (c *Config) EnsureDirs() error {
	return os.MkdirAll(c.DataDir, 0o755)
}

// DefaultPath returns the standard on-disk location of the config file.
func DefaultPath() string {
	dataDir, _ := platform.DataDir()
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = filepath.Join(home, ".ai-manager")
	}
	return filepath.Join(dataDir, "config.json")
}

// Registry holds the skill library metadata persisted to registry.json.
type Registry struct {
	Skills []SkillEntry `json:"skills"`
	Groups []GroupEntry `json:"groups"`
}

// SkillEntry is a minimal entry in the registry (full details in metadata.json).
type SkillEntry struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Version   string `json:"version"`
	Installed bool   `json:"installed"`
	Path      string `json:"path"`
}

// GroupEntry is a group definition in the registry.
type GroupEntry struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Skills []string `json:"skills"`
}

// LoadRegistry reads the registry file, returning an empty registry if not found.
func LoadRegistry(path string) (*Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Registry{}, nil
		}
		return nil, err
	}
	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, err
	}
	return &reg, nil
}

// SaveRegistry writes the registry to disk atomically.
func (r *Registry) SaveRegistry(path string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return SaveAtomic(path, data)
}

// HTTPClient builds an *http.Client that routes through the configured proxy
// when proxy is non-empty, otherwise falls back to the environment's proxy
// settings. A zero timeout means no timeout.
func HTTPClient(proxy string, timeout time.Duration) *http.Client {
	transport := &http.Transport{Proxy: http.ProxyFromEnvironment}
	if proxy != "" {
		if u, err := url.Parse(proxy); err == nil {
			transport.Proxy = http.ProxyURL(u)
		}
	}
	return &http.Client{Transport: transport, Timeout: timeout}
}
