// Package sessions: SQLite global database scanner for agents that store
// all sessions in a single SQLite file (OpenCode, Goose, OpenClaw, Kiro).
package sessions

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"ai-manager/internal/agents"
)

// sqliteScanner reads sessions from a global SQLite database.
type sqliteScanner struct {
	agent agents.AgentKind
	dbPath string
}

func newSQLiteScanner(agent agents.AgentKind, dbPath string) *sqliteScanner {
	return &sqliteScanner{agent: agent, dbPath: dbPath}
}

func (s *sqliteScanner) Agent() agents.AgentKind { return s.agent }

// openDB opens the SQLite database in read-only mode.
func (s *sqliteScanner) openDB() (*sql.DB, error) {
	if !fileExists(s.dbPath) {
		return nil, fmt.Errorf("database not found: %s", s.dbPath)
	}
	return sql.Open("sqlite", "file:"+s.dbPath+"?mode=ro")
}

// Scan returns all sessions from the SQLite database.
func (s *sqliteScanner) Scan() ([]Session, error) {
	if !fileExists(s.dbPath) {
		return []Session{}, nil
	}

	db, err := s.openDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Query sessions with message count
	rows, err := db.Query(`
		SELECT se.id, se.title, se.directory, se.time_created, se.time_updated,
		       se.tokens_input, se.tokens_output,
		       COUNT(msg.id) as msg_count
		FROM session se
		LEFT JOIN message msg ON msg.session_id = se.id
		GROUP BY se.id
		ORDER BY se.time_created DESC
		LIMIT 200
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var sess Session
		var title, directory string
		var timeCreated, timeUpdated int64
		var tokensIn, tokensOut int
		var msgCount int

		err := rows.Scan(&sess.ID, &title, &directory, &timeCreated, &timeUpdated, &tokensIn, &tokensOut, &msgCount)
		if err != nil {
			continue
		}

		sess.Agent = s.agent
		sess.Project = directory
		sess.Path = s.dbPath
		sess.Format = "sqlite"
		sess.MessageCount = msgCount
		sess.TokenEstimate = (tokensIn + tokensOut) / 4
		sess.StartedAt = unixToTime(timeCreated)
		sess.EndedAt = unixToTime(timeUpdated)
		sessions = append(sessions, sess)
	}

	return sessions, rows.Err()
}

// ReadTranscript reads the message content for a session and returns it
// as JSONL-like content that chatParser can parse.
func (s *sqliteScanner) ReadTranscript(id string) (string, error) {
	db, err := s.openDB()
	if err != nil {
		return "", err
	}
	defer db.Close()

	// Read messages for this session
	rows, err := db.Query(`
		SELECT msg.data
		FROM message msg
		WHERE msg.session_id = ?
		ORDER BY msg.time_created ASC
	`, id)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var lines []string
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			continue
		}

		// Parse the message data JSON and extract role + content
		var msg struct {
			Role    string          `json:"role"`
			Parts   []json.RawMessage `json:"parts"`
			Content json.RawMessage   `json:"content"`
		}
		if err := json.Unmarshal([]byte(data), &msg); err != nil {
			continue
		}

		// Build a JSONL line that chatParser can parse (Claude Code format)
		line := map[string]any{
			"type": "message",
			"message": map[string]any{
				"role": msg.Role,
			},
		}

		// Extract text content from parts or content
		var textParts []map[string]any
		if len(msg.Parts) > 0 {
			for _, p := range msg.Parts {
				var part struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}
				if json.Unmarshal(p, &part) == nil && part.Text != "" {
					textParts = append(textParts, map[string]any{
						"type": "text",
						"text": part.Text,
					})
				}
			}
		}
		if len(textParts) > 0 {
			line["message"].(map[string]any)["content"] = textParts
		}

		jsonLine, _ := json.Marshal(line)
		lines = append(lines, string(jsonLine))
	}

	return strings.Join(lines, "\n"), rows.Err()
}

// unixToTime converts a Unix timestamp (seconds or milliseconds) to time.Time.
func unixToTime(ts int64) time.Time {
	if ts == 0 {
		return time.Time{}
	}
	// If timestamp is in milliseconds (> 1e12), convert to seconds
	if ts > 1e12 {
		return time.Unix(ts/1000, 0)
	}
	return time.Unix(ts, 0)
}

// fileExists returns true if path is a regular file.
func fileExists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}
