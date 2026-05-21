package runnercli

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"rcloneflow/internal/adapter"
)

func sizeOf(r *adapter.CmdRunner, cfg, target string, opts map[string]any) (bytes int64, count int64, err error) {
	args := []string{"size", target, "--config", cfg, "--json"}
	args = addFilterFlags(args, opts, true)
	out, _, e := r.Run(context.Background(), args...)
	if e == nil {
		var m map[string]any
		if json.Unmarshal([]byte(out), &m) == nil {
			if v, ok := m["bytes"].(float64); ok {
				bytes = int64(v)
			}
			if v, ok := m["count"].(float64); ok {
				count = int64(v)
			}
			return bytes, count, nil
		}
	}
	args = []string{"size", target, "--config", cfg}
	args = addFilterFlags(args, opts, false)
	out, _, e = r.Run(context.Background(), args...)
	if e != nil {
		return 0, 0, e
	}
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Total objects:") {
			fmt.Sscanf(line, "Total objects: %d", &count)
		}
		if strings.HasPrefix(line, "Total size:") {
			var human string
			fmt.Sscanf(line, "Total size: %s (%d)", &human, &bytes)
		}
	}
	return bytes, count, nil
}

func sizeOfPaged(r *adapter.CmdRunner, cfg, target string, opts map[string]any) (int64, int64, error) {
	lsArgs := []string{"lsf", target, "--config", cfg, "--dirs-only", "--max-depth", "1"}
	lsArgs = addFilterFlags(lsArgs, opts, true)
	out, _, e := r.Run(context.Background(), lsArgs...)
	if e != nil {
		return sizeOf(r, cfg, target, opts)
	}
	collectDirs := func(s string) []string {
		arr := []string{}
		for _, ln := range strings.Split(s, "\n") {
			ln = strings.TrimSpace(ln)
			if ln == "" {
				continue
			}
			arr = append(arr, strings.TrimSuffix(ln, "/"))
		}
		return arr
	}
	lines := collectDirs(out)
	if len(lines) == 0 {
		ls2 := []string{"lsf", target, "--config", cfg, "--dirs-only", "--max-depth", "2"}
		ls2 = addFilterFlags(ls2, opts, true)
		if out2, _, e2 := r.Run(context.Background(), ls2...); e2 == nil {
			all := collectDirs(out2)
			sec := []string{}
			for _, d := range all {
				if strings.Count(d, "/") == 1 {
					sec = append(sec, d)
				}
			}
			lines = sec
		}
	}
	if len(lines) == 0 {
		return sizeOf(r, cfg, target, opts)
	}
	var totalBytes, totalCount int64
	rootArgs := []string{"lsjson", target, "--config", cfg, "--files-only", "--max-depth", "1"}
	rootArgs = addFilterFlags(rootArgs, opts, true)
	if out2, _, e2 := r.Run(context.Background(), rootArgs...); e2 == nil {
		var arr []map[string]any
		if json.Unmarshal([]byte(out2), &arr) == nil {
			for _, it := range arr {
				if sz, ok := it["Size"].(float64); ok {
					totalBytes += int64(sz)
					totalCount++
				}
			}
		}
	}
	for _, d := range lines {
		child := target
		if !strings.HasSuffix(child, "/") {
			child += "/"
		}
		child += d
		b, c, e3 := sizeOf(r, cfg, child, opts)
		if e3 == nil {
			totalBytes += b
			totalCount += c
		}
	}
	return totalBytes, totalCount, nil
}