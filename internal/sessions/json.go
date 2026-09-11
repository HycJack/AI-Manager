// json.go — JSON flat scanner (Gemini CLI, Continue, Kiro).
//
// Layout: <root>/*.json — each .json file is one session. Index files
// (sessions.json, index.json) are excluded; they list other sessions
// rather than being sessions themselves.

package sessions

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ai-manager/internal/agents"
)

// jsonScanner scans sessions stored as flat JSON files (Gemini CLI, Continue, Kiro).
type jsonScanner struct {
	agent    agents.AgentKind
	rootPath string
}

func newJSONScanner(agent agents.AgentKind, rootPath string) *jsonScanner {
	return &jsonScanner{agent: agent, rootPath: rootPath}
}

// Agent implements SessionScanner.
func (s *jsonScanner) Agent() agents.AgentKind { return s.agent }

// indexFiles are flat JSON files that are indexes/manifests rather than
// individual sessions.
var indexFiles = map[string]bool{
	"sessions.json": true,
	"session.json":  true, // singular variant
	"index.json":    true,
	"manifest.json": true,
	"meta.json":     true,
	"config.json":   true,
	"settings.json": true,
}

// Scan walks rootPath and collects one Session per *.json file found,
// excluding known index files. A missing rootPath returns an empty list.
func (s *jsonScanner) Scan() ([]Session, error) {
	if _, err := os.Stat(s.rootPath); err != nil {
		return []Session{}, nil
	}

	var sessions []Session
	err := filepath.Walk(s.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".json") {
			return nil
		}
		if indexFiles[info.Name()] {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		lines, chars := countLinesAndBytes(data)

		sessions = append(sessions, Session{
			Agent:         s.agent,
			ID:            strings.TrimSuffix(info.Name(), filepath.Ext(info.Name())),
			Project:       "", // flat format: no per-project directory
			StartedAt:     info.ModTime(),
			MessageCount:  lines,
			TokenEstimate: chars / 4,
			Path:          path,
			Format:        "json",
		})
		return nil
	})
	if err != nil {
		return sessions, err
	}
	return sessions, nil
}

// ReadTranscript reads the raw content of a session by ID, searching
// all subdirectories for a matching *.json file (excluding index files).
func (s *jsonScanner) ReadTranscript(id string) (string, error) {
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

// findPath returns the on-disk path of a *.json file whose basename
// (without extension) matches id, or "" if none exists.
func (s *jsonScanner) findPath(id string) (string, error) {
	var found string
	err := filepath.Walk(s.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".json") || indexFiles[info.Name()] {
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
