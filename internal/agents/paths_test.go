package agents

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDefaultAgentPaths_ReturnsAllAgents verifies that DefaultAgentPaths()
// returns an entry for every AgentKind, so no agent is silently omitted.
func TestDefaultAgentPaths_ReturnsAllAgents(t *testing.T) {
	paths := DefaultAgentPaths()
	for _, kind := range AllAgentKinds() {
		if _, ok := paths[kind]; !ok {
			t.Errorf("DefaultAgentPaths() missing entry for %q", kind)
		}
	}
}

// TestDefaultAgentPaths_VerifiedAgents checks the path values for agents
// verified in docs/agents/memory-session-paths.md (rows 1-20).
func TestDefaultAgentPaths_VerifiedAgents(t *testing.T) {
	paths := DefaultAgentPaths()

	tests := []struct {
		agent    AgentKind
		wantProv string
		wantSess string
		wantSessFmt SessionFormat
		wantMem  string
		wantMemFmt MemoryFormat
		wantStatus PathStatus
	}{
		{
			agent: AgentClaudeCode,
			wantProv: "{home}/.claude/settings.json",
			wantSess: "{home}/.claude/projects",
			wantSessFmt: SessionFormatJSONLPerProject,
			wantMem: "{home}/.claude/projects/{project}/memory",
			wantMemFmt: MemoryFormatMarkdownIndex,
			wantStatus: PathStatusVerified,
		},
		{
			agent: AgentCodex,
			wantProv: "{home}/.codex/config.toml",
			wantSess: "{home}/.codex/sessions",
			wantSessFmt: SessionFormatJSONLPerProject,
			wantMem: "{home}/.codex/memories_1.sqlite",
			wantMemFmt: MemoryFormatSQLite,
			wantStatus: PathStatusVerified,
		},
		{
			agent: AgentGemini,
			wantProv: "{home}/.gemini/settings.json",
			wantSess: "{home}/.gemini/tmp",
			wantSessFmt: SessionFormatJSONFlat,
			wantMem: "{home}/.gemini/tmp/{project}/memory",
			wantMemFmt: MemoryFormatMarkdownIndex,
			wantStatus: PathStatusVerified,
		},
		{
			agent: AgentOpenCode,
			wantProv: "{home}/.config/opencode/opencode.json",
			wantSess: "{data_dir}/opencode/opencode.db",
			wantSessFmt: SessionFormatSQLiteGlobal,
			wantMem: "",
			wantMemFmt: MemoryFormatNone,
			wantStatus: PathStatusVerified,
		},
		{
			agent: AgentContinue,
			wantProv: "{home}/.continue/config.yaml",
			wantSess: "{home}/.continue/sessions",
			wantSessFmt: SessionFormatJSONFlat,
			wantMem: "",
			wantMemFmt: MemoryFormatNone,
			wantStatus: PathStatusVerified,
		},
		{
			agent: AgentWindsurf,
			wantProv: "{home}/.codeium/windsurf/settings.json",
			wantSess: "{data_dir}/Windsurf/User/workspaceStorage",
			wantSessFmt: SessionFormatSQLiteGlobal,
			wantMem: "{home}/.codeium/windsurf/memories",
			wantMemFmt: MemoryFormatMarkdownFlat,
			wantStatus: PathStatusVerified,
		},
		{
			agent: AgentOpenClaw,
			wantProv: "{state_dir}/config.toml",
			wantSess: "{home}/.openclaw/agents/{agent}/agent/openclaw-agent.sqlite",
			wantSessFmt: SessionFormatSQLiteGlobal,
			wantMem: "{home}/.openclaw/workspace",
			wantMemFmt: MemoryFormatMarkdownIndex,
			wantStatus: PathStatusVerified,
		},
		{
			agent: AgentKiro,
			wantProv: "{home}/.kiro/config.yaml",
			wantSess: "{home}/.kiro/sessions/cli",
			wantSessFmt: SessionFormatJSONFlat,
			wantMem: "{home}/.kiro/crew/workspace/memory",
			wantMemFmt: MemoryFormatMarkdownIndex,
			wantStatus: PathStatusVerified,
		},
		{
			agent: AgentCursor,
			wantProv: "{home}/.cursor/settings.json",
			wantSess: "{home}/.cursor/projects",
			wantSessFmt: SessionFormatVSCodeStorage,
			wantMem: "",
			wantMemFmt: MemoryFormatRulesOnly,
			wantStatus: PathStatusVerified,
		},
		{
			agent: AgentCopilot,
			wantProv: "",
			wantSess: "{home}/.copilot/session-state",
			wantSessFmt: SessionFormatVSCodeStorage,
			wantMem: "",
			wantMemFmt: MemoryFormatServerSide,
			wantStatus: PathStatusUnverified,
		},
		{
			agent: AgentAntigravity,
			wantProv: "",
			wantSess: "",
			wantSessFmt: SessionFormatNone,
			wantMem: "",
			wantMemFmt: MemoryFormatNone,
			wantStatus: PathStatusNotSupported,
		},
		{
			agent: AgentAmp,
			wantProv: "{home}/.config/amp/config.json",
			wantSess: "",
			wantSessFmt: SessionFormatNone,
			wantMem: "",
			wantMemFmt: MemoryFormatServerSide,
			wantStatus: PathStatusVerified,
		},
		{
			agent: AgentCline,
			wantProv: "",
			wantSess: "{home}/.cline/data/sessions",
			wantSessFmt: SessionFormatVSCodeStorage,
			wantMem: "",
			wantMemFmt: MemoryFormatRulesOnly,
			wantStatus: PathStatusUnverified,
		},
		{
			agent: AgentGoose,
			wantProv: "{data_dir}/goose/config.yaml",
			wantSess: "{data_dir}/goose/sessions/sessions.db",
			wantSessFmt: SessionFormatSQLiteGlobal,
			wantMem: "{home}/.config/goose/memory",
			wantMemFmt: MemoryFormatMarkdownIndex,
			wantStatus: PathStatusUnverified,
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.agent), func(t *testing.T) {
			p, ok := paths[tt.agent]
			if !ok {
				t.Fatalf("DefaultAgentPaths() missing entry for %q", tt.agent)
			}
			if p.ProviderConfigPath != tt.wantProv {
				t.Errorf("ProviderConfigPath = %q, want %q", p.ProviderConfigPath, tt.wantProv)
			}
			if p.SessionRootPath != tt.wantSess {
				t.Errorf("SessionRootPath = %q, want %q", p.SessionRootPath, tt.wantSess)
			}
			if p.SessionFormat != tt.wantSessFmt {
				t.Errorf("SessionFormat = %q, want %q", p.SessionFormat, tt.wantSessFmt)
			}
			if p.MemoryRootPath != tt.wantMem {
				t.Errorf("MemoryRootPath = %q, want %q", p.MemoryRootPath, tt.wantMem)
			}
			if p.MemoryFormat != tt.wantMemFmt {
				t.Errorf("MemoryFormat = %q, want %q", p.MemoryFormat, tt.wantMemFmt)
			}
			if p.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", p.Status, tt.wantStatus)
			}
		})
	}
}

