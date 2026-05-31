package evaluate

import (
	"fmt"
	"strconv"
)

// toComparableString converts v to a string for condition matching, mirroring
// the JS evaluator's String() coercion semantics. nil → "".
func toComparableString(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case bool:
		if val {
			return "true"
		}
		return "false"
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// coerceType coerces v to the given FlagType, returning v unchanged when
// coercion fails so the caller always gets a usable value.
func coerceType(v any, ft FlagType) any {
	switch ft {
	case FlagTypeBoolean:
		b, err := CoerceBool(v)
		if err != nil {
			return v
		}
		return b
	case FlagTypeString:
		s, err := CoerceString(v)
		if err != nil {
			return v
		}
		return s
	case FlagTypeNumber:
		n, err := CoerceFloat64(v)
		if err != nil {
			return v
		}
		return n
	default: // FlagTypeJSON or unknown — return as-is
		return v
	}
}

// CoerceString coerces v to a string.
func CoerceString(v any) (string, error) {
	switch val := v.(type) {
	case string:
		return val, nil
	case bool:
		return strconv.FormatBool(val), nil
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64), nil
	case int:
		return strconv.Itoa(val), nil
	default:
		return "", fmt.Errorf("cannot coerce %T to string", v)
	}
}

// CoerceBool coerces v to a bool.
func CoerceBool(v any) (bool, error) {
	switch val := v.(type) {
	case bool:
		return val, nil
	case string:
		return strconv.ParseBool(val)
	case float64:
		return val != 0, nil
	default:
		return false, fmt.Errorf("cannot coerce %T to bool", v)
	}
}

// CoerceFloat64 coerces v to a float64.
func CoerceFloat64(v any) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case int:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(val, 64)
	case bool:
		if val {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot coerce %T to float64", v)
	}
}
