package providers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ai-manager/internal/agents"
)

// TestRegistry_AllAdaptersRegistered verifies all adapters are registered.
func TestRegistry_AllAdaptersRegistered(t *testing.T) {
	expected := []agents.AgentKind{
		agents.AgentClaudeCode, agents.AgentCodex, agents.AgentOpenCode,
		agents.AgentGemini, agents.AgentContinue, agents.AgentWindsurf,
		agents.AgentOpenClaw, agents.AgentKiro, agents.AgentAmp,
	}
	for _, kind := range expected {
		if For(kind) == nil {
			t.Errorf("adapter not registered for %q", kind)
		}
	}
}

// TestRegistry_UnregisteredAgent returns nil for agents without adapters.
func TestRegistry_UnregisteredAgent(t *testing.T) {
	if For(agents.AgentCursor) != nil {
		t.Error("Cursor should not have an adapter (T14)")
	}
	if For(agents.AgentAntigravity) != nil {
		t.Error("Antigravity should not have an adapter")
	}
}

// TestMaskAPIKey verifies API key masking.
func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{"", ""},
		{"abc", "****"},
		{"abcd", "****"},
		{"sk-ant-api03-abcdefghij1234567890", "****7890"},
		{"sk-proj-abcdefghijklmnopqrstuvwxyz", "****wxyz"},
	}
	for _, tt := range tests {
		got := MaskAPIKey(tt.key)
		if got != tt.want {
			t.Errorf("MaskAPIKey(%q) = %q, want %q", tt.key, got, tt.want)
		}
	}
}

// TestIsMaskedAPIKey verifies masked key detection.
func TestIsMaskedAPIKey(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{"", false},
		{"sk-ant-api03-abc", false},
		{"****", true},
		{"****abcd", true},
		{"not masked", false},
	}
	for _, tt := range tests {
		got := IsMaskedAPIKey(tt.key)
		if got != tt.want {
			t.Errorf("IsMaskedAPIKey(%q) = %v, want %v", tt.key, got, tt.want)
		}
	}
}

