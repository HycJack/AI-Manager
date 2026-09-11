package memory

import (
	"os"
	"path/filepath"
	"testing"

	"ai-manager/internal/agents"
)

// --- TestMarkdownIndexScanner_WithMemoryMd ---

func TestMarkdownIndexScanner_WithMemoryMd(t *testing.T) {
	dir := t.TempDir()

	// Create MEMORY.md with [[topic1]] and [[topic2]] links.
	memoryMd := filepath.Join(dir, "MEMORY.md")
	os.WriteFile(memoryMd, []byte("# Memory Index\n\n[[topic1]]\n[[topic2]]\n"), 0o644)

	// Create topic files.
	topic1Path := filepath.Join(dir, "topic1.md")
	os.WriteFile(topic1Path, []byte("# Topic 1\n\nSome content."), 0o644)

	topic2Path := filepath.Join(dir, "topic2.md")
	os.WriteFile(topic2Path, []byte("# Topic 2\n\nMore content."), 0o644)

	scanner := &markdownIndexScanner{
		agent:    agents.AgentClaudeCode,
		rootPath: dir,
	}

	entries, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	// Verify MEMORY.md is index.
	entry := findEntry(entries, memoryMd)
	if entry == nil {
		t.Fatal("MEMORY.md entry not found")
	}
	if entry.Kind != "index" {
		t.Errorf("MEMORY.md kind = %q, want %q", entry.Kind, "index")
	}
	if entry.Title != "MEMORY" {
		t.Errorf("MEMORY.md title = %q, want %q", entry.Title, "MEMORY")
	}
	if entry.Agent != agents.AgentClaudeCode {
		t.Errorf("MEMORY.md agent = %q, want %q", entry.Agent, agents.AgentClaudeCode)
	}

	// Verify topic1.md is topic.
	entry = findEntry(entries, topic1Path)
	if entry == nil {
		t.Fatal("topic1.md entry not found")
	}
	if entry.Kind != "topic" {
		t.Errorf("topic1.md kind = %q, want %q", entry.Kind, "topic")
	}
	if entry.Title != "topic1" {
		t.Errorf("topic1.md title = %q, want %q", entry.Title, "topic1")
	}

	// Verify topic2.md is topic.
	entry = findEntry(entries, topic2Path)
	if entry == nil {
		t.Fatal("topic2.md entry not found")
	}
	if entry.Kind != "topic" {
		t.Errorf("topic2.md kind = %q, want %q", entry.Kind, "topic")
	}
	if entry.Title != "topic2" {
		t.Errorf("topic2.md title = %q, want %q", entry.Title, "topic2")
	}
}

// --- TestMarkdownIndexScanner_OpenClaw ---

func TestMarkdownIndexScanner_OpenClaw(t *testing.T) {
	dir := t.TempDir()

	// Create MEMORY.md (no topic links to keep the test focused).
	os.WriteFile(filepath.Join(dir, "MEMORY.md"), []byte("# OpenClaw Memory\n"), 0o644)

	// Create workspace files.
	os.WriteFile(filepath.Join(dir, "USER.md"), []byte("# User identity\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# Agent list\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "SOUL.md"), []byte("# Soul\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "IDENTITY.md"), []byte("# Identity\n"), 0o644)

	scanner := &markdownIndexScanner{
		agent:    agents.AgentOpenClaw,
		rootPath: dir,
	}

	entries, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}

	if len(entries) != 5 {
		t.Fatalf("expected 5 entries, got %d", len(entries))
	}

	// Verify each file and its kind.
	checks := map[string]string{
		"MEMORY.md":   "index",
		"USER.md":     "workspace",
		"AGENTS.md":   "workspace",
		"SOUL.md":     "workspace",
		"IDENTITY.md": "workspace",
	}
	for name, wantKind := range checks {
		entry := findEntry(entries, filepath.Join(dir, name))
		if entry == nil {
			t.Errorf("%s entry not found", name)
			continue
		}
		if entry.Kind != wantKind {
			t.Errorf("%s kind = %q, want %q", name, entry.Kind, wantKind)
		}
		if entry.Agent != agents.AgentOpenClaw {
			t.Errorf("%s agent = %q, want %q", name, entry.Agent, agents.AgentOpenClaw)
		}
	}
}

// TestMarkdownIndexScanner_OpenClaw_Daily verifies daily log scanning.
func TestMarkdownIndexScanner_OpenClaw_Daily(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "MEMORY.md"), []byte("# Memory\n"), 0o644)

	// Create daily files in memory/ subdirectory.
	memoryDir := filepath.Join(dir, "memory")
	if err := os.MkdirAll(memoryDir, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(memoryDir, "2024-01-01.md"), []byte("Day 1 content"), 0o644)
	os.WriteFile(filepath.Join(memoryDir, "2024-01-02.md"), []byte("Day 2 content"), 0o644)

	scanner := &markdownIndexScanner{
		agent:    agents.AgentOpenClaw,
		rootPath: dir,
	}

	entries, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}

	if len(entries) != 3 { // MEMORY.md + 2 daily
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	dailyCount := 0
	for _, entry := range entries {
		if entry.Kind == "daily" {
			dailyCount++
			if entry.Title == "2024-01-01" && entry.Content != "Day 1 content" {
				t.Errorf("daily entry content mismatch")
			}
		}
	}
	if dailyCount != 2 {
		t.Errorf("expected 2 daily entries, got %d", dailyCount)
	}
}

