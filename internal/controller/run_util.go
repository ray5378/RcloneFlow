package controller

import "encoding/json"

func abs64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func anyToInt64(v any) int64 {
	switch x := v.(type) {
	case int64:
		return x
	case int:
		return int64(x)
	case float64:
		return int64(x)
	case float32:
		return int64(x)
	case json.Number:
		if n, err := x.Int64(); err == nil {
			return n
		}
	}
	return 0
}