// TestFileExists verifies file existence check.
func TestFileExists(t *testing.T) {
	if fileExists("") {
		t.Error("fileExists(\"\") should be false")
	}
	if fileExists("/nonexistent/path/xyz") {
		t.Error("fileExists(nonexistent) should be false")
	}
	tmp := t.TempDir()
	f := filepath.Join(tmp, "test.txt")
	if err := os.WriteFile(f, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !fileExists(f) {
		t.Error("fileExists(existing) should be true")
	}
}

// TestExpandHome verifies ~ expansion.
func TestExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input string
		want  string
	}{
		{"~/.claude/settings.json", home + "/.claude/settings.json"},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}
	for _, tt := range tests {
		got := expandHome(tt.input)
		if got != tt.want {
			t.Errorf("expandHome(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// TestInferProviderFromJSON verifies provider inference from JSON.
func TestInferProviderFromJSON(t *testing.T) {
	tests := []struct {
		data map[string]any
		want string
	}{
		{
			data: map[string]any{"provider": "anthropic"},
			want: "anthropic",
		},
		{
			data: map[string]any{"baseUrl": "https://api.anthropic.com"},
			want: "anthropic",
		},
		{
			data: map[string]any{"baseUrl": "https://api.openai.com"},
			want: "openai",
		},
		{
			data: map[string]any{"apiKey": "sk-123"},
			want: "openai", // default when no baseUrl
		},
		{
			data: map[string]any{},
			want: "unknown",
		},
	}
	for _, tt := range tests {
		got := inferProviderFromJSON(tt.data)
		if got != tt.want {
			t.Errorf("inferProviderFromJSON(%v) = %q, want %q", tt.data, got, tt.want)
		}
	}
}

// TestInferProviderFromRaw verifies provider inference from raw content.
func TestInferProviderFromRaw(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{"model_provider = \"anthropic\"", "anthropic"},
		{"base_url = \"https://api.openai.com\"", "openai"},
		{"api_key = \"sk-deepseek-123\"", "deepseek"},
		{"model = \"gemini-pro\"", "google"},
		{"model = \"gpt-4\"", "unknown"}, // no explicit provider
	}
	for _, tt := range tests {
		got := inferProviderFromRaw(tt.raw)
		if got != tt.want {
			t.Errorf("inferProviderFromRaw(%q) = %q, want %q", tt.raw, got, tt.want)
		}
	}
}

// TestExtractTOMLValue verifies TOML value extraction.
func TestExtractTOMLValue(t *testing.T) {
	tests := []struct {
		raw  string
		keys []string
		want string
	}{
		{"model = \"gpt-4o\"", []string{"model"}, "gpt-4o"},
		{"api_key = 'sk-123'", []string{"api_key"}, "sk-123"},
		{"base_url = \"https://api.openai.com/v1\"", []string{"base_url"}, "https://api.openai.com/v1"},
		{"# model = \"commented\"", []string{"model"}, ""},
		{"model = \"gpt-4o\"\nother = \"value\"", []string{"model"}, "gpt-4o"},
	}
	for _, tt := range tests {
		got := extractTOMLValue(tt.raw, tt.keys...)
		if got != tt.want {
			t.Errorf("extractTOMLValue(%q, %v) = %q, want %q", tt.raw, tt.keys, got, tt.want)
		}
	}
}

// TestExtractYAMLValue verifies YAML value extraction.
func TestExtractYAMLValue(t *testing.T) {
	tests := []struct {
		raw  string
		keys []string
		want string
	}{
		{"model: gpt-4o", []string{"model"}, "gpt-4o"},
		{`apiKey: "sk-123"`, []string{"apiKey"}, "sk-123"},
		{`base_url: 'https://api.openai.com'`, []string{"base_url"}, "https://api.openai.com"},
		{"model: gpt-4o\nother: value", []string{"model"}, "gpt-4o"},
		{"other: value", []string{"model"}, ""},
	}
	for _, tt := range tests {
		got := extractYAMLValue(tt.raw, tt.keys...)
		if got != tt.want {
			t.Errorf("extractYAMLValue(%q, %v) = %q, want %q", tt.raw, tt.keys, got, tt.want)
		}
	}
}

// TestClaudeCodeAdapter_Read_NotExists verifies Read() when file doesn't exist.
func TestClaudeCodeAdapter_Read_NotExists(t *testing.T) {
	a := &claudeCodeAdapter{}
	cfg, err := a.Read()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Agent != agents.AgentClaudeCode {
		t.Errorf("Agent = %q, want %q", cfg.Agent, agents.AgentClaudeCode)
	}
	// May or may not exist depending on the test machine
	if cfg.Format != "json" {
		t.Errorf("Format = %q, want json", cfg.Format)
	}
}

// TestClaudeCodeAdapter_Read_WithFile verifies Read() with a temp config file.
func TestClaudeCodeAdapter_Read_WithFile(t *testing.T) {
	// Create a temp config file at the expected path
	tmp := t.TempDir()
	// Override home dir for this test
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmp)

	settingsPath := filepath.Join(tmp, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatal(err)
	}
	content := `{"model":"claude-3.5-sonnet","apiKey":"sk-ant-api03-abcdef123456","baseUrl":"https://api.anthropic.com"}`
	if err := os.WriteFile(settingsPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	a := &claudeCodeAdapter{}
	cfg, err := a.Read()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Exists {
		t.Fatal("Exists should be true")
	}
	if cfg.Model != "claude-3.5-sonnet" {
		t.Errorf("Model = %q, want claude-3.5-sonnet", cfg.Model)
	}
	if !strings.Contains(cfg.APIKey, "3456") {
		t.Errorf("APIKey = %q, want to contain 3456 (masked)", cfg.APIKey)
	}
	if cfg.BaseURL != "https://api.anthropic.com" {
		t.Errorf("BaseURL = %q, want https://api.anthropic.com", cfg.BaseURL)
	}

	_ = origHome
}

// TestCodexAdapter_Read_NotExists verifies Read() when file doesn't exist.
func TestCodexAdapter_Read_NotExists(t *testing.T) {
	a := &codexAdapter{}
	cfg, err := a.Read()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Agent != agents.AgentCodex {
		t.Errorf("Agent = %q, want %q", cfg.Agent, agents.AgentCodex)
	}
	if cfg.Format != "toml" {
		t.Errorf("Format = %q, want toml", cfg.Format)
	}
}

// TestCodexAdapter_Read_WithFile verifies Read() with a temp TOML file.
func TestCodexAdapter_Read_WithFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	configPath := filepath.Join(tmp, ".codex", "config.toml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	content := `model = "gpt-4o"
api_key = "sk-proj-abcdefghij"
base_url = "https://api.openai.com/v1"
`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	a := &codexAdapter{}
	cfg, err := a.Read()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Exists {
		t.Fatal("Exists should be true")
	}
	if cfg.Model != "gpt-4o" {
		t.Errorf("Model = %q, want gpt-4o", cfg.Model)
	}
	if cfg.BaseURL != "https://api.openai.com/v1" {
		t.Errorf("BaseURL = %q, want https://api.openai.com/v1", cfg.BaseURL)
	}
}

// TestContinueAdapter_Read_WithFile verifies Read() with a temp YAML file.
func TestContinueAdapter_Read_WithFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	configPath := filepath.Join(tmp, ".continue", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	content := `model: gpt-4o
apiKey: sk-proj-1234567890
baseUrl: https://api.openai.com/v1
`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	a := &continueAdapter{}
	cfg, err := a.Read()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Exists {
		t.Fatal("Exists should be true")
	}
	if cfg.Model != "gpt-4o" {
		t.Errorf("Model = %q, want gpt-4o", cfg.Model)
	}
	if cfg.BaseURL != "https://api.openai.com/v1" {
		t.Errorf("BaseURL = %q, want https://api.openai.com/v1", cfg.BaseURL)
	}
}

// TestProviderConfig_WriteJSON verifies JSON write round-trip.
func TestProviderConfig_WriteJSON(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	settingsPath := filepath.Join(tmp, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatal(err)
	}
	initial := `{"model":"claude-3","apiKey":"sk-ant-1234"}`
	if err := os.WriteFile(settingsPath, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	a := &claudeCodeAdapter{}
	cfg, err := a.Read()
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}

	// Modify model and write back
	cfg.Model = "claude-3.5-sonnet"
	if err := a.Write(cfg); err != nil {
		t.Fatalf("Write error: %v", err)
	}

	// Read back and verify
	cfg2, err := a.Read()
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
	if cfg2.Model != "claude-3.5-sonnet" {
		t.Errorf("Model = %q, want claude-3.5-sonnet", cfg2.Model)
	}
	// API key should be preserved (masked value written back)
	if !strings.Contains(cfg2.APIKey, "1234") {
		t.Errorf("APIKey = %q, want to contain 1234", cfg2.APIKey)
	}
}

// TestProviderConfig_WriteJSON_NewKey verifies writing a new API key.
func TestProviderConfig_WriteJSON_NewKey(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	settingsPath := filepath.Join(tmp, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatal(err)
	}
	initial := `{"model":"claude-3"}`
	if err := os.WriteFile(settingsPath, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	a := &claudeCodeAdapter{}
	cfg, err := a.Read()
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}

	// Set a new (unmasked) API key
	cfg.APIKey = "sk-ant-api03-newkey1234567890"
	if err := a.Write(cfg); err != nil {
		t.Fatalf("Write error: %v", err)
	}

	cfg2, err := a.Read()
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
	// Should be masked with new key's last 4 chars
	if !strings.Contains(cfg2.APIKey, "7890") {
		t.Errorf("APIKey = %q, want to contain 7890 (masked)", cfg2.APIKey)
	}
}

// TestWindsurfAdapter_Read_WithFile verifies Windsurf JSON adapter.
func TestWindsurfAdapter_Read_WithFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	configPath := filepath.Join(tmp, ".codeium", "windsurf", "settings.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	content := `{"model":"gpt-4","apiKey":"sk-windsurf-123456"}`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	a := &windsurfAdapter{}
	cfg, err := a.Read()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Exists {
		t.Fatal("Exists should be true")
	}
	if cfg.Model != "gpt-4" {
		t.Errorf("Model = %q, want gpt-4", cfg.Model)
	}
}

// TestOpenClawAdapter_Read_WithFile verifies OpenClaw TOML adapter.
func TestOpenClawAdapter_Read_WithFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	configPath := filepath.Join(tmp, ".openclaw", "config.toml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	content := `model = "gpt-4o"
api_key = "sk-openclaw-123456"
`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	a := &openClawAdapter{}
	cfg, err := a.Read()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Exists {
		t.Fatal("Exists should be true")
	}
	if cfg.Model != "gpt-4o" {
		t.Errorf("Model = %q, want gpt-4o", cfg.Model)
	}
}

// TestKiroAdapter_Read_WithFile verifies Kiro YAML adapter.
func TestKiroAdapter_Read_WithFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	configPath := filepath.Join(tmp, ".kiro", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	content := `model: gpt-4o
apiKey: sk-kiro-123456
`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	a := &kiroAdapter{}
	cfg, err := a.Read()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Exists {
		t.Fatal("Exists should be true")
	}
	if cfg.Model != "gpt-4o" {
		t.Errorf("Model = %q, want gpt-4o", cfg.Model)
	}
}

// TestAmpAdapter_Read_WithFile verifies Amp JSON adapter.
func TestAmpAdapter_Read_WithFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	configPath := filepath.Join(tmp, ".config", "amp", "config.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	content := `{"model":"claude-3.5-sonnet","apiKey":"sk-amp-123456"}`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	a := &ampAdapter{}
	cfg, err := a.Read()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Exists {
		t.Fatal("Exists should be true")
	}
	if cfg.Model != "claude-3.5-sonnet" {
		t.Errorf("Model = %q, want claude-3.5-sonnet", cfg.Model)
	}
}

// TestGeminiAdapter_Read_WithFile verifies Gemini JSON adapter.
func TestGeminiAdapter_Read_WithFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	configPath := filepath.Join(tmp, ".gemini", "settings.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	content := `{"model":"gemini-2.0-pro","apiKey":"sk-gemini-123456"}`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	a := &geminiAdapter{}
	cfg, err := a.Read()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Exists {
		t.Fatal("Exists should be true")
	}
	if cfg.Model != "gemini-2.0-pro" {
		t.Errorf("Model = %q, want gemini-2.0-pro", cfg.Model)
	}
}
