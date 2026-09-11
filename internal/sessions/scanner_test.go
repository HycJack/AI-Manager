package sessions

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"ai-manager/internal/agents"
)

// TestJSONLScanner_Scan creates a temp dir tree with JSONL files and
// verifies the scanner returns the correct sessions.
func TestJSONLScanner_Scan(t *testing.T) {
	root := t.TempDir()

	// Two project directories with encoded slugs.
	projA := filepath.Join(root, "Users-yicaohuang-myproj")
	projB := filepath.Join(root, "Users-yicaohuang-other")
	for _, d := range []string{projA, projB} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Three JSONL files total.
	files := map[string]string{
		filepath.Join(projA, "session-a.jsonl"): "line1\nline2\nline3\n",
		filepath.Join(projA, "session-b.jsonl"): "one line\n",
		filepath.Join(projB, "session-c.jsonl"): "a\nb\nc\nd\n",
	}
	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	s := newJSONLScanner(agents.AgentClaudeCode, root)
	sessions, err := s.Scan()
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(sessions))
	}

	byID := map[string]Session{}
	for _, s := range sessions {
		byID[s.ID] = s
	}

	for _, tc := range []struct {
		id       string
		project  string
		messages int
	}{
		{"session-a", "/Users/yicaohuang/myproj", 3},
		{"session-b", "/Users/yicaohuang/myproj", 1},
		{"session-c", "/Users/yicaohuang/other", 4},
	} {
		sess, ok := byID[tc.id]
		if !ok {
			t.Fatalf("missing session %q", tc.id)
		}
		if sess.Agent != agents.AgentClaudeCode {
			t.Errorf("%s: agent = %s, want %s", tc.id, sess.Agent, agents.AgentClaudeCode)
		}
		if sess.Project != tc.project {
			t.Errorf("%s: project = %q, want %q", tc.id, sess.Project, tc.project)
		}
		if sess.MessageCount != tc.messages {
			t.Errorf("%s: MessageCount = %d, want %d", tc.id, sess.MessageCount, tc.messages)
		}
		if sess.Format != "jsonl" {
			t.Errorf("%s: Format = %q, want %q", tc.id, sess.Format, "jsonl")
		}
		if sess.Path == "" {
			t.Errorf("%s: Path empty", tc.id)
		}
		if sess.TokenEstimate <= 0 {
			t.Errorf("%s: TokenEstimate = %d, want > 0", tc.id, sess.TokenEstimate)
		}
		if sess.StartedAt.IsZero() {
			t.Errorf("%s: StartedAt is zero", tc.id)
		}
	}
}

