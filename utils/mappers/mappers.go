package mappers

import "strconv"

func ForceString(s any) string {
	switch v := s.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}

func ForceInt(s any) int {
	switch v := s.(type) {
	case string:
		i, err := strconv.Atoi(v)
		if err != nil {
			return 0
		}
		return i
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case bool:
		if v {
			return 1
		}
		return 0
	default:
		return 0
	}
}