// TestDefaultAgentPaths_NotSupportedAgents checks that agents without
// local paths are marked as not-supported.
func TestDefaultAgentPaths_NotSupportedAgents(t *testing.T) {
	paths := DefaultAgentPaths()

	for _, kind := range []AgentKind{
		AgentUniversal, AgentPi, AgentGrok, AgentDroid, AgentHermes, AgentCcSwitch,
	} {
		p, ok := paths[kind]
		if !ok {
			t.Fatalf("DefaultAgentPaths() missing entry for %q", kind)
		}
		if p.Status != PathStatusNotSupported {
			t.Errorf("Status for %q = %q, want %q", kind, p.Status, PathStatusNotSupported)
		}
		if p.SupportsProvider() || p.SupportsSession() || p.SupportsMemory() {
			t.Errorf("Agent %q should not support any feature", kind)
		}
	}
}

// TestAgentPaths_SupportsProvider checks the SupportsProvider() helper.
func TestAgentPaths_SupportsProvider(t *testing.T) {
	paths := DefaultAgentPaths()

	p := paths[AgentClaudeCode]
	if !p.SupportsProvider() {
		t.Error("ClaudeCode should support provider")
	}
	p = paths[AgentCursor]
	if p.SupportsProvider() == false {
		t.Error("Cursor should support provider (settings.json)")
	}
	p = paths[AgentAntigravity]
	if p.SupportsProvider() {
		t.Error("Antigravity should not support provider")
	}
}

