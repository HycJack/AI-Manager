// Package app holds the Wails3 application bootstrap and bound services.
// AgentService exposes agent provider/session/memory views to the frontend.
package app

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"ai-manager/internal/agents"
	"ai-manager/internal/memory"
	"ai-manager/internal/providers"
	"ai-manager/internal/sessions"
)

// AgentService exposes agent provider config, session list, and memory
// views to the frontend via Wails3 bindings.
type AgentService struct {
	state *State
	// enableUnverifiedPaths controls whether unverified agent paths are
	// scanned. Defaults to false.
	enableUnverifiedPaths bool
	// customPaths holds user-configured path overrides, loaded from disk.
	customPaths map[string]CustomAgentPaths
}

// CustomAgentPaths holds user-configured path overrides for a single agent.
type CustomAgentPaths struct {
	ProviderConfigPath string `json:"provider_config_path,omitempty"`
	SessionRootPath    string `json:"session_root_path,omitempty"`
	SessionFormat      string `json:"session_format,omitempty"`
	MemoryRootPath     string `json:"memory_root_path,omitempty"`
	MemoryFormat       string `json:"memory_format,omitempty"`
}

// AllCustomPaths is the root of the custom paths JSON file.
type AllCustomPaths struct {
	Agents map[string]CustomAgentPaths `json:"agents"`
}

// NewAgentService creates the agent service.
func NewAgentService(s *State) *AgentService {
	return &AgentService{
		state:                 s,
		enableUnverifiedPaths: false,
		customPaths:           loadCustomPaths(),
	}
}

// AgentInfo is the frontend-facing view of an agent with its paths and
// capability flags.
type AgentInfo struct {
	Kind             agents.AgentKind      `json:"kind"`
	Label            string                `json:"label"`
	IconType         string                `json:"icon_type"`
	Paths            agents.AgentPaths     `json:"paths"`
	SupportsProvider bool                  `json:"supports_provider"`
	SupportsSession  bool                  `json:"supports_session"`
	SupportsMemory   bool                  `json:"supports_memory"`
	Status           agents.PathStatus     `json:"status"`
	HasProviderConfig bool                 `json:"has_provider_config"`
	HasSessions      bool                  `json:"has_sessions"`
	HasMemory        bool                  `json:"has_memory"`
}

// SessionFilter filters sessions by project and/or time.
type SessionFilter struct {
	Project string `json:"project"`
	Since   string `json:"since"` // ISO 8601 timestamp (empty = no filter)
	Limit   int    `json:"limit"` // default 50
}

// ListAgents returns all known agents with their paths and capabilities.
func (s *AgentService) ListAgents() ([]AgentInfo, error) {
	defaultPaths := agents.DefaultAgentPaths()
	allKinds := agents.AllAgentKinds()

	result := make([]AgentInfo, 0, len(allKinds))
	for _, kind := range allKinds {
		p, ok := defaultPaths[kind]
		if !ok {
			p = agents.AgentPaths{Status: agents.PathStatusNotSupported}
		}

		// Apply custom path overrides
		if custom, exists := s.customPaths[string(kind)]; exists {
			p = applyCustomPaths(p, custom)
		}

		// Skip unverified agents unless the toggle is enabled
		if p.Status == agents.PathStatusUnverified && !s.enableUnverifiedPaths {
			continue
		}

		resolved := agents.ResolveAgentPaths(p, kind, "")
		hasProvider := s.hasProviderConfig(kind, resolved)
		hasSessions := s.hasSessions(kind, resolved)
		hasMemory := s.hasMemory(kind, resolved)

		info := AgentInfo{
			Kind:              kind,
			Label:             agentLabel(kind),
			IconType:          agentIcon(kind),
			Paths:             resolved,
			SupportsProvider:  resolved.SupportsProvider(),
			SupportsSession:   resolved.SupportsSession(),
			SupportsMemory:    resolved.SupportsMemory(),
			Status:            resolved.Status,
			HasProviderConfig: hasProvider,
			HasSessions:       hasSessions,
			HasMemory:         hasMemory,
		}
		result = append(result, info)
	}
	return result, nil
}

