package evaluate

import (
	"math"
	"time"
)

// applyRollout dispatches to the correct rollout strategy and returns the
// resulting flag value. fallback is used when the rollout yields no value.
func applyRollout(
	rollout *Rollout,
	flatCtx map[string]any,
	flagType FlagType,
	variationsByID map[string]Variation,
	fallback any,
) (any, error) {
	switch rollout.Strategy {
	case RolloutOff:
		return coerceType(fallback, flagType), nil
	case RolloutPercentage:
		return applyPercentageRollout(rollout, flatCtx, flagType, fallback)
	case RolloutVariant:
		return applyVariantRollout(rollout, flatCtx, variationsByID, fallback)
	case RolloutGradual:
		return applyGradualRollout(rollout, flatCtx, flagType, fallback)
	default:
		return coerceType(fallback, flagType), nil
	}
}

// getStableRolloutKey extracts a stable bucketing key from the flat context.
// For single-kind contexts it returns "<kind>.<key>"; for legacy contexts it
// falls back to "user_id". Returns "" when no key can be determined.
func getStableRolloutKey(flatCtx map[string]any) string {
	kind, hasKind := flatCtx["kind"]
	key, hasKey := flatCtx["key"]
	if hasKind && hasKey && kind != nil && key != nil {
		return toComparableString(kind) + "." + toComparableString(key)
	}
	if userID, ok := flatCtx["user_id"]; ok && userID != nil {
		return toComparableString(userID)
	}
	return ""
}

// applyPercentageRollout evaluates a percentage rollout (boolean flags only).
func applyPercentageRollout(rollout *Rollout, flatCtx map[string]any, flagType FlagType, fallback any) (any, error) {
	if flagType != FlagTypeBoolean {
		return coerceType(fallback, flagType), ErrPercentageNonBoolean
	}
	key := getStableRolloutKey(flatCtx)
	if key == "" {
		return coerceType(fallback, flagType), ErrMissingRolloutKey
	}
	if rollout.Percentage >= 100 {
		return true, nil
	}
	if rollout.Percentage <= 0 {
		return false, nil
	}
	bucket := HashPercent(key + rollout.Salt)
	return bucket < int(rollout.Percentage), nil
}

// applyVariantRollout evaluates a weighted variant rollout (any flag type).
func applyVariantRollout(rollout *Rollout, flatCtx map[string]any, variationsByID map[string]Variation, fallback any) (any, error) {
	if len(rollout.Variants) == 0 {
		return fallback, nil
	}

	// Validate that weights sum to 100 (defensive; primary validation is server-side).
	var totalWeight float64
	for _, vw := range rollout.Variants {
		totalWeight += vw.Weight
	}
	if math.Abs(totalWeight-100) > 0.001 {
		return fallback, ErrInvalidVariantWeights
	}

	key := getStableRolloutKey(flatCtx)
	if key == "" {
		return fallback, ErrMissingRolloutKey
	}

	bucket := float64(HashPercent(key + rollout.Salt))

	var cumulative float64
	for _, vw := range rollout.Variants {
		cumulative += vw.Weight
		if bucket < cumulative {
			variation, ok := variationsByID[vw.VariationID]
			if !ok {
				return fallback, ErrMissingVariation
			}
			return variation.Value, nil
		}
	}
	return fallback, nil
}

// computeCurrentPercentage calculates the effective rollout percentage for a
// gradual rollout at the given point in time.
func computeCurrentPercentage(rollout *Rollout, now time.Time) float64 {
	if rollout.StartAt == nil || now.Before(*rollout.StartAt) {
		return 0
	}
	if rollout.IntervalHours <= 0 || rollout.Increment <= 0 {
		return rollout.TargetPercentage
	}
	elapsedHours := now.Sub(*rollout.StartAt).Hours()
	intervalsElapsed := math.Floor(elapsedHours / rollout.IntervalHours)
	current := intervalsElapsed * rollout.Increment
	if current > rollout.TargetPercentage {
		return rollout.TargetPercentage
	}
	return current
}

// applyGradualRollout evaluates a time-based progressive rollout (boolean only).
func applyGradualRollout(rollout *Rollout, flatCtx map[string]any, flagType FlagType, fallback any) (any, error) {
	if flagType != FlagTypeBoolean {
		return coerceType(fallback, flagType), ErrGradualNonBoolean
	}
	pct := computeCurrentPercentage(rollout, time.Now())
	effective := &Rollout{
		Strategy:   RolloutPercentage,
		Percentage: pct,
		Salt:       rollout.Salt,
	}
	return applyPercentageRollout(effective, flatCtx, flagType, fallback)
}

// InRollout returns true when the 0-based hash bucket for key falls within
// the [0, percentage) range (percentage expressed as 0–100).
//
// Deprecated: uses FNV-1a (StableHash). New callers should use the Evaluator
// with a RolloutPercentage strategy, which uses StringHash for cross-SDK parity.
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
//
// Deprecated: uses FNV-1a (StableHash). New callers should use the Evaluator
// with a RolloutVariant strategy, which uses StringHash for cross-SDK parity.
func VariantForKey(key string, variants []string) string {
	if len(variants) == 0 {
		return ""
	}
	idx := int(StableHash(key)) % len(variants)
	return variants[idx]
}