// TestAgentPaths_SupportsSession checks the SupportsSession() helper.
func TestAgentPaths_SupportsSession(t *testing.T) {
	paths := DefaultAgentPaths()

	p := paths[AgentClaudeCode]
	if !p.SupportsSession() {
		t.Error("ClaudeCode should support session")
	}
	p = paths[AgentOpenCode]
	if !p.SupportsSession() {
		t.Error("OpenCode should support session")
	}
	p = paths[AgentAmp]
	if p.SupportsSession() {
		t.Error("Amp should not support local session (server-side)")
	}
}

// TestAgentPaths_SupportsMemory checks the SupportsMemory() helper.
func TestAgentPaths_SupportsMemory(t *testing.T) {
	paths := DefaultAgentPaths()

	p := paths[AgentClaudeCode]
	if !p.SupportsMemory() {
		t.Error("ClaudeCode should support memory")
	}
	p = paths[AgentCodex]
	if !p.SupportsMemory() {
		t.Error("Codex should support memory")
	}
	p = paths[AgentContinue]
	if p.SupportsMemory() {
		t.Error("Continue should not support memory (uses Rules)")
	}
	p = paths[AgentOpenCode]
	if p.SupportsMemory() {
		t.Error("OpenCode should not support memory (no feature)")
	}
}

// TestResolveAgentPaths_HomePlaceholder verifies {home} is replaced.
func TestResolveAgentPaths_HomePlaceholder(t *testing.T) {
	home, _ := os.UserHomeDir()

	p := AgentPaths{
		ProviderConfigPath: "{home}/.claude/settings.json",
		SessionRootPath:    "{home}/.claude/projects",
	}
	got := ResolveAgentPaths(p, AgentClaudeCode, "")

	if !strings.HasPrefix(got.ProviderConfigPath, home+"/.claude/settings.json") {
		t.Errorf("ProviderConfigPath = %q, want prefix %q", got.ProviderConfigPath, home)
	}
	if !strings.HasPrefix(got.SessionRootPath, home+"/.claude/projects") {
		t.Errorf("SessionRootPath = %q, want prefix %q", got.SessionRootPath, home)
	}
}

// TestResolveAgentPaths_ProjectPlaceholder verifies {project} is replaced.
func TestResolveAgentPaths_ProjectPlaceholder(t *testing.T) {
	p := AgentPaths{
		MemoryRootPath: "{home}/.claude/projects/{project}/memory",
	}
	got := ResolveAgentPaths(p, AgentClaudeCode, "/workspace/my-project")

	if !strings.Contains(got.MemoryRootPath, "/workspace/my-project/memory") {
		t.Errorf("MemoryRootPath = %q, want to contain %q", got.MemoryRootPath, "/workspace/my-project/memory")
	}
}

// TestResolveAgentPaths_StateDirPlaceholder verifies {state_dir} is replaced
// with ~/.<agent>.
func TestResolveAgentPaths_StateDirPlaceholder(t *testing.T) {
	home, _ := os.UserHomeDir()

	p := AgentPaths{
		ProviderConfigPath: "{state_dir}/config.toml",
	}
	got := ResolveAgentPaths(p, AgentOpenClaw, "")

	want := filepath.Join(home, ".openclaw", "config.toml")
	if got.ProviderConfigPath != want {
		t.Errorf("ProviderConfigPath = %q, want %q", got.ProviderConfigPath, want)
	}
}

