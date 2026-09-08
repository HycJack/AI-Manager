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
// directory.
func Default() *Config {
	home, _ := os.UserHomeDir()
	base := filepath.Join(home, ".ai-manager")
	return &Config{
		AppName:  "Wails3 Skeleton",
		DataDir:  base,
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

// Save writes the config to disk.
func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// EnsureDirs creates all configured directories.
func (c *Config) EnsureDirs() error {
	return os.MkdirAll(c.DataDir, 0o755)
}

// DefaultPath returns the standard on-disk location of the config file.
func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ai-manager", "config.json")
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