// GetAgent returns a single agent's info by kind string.
func (s *AgentService) GetAgent(kind string) (*AgentInfo, error) {
	ak := agents.AgentKind(kind)
	all, err := s.ListAgents()
	if err != nil {
		return nil, err
	}
	for _, a := range all {
		if a.Kind == ak {
			return &a, nil
		}
	}
	return nil, nil
}

// GetProviderConfig returns the provider configuration for an agent.
func (s *AgentService) GetProviderConfig(kind string) (*providers.ProviderConfig, error) {
	ak := agents.AgentKind(kind)
	adapter := providers.For(ak)
	if adapter == nil {
		return &providers.ProviderConfig{
			Agent: ak, Format: "none", Exists: false,
		}, nil
	}
	return adapter.Read()
}

// UpdateProviderConfig updates the provider configuration for an agent.
func (s *AgentService) UpdateProviderConfig(kind string, patch *providers.ProviderConfig) (*providers.ProviderConfig, error) {
	ak := agents.AgentKind(kind)
	adapter := providers.For(ak)
	if adapter == nil {
		return nil, nil
	}
	if patch == nil {
		return nil, nil
	}
	// Merge: read current, apply patch, write
	current, err := adapter.Read()
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, nil
	}
	if patch.Model != "" {
		current.Model = patch.Model
	}
	if patch.BaseURL != "" {
		current.BaseURL = patch.BaseURL
	}
	if patch.APIKey != "" && !providers.IsMaskedAPIKey(patch.APIKey) {
		current.APIKey = patch.APIKey
	}
	if patch.Provider != "" && patch.Provider != "unknown" {
		current.Provider = patch.Provider
	}
	if err := adapter.Write(current); err != nil {
		return nil, err
	}
	return adapter.Read()
}

// getResolvedPaths returns the resolved paths for an agent, applying
// custom overrides if present.
func (s *AgentService) getResolvedPaths(kind agents.AgentKind) agents.AgentPaths {
	defaultPaths := agents.DefaultAgentPaths()
	p, ok := defaultPaths[kind]
	if !ok {
		return agents.AgentPaths{Status: agents.PathStatusNotSupported}
	}
	if custom, exists := s.customPaths[string(kind)]; exists {
		p = applyCustomPaths(p, custom)
	}
	return agents.ResolveAgentPaths(p, kind, "")
}

// ListSessions returns sessions for an agent, optionally filtered.
func (s *AgentService) ListSessions(kind string, filter *SessionFilter) ([]sessions.Session, error) {
	ak := agents.AgentKind(kind)
	paths := s.getResolvedPaths(ak)
	scanner := sessions.ForWithPaths(ak, paths)
	if scanner == nil {
		return nil, nil
	}
	all, err := scanner.Scan()
	if err != nil {
		return nil, err
	}

	// Apply filters
	if filter != nil {
		if filter.Project != "" {
			filtered := make([]sessions.Session, 0)
			for _, sess := range all {
				if sess.Project == filter.Project || containsStr(sess.Project, filter.Project) {
					filtered = append(filtered, sess)
				}
			}
			all = filtered
		}
		if filter.Since != "" {
			since, err := parseTime(filter.Since)
			if err == nil {
				filtered := make([]sessions.Session, 0)
				for _, sess := range all {
					if sess.StartedAt.After(since) || sess.StartedAt.Equal(since) {
						filtered = append(filtered, sess)
					}
				}
				all = filtered
			}
		}
	}
	sortSessions(all)

	// Apply limit
	if filter != nil && filter.Limit > 0 && len(all) > filter.Limit {
		all = all[:filter.Limit]
	} else if len(all) > 50 {
		all = all[:50]
	}

	return all, nil
}

// GetSessionTranscript returns the raw transcript content for a session.
func (s *AgentService) GetSessionTranscript(kind, id string) (string, error) {
	ak := agents.AgentKind(kind)
	paths := s.getResolvedPaths(ak)
	scanner := sessions.ForWithPaths(ak, paths)
	if scanner == nil {
		return "", nil
	}
	return scanner.ReadTranscript(id)
}

