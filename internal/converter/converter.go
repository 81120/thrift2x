package converter

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/81120/thrift2x/internal/targets"
	_ "github.com/81120/thrift2x/internal/targets/typescript"
)

type Config struct {
	InDir         string
	OutDir        string
	Exclude       []string
	Target        string
	TargetOptions map[string]string
	Jobs          string
}

type Stats struct {
	Total           int
	Success         int
	Failed          int
	Jobs            int
	ScanDuration    time.Duration
	ConvertDuration time.Duration
	TotalDuration   time.Duration
}

func Run(cfg Config) (Stats, error) {
	var stats Stats
	if strings.TrimSpace(cfg.InDir) == "" || strings.TrimSpace(cfg.OutDir) == "" {
		return stats, fmt.Errorf("in and out directory are required")
	}

	targetName := strings.TrimSpace(cfg.Target)
	target, ok := targets.Get(targetName)
	if !ok {
		return stats, fmt.Errorf("unknown target %q (available: %s)", targetName, strings.Join(targets.List(), ", "))
	}
	gen, err := target.NewGenerator(cfg.TargetOptions)
	if err != nil {
		return stats, err
	}

	scanStart := time.Now()
	paths, err := collectThriftFiles(cfg.InDir, cfg.Exclude)
	if err != nil {
		return stats, err
	}
	stats.ScanDuration = time.Since(scanStart)
	stats.Total = len(paths)

	jobs, err := resolveJobs(cfg.Jobs, stats.Total)
	if err != nil {
		return stats, fmt.Errorf("invalid jobs value %q: %w", cfg.Jobs, err)
	}
	stats.Jobs = jobs

	convertStart := time.Now()
	results := runWorkers(paths, jobs, func(path string) fileResult {
		return processFile(cfg.InDir, cfg.OutDir, path, gen, target.FileExtension())
	})
	stats.ConvertDuration = time.Since(convertStart)

	var errs []string
	for _, r := range results {
		if r.err != nil {
			stats.Failed++
			errs = append(errs, fmt.Sprintf("%s: %v", r.path, r.err))
			continue
		}
		if !r.skip {
			stats.Success++
		}
	}

	stats.TotalDuration = stats.ScanDuration + stats.ConvertDuration
	if len(errs) > 0 {
		errItems := make([]error, 0, len(errs))
		for _, msg := range errs {
			errItems = append(errItems, errors.New(msg))
		}
		return stats, errors.Join(errItems...)
	}
	return stats, nil
}

func FormatDuration(d time.Duration) string {
	return d.Truncate(time.Millisecond).String()
}
