package runnercli

type anyMap = map[string]any

func existsStr(m anyMap, key string) bool {
	if m == nil { return false }
	if v, ok := m[key]; ok {
		_, ok2 := v.(string)
		return ok2
	}
	return false
}

func existsBool(m anyMap, key string) bool {
	if m == nil { return false }
	if v, ok := m[key]; ok {
		_, ok2 := v.(bool)
		return ok2
	}
	return false
}

func eff(cur anyMap) anyMap {
	if cur == nil { return nil }
	if v, ok := cur["effectiveOptions"]; ok {
		if m, ok2 := v.(map[string]any); ok2 { return m }
	}
	return nil
}