// ListMemory returns memory entries for an agent.
func (s *AgentService) ListMemory(kind string) ([]memory.MemoryEntry, error) {
	ak := agents.AgentKind(kind)
	paths := s.getResolvedPaths(ak)
	scanner := memory.ForWithPaths(ak, paths)
	if scanner == nil {
		return nil, nil
	}
	return scanner.Scan()
}

// GetMemoryEntry returns a single memory entry's full content.
func (s *AgentService) GetMemoryEntry(kind, path string) (*memory.MemoryEntry, error) {
	ak := agents.AgentKind(kind)
	paths := s.getResolvedPaths(ak)
	scanner := memory.ForWithPaths(ak, paths)
	if scanner == nil {
		return nil, nil
	}
	content, err := scanner.ReadEntry(path)
	if err != nil {
		return nil, err
	}
	return &memory.MemoryEntry{
		Agent:  ak,
		Path:   path,
		Title:  titleFromPath(path),
		Content: content,
	}, nil
}

// OpenSessionInEditor opens a session file in the system's default editor.
func (s *AgentService) OpenSessionInEditor(kind, id string) error {
	ak := agents.AgentKind(kind)
	paths := s.getResolvedPaths(ak)
	scanner := sessions.ForWithPaths(ak, paths)
	if scanner == nil {
		return nil
	}
	// Find the session's file path
	all, err := scanner.Scan()
	if err != nil {
		return err
	}
	for _, sess := range all {
		if sess.ID == id {
			return s.openInEditor(sess.Path)
		}
	}
	return nil
}

// OpenMemoryInEditor opens a memory file in the system's default editor.
func (s *AgentService) OpenMemoryInEditor(kind, path string) error {
	ak := agents.AgentKind(kind)
	paths := s.getResolvedPaths(ak)
	scanner := memory.ForWithPaths(ak, paths)
	if scanner == nil {
		return nil
	}
	return s.openInEditor(path)
}

// SetEnableUnverifiedPaths toggles scanning of unverified agent paths.
func (s *AgentService) SetEnableUnverifiedPaths(enabled bool) {
	s.enableUnverifiedPaths = enabled
}

// GetCustomPaths returns all user-configured path overrides.
func (s *AgentService) GetCustomPaths() map[string]CustomAgentPaths {
	return s.customPaths
}

// SaveAgentPaths saves custom path overrides for a single agent.
func (s *AgentService) SaveAgentPaths(kind string, paths *CustomAgentPaths) error {
	if paths == nil {
		// Delete the entry
		delete(s.customPaths, kind)
	} else {
		s.customPaths[kind] = *paths
	}
	return saveCustomPaths(s.customPaths)
}

// --- custom path persistence ---

func customPathsFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".aimanager", "agent_paths.json")
}

func loadCustomPaths() map[string]CustomAgentPaths {
	result := make(map[string]CustomAgentPaths)
	path := customPathsFile()
	if path == "" {
		return result
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return result
	}
	var all AllCustomPaths
	if err := json.Unmarshal(data, &all); err != nil {
		return result
	}
	if all.Agents != nil {
		for k, v := range all.Agents {
			result[k] = v
		}
	}
	return result
}

