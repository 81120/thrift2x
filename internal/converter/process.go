package converter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/81120/thrift2x/internal/parser"
	"github.com/81120/thrift2x/internal/targets"
)

func processFile(inDir, outDir, path string, gen targets.Generator, ext string) fileResult {
	content, err := os.ReadFile(path)
	if err != nil {
		return fileResult{path: path, err: fmt.Errorf("read error: %w", err)}
	}

	parsed, err := parser.Parse(path, string(content))
	if err != nil {
		return fileResult{path: path, err: fmt.Errorf("parse error: %w", err)}
	}

	out := gen.Generate(parsed)
	rel, err := filepath.Rel(inDir, path)
	if err != nil {
		return fileResult{path: path, err: fmt.Errorf("rel path error: %w", err)}
	}
	if strings.TrimSpace(ext) == "" {
		ext = ".ts"
	}
	dst := filepath.Join(outDir, strings.TrimSuffix(rel, filepath.Ext(rel))+ext)
	if strings.TrimSpace(out) == "" {
		_ = os.Remove(dst)
		return fileResult{path: path, skip: true}
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fileResult{path: path, err: fmt.Errorf("mkdir error: %w", err)}
	}
	if err := os.WriteFile(dst, []byte(out), 0o644); err != nil {
		return fileResult{path: path, err: fmt.Errorf("write error: %w", err)}
	}
	return fileResult{path: path}
}
