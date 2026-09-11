// Package agents defines agent kinds, install targets, skill discovery paths,
// and (from T12+) the provider/session/memory path table for runtime data.
package agents

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// SessionFormat describes how an agent organizes its session files.
type SessionFormat string

const (
	// SessionFormatJSONLPerProject: sessions are JSONL files under
	// <root>/<encoded-project-path>/*.jsonl (Claude Code, Codex).
	SessionFormatJSONLPerProject SessionFormat = "jsonl-per-project"

	// SessionFormatSQLiteGlobal: all sessions share one global SQLite DB
	// (OpenCode, Windsurf, Goose, OpenClaw, Kiro).
	SessionFormatSQLiteGlobal SessionFormat = "sqlite-global"

	// SessionFormatJSONFlat: sessions are independent JSON files under
	// <root>/*.json (Gemini CLI, Continue).
	SessionFormatJSONFlat SessionFormat = "json-flat"

	// SessionFormatVSCodeStorage: sessions live inside the IDE's
	// workspaceStorage SQLite (Cursor, Copilot, Cline, Roo Code).
	// Opt-in via EnableUnverifiedAgentPaths.
	SessionFormatVSCodeStorage SessionFormat = "vscode-storage"

	// SessionFormatNone: the agent does not store local session files
	// (Amp is server-side; Replit AI is cloud-managed).
	SessionFormatNone SessionFormat = "none"
)

// MemoryFormat describes how an agent stores its persistent memory.
type MemoryFormat string

const (
	// MemoryFormatMarkdownIndex: memory is a MEMORY.md index with
	// [[topic]] links to topic files (Claude Code, OpenClaw, Kiro,
	// Gemini CLI).
	MemoryFormatMarkdownIndex MemoryFormat = "markdown-index"

	// MemoryFormatSQLite: memory is stored in a SQLite database
	// (Codex, Kiro FTS5).
	MemoryFormatSQLite MemoryFormat = "sqlite"

	// MemoryFormatMarkdownFlat: memory is a single Markdown file
	// (Windsurf: global_rules.md).
	MemoryFormatMarkdownFlat MemoryFormat = "markdown-flat"

	// MemoryFormatRulesOnly: the agent has no native memory; persistent
	// context is provided via Rules files (Cursor, Cline, Tabnine).
	MemoryFormatRulesOnly MemoryFormat = "rules-only"

	// MemoryFormatServerSide: memory is stored server-side, not on
	// local disk (Copilot, Amp).
	MemoryFormatServerSide MemoryFormat = "server-side"

	// MemoryFormatNone: the agent has no memory feature
	// (Continue, OpenCode, DeepSeek).
	MemoryFormatNone MemoryFormat = "none"
)

// PathStatus indicates whether an agent's paths are verified against
// official documentation.
type PathStatus string

const (
	// PathStatusVerified: paths are verified against official docs or
	// source code (docs/agents/memory-session-paths.md, verified
	// 2026-09-11).
	PathStatusVerified PathStatus = "verified"

	// PathStatusUnverified: paths are community-confirmed or plausible
	// but not found in official documentation. Opt-in via
	// EnableUnverifiedAgentPaths.
	PathStatusUnverified PathStatus = "unverified"

	// PathStatusNotSupported: the agent does not have local
	// provider/session/memory paths (server-side, cloud-managed, or
	// not in the docs).
	PathStatusNotSupported PathStatus = "not-supported"
)

// AgentPaths describes the on-disk persistence locations for a single agent.
// All paths use {home}, {project}, {state_dir}, {data_dir} placeholders that
// are resolved at runtime by ResolveAgentPaths.
//
// Source: docs/agents/memory-session-paths.md (verified 2026-09-11).
type AgentPaths struct {
	// ProviderConfigPath is the agent's provider configuration entry point.
	// Empty if the agent has no local provider config (server-side, VS Code
	// settings, etc.).
	ProviderConfigPath string `json:"provider_config_path"`

	// SessionRootPath is the root directory (or file) for session data.
	// Empty if the agent has no local session storage.
	SessionRootPath string `json:"session_root_path"`

	// SessionFormat describes how session files are organized.
	SessionFormat SessionFormat `json:"session_format"`

	// MemoryRootPath is the root directory (or file) for memory data.
	// Empty if the agent has no memory feature.
	MemoryRootPath string `json:"memory_root_path"`

	// MemoryFormat describes the memory data format.
	MemoryFormat MemoryFormat `json:"memory_format"`

	// Status indicates whether the paths are verified, unverified, or
	// not supported.
	Status PathStatus `json:"status"`
}

// SupportsProvider returns true if the agent has a local provider config.
func (p *AgentPaths) SupportsProvider() bool {
	return p.ProviderConfigPath != ""
}

// SupportsSession returns true if the agent has local session storage.
func (p *AgentPaths) SupportsSession() bool {
	return p.SessionRootPath != "" && p.SessionFormat != SessionFormatNone
}

