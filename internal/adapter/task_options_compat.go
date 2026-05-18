package adapter

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ParseTaskOptionsCompat tolerates legacy mixed-shape task option payloads stored in DB.
// Historical data may use string-or-array for filters and bool-or-int for multiThreadStreams.
func ParseTaskOptionsCompat(raw []byte) (*TaskOptions, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, err
	}
	obj = normalizeTaskOptionsMap(obj)
	bs, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	var opts TaskOptions
	if err := json.Unmarshal(bs, &opts); err != nil {
		return nil, err
	}
	return &opts, nil
}

func normalizeTaskOptionsMap(obj map[string]any) map[string]any {
	if obj == nil {
		return nil
	}
	for _, k := range []string{"exclude", "excludeFrom", "excludeIfPresent", "include", "includeFrom", "filter", "filterFrom", "filesFrom", "filesFromRaw"} {
		if v, ok := obj[k]; ok {
			obj[k] = normalizeStringSliceLike(v)
		}
	}
	if v, ok := obj["multiThreadStreams"]; ok {
		obj["multiThreadStreams"] = normalizeIntLike(v)
	}
	return obj
}

func normalizeStringSliceLike(v any) any {
	switch vv := v.(type) {
	case nil:
		return []string{}
	case string:
		s := strings.TrimSpace(vv)
		if s == "" {
			return []string{}
		}
		return []string{s}
	case []any:
		out := make([]string, 0, len(vv))
		for _, item := range vv {
			s := strings.TrimSpace(fmt.Sprint(item))
			if s != "" {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return vv
	default:
		s := strings.TrimSpace(fmt.Sprint(v))
		if s == "" {
			return []string{}
		}
		return []string{s}
	}
}

func normalizeIntLike(v any) int {
	switch vv := v.(type) {
	case nil:
		return 0
	case bool:
		if vv {
			return 4
		}
		return 0
	case float64:
		return int(vv)
	case float32:
		return int(vv)
	case int:
		return vv
	case int64:
		return int(vv)
	case json.Number:
		i, _ := vv.Int64()
		return int(i)
	case string:
		s := strings.TrimSpace(vv)
		if s == "" {
			return 0
		}
		if strings.EqualFold(s, "true") {
			return 4
		}
		if strings.EqualFold(s, "false") {
			return 0
		}
		if i, err := strconv.Atoi(s); err == nil {
			return i
		}
		return 0
	default:
		return 0
	}
}
