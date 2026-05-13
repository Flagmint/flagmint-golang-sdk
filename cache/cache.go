package cache

// Adapter is the interface for pluggable flag-value caches.
// Implementations must be safe for concurrent use.
type Adapter interface {
	// Get returns the cached FeatureFlags for the given context key.
	// ok is false when the key is not present or the entry has expired.
	Get(contextKey string) (value any, ok bool)

	// Set stores value for contextKey, replacing any existing entry.
	Set(contextKey string, value any)

	// Delete removes the entry for contextKey, if present.
	Delete(contextKey string)

	// Flush removes all cached entries.
	Flush()
}
