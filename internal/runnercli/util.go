package runnercli

import "strings"

type anyMap = map[string]any

func existsStr(m anyMap, key string) bool {
	if m == nil {
		return false
	}
	if v, ok := m[key]; ok {
		_, ok2 := v.(string)
		return ok2
	}
	return false
}

func existsBool(m anyMap, key string) bool {
	if m == nil {
		return false
	}
	if v, ok := m[key]; ok {
		_, ok2 := v.(bool)
		return ok2
	}
	return false
}

func eff(m anyMap) anyMap {
	if m == nil {
		return nil
	}
	if v, ok := m["effectiveOptions"]; ok {
		if mm, ok2 := v.(map[string]any); ok2 {
			return mm
		}
	}
	return nil
}

func forceFlagValue(args []string, flag, value string) []string {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == flag {
			args[i+1] = value
			return args
		}
	}
	return append(args, flag, value)
}

func joinRemotePath(base, rel string) string {
	rel = strings.TrimPrefix(strings.ReplaceAll(rel, "\\", "/"), "/")
	if rel == "" {
		return base
	}
	if strings.HasSuffix(base, ":") || strings.HasSuffix(base, "/") {
		return base + rel
	}
	return base + "/" + rel
}

func anyString(v any) string {
	s, _ := v.(string)
	return s
}

func anyFloat64(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int64:
		return float64(x)
	case int:
		return float64(x)
	default:
		return 0
	}
}
