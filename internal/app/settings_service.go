package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Preferences is the typed shape persisted to disk and exposed to the frontend
// via the Wails3 bindings. Adding a field here + in DEFAULT_PREFERENCES is the
// only change needed to introduce a new setting — the frontend gets a typed
// model automatically after `wails3 generate bindings`.
type Preferences struct {
	// Light/dark/system. "system" follows the OS preference.
	Theme string `json:"theme"`
	// Id of the active theme (see modules/theme/themes).
	ThemeID string `json:"themeId"`
	// UI zoom level, e.g. 1.0 === 100%.
	ZoomLevel float64 `json:"zoomLevel"`
	// Show dot-prefixed files in the (future) file explorer.
	ShowHidden bool `json:"showHidden"`
	// Open automatically at login.
	LaunchAtLogin bool `json:"launchAtLogin"`
	// UI language: "en" or "zh".
	Language string `json:"language"`
}

// DefaultPreferences is the zero-value preferences used when no file exists yet.
var DefaultPreferences = Preferences{
	Theme:         "system",
	ThemeID:       "claude",
	ZoomLevel:     1,
	ShowHidden:    false,
	LaunchAtLogin: false,
	Language:      "en",
}

// SettingsService persists the frontend's preferences as a single JSON file on
// disk. The frontend keeps the authoritative settings object (theme mode,
// theme id, zoom level, …) in a zustand store and calls SavePrefs whenever it
// changes (debounced), then LoadPrefs on startup. The backend is just dumb
// JSON persistence — no domain knowledge — so adding a new setting never
// requires touching Go beyond the Preferences struct above.
type SettingsService struct {
	state *State
	mu    sync.Mutex
}

// NewSettingsService creates the settings service.
func NewSettingsService(s *State) *SettingsService { return &SettingsService{state: s} }

// prefsPath is the on-disk location of the preferences file.
func (s *SettingsService) prefsPath() string {
	return filepath.Join(s.state.cfg.DataDir, "preferences.json")
}

// LoadPrefs returns the saved preferences, falling back to defaults if the
// file does not exist yet.
func (s *SettingsService) LoadPrefs() (Preferences, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	prefs := DefaultPreferences
	data, err := os.ReadFile(s.prefsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return prefs, nil
		}
		return prefs, err
	}
	// Unmarshal on top of the defaults so missing fields keep their default
	// values (forward-compatible with older preference files).
	if err := json.Unmarshal(data, &prefs); err != nil {
		return prefs, err
	}
	return prefs, nil
}

// SavePrefs atomically writes the preferences to disk.
func (s *SettingsService) SavePrefs(prefs Preferences) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	buf, err := json.MarshalIndent(prefs, "", "  ")
	if err != nil {
		return err
	}
	p := s.prefsPath()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	// Atomic write via temp file so a crash mid-save never corrupts prefs.
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, buf, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}
