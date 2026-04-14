package runnercli

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
