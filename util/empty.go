package util

func IsEmptyValue(val any) bool {
	switch v := val.(type) {
	case nil:
		return true
	case string:
		return v == ""
	case int, int8, int16, int32, int64:
		return v == 0
	case uint, uint8, uint16, uint32, uint64:
		return v == 0
	case float32, float64:
		return v == 0
	case bool:
		return !v
	case map[string]any:
		return len(v) == 0
	case map[string]string:
		return len(v) == 0
	case map[string][]string:
		return len(v) == 0
	case []any:
		return len(v) == 0
	case []string:
		return len(v) == 0
	default:
		return false
	}
}
