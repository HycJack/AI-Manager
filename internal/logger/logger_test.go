package logger

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// newRotator builds a rotatingWriter with a tiny maxSize (in MB) so tests can
// trigger rotation without writing megabytes.
func newRotator(t *testing.T, maxSizeMB, maxBackups int) (*rotatingWriter, string) {
	t.Helper()
	dir := t.TempDir()
	// Mirror the real layout (dataDir/logs/app.log) so ReadRecent works.
	err := os.MkdirAll(filepath.Join(dir, "logs"), 0o755)
	if err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	path := filepath.Join(dir, "logs", "app.log")
	w := &rotatingWriter{
		filename:   path,
		maxSize:    maxSizeMB,
		maxBackups: maxBackups,
		maxAge:     30,
	}
	// Ensure the file handle is closed before the temp dir is removed, so
	// Windows doesn't fail the TempDir cleanup with "file in use".
	t.Cleanup(func() { _ = w.Close() })
	return w, dir
}

func TestRotatingWriter_WritesToFile(t *testing.T) {
	w, _ := newRotator(t, 10, 5)
	if _, err := w.Write([]byte("hello")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	content, err := os.ReadFile(w.filename)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(content) != "hello" {
		t.Fatalf("got %q, want %q", content, "hello")
	}
}

func TestRotatingWriter_RotatesOnSize(t *testing.T) {
	w, dir := newRotator(t, 1, 5) // 1MB cap
	logsDir := filepath.Join(dir, "logs")
	// A single ~600KB write shouldn't rotate if it stays under the cap.
	chunk := strings.Repeat("a", 600*1024)
	if _, err := w.Write([]byte(chunk)); err != nil {
		t.Fatalf("Write 1: %v", err)
	}
	if rotated, files := countBackups(logsDir); rotated != 0 {
		t.Fatalf("expected no rotation yet, got %d backups (files=%v) maxSize=%d cur=%d", rotated, files, w.maxSize, w.currentSize)
	}
	// Push past 1MB → rotate.
	if _, err := w.Write([]byte(chunk)); err != nil {
		t.Fatalf("Write 2: %v", err)
	}
	if rotated, files := countBackups(logsDir); rotated != 1 {
		t.Fatalf("expected 1 backup after overflow, got %d (backups=%v cur=%d)", rotated, files, w.currentSize)
	}
}

func TestRotatingWriter_ConcurrentWrites(t *testing.T) {
	w, dir := newRotator(t, 10, 5) // 10MB cap — no rotation expected
	const goroutines = 16
	const perG = 200

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < perG; j++ {
				line := strings.Repeat("x", 50)
				if _, err := w.Write([]byte(line)); err != nil {
					t.Errorf("concurrent write: %v", err)
					return
				}
			}
		}(i)
	}
	wg.Wait()

	content, err := os.ReadFile(w.filename)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	// Total bytes = goroutines * perG * 50, all on one file (no rotation at 10MB).
	want := goroutines * perG * 50
	if len(content) != want {
		t.Fatalf("expected %d bytes, got %d", want, len(content))
	}
	_ = dir
}

func TestReadRecent_ReturnsTail(t *testing.T) {
	w, dir := newRotator(t, 10, 5)
	for i := 0; i < 50; i++ {
		if _, err := w.Write([]byte("line\n")); err != nil {
			t.Fatalf("Write: %v", err)
		}
	}
	tail, err := ReadRecent(dir, 10)
	if err != nil {
		t.Fatalf("ReadRecent: %v", err)
	}
	lines := strings.Split(strings.TrimRight(tail, "\n"), "\n")
	if len(lines) != 10 {
		t.Fatalf("expected 10 tail lines, got %d", len(lines))
	}
}

func countBackups(dir string) (int, []string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, nil
	}
	var files []string
	for _, e := range entries {
		// Rotated backups are "app.<timestamp>.log" — two dots — while the
		// live file is "app.log" (one dot). Match only the backups.
		name := e.Name()
		if name != "app.log" && strings.HasPrefix(name, "app.") && strings.HasSuffix(name, ".log") {
			files = append(files, name)
		}
	}
	return len(files), files
}
