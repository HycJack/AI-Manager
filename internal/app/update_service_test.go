package app

import (
	"strings"
	"testing"

	"ai-manager/internal/version"
)

// withVersion temporarily overrides the package-level build version and
// restores it when the test finishes.
func withVersion(t *testing.T, v string) {
	t.Helper()
	orig := version.Version
	version.Version = v
	t.Cleanup(func() { version.Version = orig })
}

func TestUpdateService_Disabled_DevBuild(t *testing.T) {
	private := NewUpdateService(newTestState(t))
	// No repo configured + dev version → disabled.
	withVersion(t, "dev")
	if !private.disabled() {
		t.Fatal("expected dev build with empty repo to be disabled")
	}
}

func TestUpdateService_Disabled_UnknownVersion(t *testing.T) {
	private := NewUpdateService(newTestState(t))
	withVersion(t, "unknown")
	if !private.disabled() {
		t.Fatal("expected unknown version to be disabled")
	}
}

func TestUpdateService_Disabled_EmptyRepoButRealVersion(t *testing.T) {
	private := NewUpdateService(newTestState(t))
	withVersion(t, "1.2.3")
	// A real version but no repo → still disabled (opt-in updates).
	if !private.disabled() {
		t.Fatal("expected empty repo to disable updates even with a real version")
	}
}

func TestUpdateService_NotDisabled_Configured(t *testing.T) {
	s := NewUpdateService(newTestState(t))
	withVersion(t, "1.2.3")
	s.state.cfg.UpdateRepo = "owner/repo"
	if s.disabled() {
		t.Fatal("expected configured repo + version to be enabled")
	}
}

func TestUpdateService_Check_ReturnsErrorWhenAppNil(t *testing.T) {
	s := NewUpdateService(newTestState(t))
	withVersion(t, "1.2.3")
	s.state.cfg.UpdateRepo = "owner/repo"
	// state.app is nil (New doesn't set it) → ensureInit returns a clear error.
	_, err := s.CheckForUpdates()
	if err == nil {
		t.Fatal("expected error when updater unavailable (app not set)")
	}
	if !strings.Contains(err.Error(), "not available on this platform") {
		t.Fatalf("unexpected error: %v", err)
	}
}
