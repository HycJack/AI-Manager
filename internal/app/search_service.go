package app

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// SearchOptions controls a recursive file search.
//
//   - Root is the directory to start from ("" defaults to the app data dir).
//   - FileTypes restricts results to the given extensions (e.g. ["go", "ts"]),
//     matched case-insensitively without the leading dot; empty means all files.
//   - SkipDirs lists directory names to prune while walking (compared against
//     the base name, so "node_modules" also skips ".node_modules"). If empty,
//     a set of common dependency/build/output directories is skipped.
//   - IncludeHidden, when false, skips dot-prefixed entries (e.g. .git).
type SearchOptions struct {
	Root          string   `json:"root"`
	FileTypes     []string `json:"fileTypes"`
	SkipDirs      []string `json:"skipDirs"`
	IncludeHidden bool     `json:"includeHidden"`
}

// SearchResult is one file (or directory, when Recursive or dirs are kept)
// found during the walk.
type SearchResult struct {
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`
	Size  int64  `json:"size"`
}

// defaultSkipDirs is used when the caller doesn't provide SkipDirs. It prunes
// dependency folders, VCS metadata, and build output so searches don't wade
// through e.g. node_modules.
var defaultSkipDirs = []string{
	"node_modules", ".node_modules",
	".git", ".hg", ".svn",
	".next", ".nuxt", ".cache", ".turbo",
	"dist", "build", "out", "coverage", "bin", "obj",
	"vendor",
}

// SearchService walks a directory tree and returns matching files. It lets the
// caller scope a search to a specific directory, restrict results to certain
// file types, and prune directories (like node_modules) that shouldn't be
// scanned. This is the backend half of the file-explorer / search feature.
type SearchService struct {
	state *State
}

// NewSearchService creates the search service.
func NewSearchService(s *State) *SearchService { return &SearchService{state: s} }

// Search walks Root (recursively), returning every file that passes the
// FileTypes filter while skipping any directory whose base name is in the
// effective skip set. It returns paths relative to Root.
func (s *SearchService) Search(opts SearchOptions) ([]SearchResult, error) {
	root := strings.TrimSpace(opts.Root)
	if root == "" {
		root = s.state.cfg.DataDir
	}
	root = filepath.Clean(root)

	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	types := normalizeTypes(opts.FileTypes)

	if !info.IsDir() {
		// A single file was given: respect the type filter and return it.
		if !matchType(root, types) {
			return nil, nil
		}
		return []SearchResult{{Path: filepath.Base(root), Size: info.Size()}}, nil
	}

	skip := make(map[string]struct{})
	if len(opts.SkipDirs) > 0 {
		for _, d := range opts.SkipDirs {
			skip[strings.TrimSpace(d)] = struct{}{}
		}
	} else {
		for _, d := range defaultSkipDirs {
			skip[d] = struct{}{}
		}
	}

	var results []SearchResult
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries rather than aborting the walk
		}
		if d.IsDir() {
			if path == root {
				return nil
			}
			if _, ok := skip[d.Name()]; ok {
				return filepath.SkipDir
			}
			if !opts.IncludeHidden && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !opts.IncludeHidden && strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		if len(types) > 0 && !matchType(path, types) {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			rel = path
		}
		info, err := d.Info()
		var size int64
		if err == nil {
			size = info.Size()
		}
		results = append(results, SearchResult{Path: rel, Size: size})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Path < results[j].Path })
	return results, nil
}

// normalizeTypes strips leading dots and lowercases the provided extensions.
func normalizeTypes(types []string) []string {
	out := make([]string, 0, len(types))
	for _, t := range types {
		t = strings.ToLower(strings.TrimSpace(t))
		t = strings.TrimPrefix(t, ".")
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

// matchType reports whether a path's extension is in the (normalized) list.
// Pass an empty/nil list to match everything.
func matchType(path string, types []string) bool {
	if len(types) == 0 {
		return true
	}
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	for _, t := range types {
		if t == ext {
			return true
		}
	}
	return false
}