// --- TestMarkdownIndexScanner_Kiro ---

func TestMarkdownIndexScanner_Kiro(t *testing.T) {
	dir := t.TempDir()

	// Create preferences.md.
	prefPath := filepath.Join(dir, "preferences.md")
	os.WriteFile(prefPath, []byte("# User Preferences\n\n- Dark mode\n"), 0o644)

	// Create projects.md.
	projPath := filepath.Join(dir, "projects.md")
	os.WriteFile(projPath, []byte("# Projects\n\n- AI-Manager\n"), 0o644)

	// Create history/2024-01-01.md.
	historyDir := filepath.Join(dir, "history")
	if err := os.MkdirAll(historyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	historyPath := filepath.Join(historyDir, "2024-01-01.md")
	os.WriteFile(historyPath, []byte("# Daily History\n\n- Fixed a bug\n"), 0o644)

	scanner := &markdownIndexScanner{
		agent:    agents.AgentKiro,
		rootPath: dir,
	}

	entries, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	// Verify preferences.md.
	entry := findEntry(entries, prefPath)
	if entry == nil {
		t.Fatal("preferences.md entry not found")
	}
	if entry.Kind != "preference" {
		t.Errorf("preferences.md kind = %q, want %q", entry.Kind, "preference")
	}
	if entry.Title != "preferences" {
		t.Errorf("preferences.md title = %q, want %q", entry.Title, "preferences")
	}
	if entry.Agent != agents.AgentKiro {
		t.Errorf("preferences.md agent = %q, want %q", entry.Agent, agents.AgentKiro)
	}

	// Verify projects.md.
	entry = findEntry(entries, projPath)
	if entry == nil {
		t.Fatal("projects.md entry not found")
	}
	if entry.Kind != "topic" {
		t.Errorf("projects.md kind = %q, want %q", entry.Kind, "topic")
	}
	if entry.Title != "projects" {
		t.Errorf("projects.md title = %q, want %q", entry.Title, "projects")
	}

	// Verify history/2024-01-01.md.
	entry = findEntry(entries, historyPath)
	if entry == nil {
		t.Fatal("history/2024-01-01.md entry not found")
	}
	if entry.Kind != "daily" {
		t.Errorf("history/2024-01-01.md kind = %q, want %q", entry.Kind, "daily")
	}
	if entry.Title != "2024-01-01" {
		t.Errorf("history/2024-01-01.md title = %q, want %q", entry.Title, "2024-01-01")
	}
}

// --- TestMarkdownFlatScanner ---

func TestMarkdownFlatScanner(t *testing.T) {
	dir := t.TempDir()

	rulesContent := "# Global Rules\n\nAlways be helpful and concise."
	rulesPath := filepath.Join(dir, "global_rules.md")
	os.WriteFile(rulesPath, []byte(rulesContent), 0o644)

	scanner := &markdownFlatScanner{
		agent:    agents.AgentWindsurf,
		rootPath: dir,
	}

	entries, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Kind != "index" {
		t.Errorf("kind = %q, want %q", entry.Kind, "index")
	}
	if entry.Title != "global_rules" {
		t.Errorf("title = %q, want %q", entry.Title, "global_rules")
	}
	if entry.Content != rulesContent {
		t.Errorf("content = %q, want %q", entry.Content, rulesContent)
	}
	if entry.Agent != agents.AgentWindsurf {
		t.Errorf("agent = %q, want %q", entry.Agent, agents.AgentWindsurf)
	}
	if entry.Path != rulesPath {
		t.Errorf("path = %q, want %q", entry.Path, rulesPath)
	}
	if entry.SizeBytes != len(rulesContent) {
		t.Errorf("size_bytes = %d, want %d", entry.SizeBytes, len(rulesContent))
	}
}

// --- TestFor_DispatchesByFormat ---

func TestFor_DispatchesByFormat(t *testing.T) {
	// Markdown-index agents should return *markdownIndexScanner.
	indexAgents := []agents.AgentKind{
		agents.AgentClaudeCode,
		agents.AgentOpenClaw,
		agents.AgentKiro,
		agents.AgentGemini,
		agents.AgentGoose,
	}
	for _, agent := range indexAgents {
		s := For(agent)
		if s == nil {
			t.Errorf("For(%q) returned nil, want *markdownIndexScanner", agent)
			continue
		}
		if _, ok := s.(*markdownIndexScanner); !ok {
			t.Errorf("For(%q) returned %T, want *markdownIndexScanner", agent, s)
		}
		if s.Agent() != agent {
			t.Errorf("For(%q).Agent() = %q, want %q", agent, s.Agent(), agent)
		}
	}

	// Markdown-flat agents should return *markdownFlatScanner.
	s := For(agents.AgentWindsurf)
	if s == nil {
		t.Fatal("For(AgentWindsurf) returned nil, want *markdownFlatScanner")
	}
	if _, ok := s.(*markdownFlatScanner); !ok {
		t.Errorf("For(AgentWindsurf) returned %T, want *markdownFlatScanner", s)
	}
	if s.Agent() != agents.AgentWindsurf {
		t.Errorf("For(AgentWindsurf).Agent() = %q, want %q", s.Agent(), agents.AgentWindsurf)
	}

	// Agents with no local memory scanning should return nil.
	nilAgents := []agents.AgentKind{
		agents.AgentCodex,     // SQLite (T19)
		agents.AgentOpenCode,  // None
		agents.AgentContinue,  // None
		agents.AgentCursor,    // RulesOnly
		agents.AgentCopilot,   // ServerSide
		agents.AgentAmp,       // ServerSide
		agents.AgentUniversal, // NotSupported
		agents.AgentPi,        // NotSupported
		agents.AgentGrok,      // NotSupported
		agents.AgentAntigravity, // NotSupported
		agents.AgentDroid,     // NotSupported
		agents.AgentHermes,    // NotSupported
		agents.AgentCcSwitch,  // NotSupported
		agents.AgentCline,     // RulesOnly
		agents.AgentRooCode,   // None
	}
	for _, agent := range nilAgents {
		s := For(agent)
		if s != nil {
			t.Errorf("For(%q) returned %T, want nil", agent, s)
		}
	}
}

// --- TestAllScanners ---

func TestAllScanners(t *testing.T) {
	scanners := AllScanners()
	if len(scanners) == 0 {
		t.Fatal("AllScanners() returned empty list")
	}

	// Should include at least: ClaudeCode, OpenClaw, Kiro, Gemini, Goose (markdown-index)
	// and Windsurf (markdown-flat).
	if len(scanners) < 6 {
		t.Errorf("expected at least 6 scanners, got %d", len(scanners))
	}

	// Each scanner should have a valid agent.
	for _, s := range scanners {
		if s.Agent() == "" {
			t.Error("scanner with empty agent kind")
		}
	}
}

// --- Edge-case tests ---

func TestScan_MissingDirectory_ReturnsEmpty(t *testing.T) {
	scanner := &markdownIndexScanner{
		agent:    agents.AgentClaudeCode,
		rootPath: filepath.Join(t.TempDir(), "nonexistent"),
	}

	entries, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() on missing dir returned error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for missing dir, got %d", len(entries))
	}
}

