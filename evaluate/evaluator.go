// Package evaluate provides local flag evaluation logic for the Flagmint Go SDK.
// It ports the flag-evaluator.ts logic from the JS SDK, enabling server-side
// applications to evaluate flags without a network round-trip.
package evaluate

import (
	"fmt"
	"log/slog"
	"sort"
)

// Evaluator evaluates feature flags locally against FlagConfig objects.
// It is safe for concurrent use by multiple goroutines.
type Evaluator struct {
	logger *slog.Logger
}

// NewEvaluator returns a new Evaluator using the default slog logger.
func NewEvaluator() *Evaluator {
	return &Evaluator{logger: slog.Default()}
}

// NewEvaluatorWithLogger returns a new Evaluator with the provided logger.
func NewEvaluatorWithLogger(logger *slog.Logger) *Evaluator {
	return &Evaluator{logger: logger}
}

// Evaluate evaluates config against the pre-flattened context map flatCtx and
// returns the resulting flag value. flatCtx should be produced by calling
// EvaluationContext.Flatten() before passing here.
//
// On any evaluation error (missing segment, bad config, etc.) the method
// returns the coerced default value and the error so the caller can log it.
// Evaluation never panics.
func (e *Evaluator) Evaluate(config *FlagConfig, flatCtx map[string]any) (any, error) {
	if config == nil {
		return nil, fmt.Errorf("evaluate: nil FlagConfig")
	}

	// --- 1. Kill switch -------------------------------------------------------
	if !config.IsActive {
		if config.OffVariationID != nil {
			variation, ok := config.VariationsByID[*config.OffVariationID]
			if !ok {
				e.warn("off_variation_id references unknown variation", config.Key, *config.OffVariationID)
				return coerceType(config.DefaultValue, config.Type), ErrMissingVariation
			}
			return coerceType(variation.Value, config.Type), nil
		}
		return coerceType(config.DefaultValue, config.Type), nil
	}

	// --- 2. On-variation check ------------------------------------------------
	// Boolean flags with no targeting rules and no rollouts immediately return
	// the variation whose value is true (or false if none matches).
	if config.Type == FlagTypeBoolean &&
		len(config.TargetingRules) == 0 &&
		len(config.Rollouts) == 0 {
		for _, v := range config.Variations {
			if b, ok := v.Value.(bool); ok && b {
				return true, nil
			}
		}
		if len(config.Variations) > 0 {
			return false, nil
		}
	}

	// --- 3. Context is already flat (flatCtx passed in) ----------------------

	// --- 4. Targeting rules (sorted ascending by order_index; first match wins)
	rules := make([]TargetingRule, len(config.TargetingRules))
	copy(rules, config.TargetingRules)
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].OrderIndex < rules[j].OrderIndex
	})

	for _, rule := range rules {
		matched, err := e.matchRule(rule, config.SegmentsByID, flatCtx)
		if err != nil {
			e.warn("rule evaluation error", config.Key, err.Error())
			// Fail-open: skip this rule and continue.
			continue
		}
		if !matched {
			continue
		}

		// Rule matched — apply variation or rollout.
		if rule.VariationID != nil {
			variation, ok := config.VariationsByID[*rule.VariationID]
			if !ok {
				e.warn("targeting rule references unknown variation", config.Key, *rule.VariationID)
				return coerceType(config.DefaultValue, config.Type), ErrMissingVariation
			}
			return coerceType(variation.Value, config.Type), nil
		}

		if rule.RolloutID != nil {
			rollout, ok := config.Rollouts[*rule.RolloutID]
			if !ok {
				e.warn("targeting rule references unknown rollout", config.Key, *rule.RolloutID)
				return coerceType(config.DefaultValue, config.Type), ErrMissingRollout
			}
			val, err := applyRollout(rollout, flatCtx, config.Type, config.VariationsByID, config.DefaultValue)
			if err != nil {
				e.warn("rollout error", config.Key, err.Error())
				return coerceType(config.DefaultValue, config.Type), err
			}
			return val, nil
		}

		// Rule matched but has neither variation_id nor rollout_id — misconfiguration.
		e.warn("targeting rule has neither variation_id nor rollout_id", config.Key, rule.ID)
	}

	// --- 5. No rules matched — return coerced default value ------------------
	return coerceType(config.DefaultValue, config.Type), nil
}

// matchRule evaluates whether a single targeting rule matches the flat context.
func (e *Evaluator) matchRule(rule TargetingRule, segmentsByID map[string]*Segment, flatCtx map[string]any) (bool, error) {
	switch rule.Kind {
	case "segment":
		if segmentsByID == nil {
			return false, fmt.Errorf("%w: %s", ErrMissingSegment, rule.SegmentID)
		}
		seg, ok := segmentsByID[rule.SegmentID]
		if !ok {
			return false, fmt.Errorf("%w: %s", ErrMissingSegment, rule.SegmentID)
		}
		return evaluateSegment(seg, flatCtx), nil
	case "custom":
		return evaluateConditions(rule.Conditions, rule.LogicalOp, flatCtx), nil
	default:
		return false, nil
	}
}

// evaluateSegment evaluates a segment's conditions against the flat context.
// When Force is true the segment always matches regardless of conditions.
func evaluateSegment(seg *Segment, flatCtx map[string]any) bool {
	if seg.Force {
		return true
	}
	return evaluateConditions(seg.Rules, seg.LogicalOp, flatCtx)
}

// warn emits a structured warning. Safe when logger is nil.
func (e *Evaluator) warn(msg, flagKey, detail string) {
	if e.logger != nil {
		e.logger.Warn("flagmint/evaluate: "+msg, "flag", flagKey, "detail", detail)
	}
}
