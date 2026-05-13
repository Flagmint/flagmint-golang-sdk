package evaluate

import (
	"fmt"
	"strconv"
)

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
