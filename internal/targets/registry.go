package targets

import (
	"sort"
	"strings"
	"sync"
)

var (
	mu       sync.RWMutex
	registry = map[string]Target{}
)

func Register(t Target) {
	if t == nil {
		return
	}
	name := strings.ToLower(strings.TrimSpace(t.Name()))
	if name == "" {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	registry[name] = t
}

func Get(name string) (Target, bool) {
	mu.RLock()
	defer mu.RUnlock()
	t, ok := registry[strings.ToLower(strings.TrimSpace(name))]
	return t, ok
}

func List() []string {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]string, 0, len(registry))
	for k := range registry {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