// TestResolveAgentPaths_DataDirPlaceholder verifies {data_dir} is replaced.
func TestResolveAgentPaths_DataDirPlaceholder(t *testing.T) {
	home, _ := os.UserHomeDir()

	p := AgentPaths{
		SessionRootPath: "{data_dir}/opencode/opencode.db",
	}
	got := ResolveAgentPaths(p, AgentOpenCode, "")

	if got.SessionRootPath == "" {
		t.Fatal("SessionRootPath should not be empty")
	}
	if strings.Contains(got.SessionRootPath, "{data_dir}") {
		t.Errorf("SessionRootPath = %q, placeholder not replaced", got.SessionRootPath)
	}
	if !strings.Contains(got.SessionRootPath, home) {
		t.Errorf("SessionRootPath = %q, want to contain home %q", got.SessionRootPath, home)
	}
}

// TestResolveAgentPaths_AgentPlaceholder verifies {agent} is replaced with
// the AgentKind string.
func TestResolveAgentPaths_AgentPlaceholder(t *testing.T) {
	p := AgentPaths{
		SessionRootPath: "{home}/.openclaw/agents/{agent}/agent/openclaw-agent.sqlite",
	}
	got := ResolveAgentPaths(p, AgentOpenClaw, "")

	if !strings.Contains(got.SessionRootPath, "agents/openclaw/agent") {
		t.Errorf("SessionRootPath = %q, want to contain %q", got.SessionRootPath, "agents/openclaw/agent")
	}
}

// TestResolveAgentPaths_EmptyTemplate verifies empty paths stay empty.
func TestResolveAgentPaths_EmptyTemplate(t *testing.T) {
	p := AgentPaths{
		ProviderConfigPath: "",
		SessionRootPath:    "",
	}
	got := ResolveAgentPaths(p, AgentAntigravity, "")

	if got.ProviderConfigPath != "" {
		t.Errorf("ProviderConfigPath = %q, want empty", got.ProviderConfigPath)
	}
	if got.SessionRootPath != "" {
		t.Errorf("SessionRootPath = %q, want empty", got.SessionRootPath)
	}
}

// TestResolveAgentPaths_PreservesFormat verifies the format fields are
// preserved through resolution.
func TestResolveAgentPaths_PreservesFormat(t *testing.T) {
	p := AgentPaths{
		SessionFormat: SessionFormatJSONLPerProject,
		MemoryFormat:  MemoryFormatMarkdownIndex,
		Status:        PathStatusVerified,
	}
	got := ResolveAgentPaths(p, AgentClaudeCode, "")

	if got.SessionFormat != SessionFormatJSONLPerProject {
		t.Errorf("SessionFormat = %q, want %q", got.SessionFormat, SessionFormatJSONLPerProject)
	}
	if got.MemoryFormat != MemoryFormatMarkdownIndex {
		t.Errorf("MemoryFormat = %q, want %q", got.MemoryFormat, MemoryFormatMarkdownIndex)
	}
	if got.Status != PathStatusVerified {
		t.Errorf("Status = %q, want %q", got.Status, PathStatusVerified)
	}
}

// TestPlatformDataDir verifies platformDataDir() returns a non-empty path.
func TestPlatformDataDir(t *testing.T) {
	d := platformDataDir()
	if d == "" {
		t.Fatal("platformDataDir() returned empty string")
	}
	// Should contain the user's home directory on all platforms
	home, _ := os.UserHomeDir()
	if !strings.HasPrefix(d, home) && !strings.HasPrefix(d, "/") {
		t.Errorf("platformDataDir() = %q, does not look like a valid path", d)
	}
}
