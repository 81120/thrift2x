package converter

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

type fileResult struct {
	path string
	err  error
	skip bool
}

func runWorkers(paths []string, jobs int, fn func(string) fileResult) []fileResult {
	if len(paths) == 0 {
		return nil
	}
	jobsCh := make(chan string)
	resultsCh := make(chan fileResult, len(paths))
	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()
		for p := range jobsCh {
			resultsCh <- fn(p)
		}
	}

	for i := 0; i < jobs; i++ {
		wg.Add(1)
		go worker()
	}
	for _, p := range paths {
		jobsCh <- p
	}
	close(jobsCh)
	wg.Wait()
	close(resultsCh)

	results := make([]fileResult, 0, len(paths))
	for r := range resultsCh {
		results = append(results, r)
	}
	return results
}

func resolveJobs(raw string, total int) (int, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" || raw == "auto" {
		return autoJobs(total), nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	if n <= 0 {
		return 0, fmt.Errorf("must be > 0")
	}
	return n, nil
}

func autoJobs(total int) int {
	if total <= 0 {
		return 1
	}
	cpu := runtime.NumCPU()
	j := cpu * 2
	if j < 2 {
		j = 2
	}
	if j > total {
		j = total
	}
	if j < 1 {
		j = 1
	}
	return j
}