// TestJSONLScanner_ReadTranscript verifies the raw file content is returned.
func TestJSONLScanner_ReadTranscript(t *testing.T) {
	root := t.TempDir()
	proj := filepath.Join(root, "Users-foo-bar")
	if err := os.MkdirAll(proj, 0755); err != nil {
		t.Fatal(err)
	}
	content := `{"type":"user","content":"hi"}` + "\n" +
		`{"type":"assistant","content":"hello"}` + "\n"
	if err := os.WriteFile(filepath.Join(proj, "my-session.jsonl"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	s := newJSONLScanner(agents.AgentClaudeCode, root)
	got, err := s.ReadTranscript("my-session")
	if err != nil {
		t.Fatal(err)
	}
	if got != content {
		t.Errorf("got %q, want %q", got, content)
	}

	// Non-existent ID returns an error.
	if _, err := s.ReadTranscript("nope"); err == nil {
		t.Error("expected error for missing session, got nil")
	}
}

// TestJSONLScanner_MissingDir verifies a missing root dir is handled gracefully.
func TestJSONLScanner_MissingDir(t *testing.T) {
	s := newJSONLScanner(agents.AgentClaudeCode, filepath.Join(t.TempDir(), "does-not-exist"))
	sessions, err := s.Scan()
	if err != nil {
		t.Fatalf("expected no error for missing dir, got %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(sessions))
	}
}

// TestProjectFromSlug verifies the slug-decoding helper across several
// inputs including bare names (no leading dash).
func TestProjectFromSlug(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"-Users-foo-bar", "/Users/foo/bar"},
		{"-home-foo-proj", "/home/foo/proj"},
		{"my-project", "/my/project"},
		{"simple", "/simple"},
		{"", "/"},
	}
	for _, tc := range cases {
		if got := projectFromSlug(tc.in); got != tc.want {
			t.Errorf("projectFromSlug(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestJSONScanner_Scan creates a temp dir with JSON files and verifies
// the scanner returns sessions while excluding index files.
func TestJSONScanner_Scan(t *testing.T) {
	root := t.TempDir()

	files := map[string]string{
		"session-1.json": "{}\n{}\n",  // 2 lines
		"session-2.json": "{\"k\":1}", // 1 line, no trailing newline
		"sessions.json":  "{}",        // index file, excluded
		"index.json":     "{}",        // index file, excluded
		"readme.md":      "not json",  // not .json, excluded
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	s := newJSONScanner(agents.AgentGemini, root)
	sessions, err := s.Scan()
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions (excluding index/non-json), got %d", len(sessions))
	}

	byID := map[string]Session{}
	for _, sess := range sessions {
		byID[sess.ID] = sess
	}

	if sess, ok := byID["session-1"]; ok {
		if sess.Agent != agents.AgentGemini {
			t.Errorf("agent = %s, want %s", sess.Agent, agents.AgentGemini)
		}
		if sess.MessageCount != 2 {
			t.Errorf("MessageCount = %d, want 2", sess.MessageCount)
		}
		if sess.Format != "json" {
			t.Errorf("Format = %q, want %q", sess.Format, "json")
		}
		if sess.Project != "" {
			t.Errorf("flat scanner should not derive Project, got %q", sess.Project)
		}
	} else {
		t.Error("missing session-1")
	}

	if sess, ok := byID["session-2"]; ok {
		if sess.MessageCount != 1 {
			t.Errorf("session-2 MessageCount = %d, want 1", sess.MessageCount)
		}
		if sess.TokenEstimate <= 0 {
			t.Errorf("TokenEstimate = %d, want > 0", sess.TokenEstimate)
		}
	} else {
		t.Error("missing session-2")
	}
}

// TestJSONScanner_MissingDir verifies a missing root dir is handled gracefully.
func TestJSONScanner_MissingDir(t *testing.T) {
	s := newJSONScanner(agents.AgentGemini, filepath.Join(t.TempDir(), "does-not-exist"))
	sessions, err := s.Scan()
	if err != nil {
		t.Fatalf("expected no error for missing dir, got %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(sessions))
	}
}

// TestJSONScanner_ReadTranscript verifies the raw content is returned.
func TestJSONScanner_ReadTranscript(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "sess-a.json"), []byte(`{"x":1}`), 0644); err != nil {
		t.Fatal(err)
	}
	s := newJSONScanner(agents.AgentContinue, root)
	got, err := s.ReadTranscript("sess-a")
	if err != nil {
		t.Fatal(err)
	}
	if got != `{"x":1}` {
		t.Errorf("got %q, want %q", got, `{"x":1}`)
	}
}

// TestNoopScanner verifies the T19 placeholder behaves as documented.
func TestNoopScanner(t *testing.T) {
	s := newNoopScanner(agents.AgentOpenCode, agents.SessionFormatSQLiteGlobal)
	if s.Agent() != agents.AgentOpenCode {
		t.Errorf("Agent() = %s, want %s", s.Agent(), agents.AgentOpenCode)
	}
	sessions, err := s.Scan()
	if err != nil {
		t.Fatalf("Scan() err = %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("Scan() returned %d sessions, want 0", len(sessions))
	}
	if _, err := s.ReadTranscript("any"); err == nil {
		t.Error("ReadTranscript should return error for no-op scanner")
	}
}

// TestFor_DispatchesByFormat verifies For() returns the correct scanner
// type for every agent based on its SessionFormat.
func TestFor_DispatchesByFormat(t *testing.T) {
	cases := []struct {
		agent agents.AgentKind
		want  string
	}{
		// JSONL per-project
		{agents.AgentClaudeCode, "jsonl"},
		{agents.AgentCodex, "jsonl"},

		// JSON flat
		{agents.AgentGemini, "json"},
		{agents.AgentContinue, "json"},
		{agents.AgentKiro, "json"},

		// SQLite global → no-op (T19)
		{agents.AgentOpenCode, "noop"},
		{agents.AgentWindsurf, "noop"},
		{agents.AgentOpenClaw, "noop"},
		{agents.AgentGoose, "noop"},

		// VS Code storage → no-op (T19). Roo Code has an empty
		// SessionRootPath so SupportsSession() is false → nil.
		{agents.AgentCursor, "noop"},
		{agents.AgentCopilot, "noop"},
		{agents.AgentCline, "noop"},
		{agents.AgentRooCode, "nil"},

		// No local session storage → nil
		{agents.AgentAmp, "nil"},
		{agents.AgentAntigravity, "nil"},
		{agents.AgentUniversal, "nil"},
		{agents.AgentPi, "nil"},
		{agents.AgentGrok, "nil"},
		{agents.AgentDroid, "nil"},
		{agents.AgentHermes, "nil"},
		{agents.AgentCcSwitch, "nil"},
	}
	for _, tc := range cases {
		t.Run(string(tc.agent), func(t *testing.T) {
			got := scannerKind(For(tc.agent))
			if got != tc.want {
				t.Errorf("For(%s) = %q, want %q", tc.agent, got, tc.want)
			}
		})
	}
}

// TestAllScanners returns a scanner for every agent that has local
// session storage (SessionFormat != none), and verifies no nil entries.
func TestAllScanners(t *testing.T) {
	all := AllScanners()
	if len(all) == 0 {
		t.Fatal("AllScanners returned empty")
	}
	seen := map[agents.AgentKind]bool{}
	for _, s := range all {
		if s == nil {
			t.Fatal("AllScanners contains nil")
		}
		if seen[s.Agent()] {
			t.Errorf("duplicate scanner for %s", s.Agent())
		}
		seen[s.Agent()] = true
	}
	// Spot-check that all JSONL/JSON agents are present.
	for _, want := range []agents.AgentKind{
		agents.AgentClaudeCode, agents.AgentCodex,
		agents.AgentGemini, agents.AgentContinue, agents.AgentKiro,
	} {
		if !seen[want] {
			t.Errorf("AllScanners missing scanner for %s", want)
		}
	}
}

// TestApplyFilter exercises the SessionFilter helpers.
func TestApplyFilter(t *testing.T) {
	now := time.Now()
	sessions := []Session{
		{Agent: agents.AgentClaudeCode, ID: "a", Project: "/p/one", StartedAt: now.Add(-3 * time.Hour)},
		{Agent: agents.AgentClaudeCode, ID: "b", Project: "/p/one", StartedAt: now.Add(-2 * time.Hour)},
		{Agent: agents.AgentClaudeCode, ID: "c", Project: "/p/two", StartedAt: now.Add(-1 * time.Hour)},
	}

	// Project filter.
	got := ApplyFilter(sessions, SessionFilter{Project: "/p/one"})
	if len(got) != 2 {
		t.Errorf("project filter: got %d, want 2", len(got))
	}

	// Since filter.
	got = ApplyFilter(sessions, SessionFilter{Since: now.Add(-2 * time.Hour)})
	if len(got) != 2 {
		t.Errorf("since filter: got %d, want 2", len(got))
	}

	// Limit.
	got = ApplyFilter(sessions, SessionFilter{Limit: 1})
	if len(got) != 1 {
		t.Errorf("limit filter: got %d, want 1", len(got))
	}

	// EffectiveLimit defaults to 50.
	var f SessionFilter
	if f.EffectiveLimit() != 50 {
		t.Errorf("EffectiveLimit = %d, want 50", f.EffectiveLimit())
	}
	var f7 SessionFilter
	f7.Limit = 7
	if f7.EffectiveLimit() != 7 {
		t.Errorf("EffectiveLimit = %d, want 7", f7.EffectiveLimit())
	}

	// Empty filter returns everything.
	got = ApplyFilter(sessions, SessionFilter{})
	if !reflect.DeepEqual(got, sessions) {
		t.Errorf("empty filter should return all; got %v", got)
	}
}

// TestCountLinesAndBytes verifies the line-count helper across edge cases.
func TestCountLinesAndBytes(t *testing.T) {
	cases := []struct {
		data      string
		wantLines int
		wantBytes int
	}{
		{"", 0, 0},
		{"a", 1, 1},
		{"a\n", 1, 2},
		{"a\nb", 2, 3},
		{"a\nb\n", 2, 4},
		{"a\n\nb", 3, 4},
	}
	for _, tc := range cases {
		lines, bytes := countLinesAndBytes([]byte(tc.data))
		if lines != tc.wantLines || bytes != tc.wantBytes {
			t.Errorf("countLinesAndBytes(%q) = (%d, %d), want (%d, %d)",
				tc.data, lines, bytes, tc.wantLines, tc.wantBytes)
		}
	}
}

// --- helpers ---

// scannerKind returns a short label for a SessionScanner, useful in tests.
func scannerKind(s SessionScanner) string {
	if s == nil {
		return "nil"
	}
	switch s.(type) {
	case *jsonlScanner:
		return "jsonl"
	case *jsonScanner:
		return "json"
	case *noopScanner:
		return "noop"
	default:
		return "unknown"
	}
}
