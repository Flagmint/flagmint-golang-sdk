package evaluate

import (
	"strconv"
	"strings"
)

// Operator is a comparison operator used in targeting-rule conditions.
type Operator string

const (
	OpEquals      Operator = "eq"
	OpNotEquals   Operator = "neq"
	OpContains    Operator = "contains"
	OpNotContains Operator = "not_contains"
	OpGreaterThan Operator = "gt"
	OpLessThan    Operator = "lt"
	OpIn          Operator = "in"
	// OpNin is the "not in" operator. Values is expected to be []any.
	OpNin        Operator = "nin"
	OpStartsWith Operator = "startsWith"
	OpEndsWith   Operator = "endsWith"
	// OpExists is true when the attribute is present in the context and not nil.
	OpExists Operator = "exists"
	// OpNotExists is true when the attribute is absent from the context or nil.
	OpNotExists Operator = "not_exists"
)

// Condition is a single attribute comparison in a targeting rule.
type Condition struct {
	Attribute string   `json:"attribute"`
	Operator  Operator `json:"operator"`
	Value     any      `json:"value"`
}

// Match returns true when the flat context attributes satisfy the condition.
func (c Condition) Match(attrs map[string]any) bool {
	// exists / not_exists do not require the attribute to be present.
	attrVal, attrPresent := attrs[c.Attribute]

	switch c.Operator {
	case OpExists:
		return attrPresent && attrVal != nil
	case OpNotExists:
		return !attrPresent || attrVal == nil
	}

	if !attrPresent {
		return false
	}

	switch c.Operator {
	case OpEquals:
		return matchEq(attrVal, c.Value)
	case OpNotEquals:
		return !matchEq(attrVal, c.Value)

	case OpContains:
		return strings.Contains(toComparableString(attrVal), toComparableString(c.Value))
	case OpNotContains:
		return !strings.Contains(toComparableString(attrVal), toComparableString(c.Value))
	case OpStartsWith:
		return strings.HasPrefix(toComparableString(attrVal), toComparableString(c.Value))
	case OpEndsWith:
		return strings.HasSuffix(toComparableString(attrVal), toComparableString(c.Value))

	case OpGreaterThan:
		a, err1 := CoerceFloat64(attrVal)
		b, err2 := CoerceFloat64(c.Value)
		return err1 == nil && err2 == nil && a > b
	case OpLessThan:
		a, err1 := CoerceFloat64(attrVal)
		b, err2 := CoerceFloat64(c.Value)
		return err1 == nil && err2 == nil && a < b

	case OpIn:
		vals, ok := c.Value.([]any)
		if !ok {
			return false
		}
		attrStr := toComparableString(attrVal)
		for _, v := range vals {
			if toComparableString(v) == attrStr {
				return true
			}
		}
		return false
	case OpNin:
		vals, ok := c.Value.([]any)
		if !ok {
			// nil / non-list value → attribute is not in an empty set → true
			return true
		}
		attrStr := toComparableString(attrVal)
		for _, v := range vals {
			if toComparableString(v) == attrStr {
				return false
			}
		}
		return true

	default:
		return false
	}
}

// matchEq compares attrVal and condVal for equality, mirroring the JS
// evaluateRule "eq" special-case logic:
//   - bool attribute vs string "true"/"false" → compare as bool string
//   - numeric attribute vs string value → parse string as float64 and compare numerically
//   - otherwise → toComparableString on both sides
func matchEq(attrVal, condVal any) bool {
	switch attr := attrVal.(type) {
	case bool:
		if strVal, ok := condVal.(string); ok {
			expected := "false"
			if attr {
				expected = "true"
			}
			return expected == strVal
		}
	case float64:
		if strVal, ok := condVal.(string); ok {
			numVal, err := strconv.ParseFloat(strVal, 64)
			if err == nil {
				return attr == numVal
			}
		}
	case int:
		if strVal, ok := condVal.(string); ok {
			numVal, err := strconv.ParseFloat(strVal, 64)
			if err == nil {
				return float64(attr) == numVal
			}
		}
	}
	return toComparableString(attrVal) == toComparableString(condVal)
}

// evaluateConditions evaluates a slice of conditions with the given logical
// operator and returns whether the overall condition group matches.
func evaluateConditions(conditions []Condition, logicalOp LogicalOperator, attrs map[string]any) bool {
	switch logicalOp {
	case LogicalOR:
		for _, cond := range conditions {
			if cond.Match(attrs) {
				return true
			}
		}
		return false
	case LogicalNOT:
		// NOT: all conditions must fail (none may match).
		for _, cond := range conditions {
			if cond.Match(attrs) {
				return false
			}
		}
		return true
	default: // AND (also the zero-value / default)
		for _, cond := range conditions {
			if !cond.Match(attrs) {
				return false
			}
		}
		return true
	}
}

// MatchAll returns true when every condition in rules matches attrs.
// Kept for backward compatibility; new code should use evaluateConditions.
func MatchAll(rules []Condition, attrs map[string]any) bool {
	for _, r := range rules {
		if !r.Match(attrs) {
			return false
		}
	}
	return true
}
