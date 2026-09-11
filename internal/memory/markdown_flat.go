package memory

import (
	"os"
	"path/filepath"

	"ai-manager/internal/agents"
)

// markdownFlatScanner handles agents that store memory as a single flat
// Markdown file (e.g. Windsurf's global_rules.md).
//
// Used by: Windsurf.
type markdownFlatScanner struct {
	agent    agents.AgentKind
	rootPath string
}

// Agent returns the agent kind this scanner handles.
func (s *markdownFlatScanner) Agent() agents.AgentKind {
	return s.agent
}

// Scan reads the single memory file and returns it as one entry.
// Returns an empty slice (not error) if the file does not exist.
func (s *markdownFlatScanner) Scan() ([]MemoryEntry, error) {
	entries := make([]MemoryEntry, 0, 1)

	if s.rootPath == "" {
		return entries, nil
	}

	// Determine the actual file path: rootPath may be a directory
	// (containing global_rules.md) or the file itself.
	var filePath string
	info, err := os.Stat(s.rootPath)
	if err != nil {
		return entries, nil
	}

	if info.IsDir() {
		filePath = filepath.Join(s.rootPath, "global_rules.md")
	} else {
		filePath = s.rootPath
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return entries, nil
	}

	entries = append(entries, MemoryEntry{
		Agent:      s.agent,
		Path:       filePath,
		Title:      "global_rules",
		Kind:       "index",
		Content:    string(content),
		SizeBytes:  len(content),
		ModifiedAt: fileModTime(filePath),
	})
	return entries, nil
}

// ReadEntry reads the content of a single memory file.
func (s *markdownFlatScanner) ReadEntry(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
