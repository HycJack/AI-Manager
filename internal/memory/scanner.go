// Package memory implements memory scanning and reading for various agent formats.
//
// This package provides:
//   - MemoryEntry: a common data structure for memory items across agents.
//   - MemoryScanner: an interface for scanning and reading an agent's memory.
//   - For(): dispatch to the correct scanner based on the agent's MemoryFormat.
//   - AllScanners(): returns scanners for all agents with memory support.
//
// Currently implemented formats:
//   - MemoryFormatMarkdownIndex (Claude Code, OpenClaw, Kiro, Gemini CLI)
//   - MemoryFormatMarkdownFlat  (Windsurf)
//
// SQLite scanning is deferred to T19.
package memory

import (
	"os"
	"time"

	"ai-manager/internal/agents"
)

// MemoryEntry describes a single memory item found during scanning.
type MemoryEntry struct {
	// Agent is the agent that owns this memory entry.
	Agent agents.AgentKind `json:"agent"`

	// Path is the absolute file path of this memory entry.
	Path string `json:"path"`

	// Title is a human-readable name for this entry (derived from the file name).
	Title string `json:"title"`

	// Kind classifies the entry:
	//   - "index"      for MEMORY.md (the index file)
	//   - "topic"      for topic files linked from MEMORY.md (e.g. topic.md)
	//   - "workspace"  for OpenClaw workspace files (USER.md, AGENTS.md, etc.)
	//   - "daily"      for date-stamped log files (memory/YYYY-MM-DD.md)
	//   - "preference" for preferences.md (Kiro)
	Kind string `json:"kind"`

	// Content is the full text content of the file.
	Content string `json:"content"`

	// SizeBytes is the size of the file in bytes.
	SizeBytes int `json:"size_bytes"`

	// ModifiedAt is the file's last modification time.
	ModifiedAt time.Time `json:"modified_at"`
}

// MemoryScanner scans and reads memory entries for a specific agent.
type MemoryScanner interface {
	// Agent returns the agent kind this scanner handles.
	Agent() agents.AgentKind

	// Scan returns all memory entries found under the scanner's root path.
	// Returns an empty slice (not nil, not error) if the root path does not exist.
	Scan() ([]MemoryEntry, error)

	// ReadEntry reads the content of a single memory file at the given path.
	ReadEntry(path string) (string, error)
}

// For returns the appropriate MemoryScanner for the given agent, based on
// the agent's MemoryFormat from DefaultAgentPaths.
//
// Returns nil if the agent is not in the table, has no MemoryRootPath, or
// uses a format that is not yet implemented (SQLite is T19).
func For(agent agents.AgentKind) MemoryScanner {
	paths := agents.DefaultAgentPaths()
	p, ok := paths[agent]
	if !ok {
		return nil
	}
	return ForWithPaths(agent, p)
}

// ForWithPaths returns a MemoryScanner for the given agent using
// explicit paths (allows custom path overrides).
func ForWithPaths(agent agents.AgentKind, p agents.AgentPaths) MemoryScanner {
	if p.MemoryRootPath == "" {
		return nil
	}
	resolved := agents.ResolveAgentPaths(p, agent, "")

	switch p.MemoryFormat {
	case agents.MemoryFormatMarkdownIndex:
		return &markdownIndexScanner{
			agent:    agent,
			rootPath: resolved.MemoryRootPath,
		}
	case agents.MemoryFormatMarkdownFlat:
		return &markdownFlatScanner{
			agent:    agent,
			rootPath: resolved.MemoryRootPath,
		}
	default:
		return nil
	}
}

// AllScanners returns scanners for all agents that have memory scanning support.
// Agents with unimplemented formats (SQLite, none, rules-only, server-side) are excluded.
func AllScanners() []MemoryScanner {
	scanners := make([]MemoryScanner, 0, 8)
	for _, agent := range agents.AllAgentKinds() {
		s := For(agent)
		if s != nil {
			scanners = append(scanners, s)
		}
	}
	return scanners
}

// fileModTime returns the modification time of the file at path,
// or zero time if the file cannot be stat'd.
func fileModTime(path string) time.Time {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}
