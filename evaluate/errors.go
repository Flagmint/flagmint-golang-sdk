package evaluate

import "errors"

// Sentinel errors returned by the local evaluator.
// All evaluation errors are non-fatal; the caller should return the flag's
// fallback value and log the error rather than propagating it.
var (
	// ErrMissingRolloutKey is returned when a rollout requires a stable user
	// key (kind+key or user_id) but none is present in the context.
	ErrMissingRolloutKey = errors.New("rollout requires stable user key (kind+key or user_id)")

	// ErrPercentageNonBoolean is returned when a percentage rollout is applied
	// to a non-boolean flag. Use variant rollout for multi-value flags.
	ErrPercentageNonBoolean = errors.New("percentage rollout is boolean-only; use variant rollout for multi-value flags")

	// ErrGradualNonBoolean is returned when a gradual rollout is applied to a
	// non-boolean flag.
	ErrGradualNonBoolean = errors.New("gradual rollout is boolean-only; use variant rollout for multi-value flags")

	// ErrMissingSegment is returned when a targeting rule references a segment
	// ID that is not present in FlagConfig.SegmentsByID.
	ErrMissingSegment = errors.New("targeting rule references unknown segment")

	// ErrMissingVariation is returned when a targeting rule or off-variation
	// ID references a variation that is not present in FlagConfig.VariationsByID.
	ErrMissingVariation = errors.New("targeting rule references unknown variation")

	// ErrMissingRollout is returned when a targeting rule references a rollout
	// ID that is not present in FlagConfig.Rollouts.
	ErrMissingRollout = errors.New("targeting rule references unknown rollout")

	// ErrInvalidVariantWeights is returned when variant rollout weights do not
	// sum to 100.
	ErrInvalidVariantWeights = errors.New("variant weights must sum to 100")
)
