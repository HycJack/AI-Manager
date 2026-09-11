// Package sessions scans agent session files on disk and extracts metadata.
//
// Sessions are organized differently per agent: JSONL per-project files
// (Claude Code, Codex), flat JSON files (Gemini CLI, Continue, Kiro), or
// SQLite/VS Code storage (handled by the no-op scanner until T19).
//
// See docs/agent-provider-session-memory.md §"Session" for the design.
package sessions

import (
	"fmt"
	"time"

	"ai-manager/internal/agents"
)

// Session describes metadata + content entry for one agent session.
type Session struct {
	// Agent is the agent that produced this session.
	Agent agents.AgentKind `json:"agent"`

	// ID is the session identifier. For JSONL/JSON scanners this is the
	// filename without extension.
	ID string `json:"id"`

	// Project is the project path this session belongs to. Derived from
	// the parent directory slug for JSONL scanners; empty for flat JSON
	// and other formats.
	Project string `json:"project,omitempty"`

	// StartedAt is the session start time. Currently derived from the
	// file's mtime.
	StartedAt time.Time `json:"started_at"`

	// EndedAt is the session end time. Zero if not derivable.
	EndedAt time.Time `json:"ended_at,omitempty"`

	// MessageCount is the number of message entries in the session.
	MessageCount int `json:"message_count"`

	// TokenEstimate is a rough token count (total bytes / 4).
	TokenEstimate int `json:"token_estimate,omitempty"`

	// Path is the on-disk path to the session file.
	Path string `json:"path"`

	// Format is the file format identifier: "jsonl", "json", "sqlite".
	Format string `json:"format"`
}

// SessionScanner scans sessions for a single agent.
type SessionScanner interface {
	// Agent returns the agent this scanner serves.
	Agent() agents.AgentKind
	// Scan returns all sessions found on disk for this agent.
	// A missing root directory returns an empty list, not an error.
	Scan() ([]Session, error)
	// ReadTranscript returns the raw session content (read-only).
	ReadTranscript(id string) (string, error)
}

// SessionFilter constrains a session list to a subset.
//
// Limit defaults to 50 when set to zero.
type SessionFilter struct {
	Project string    `json:"project,omitempty"`
	Since   time.Time `json:"since,omitempty"`
	Limit   int       `json:"limit,omitempty"`
}

// EffectiveLimit returns Limit, defaulting to 50 when unset.
func (f *SessionFilter) EffectiveLimit() int {
	if f.Limit > 0 {
		return f.Limit
	}
	return 50
}

// ApplyFilter returns a subset of sessions matching the filter.
//
// Filters are ANDed: a session must match both the project (if set)
// and the Since cutoff (if set) to be included. The Limit is applied
// after filtering and truncates the result to at most EffectiveLimit().
func ApplyFilter(sessions []Session, f SessionFilter) []Session {
	out := make([]Session, 0, len(sessions))
	for _, s := range sessions {
		if f.Project != "" && s.Project != f.Project {
			continue
		}
		if !f.Since.IsZero() && s.StartedAt.Before(f.Since) {
			continue
		}
		out = append(out, s)
	}
	if limit := f.EffectiveLimit(); limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

// For returns a SessionScanner for the given agent, dispatching on the
// agent's SessionFormat from DefaultAgentPaths.
//
// Returns nil when the agent has no local session storage (SessionFormatNone
// or missing from the table). For formats not yet implemented (SQLite,
// VS Code storage — T19), returns a noopScanner that yields an empty list.
func For(agent agents.AgentKind) SessionScanner {
	paths := agents.DefaultAgentPaths()
	p, ok := paths[agent]
	if !ok {
		return nil
	}
	return ForWithPaths(agent, p)
}

// ForWithPaths returns a SessionScanner for the given agent using
// explicit paths (allows custom path overrides).
func ForWithPaths(agent agents.AgentKind, p agents.AgentPaths) SessionScanner {
	if !p.SupportsSession() {
		return nil
	}
	switch p.SessionFormat {
	case agents.SessionFormatJSONLPerProject:
		return newJSONLScanner(agent, p.SessionRootPath)
	case agents.SessionFormatJSONFlat:
		return newJSONScanner(agent, p.SessionRootPath)
	case agents.SessionFormatSQLiteGlobal:
		return newSQLiteScanner(agent, p.SessionRootPath)
	case agents.SessionFormatVSCodeStorage:
		return newNoopScanner(agent, p.SessionFormat)
	default:
		return nil
	}
}

// AllScanners returns scanners for every agent that has local session
// storage. Includes no-op scanners for agents whose format is deferred
// to T19; callers may filter these out if they need only live scanners.
func AllScanners() []SessionScanner {
	out := make([]SessionScanner, 0, len(agents.AllAgentKinds()))
	for _, kind := range agents.AllAgentKinds() {
		if s := For(kind); s != nil {
			out = append(out, s)
		}
	}
	return out
}

// noopScanner is returned for session formats not yet implemented (T19).
// It returns an empty list from Scan() and an error from ReadTranscript
// that names the deferred format.
type noopScanner struct {
	agent  agents.AgentKind
	format agents.SessionFormat
}

func newNoopScanner(agent agents.AgentKind, format agents.SessionFormat) *noopScanner {
	return &noopScanner{agent: agent, format: format}
}

// Agent implements SessionScanner.
func (s *noopScanner) Agent() agents.AgentKind { return s.agent }

// Scan returns an empty list; the underlying format is not implemented yet.
func (s *noopScanner) Scan() ([]Session, error) {
	// T19: SQLite / VS Code storage scanning not yet implemented.
	return []Session{}, nil
}

// ReadTranscript returns an error explaining the format is deferred to T19.
func (s *noopScanner) ReadTranscript(id string) (string, error) {
	return "", fmt.Errorf("session format %q not implemented (T19)", s.format)
}
