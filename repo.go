package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func findRepos(dirs []string) []string {
	seen := make(map[string]struct{})

	for _, dir := range dirs {
		abs, err := filepath.Abs(dir)
		if err != nil {
			continue
		}
		abs, _ = filepath.EvalSymlinks(abs)

		filepath.WalkDir(abs, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if !d.IsDir() {
				return nil
			}

			// Check depth relative to search root
			rel, _ := filepath.Rel(abs, path)
			depth := 0
			if rel != "." {
				depth = strings.Count(rel, string(os.PathSeparator)) + 1
			}
			if depth > 3 {
				return fs.SkipDir
			}

			if d.Name() == ".git" {
				seen[filepath.Dir(path)] = struct{}{}
				return fs.SkipDir
			}
			return nil
		})
	}

	repos := make([]string, 0, len(seen))
	for r := range seen {
		repos = append(repos, r)
	}
	sort.Strings(repos)
	return repos
}