// SupportsMemory returns true if the agent has a memory feature.
func (p *AgentPaths) SupportsMemory() bool {
	return p.MemoryRootPath != "" && p.MemoryFormat != MemoryFormatNone
}

// DefaultAgentPaths returns the verified agent path table.
//
// This table is the Go equivalent of docs/agents/memory-session-paths.md's
// "Verified table" (rows 1-20). Any change to the docs MUST be accompanied by
// a corresponding change here; paths_verify_test.go asserts the two are in sync.
//
// Rows marked "⚠️ unverified" in the docs are set to PathStatusUnverified and
// only scanned when EnableUnverifiedAgentPaths is true.
// Rows marked "🚫 deprecated" or "❌ no feature" are set to PathStatusNotSupported.
func DefaultAgentPaths() map[AgentKind]AgentPaths {
	return map[AgentKind]AgentPaths{
		// --- Verified (from docs rows 1-20, ✅ or ✅-with-note) ---

		AgentClaudeCode: {
			ProviderConfigPath: "{home}/.claude/settings.json",
			SessionRootPath:    "{home}/.claude/projects",
			SessionFormat:      SessionFormatJSONLPerProject,
			MemoryRootPath:     "{home}/.claude/projects/{project}/memory",
			MemoryFormat:       MemoryFormatMarkdownIndex,
			Status:             PathStatusVerified,
		},

		AgentCodex: {
			ProviderConfigPath: "{home}/.codex/config.toml",
			SessionRootPath:    "{home}/.codex/sessions",
			SessionFormat:      SessionFormatJSONLPerProject,
			MemoryRootPath:     "{home}/.codex/memories_1.sqlite",
			MemoryFormat:       MemoryFormatSQLite,
			Status:             PathStatusVerified,
		},

		AgentGemini: {
			ProviderConfigPath: "{home}/.gemini/settings.json",
			SessionRootPath:    "{home}/.gemini/tmp",
			SessionFormat:      SessionFormatJSONFlat,
			MemoryRootPath:     "{home}/.gemini/tmp/{project}/memory",
			MemoryFormat:       MemoryFormatMarkdownIndex,
			Status:             PathStatusVerified,
		},

		AgentOpenCode: {
			ProviderConfigPath: "{home}/.config/opencode/opencode.json",
			SessionRootPath:    "{data_dir}/opencode/opencode.db",
			SessionFormat:      SessionFormatSQLiteGlobal,
			MemoryRootPath:     "",
			MemoryFormat:       MemoryFormatNone,
			Status:             PathStatusVerified,
		},

		AgentContinue: {
			ProviderConfigPath: "{home}/.continue/config.yaml",
			SessionRootPath:    "{home}/.continue/sessions",
			SessionFormat:      SessionFormatJSONFlat,
			MemoryRootPath:     "",
			MemoryFormat:       MemoryFormatNone,
			Status:             PathStatusVerified,
		},

		AgentWindsurf: {
			ProviderConfigPath: "{home}/.codeium/windsurf/settings.json",
			SessionRootPath:    "{data_dir}/Windsurf/User/workspaceStorage",
			SessionFormat:      SessionFormatSQLiteGlobal,
			MemoryRootPath:     "{home}/.codeium/windsurf/memories",
			MemoryFormat:       MemoryFormatMarkdownFlat,
			Status:             PathStatusVerified,
		},

		AgentOpenClaw: {
			ProviderConfigPath: "{state_dir}/config.toml",
			SessionRootPath:    "{home}/.openclaw/agents/{agent}/agent/openclaw-agent.sqlite",
			SessionFormat:      SessionFormatSQLiteGlobal,
			MemoryRootPath:     "{home}/.openclaw/workspace",
			MemoryFormat:       MemoryFormatMarkdownIndex,
			Status:             PathStatusVerified,
		},

		AgentKiro: {
			ProviderConfigPath: "{home}/.kiro/config.yaml",
			SessionRootPath:    "{home}/.kiro/sessions/cli",
			SessionFormat:      SessionFormatJSONFlat,
			MemoryRootPath:     "{home}/.kiro/crew/workspace/memory",
			MemoryFormat:       MemoryFormatMarkdownIndex,
			Status:             PathStatusVerified,
		},

		// --- Cursor: session verified, memory is Rules-only ---
		AgentCursor: {
			ProviderConfigPath: "{home}/.cursor/settings.json",
			SessionRootPath:    "{home}/.cursor/projects",
			SessionFormat:      SessionFormatVSCodeStorage,
			MemoryRootPath:     "",
			MemoryFormat:       MemoryFormatRulesOnly,
			Status:             PathStatusVerified,
		},

		// --- Copilot: session community-confirmed, memory server-side ---
		AgentCopilot: {
			ProviderConfigPath: "",
			SessionRootPath:    "{home}/.copilot/session-state",
			SessionFormat:      SessionFormatVSCodeStorage,
			MemoryRootPath:     "",
			MemoryFormat:       MemoryFormatServerSide,
			Status:             PathStatusUnverified,
		},

		// --- Antigravity: skill paths verified (skill-paths.md row 20),
		//     but no memory/session paths documented ---
		AgentAntigravity: {
			ProviderConfigPath: "",
			SessionRootPath:    "",
			SessionFormat:      SessionFormatNone,
			MemoryRootPath:     "",
			MemoryFormat:       MemoryFormatNone,
			Status:             PathStatusNotSupported,
		},

		// --- Agents not in memory-session-paths.md docs ---
		AgentUniversal: {Status: PathStatusNotSupported},
		AgentPi: {
			ProviderConfigPath: "{home}/.pi/agent/settings.json",
			SessionRootPath:    "{home}/.pi/agent/sessions",
			SessionFormat:      SessionFormatJSONLPerProject,
			MemoryRootPath:     "{home}/.pi/agent/skills",
			MemoryFormat:       MemoryFormatNone,
			Status:             PathStatusVerified,
		},
		AgentGrok:      {Status: PathStatusNotSupported},
		AgentDroid:     {Status: PathStatusNotSupported},
		AgentHermes:    {Status: PathStatusNotSupported},
		AgentCcSwitch:  {Status: PathStatusNotSupported},

		// --- Agents with unverified paths (community-confirmed, not in docs) ---
		AgentCline: {
			ProviderConfigPath: "", // VS Code settings, not a file
			SessionRootPath:    "{home}/.cline/data/sessions",
			SessionFormat:      SessionFormatVSCodeStorage,
			MemoryRootPath:     "",
			MemoryFormat:       MemoryFormatRulesOnly,
			Status:             PathStatusUnverified,
		},

		AgentAmp: {
			ProviderConfigPath: "{home}/.config/amp/config.json",
			SessionRootPath:    "", // server-side; local mirror unverified
			SessionFormat:      SessionFormatNone,
			MemoryRootPath:     "",
			MemoryFormat:       MemoryFormatServerSide,
			Status:             PathStatusVerified, // Amp settings paths are documented
		},

		AgentGoose: {
			ProviderConfigPath: "{data_dir}/goose/config.yaml",
			SessionRootPath:    "{data_dir}/goose/sessions/sessions.db",
			SessionFormat:      SessionFormatSQLiteGlobal,
			MemoryRootPath:     "{home}/.config/goose/memory",
			MemoryFormat:       MemoryFormatMarkdownIndex,
			Status:             PathStatusUnverified, // memory path not confirmed in source
		},

		AgentRooCode: {
			ProviderConfigPath: "", // VS Code settings
			SessionRootPath:    "", // VS Code globalStorage
			SessionFormat:      SessionFormatVSCodeStorage,
			MemoryRootPath:     "",
			MemoryFormat:       MemoryFormatNone,
			Status:             PathStatusUnverified,
		},
	}
}

