package app

import (
	"os"
	"path/filepath"
	"testing"

	"ai-manager/internal/config"
)

// newTestState builds a State whose config points at a temp data dir.
func newTestState(t *testing.T) *State {
	t.Helper()
	dir := t.TempDir()
	cfg := &config.Config{AppName: "test", DataDir: dir}
	return New(cfg)
}

func TestSettingsService_LoadMissingReturnsDefaults(t *testing.T) {
	s := NewSettingsService(newTestState(t))
	got, err := s.LoadPrefs()
	if err != nil {
		t.Fatalf("LoadPrefs: %v", err)
	}
	if got != DefaultPreferences {
		t.Fatalf("expected defaults, got %+v", got)
	}
}

func TestSettingsService_SaveLoadRoundTrip(t *testing.T) {
	s := NewSettingsService(newTestState(t))
	in := Preferences{
		Theme:         "dark",
		ThemeID:       "dracula",
		ZoomLevel:     1.25,
		ShowHidden:    true,
		LaunchAtLogin: true,
	}
	if err := s.SavePrefs(in); err != nil {
		t.Fatalf("SavePrefs: %v", err)
	}
	got, err := s.LoadPrefs()
	if err != nil {
		t.Fatalf("LoadPrefs: %v", err)
	}
	if got != in {
		t.Fatalf("round-trip mismatch:\nwant %+v\ngot  %+v", in, got)
	}
}

func TestSettingsService_SaveIsAtomic(t *testing.T) {
	s := NewSettingsService(newTestState(t))
	// After a successful save there must be no leftover .tmp file.
	if err := s.SavePrefs(DefaultPreferences); err != nil {
		t.Fatalf("SavePrefs: %v", err)
	}
	if _, err := os.Stat(s.prefsPath() + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("expected no .tmp leftover, got %v", err)
	}
	// And the real file exists.
	if _, err := os.Stat(s.prefsPath()); err != nil {
		t.Fatalf("expected preferences.json to exist: %v", err)
	}
}

func TestSettingsService_LoadFillsMissingFields(t *testing.T) {
	// Simulate an older preference file that's missing newer fields: the
	// loader should fill them from DefaultPreferences (forward-compatible).
	s := NewSettingsService(newTestState(t))
	dir := s.state.cfg.DataDir
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Only write theme + themeId; leave the rest absent.
	if err := os.WriteFile(
		filepath.Join(dir, "preferences.json"),
		[]byte(`{"theme":"light","themeId":"nord"}`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	got, err := s.LoadPrefs()
	if err != nil {
		t.Fatalf("LoadPrefs: %v", err)
	}
	if got.Theme != "light" || got.ThemeID != "nord" {
		t.Fatalf("unexpected themed fields: %+v", got)
	}
	if got.ZoomLevel != DefaultPreferences.ZoomLevel {
		t.Fatalf("expected default zoom, got %v", got.ZoomLevel)
	}
	if got.ShowHidden != DefaultPreferences.ShowHidden {
		t.Fatalf("expected default showHidden, got %v", got.ShowHidden)
	}
}
