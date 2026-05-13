package evaluate

// RolloutKind enumerates the supported rollout strategies.
type RolloutKind string

const (
	// RolloutPercentage enables a flag for a percentage of contexts.
	RolloutPercentage RolloutKind = "percentage"
	// RolloutVariant distributes contexts among named variants.
	RolloutVariant RolloutKind = "variant"
	// RolloutGradual slowly ramps a flag from 0 % to 100 % over time.
	RolloutGradual RolloutKind = "gradual"
)

// InRollout returns true when the 0-based hash bucket for key falls within
// the [0, percentage) range (percentage expressed as 0–100).
func InRollout(key string, percentage float64) bool {
	if percentage <= 0 {
		return false
	}
	if percentage >= 100 {
		return true
	}
	bucket := float64(StableHash(key) % 100)
	return bucket < percentage
}

// VariantForKey maps key to one of the provided variant names using a stable
// hash. The variants slice must be non-empty.
func VariantForKey(key string, variants []string) string {
	if len(variants) == 0 {
		return ""
	}
	idx := int(StableHash(key)) % len(variants)
	return variants[idx]
}
