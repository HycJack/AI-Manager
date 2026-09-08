// Package logger provides a leveled, file-rotating logger backed by slog.
//
// It writes to both stderr (for dev / console visibility) and a rotating
// log file under <dataDir>/logs/app-YYYY-MM-DD.log. The rotation is handled
// by lumberjack-style size capping: when the current file exceeds MaxSize
// MB it is renamed with a timestamp suffix and a new file is started.
//
// Usage:
//
//	logger.Init(dataDir, "debug")  // call once at startup
//	logger.Default()               // get the *slog.Logger
//	logger.Info("msg", "key", val)
//	logger.Error("msg", "err", err)
package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	// MaxSize is the cap in MB before a log file is rotated.
	MaxSize = 10
	// MaxBackups is how many rotated files to keep.
	MaxBackups = 5
	// MaxAge is how many days to keep a rotated file.
	MaxAge = 30
)

var (
	mu    sync.Mutex
	std   *slog.Logger
	level slog.Level
)

func init() {
	// Before Init is called, fall back to a basic stderr handler so early
	// startup messages aren't lost.
	std = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	level = slog.LevelInfo
}

// Init sets up the dual-writer logger. dataDir is the app data root; log
// files are written to <dataDir>/logs/. levelStr controls verbosity
// (debug/info/warn/error); unknown values fall back to info.
func Init(dataDir, levelStr string) error {
	mu.Lock()
	defer mu.Unlock()

	level = parseLevel(levelStr)

	logDir := filepath.Join(dataDir, "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return fmt.Errorf("logger: create log dir: %w", err)
	}

	logFile := filepath.Join(logDir, "app.log")
	rotator := &rotatingWriter{
		filename:   logFile,
		maxSize:    MaxSize,
		maxBackups: MaxBackups,
		maxAge:     MaxAge,
	}

	// stderr + file
	w := io.MultiWriter(os.Stderr, rotator)
	std = slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Format time as ISO8601 for easier grepping.
			if a.Key == slog.TimeKey && len(groups) == 0 {
				a.Value = slog.StringValue(time.Now().Format(time.RFC3339))
			}
			return a
		},
	}))
	slog.SetDefault(std)
	return nil
}

// Default returns the package-level logger.
func Default() *slog.Logger {
	mu.Lock()
	defer mu.Unlock()
	return std
}

// Debug logs at DEBUG level.
func Debug(msg string, args ...any) { Default().Debug(msg, args...) }

// Info logs at INFO level.
func Info(msg string, args ...any) { Default().Info(msg, args...) }

// Warn logs at WARN level.
func Warn(msg string, args ...any) { Default().Warn(msg, args...) }

// Error logs at ERROR level.
func Error(msg string, args ...any) { Default().Error(msg, args...) }

func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// rotatingWriter is a minimal size-based log rotator. When the current
// file exceeds maxBytes, it is renamed with a timestamp suffix and a new
// file is opened. Old files beyond maxBackups or maxAge days are removed.
type rotatingWriter struct {
	mu         sync.Mutex
	filename   string
	f          *os.File
	currentSize int64
	maxSize    int
	maxBackups int
	maxAge     int
}

func (w *rotatingWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.f == nil {
		if err := w.openNew(); err != nil {
			return 0, err
		}
	}

	// Check if rotation is needed.
	if w.currentSize+int64(len(p)) > int64(w.maxSize)*1024*1024 {
		if err := w.rotate(); err != nil {
			// Best effort: keep writing to the old file.
			return 0, err
		}
	}

	n, err := w.f.Write(p)
	w.currentSize += int64(n)
	return n, err
}

// Close flushes and closes the current log file. Safe to call on an
// uninitialised writer; returns the first error encountered. Idempotent.
func (w *rotatingWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.f == nil {
		return nil
	}
	err := w.f.Close()
	w.f = nil
	w.currentSize = 0
	return err
}

func (w *rotatingWriter) openNew() error {
	f, err := os.OpenFile(w.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	info, _ := f.Stat()
	w.f = f
	if info != nil {
		w.currentSize = info.Size()
	}
	return nil
}

func (w *rotatingWriter) rotate() error {
	if w.f != nil {
		w.f.Close()
		w.f = nil
	}

	// Rename current file with timestamp.
	ts := time.Now().Format("20060102-150405")
	rotatedName := strings.TrimSuffix(w.filename, ".log") + "." + ts + ".log"
	if err := os.Rename(w.filename, rotatedName); err != nil {
		// Best effort: reopen the original file so logging keeps working even
		// if the rename failed (e.g. the file is momentarily locked). The
		// error is surfaced to the caller rather than silently swallowed.
		if openErr := w.openNew(); openErr != nil {
			return fmt.Errorf("logger: rotate rename failed (%v) and reopen failed: %w", err, openErr)
		}
		return fmt.Errorf("logger: rotate rename failed: %w", err)
	}

	// Clean up old backups.
	w.cleanBackups()

	if err := w.openNew(); err != nil {
		return fmt.Errorf("logger: open new log after rotate: %w", err)
	}
	return nil
}

func (w *rotatingWriter) cleanBackups() {
	dir := filepath.Dir(w.filename)
	base := strings.TrimSuffix(filepath.Base(w.filename), ".log")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	now := time.Now()
	cutoff := now.AddDate(0, 0, -w.maxAge)
	var backups []string
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, base+".") || !strings.HasSuffix(name, ".log") {
			continue
		}
		full := filepath.Join(dir, name)
		info, err := os.Stat(full)
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			os.Remove(full)
			continue
		}
		backups = append(backups, full)
	}
	// If still too many, remove the oldest.
	if len(backups) > w.maxBackups {
		// Sort by modtime (oldest first) — simple insertion sort.
		for i := 1; i < len(backups); i++ {
			for j := i; j > 0; j-- {
				fi1, _ := os.Stat(backups[j-1])
				fi2, _ := os.Stat(backups[j])
				if fi1 != nil && fi2 != nil && fi1.ModTime().After(fi2.ModTime()) {
					backups[j-1], backups[j] = backups[j], backups[j-1]
				}
			}
		}
		for len(backups) > w.maxBackups {
			os.Remove(backups[0])
			backups = backups[1:]
		}
	}
}

// LogPath returns the on-disk path of the current log file, for the
// frontend's log viewer / error reporting.
func LogPath(dataDir string) string {
	return filepath.Join(dataDir, "logs", "app.log")
}

// ReadRecent returns the last n lines of the current log file. dataDir is
// the app data root passed to Init.
func ReadRecent(dataDir string, n int) (string, error) {
	data, err := os.ReadFile(LogPath(dataDir))
	if err != nil {
		return "", err
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) <= n {
		return string(data), nil
	}
	return strings.Join(lines[len(lines)-n:], "\n"), nil
}
