// Package evaluate provides local flag evaluation logic.
package evaluate

import "github.com/flagmint/flagmint-go/internal/syncutil"

// Evaluator evaluates feature flags locally against a cached rule set.
//
// NOTE: Full implementation is tracked in Ticket 5.
type Evaluator struct {
	rules syncutil.RWValue[map[string]any]
}

// NewEvaluator returns a new Evaluator with an empty rule set.
func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

// SetRules replaces the current rule set.
func (e *Evaluator) SetRules(rules map[string]any) {
	e.rules.Store(rules)
}

// Evaluate returns the flag value for key given the flat attribute map ctx.
// Returns (nil, false) when key is unknown or evaluation fails.
func (e *Evaluator) Evaluate(key string, ctx map[string]any) (any, bool) {
	rules := e.rules.Load()
	if rules == nil {
		return nil, false
	}
	val, ok := rules[key]
	if !ok {
		return nil, false
	}
	_ = ctx // future: apply targeting rules
	return val, true
}
