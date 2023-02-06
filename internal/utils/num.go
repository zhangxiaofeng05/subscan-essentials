package utils

import "strconv"

func Int64FromInterface(i interface{}) int64 {
	switch i := i.(type) {
	case int:
		return int64(i)
	case int64:
		return i
	case uint64:
		return int64(i)
	case float64:
		return int64(i)
	case string:
		r, _ := strconv.ParseInt(i, 10, 64)
		return r
	}
	return 0
}

// Convert int64, uint64, float64, string to int, return 0 if other types
func IntFromInterface(i interface{}) int {
	switch i := i.(type) {
	case int:
		return i
	case int64:
		return int(i)
	case uint64:
		return int(i)
	case float64:
		return int(i)
	case string:
		return StringToInt(i)
	}
	return 0
}

// Convert str to int, return 0 if error
func StringToInt(s string) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return 0
}