func TestScan_EmptyRootPath_ReturnsEmpty(t *testing.T) {
	scanner := &markdownIndexScanner{
		agent:    agents.AgentClaudeCode,
		rootPath: "",
	}

	entries, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() with empty root returned error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for empty root, got %d", len(entries))
	}
}

func TestMarkdownIndexScanner_FileAsRootPath(t *testing.T) {
	dir := t.TempDir()

	filePath := filepath.Join(dir, "MEMORY.md")
	content := "# Memory File\n\n[[topic]]\n"
	os.WriteFile(filePath, []byte(content), 0o644)

	scanner := &markdownIndexScanner{
		agent:    agents.AgentClaudeCode,
		rootPath: filePath, // point to file, not directory
	}

	entries, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Kind != "index" {
		t.Errorf("kind = %q, want %q", entries[0].Kind, "index")
	}
	if entries[0].Content != content {
		t.Errorf("content mismatch")
	}
}

// --- ReadEntry tests ---

func TestMarkdownIndexScanner_ReadEntry(t *testing.T) {
	dir := t.TempDir()

	filePath := filepath.Join(dir, "topic.md")
	expected := "# Test Topic\n\nThis is the content."
	os.WriteFile(filePath, []byte(expected), 0o644)

	scanner := &markdownIndexScanner{
		agent:    agents.AgentClaudeCode,
		rootPath: dir,
	}

	result, err := scanner.ReadEntry(filePath)
	if err != nil {
		t.Fatalf("ReadEntry() returned error: %v", err)
	}
	if result != expected {
		t.Errorf("ReadEntry() = %q, want %q", result, expected)
	}
}

