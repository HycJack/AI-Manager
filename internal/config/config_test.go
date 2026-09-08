package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault_Values(t *testing.T) {
	c := Default()
	if c.AppName == "" {
		t.Error("expected non-empty AppName")
	}
	if c.DataDir == "" {
		t.Error("expected non-empty DataDir")
	}
	if c.LogLevel != "info" {
		t.Errorf("expected default LogLevel=info, got %q", c.LogLevel)
	}
	if c.UpdateRepo != "" {
		t.Errorf("expected default UpdateRepo empty (opt-in), got %q", c.UpdateRepo)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	in := Default()
	in.AppName = "MyApp"
	in.Proxy = "http://127.0.0.1:7890"
	in.LogLevel = "debug"
	in.UpdateRepo = "owner/repo"

	if err := in.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if *got != *in {
		t.Fatalf("round-trip mismatch:\nwant %+v\ngot  %+v", *in, *got)
	}
}

func TestLoad_MissingFileDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nope.json")
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load on missing file should not error, got %v", err)
	}
	if got.DataDir == "" || got.LogLevel == "" {
		t.Fatalf("expected defaults back, got %+v", got)
	}
}

func TestLoad_MissingFieldsStayDefault(t *testing.T) {
	// Old config file missing newly-added fields (logLevel/updateRepo) should
	// be filled from Default() — forward compatible.
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{"appName":"Old","dataDir":"/tmp/x"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.AppName != "Old" {
		t.Errorf("expected existing field preserved, got %q", got.AppName)
	}
	if got.LogLevel != "info" {
		t.Errorf("expected missing logLevel to default to info, got %q", got.LogLevel)
	}
	if got.UpdateRepo != "" {
		t.Errorf("expected missing updateRepo to default to empty, got %q", got.UpdateRepo)
	}
}
