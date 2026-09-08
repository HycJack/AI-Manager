package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/updater"
	"github.com/wailsapp/wails/v3/pkg/updater/providers/github"
	"ai-manager/internal/version"
)

// UpdateService wraps the Wails3 updater for frontend consumption.
// The frontend calls CheckForUpdates / DownloadAndInstall / Restart; the
// updater's own events (EventUpdateAvailable, EventUpdateReady, …) are
// emitted on the app's event bus and can be subscribed to directly.
type UpdateService struct {
	state *State
	init  sync.Once
}

// NewUpdateService creates the update service.
func NewUpdateService(s *State) *UpdateService { return &UpdateService{state: s} }

// CheckForUpdates queries the GitHub releases endpoint for a newer version.
// Returns nil if already up to date.
func (s *UpdateService) CheckForUpdates() (*updater.Release, error) {
	if err := s.ensureInit(); err != nil {
		return nil, err
	}
	return s.state.app.Updater.Check(context.Background())
}

// CheckAndInstall opens the built-in update window and runs the full flow
// (check → download → verify → stage). The user dismisses the window.
func (s *UpdateService) CheckAndInstall() error {
	if err := s.ensureInit(); err != nil {
		return err
	}
	return s.state.app.Updater.CheckAndInstall(context.Background())
}

// Restart quits the app and swaps in the staged update. Only valid after a
// successful DownloadAndInstall.
func (s *UpdateService) Restart() error {
	if err := s.ensureInit(); err != nil {
		return err
	}
	return s.state.app.Updater.Restart(context.Background())
}

// CurrentVersion returns the version string the updater was configured with.
func (s *UpdateService) CurrentVersion() string {
	return version.Version
}

// UpdateState reports the updater's current lifecycle phase as a string,
// so the frontend can render the right UI without importing the enum.
func (s *UpdateService) UpdateState() string {
	if s.state.app == nil || s.state.app.Updater == nil {
		return "unconfigured"
	}
	return string(s.state.app.Updater.State())
}

// disabled reports whether the updater should be skipped entirely. In dev
// builds version.Version is "dev" (no ldflags), which is not a semver — the
// GitHub provider would always conclude an update exists, and the user
// shouldn't be prompted while developing. It also covers an unset repo.
func (s *UpdateService) disabled() bool {
	v := version.Version
	return v == "" || v == "dev" || v == "unknown" || s.state.cfg.UpdateRepo == ""
}

// ensureInit lazily configures the updater exactly once. The repo owner and
// name come from config.json (UpdateRepo). sync.Once makes concurrent calls
// from the frontend safe — only the first one performs the Init.
func (s *UpdateService) ensureInit() error {
	if s.disabled() {
		return fmt.Errorf("updater: disabled in this build")
	}
	// Resolve the updater handle up front; a State whose app was never set
	// must surface a clear error instead of nil-deref panicking the Once body.
	if s.state.app == nil || s.state.app.Updater == nil {
		return fmt.Errorf("updater: not available on this platform")
	}
	var err error
	s.init.Do(func() {
		provider, perr := github.New(github.Config{
			Repository: s.state.cfg.UpdateRepo,
		})
		if perr != nil {
			err = fmt.Errorf("updater: create github provider: %w", perr)
			return
		}
		if ierr := s.state.app.Updater.Init(updater.Config{
			CurrentVersion: version.Version,
			Providers:      []updater.Provider{provider},
			CheckInterval:  24 * time.Hour, // daily background check
		}); ierr != nil {
			err = fmt.Errorf("updater: init: %w", ierr)
		}
	})
	return err
}