func saveCustomPaths(paths map[string]CustomAgentPaths) error {
	path := customPathsFile()
	if path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	all := AllCustomPaths{Agents: paths}
	data, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// --- helpers ---

// applyCustomPaths merges user-configured path overrides onto default paths.
func applyCustomPaths(p agents.AgentPaths, custom CustomAgentPaths) agents.AgentPaths {
	if custom.ProviderConfigPath != "" {
		p.ProviderConfigPath = custom.ProviderConfigPath
	}
	if custom.SessionRootPath != "" {
		p.SessionRootPath = custom.SessionRootPath
	}
	if custom.SessionFormat != "" {
		p.SessionFormat = agents.SessionFormat(custom.SessionFormat)
	}
	if custom.MemoryRootPath != "" {
		p.MemoryRootPath = custom.MemoryRootPath
	}
	if custom.MemoryFormat != "" {
		p.MemoryFormat = agents.MemoryFormat(custom.MemoryFormat)
	}
	// Custom paths are always "verified" (user explicitly configured them)
	p.Status = agents.PathStatusVerified
	return p
}

func (s *AgentService) hasProviderConfig(kind agents.AgentKind, paths agents.AgentPaths) bool {
	adapter := providers.For(kind)
	if adapter == nil {
		return false
	}
	cfg, err := adapter.Read()
	if err != nil || cfg == nil {
		return false
	}
	return cfg.Exists
}

func (s *AgentService) hasSessions(kind agents.AgentKind, paths agents.AgentPaths) bool {
	scanner := sessions.ForWithPaths(kind, paths)
	if scanner == nil {
		return false
	}
	all, err := scanner.Scan()
	if err != nil {
		return false
	}
	return len(all) > 0
}

func (s *AgentService) hasMemory(kind agents.AgentKind, paths agents.AgentPaths) bool {
	scanner := memory.ForWithPaths(kind, paths)
	if scanner == nil {
		return false
	}
	all, err := scanner.Scan()
	if err != nil {
		return false
	}
	return len(all) > 0
}

func (s *AgentService) openInEditor(path string) error {
	if path == "" {
		return nil
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", "-a", "TextEdit", path)
	case "windows":
		cmd = exec.Command("notepad", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}

func agentLabel(kind agents.AgentKind) string {
	switch kind {
	case agents.AgentClaudeCode:
		return "Claude Code"
	case agents.AgentCodex:
		return "Codex CLI"
	case agents.AgentCursor:
		return "Cursor"
	case agents.AgentOpenCode:
		return "OpenCode"
	case agents.AgentGemini:
		return "Gemini CLI"
	case agents.AgentContinue:
		return "Continue.dev"
	case agents.AgentCline:
		return "Cline"
	case agents.AgentWindsurf:
		return "Windsurf"
	case agents.AgentOpenClaw:
		return "OpenClaw"
	case agents.AgentKiro:
		return "Kiro"
	case agents.AgentAmp:
		return "Amp"
	case agents.AgentGoose:
		return "Goose"
	case agents.AgentRooCode:
		return "Roo Code"
	case agents.AgentCopilot:
		return "GitHub Copilot"
	case agents.AgentAntigravity:
		return "Antigravity"
	case agents.AgentUniversal:
		return "Universal"
	default:
		return string(kind)
	}
}

func agentIcon(kind agents.AgentKind) string {
	switch kind {
	case agents.AgentClaudeCode:
		return "claude"
	case agents.AgentCodex:
		return "codex"
	case agents.AgentCursor:
		return "cursor"
	case agents.AgentOpenCode:
		return "opencode"
	case agents.AgentGemini:
		return "gemini"
	case agents.AgentContinue:
		return "continue"
	case agents.AgentCline:
		return "cline"
	case agents.AgentWindsurf:
		return "windsurf"
	case agents.AgentOpenClaw:
		return "openclaw"
	case agents.AgentKiro:
		return "kiro"
	case agents.AgentAmp:
		return "amp"
	case agents.AgentGoose:
		return "goose"
	case agents.AgentRooCode:
		return "roocode"
	case agents.AgentCopilot:
		return "copilot"
	case agents.AgentAntigravity:
		return "antigravity"
	default:
		return "agent"
	}
}

func containsStr(s, substr string) bool {
	return strings.Contains(s, substr)
}

func titleFromPath(path string) string {
	// Extract filename from path
	parts := splitPath(path)
	if len(parts) == 0 {
		return ""
	}
	name := parts[len(parts)-1]
	// Strip extension
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '.' {
			return name[:i]
		}
	}
	return name
}

func splitPath(path string) []string {
	var parts []string
	current := ""
	for _, c := range path {
		if c == '/' || c == '\\' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func parseTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, nil
}

func sortSessions(all []sessions.Session) {
	sort.Slice(all, func(i, j int) bool {
		return all[i].StartedAt.After(all[j].StartedAt)
	})
}
