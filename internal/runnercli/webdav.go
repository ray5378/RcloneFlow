package runnercli

import (
	"context"
	"encoding/json"
	"strings"

	"rcloneflow/internal/adapter"
)

func isWebDAVUnderlying(cfgPath, remote string) bool {
	cr := &adapter.CmdRunner{}
	out, _, err := cr.Run(context.Background(), []string{"config", "dump", "--config", cfgPath}...)
	if err != nil {
		return false
	}
	var dump map[string]any
	if json.Unmarshal([]byte(out), &dump) != nil {
		return false
	}
	name := remote
	for depth := 0; depth < 4; depth++ {
		sec, _ := dump[name].(map[string]any)
		if sec == nil {
			break
		}
		if t, _ := sec["type"].(string); strings.EqualFold(t, "webdav") {
			return true
		}
		if base, _ := sec["remote"].(string); base != "" {
			if i := strings.Index(base, ":"); i > 0 {
				name = base[:i]
				continue
			}
		}
		break
	}
	return false
}

func normalizeVisibleTargetPaths(arr []map[string]any, openlistCASCompatible bool) map[string]struct{} {
	m := map[string]struct{}{}
	for _, it := range arr {
		p, _ := it["Path"].(string)
		if p == "" {
			p, _ = it["path"].(string)
		}
		p = strings.TrimSpace(strings.ReplaceAll(p, "\\", "/"))
		if p == "" {
			continue
		}
		m[p] = struct{}{}
		if openlistCASCompatible && isCASPath(p) {
			m[trimCASSuffix(p)] = struct{}{}
		}
	}
	return m
}

func expectedVisibleDestinationPaths(plan *openlistCASCompatPlan, originalMode string) []string {
	if plan == nil {
		return nil
	}
	if originalMode == "move" {
		return append([]string(nil), plan.MatchedSource...)
	}
	return append([]string(nil), plan.SourceFiles...)
}

func areAllExpectedPathsVisible(expected []string, visible map[string]struct{}) bool {
	for _, p := range expected {
		if _, ok := visible[p]; !ok {
			return false
		}
	}
	return true
}