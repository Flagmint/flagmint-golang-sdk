package evaluate

// StableHash returns a deterministic 32-bit hash for s.
// The algorithm is FNV-1a, matching the JS SDK implementation.
func StableHash(s string) uint32 {
	const (
		offset32 uint32 = 2166136261
		prime32  uint32 = 16777619
	)
	h := offset32
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= prime32
	}
	return h
}
