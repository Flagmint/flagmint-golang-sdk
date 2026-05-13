package evaluate

import "strings"

// Operator is a comparison operator used in targeting rules.
type Operator string

const (
	OpEquals       Operator = "eq"
	OpNotEquals    Operator = "neq"
	OpContains     Operator = "contains"
	OpNotContains  Operator = "not_contains"
	OpGreaterThan  Operator = "gt"
	OpLessThan     Operator = "lt"
	OpIn           Operator = "in"
	OpNotIn        Operator = "not_in"
)

// Condition is a single targeting rule condition.
type Condition struct {
	Attribute string   `json:"attribute"`
	Op        Operator `json:"op"`
	Value     any      `json:"value"`
}

// Match returns true when the flat context attributes satisfy the condition.
func (c Condition) Match(attrs map[string]any) bool {
	attrVal, ok := attrs[c.Attribute]
	if !ok {
		return false
	}

	attrStr, _ := CoerceString(attrVal)
	valStr, _ := CoerceString(c.Value)

	switch c.Op {
	case OpEquals:
		return attrStr == valStr
	case OpNotEquals:
		return attrStr != valStr
	case OpContains:
		return strings.Contains(attrStr, valStr)
	case OpNotContains:
		return !strings.Contains(attrStr, valStr)
	case OpGreaterThan:
		a, err1 := CoerceFloat64(attrVal)
		b, err2 := CoerceFloat64(c.Value)
		return err1 == nil && err2 == nil && a > b
	case OpLessThan:
		a, err1 := CoerceFloat64(attrVal)
		b, err2 := CoerceFloat64(c.Value)
		return err1 == nil && err2 == nil && a < b
	case OpIn:
		if vals, ok := c.Value.([]any); ok {
			for _, v := range vals {
				s, _ := CoerceString(v)
				if attrStr == s {
					return true
				}
			}
		}
		return false
	case OpNotIn:
		if vals, ok := c.Value.([]any); ok {
			for _, v := range vals {
				s, _ := CoerceString(v)
				if attrStr == s {
					return false
				}
			}
		}
		return true
	default:
		return false
	}
}

// MatchAll returns true when every condition in rules matches attrs.
func MatchAll(rules []Condition, attrs map[string]any) bool {
	for _, r := range rules {
		if !r.Match(attrs) {
			return false
		}
	}
	return true
}
