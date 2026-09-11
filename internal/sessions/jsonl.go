// jsonl.go — JSONL per-project scanner (Claude Code, Codex).
//
// Layout: <root>/<encoded-project-path>/*.jsonl
// The encoded project path is the on-disk slug form of the project
// path (slashes → dashes, leading "/" becomes a leading "-"). Reverse
// by replacing "-" with "/" (the leading dash yields the leading slash).

package sessions

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ai-manager/internal/agents"
)

// jsonlScanner scans sessions stored as JSONL files in per-project
// subdirectories (Claude Code, Codex).
type jsonlScanner struct {
	agent    agents.AgentKind
	rootPath string
}

func newJSONLScanner(agent agents.AgentKind, rootPath string) *jsonlScanner {
	return &jsonlScanner{agent: agent, rootPath: rootPath}
}

// Agent implements SessionScanner.
func (s *jsonlScanner) Agent() agents.AgentKind { return s.agent }

// Scan walks rootPath and collects one Session per *.jsonl file found.
// A missing rootPath returns an empty list, not an error.
func (s *jsonlScanner) Scan() ([]Session, error) {
	if _, err := os.Stat(s.rootPath); err != nil {
		return []Session{}, nil
	}

	var sessions []Session
	err := filepath.Walk(s.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".jsonl") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			// Skip files we can't read; don't abort the whole walk.
			return nil
		}
		lines, chars := countLinesAndBytes(data)

		sessions = append(sessions, Session{
			Agent:         s.agent,
			ID:            strings.TrimSuffix(info.Name(), filepath.Ext(info.Name())),
			Project:       projectFromSlug(filepath.Base(filepath.Dir(path))),
			StartedAt:     info.ModTime(),
			MessageCount:  lines,
			TokenEstimate: chars / 4,
			Path:          path,
			Format:        "jsonl",
		})
		return nil
	})
	if err != nil {
		return sessions, err
	}
	return sessions, nil
}

// ReadTranscript reads the raw content of a session by ID, searching
// all subdirectories for a matching *.jsonl file.
func (s *jsonlScanner) ReadTranscript(id string) (string, error) {
	found, err := s.findPath(id)
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("session %q not found", id)
	}
	data, err := os.ReadFile(found)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// findPath returns the on-disk path of a *.jsonl file whose basename
// (without extension) matches id, or "" if none exists.
func (s *jsonlScanner) findPath(id string) (string, error) {
	var found string
	err := filepath.Walk(s.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".jsonl") {
			return nil
		}
		base := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
		if base == id && found == "" {
			found = path
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return found, nil
}

// projectFromSlug reverses the on-disk slug encoding of a project path:
// replace dashes with slashes, then ensure a leading slash.
//
// Examples:
//
//	"-Users-foo-bar" → "/Users/foo/bar"  (Claude Code style)
//	"my-project"     → "/my-project"     (bare name → absolute-ish)
func projectFromSlug(slug string) string {
	path := strings.ReplaceAll(slug, "-", "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

// countLinesAndBytes returns (lineCount, byteCount) for data.
// A trailing newline does not start a new empty line, but a non-empty
// final line without a trailing newline is counted.
func countLinesAndBytes(data []byte) (int, int) {
	if len(data) == 0 {
		return 0, 0
	}
	newlines := strings.Count(string(data), "\n")
	lines := newlines
	if data[len(data)-1] != '\n' {
		lines++ // final unterminated line counts
	}
	return lines, len(data)
}
