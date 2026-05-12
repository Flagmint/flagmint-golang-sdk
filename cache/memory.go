package cache

import "sync"

// entry holds a cached value.
type entry struct {
	value any
}

// MemoryCache is a simple in-memory cache backed by a sync.Map.
// It has no capacity limit or TTL; it is suitable for use in environments
// where the flag set is small and bounded.
// For production use cases consider supplying a proper LRU adapter via
// [WithCacheAdapter].
type MemoryCache struct {
	mu   sync.RWMutex
	data map[string]entry
}

// NewMemoryCache returns an initialised MemoryCache.
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{data: make(map[string]entry)}
}

// Get returns the cached value for contextKey.
func (m *MemoryCache) Get(contextKey string) (any, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.data[contextKey]
	if !ok {
		return nil, false
	}
	return e.value, true
}

// Set stores value under contextKey.
func (m *MemoryCache) Set(contextKey string, value any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[contextKey] = entry{value: value}
}

// Delete removes the entry for contextKey.
func (m *MemoryCache) Delete(contextKey string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, contextKey)
}

// Flush clears all cached entries.
func (m *MemoryCache) Flush() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[string]entry)
}
