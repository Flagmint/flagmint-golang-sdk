package evaluate

// StringHash produces a 32-bit hash identical to the npm `string-hash` package
// (djb2 variant). This is the hash used for rollout bucketing to guarantee
// cross-SDK consistency: a user bucketed at 30% in the JS SDK must land in the
// same bucket in the Go SDK.
func StringHash(s string) uint32 {
	var hash uint32 = 5381
	for i := 0; i < len(s); i++ {
		hash = (hash << 5) + hash + uint32(s[i])
	}
	return hash
}

// HashPercent returns a stable bucket value in [0, 99] for rollout bucketing.
// Two callers with the same input always receive the same bucket.
func HashPercent(input string) int {
	return int(StringHash(input) % 100)
}

// StableHash returns a deterministic 32-bit hash for s using FNV-1a.
//
// Deprecated: use StringHash for cross-SDK hash parity with the JS evaluator.
// StableHash is retained for callers that rely on InRollout / VariantForKey.
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
