package converter

import (
	"io/fs"
	"path/filepath"
	"strings"
)

func collectThriftFiles(inDir string, excludes []string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(inDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if shouldExclude(path, excludes) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".thrift") {
			return nil
		}
		if shouldExclude(path, excludes) {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	return paths, err
}

func shouldExclude(path string, patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}
	norm := filepath.ToSlash(path)
	for _, p := range patterns {
		if strings.Contains(norm, filepath.ToSlash(p)) {
			return true
		}
	}
	return false
}