// ResolveAgentPaths replaces path placeholders with actual values.
//
// Placeholders:
//   - {home}       → os.UserHomeDir()
//   - {project}    → the given projectPath (empty if not provided)
//   - {state_dir}  → filepath.Join(home, ".<agent>") — agent's own state dir
//   - {data_dir}   → platform data dir (~/Library/Application Support on macOS,
//                     %LOCALAPPDATA% on Windows, ~/.local/share on Linux)
//   - {agent}      → the AgentKind string (for per-agent subpaths)
func ResolveAgentPaths(paths AgentPaths, agent AgentKind, projectPath string) AgentPaths {
	home, _ := os.UserHomeDir()
	dataDir := platformDataDir()

	resolved := func(template string) string {
		if template == "" {
			return ""
		}
		result := template
		result = strings.ReplaceAll(result, "{home}", home)
		result = strings.ReplaceAll(result, "{data_dir}", dataDir)
		result = strings.ReplaceAll(result, "{state_dir}", filepath.Join(home, "."+string(agent)))
		result = strings.ReplaceAll(result, "{agent}", string(agent))
		result = strings.ReplaceAll(result, "{project}", projectPath)
		return result
	}

	return AgentPaths{
		ProviderConfigPath: resolved(paths.ProviderConfigPath),
		SessionRootPath:    resolved(paths.SessionRootPath),
		SessionFormat:      paths.SessionFormat,
		MemoryRootPath:     resolved(paths.MemoryRootPath),
		MemoryFormat:       paths.MemoryFormat,
		Status:             paths.Status,
	}
}

// platformDataDir returns the platform-specific data directory:
//   - macOS:   ~/Library/Application Support
//   - Windows: %LOCALAPPDATA%
//   - Linux:   ~/.local/share
func platformDataDir() string {
	home, _ := os.UserHomeDir()
	switch runtimeGOOS() {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support")
	case "windows":
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			return localAppData
		}
		return filepath.Join(home, "AppData", "Local")
	default:
		if xdgDataHome := os.Getenv("XDG_DATA_HOME"); xdgDataHome != "" {
			return xdgDataHome
		}
		return filepath.Join(home, ".local", "share")
	}
}

// runtimeGOOS returns the current operating system name.
// Extracted so it can be overridden in tests.
var runtimeGOOS = func() string { return runtime.GOOS }
