package app

import (
	"fmt"
	"strings"
	"time"
)

// GreeterService is a minimal example service showing the Wails3 pattern:
//
//   - NewXService(state) constructor takes the shared *State.
//   - Every exported method on the struct becomes a method callable from the
//     frontend (after `wails3 generate bindings`).
//   - Methods return plain types (or types from this module) that get
//     marshalled over the JS bridge; multi-return values become the resolve
//     value + a possible error.
//
// It deliberately does NOT touch config — that's SettingsService's job — so
// the two services stay orthogonal. Replace / extend this with your own
// services. Register the service in main.go's application.Services slice.
type GreeterService struct {
	state *State
}

// NewGreeterService creates the example service.
func NewGreeterService(s *State) *GreeterService { return &GreeterService{state: s} }

// Greet returns a friendly message for the given name. It demonstrates a
// simple request/response call from the frontend and emits a "greeting" event
// the frontend can subscribe to (see State.Emit).
func (s *GreeterService) Greet(name string) string {
	if name == "" {
		name = "World"
	}
	s.state.Emit("greeting", map[string]any{
		"name": name,
		"at":   time.Now().Format(time.RFC3339),
	})
	return fmt.Sprintf("Hello, %s! 👋", strings.TrimSpace(name))
}

// Ping is a no-op used to verify the bridge is alive.
func (s *GreeterService) Ping() string { return "pong" }