func TestMarkdownFlatScanner_ReadEntry(t *testing.T) {
	dir := t.TempDir()

	filePath := filepath.Join(dir, "global_rules.md")
	expected := "# Rules\n\nAlways test."
	os.WriteFile(filePath, []byte(expected), 0o644)

	scanner := &markdownFlatScanner{
		agent:    agents.AgentWindsurf,
		rootPath: dir,
	}

	result, err := scanner.ReadEntry(filePath)
	if err != nil {
		t.Fatalf("ReadEntry() returned error: %v", err)
	}
	if result != expected {
		t.Errorf("ReadEntry() = %q, want %q", result, expected)
	}
}

func TestReadEntry_NonexistentFile_ReturnsError(t *testing.T) {
	dir := t.TempDir()

	scanner := &markdownIndexScanner{
		agent:    agents.AgentClaudeCode,
		rootPath: dir,
	}

	_, err := scanner.ReadEntry(filepath.Join(dir, "nope.md"))
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

// --- extractTopics unit test ---

func TestExtractTopics(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		want     []string
	}{
		{
			name:    "double-bracket links",
			content: "Hello [[topic1]] and [[topic2]]",
			want:    []string{"topic1", "topic2"},
		},
		{
			name:    "single-bracket links",
			content: "See [topic1] for details",
			want:    []string{"topic1"},
		},
		{
			name:    "deduplicated",
			content: "[[topic1]] [[topic1]]",
			want:    []string{"topic1"},
		},
		{
			name:    "no links",
			content: "Just plain text",
			want:    []string{},
		},
		{
			name:    "empty content",
			content: "",
			want:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTopics(tt.content)
			if len(got) != len(tt.want) {
				t.Fatalf("extractTopics() = %v (len %d), want %v (len %d)", got, len(got), tt.want, len(tt.want))
			}
			for i, w := range tt.want {
				if got[i] != w {
					t.Errorf("extractTopics()[%d] = %q, want %q", i, got[i], w)
				}
			}
		})
	}
}

// --- helper ---

func findEntry(entries []MemoryEntry, path string) *MemoryEntry {
	for i := range entries {
		if entries[i].Path == path {
			return &entries[i]
		}
	}
	return nil
}
