package memory

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"ai-manager/internal/agents"
)

// topicLinkRe extracts topic names from [[topic]] wiki-links.
// Matches [text] where text contains no closing bracket.
var topicLinkRe = regexp.MustCompile(`\[([^\]]+)\]`)

// markdownIndexScanner handles agents that use a MEMORY.md index with
// [[topic]] links (Obsidian style) to reference topic files.
//
// Used by: Claude Code, OpenClaw, Kiro, Gemini CLI.
type markdownIndexScanner struct {
	agent    agents.AgentKind
	rootPath string
}

// Agent returns the agent kind this scanner handles.
func (s *markdownIndexScanner) Agent() agents.AgentKind {
	return s.agent
}

// Scan returns all memory entries found under the root path.
// Returns an empty slice (not error) if the root path does not exist.
func (s *markdownIndexScanner) Scan() ([]MemoryEntry, error) {
	entries := make([]MemoryEntry, 0, 8)

	if s.rootPath == "" {
		return entries, nil
	}

	info, err := os.Stat(s.rootPath)
	if err != nil {
		// Root path does not exist — return empty list gracefully.
		return entries, nil
	}

	if !info.IsDir() {
		// Root path is a file: read as a single entry.
		content, err := os.ReadFile(s.rootPath)
		if err != nil {
			return entries, nil
		}
		entries = append(entries, MemoryEntry{
			Agent:      s.agent,
			Path:       s.rootPath,
			Title:      strings.TrimSuffix(filepath.Base(s.rootPath), ".md"),
			Kind:       "index",
			Content:    string(content),
			SizeBytes:  len(content),
			ModifiedAt: fileModTime(s.rootPath),
		})
		return entries, nil
	}

	// --- Directory mode ---

	// 1. Check for MEMORY.md — parse as index, extract [[topic]] links.
	memoryMdPath := filepath.Join(s.rootPath, "MEMORY.md")
	if content, err := os.ReadFile(memoryMdPath); err == nil {
		text := string(content)
		entries = append(entries, MemoryEntry{
			Agent:      s.agent,
			Path:       memoryMdPath,
			Title:      "MEMORY",
			Kind:       "index",
			Content:    text,
			SizeBytes:  len(content),
			ModifiedAt: fileModTime(memoryMdPath),
		})

		// Extract [[topic]] links and scan for topic files.
		topics := extractTopics(text)
		for _, topic := range topics {
			topicPath := filepath.Join(s.rootPath, topic+".md")
			if tc, err := os.ReadFile(topicPath); err == nil {
				entries = append(entries, MemoryEntry{
					Agent:      s.agent,
					Path:       topicPath,
					Title:      topic,
					Kind:       "topic",
					Content:    string(tc),
					SizeBytes:  len(tc),
					ModifiedAt: fileModTime(topicPath),
				})
			}
		}
	}

	// 2. Agent-specific additional files.
	switch s.agent {
	case agents.AgentOpenClaw:
		// USER.md, AGENTS.md, SOUL.md, IDENTITY.md as workspace files.
		for _, name := range []string{"USER.md", "AGENTS.md", "SOUL.md", "IDENTITY.md"} {
			p := filepath.Join(s.rootPath, name)
			if content, err := os.ReadFile(p); err == nil {
				entries = append(entries, MemoryEntry{
					Agent:      s.agent,
					Path:       p,
					Title:      strings.TrimSuffix(name, ".md"),
					Kind:       "workspace",
					Content:    string(content),
					SizeBytes:  len(content),
					ModifiedAt: fileModTime(p),
				})
			}
		}
		// memory/YYYY-MM-DD.md daily logs.
		scanMDFiles(filepath.Join(s.rootPath, "memory"), s.agent, "daily", &entries)

	case agents.AgentKiro:
		// preferences.md
		prefPath := filepath.Join(s.rootPath, "preferences.md")
		if content, err := os.ReadFile(prefPath); err == nil {
			entries = append(entries, MemoryEntry{
				Agent:      s.agent,
				Path:       prefPath,
				Title:      "preferences",
				Kind:       "preference",
				Content:    string(content),
				SizeBytes:  len(content),
				ModifiedAt: fileModTime(prefPath),
			})
		}
		// projects.md
		projPath := filepath.Join(s.rootPath, "projects.md")
		if content, err := os.ReadFile(projPath); err == nil {
			entries = append(entries, MemoryEntry{
				Agent:      s.agent,
				Path:       projPath,
				Title:      "projects",
				Kind:       "topic",
				Content:    string(content),
				SizeBytes:  len(content),
				ModifiedAt: fileModTime(projPath),
			})
		}
		// history/*.md
		scanMDFiles(filepath.Join(s.rootPath, "history"), s.agent, "daily", &entries)
	}

	return entries, nil
}

// ReadEntry reads the content of a single memory file.
func (s *markdownIndexScanner) ReadEntry(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// extractTopics parses [[topic]] wiki-links from content and returns deduplicated topic names.
func extractTopics(content string) []string {
	matches := topicLinkRe.FindAllStringSubmatch(content, -1)
	seen := make(map[string]bool, len(matches))
	topics := make([]string, 0, len(matches))

	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		// The regex \[([^\]]+)\] on [[topic]] captures [topic (includes 2nd bracket).
		// Strip any leading '[' characters.
		topic := strings.TrimLeft(m[1], "[")
		if topic == "" {
			continue
		}
		if !seen[topic] {
			seen[topic] = true
			topics = append(topics, topic)
		}
	}
	return topics
}

// scanMDFiles reads all .md files in a directory and appends them as entries.
func scanMDFiles(dir string, agent agents.AgentKind, kind string, entries *[]MemoryEntry) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".md") {
			continue
		}
		p := filepath.Join(dir, f.Name())
		content, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		*entries = append(*entries, MemoryEntry{
			Agent:      agent,
			Path:       p,
			Title:      strings.TrimSuffix(f.Name(), ".md"),
			Kind:       kind,
			Content:    string(content),
			SizeBytes:  len(content),
			ModifiedAt: fileModTime(p),
		})
	}
}
