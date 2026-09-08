package app

import (
	"os"
	"path/filepath"
	"testing"
)

// buildTree creates a small fixture directory and returns its path.
//
//	<root>/
//	  main.go
//	  README.md
//	  node_modules/pkg/index.js   (should be pruned via default skip)
//	  .git/config                 (hidden dir, pruned by default)
//	  src/util.go
//	  src/util_test.go
func buildTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, rel := range []string{
		"main.go",
		"README.md",
		"node_modules/pkg/index.js",
		".git/config",
		"src/util.go",
		"src/util_test.go",
		"src/data.json",
	} {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestSearchService_DefaultSkipDirs(t *testing.T) {
	s := NewSearchService(newTestState(t))
	got, err := s.Search(SearchOptions{Root: buildTree(t)})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	// node_modules and .git must be pruned; everything else is returned.
	want := []string{"README.md", "main.go", "src/data.json", "src/util.go", "src/util_test.go"}
	if len(got) != len(want) {
		t.Fatalf("expected %d files, got %d: %+v", len(want), len(got), got)
	}
	for i, w := range want {
		if got[i].Path != filepath.FromSlash(w) {
			t.Fatalf("result[%d] = %q, want %q (all: %+v)", i, got[i].Path, w, got)
		}
	}
}

func TestSearchService_FileTypeFilter(t *testing.T) {
	s := NewSearchService(newTestState(t))
	got, err := s.Search(SearchOptions{
		Root:      buildTree(t),
		FileTypes: []string{".go"},
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	want := []string{"main.go", "src/util.go", "src/util_test.go"}
	if len(got) != len(want) {
		t.Fatalf("expected %d files, got %d: %+v", len(want), len(got), got)
	}
	for i, w := range want {
		if got[i].Path != filepath.FromSlash(w) {
			t.Fatalf("result[%d] = %q, want %q", i, got[i].Path, w)
		}
	}
}

func TestSearchService_CustomSkipDirs(t *testing.T) {
	s := NewSearchService(newTestState(t))
	// Override the skip list: only skip "src" (which should also drop its .go
	// and .json children) — node_modules is now included.
	got, err := s.Search(SearchOptions{
		Root:     buildTree(t),
		SkipDirs: []string{"src"},
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	for _, r := range got {
		if filepath.Base(filepath.Clean(r.Path)) == "config" {
			t.Fatalf("expected .git to be pruned (hidden default), got %q", r.Path)
		}
		if r.Path == "src/util.go" || r.Path == "src/data.json" {
			t.Fatalf("expected src to be pruned, got %q", r.Path)
		}
	}
}

func TestSearchService_SingleFile(t *testing.T) {
	s := NewSearchService(newTestState(t))
	root := buildTree(t)
	single := filepath.Join(root, "main.go")
	got, err := s.Search(SearchOptions{Root: single, FileTypes: []string{"go"}})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(got) != 1 || got[0].Path != "main.go" {
		t.Fatalf("single-file search: %+v", got)
	}

	// A mismatched type on a single file yields nothing.
	got, err = s.Search(SearchOptions{Root: single, FileTypes: []string{"ts"}})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected no match, got %+v", got)
	}
}
