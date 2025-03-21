package gonfig

import (
	"strings"
	"sync"
)

// PathCache provides thread-safe caching for split paths.
type PathCache struct {
	cache sync.Map
}

// NewPathCache creates a new path cache instance.
func NewPathCache() *PathCache {
	return &PathCache{}
}

// Get retrieves or creates a split path.
func (pc *PathCache) Get(path string) []string {
	if cached, ok := pc.cache.Load(path); ok {
		return cached.([]string)
	}

	parts := strings.Split(path, ".")
	pc.cache.Store(path, parts)
	return parts
}
