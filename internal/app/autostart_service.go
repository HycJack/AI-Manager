package app

// AutostartService wraps the Wails3 AutostartManager for frontend control
// of the "launch at login" preference. The actual platform registration
// (registry on Windows, LaunchAgent on macOS, .desktop on Linux) is
// handled by the framework.
type AutostartService struct {
	state *State
}

// NewAutostartService creates the autostart service.
func NewAutostartService(s *State) *AutostartService { return &AutostartService{state: s} }

// IsEnabled reports whether the app is registered to launch at login.
func (s *AutostartService) IsEnabled() (bool, error) {
	return s.state.app.Autostart.IsEnabled()
}

// Enable registers the app to launch at login.
func (s *AutostartService) Enable() error {
	return s.state.app.Autostart.Enable()
}

// Disable removes the login registration.
func (s *AutostartService) Disable() error {
	return s.state.app.Autostart.Disable()
}

// SetEnabled is a convenience toggle for the frontend: enable == true calls
// Enable, false calls Disable.
func (s *AutostartService) SetEnabled(enable bool) error {
	if enable {
		return s.Enable()
	}
	return s.Disable()
}
