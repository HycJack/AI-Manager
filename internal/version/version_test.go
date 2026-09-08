package version

import "testing"

func withVersionInfo(t *testing.T, v, commit, buildTime string) {
	t.Helper()
	origV, origC, origB := Version, Commit, BuildTime
	Version, Commit, BuildTime = v, commit, buildTime
	t.Cleanup(func() { Version, Commit, BuildTime = origV, origC, origB })
}

func TestGet_DefaultsDevBuild(t *testing.T) {
	withVersionInfo(t, "dev", "unknown", "unknown")
	info := Get()
	if info.Version != "dev" || info.Commit != "unknown" || info.BuildTime != "unknown" {
		t.Fatalf("unexpected defaults: %+v", info)
	}
}

func TestGet_InjectedValues(t *testing.T) {
	withVersionInfo(t, "1.2.3", "abc12345", "2026-01-01T00:00:00Z")
	info := Get()
	if info.Version != "1.2.3" || info.Commit != "abc12345" {
		t.Fatalf("unexpected values: %+v", info)
	}
	if info.BuildTime != "2026-01-01T00:00:00Z" {
		t.Fatalf("unexpected build time: %q", info.BuildTime)
	}
}
