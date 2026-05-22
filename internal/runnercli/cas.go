package runnercli

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"rcloneflow/internal/adapter"
	"rcloneflow/internal/logutil"
	"rcloneflow/internal/store"
)

type openlistCASCompatPlan struct {
	ExcludeFrom       string
	SourceFiles       []string
	MatchedSource     []string
	DestinationExtras []string
}

func isOpenlistCASCompatible(run store.Run) bool {
	if run.Summary == nil {
		return false
	}
	if raw, ok := run.Summary["effectiveOptions"].(map[string]any); ok {
		if v, ok := raw["openlistCasCompatible"].(bool); ok {
			return v
		}
	}
	return false
}

func buildOpenlistCASCompatPlan(cfg, src, dst, mode string) (*openlistCASCompatPlan, error) {
	srcFiles, err := listRecursiveFilePaths(cfg, src)
	if err != nil {
		return nil, err
	}
	dstFiles, err := listRecursiveFilePaths(cfg, dst)
	if err != nil {
		return nil, err
	}
	plan := buildOpenlistCASCompatPlanFromPaths(srcFiles, dstFiles, mode)
	if len(plan.MatchedSource) > 0 {
		f, err := os.CreateTemp("", "rcloneflow-openlist-cas-*.txt")
		if err != nil {
			return nil, err
		}
		for _, p := range plan.MatchedSource {
			if _, err := f.WriteString(p + "\n"); err != nil {
				f.Close()
				os.Remove(f.Name())
				return nil, err
			}
		}
		if err := f.Close(); err != nil {
			os.Remove(f.Name())
			return nil, err
		}
		plan.ExcludeFrom = f.Name()
	}
	return plan, nil
}

func buildOpenlistCASCompatPlanFromPaths(srcFiles, dstFiles []string, mode string) *openlistCASCompatPlan {
	dstExact := make(map[string]struct{}, len(dstFiles))
	for _, p := range dstFiles {
		dstExact[p] = struct{}{}
	}
	srcExact := make(map[string]struct{}, len(srcFiles))
	for _, p := range srcFiles {
		srcExact[p] = struct{}{}
	}
	matched := make([]string, 0)
	for _, p := range srcFiles {
		if isCASPath(p) {
			if _, ok := dstExact[p]; ok {
				matched = append(matched, p)
			}
			continue
		}
		if _, ok := dstExact[p+".cas"]; ok {
			matched = append(matched, p)
		}
	}
	extras := make([]string, 0)
	if mode == "sync" {
		for _, p := range dstFiles {
			if isCASPath(p) {
				if _, ok := srcExact[p]; ok {
					continue
				}
				if _, ok := srcExact[trimCASSuffix(p)]; ok {
					continue
				}
				extras = append(extras, p)
				continue
			}
			if _, ok := srcExact[p]; ok {
				continue
			}
			extras = append(extras, p)
		}
	}
	return &openlistCASCompatPlan{SourceFiles: append([]string(nil), srcFiles...), MatchedSource: matched, DestinationExtras: extras}
}

func (p *openlistCASCompatPlan) ApplyPostActions(cfg, src, dst, originalMode string) error {
	if p == nil {
		return nil
	}
	cr := &adapter.CmdRunner{}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()
	if originalMode == "move" {
		for _, rel := range p.MatchedSource {
			if _, _, err := cr.Run(ctx, []string{"deletefile", joinRemotePath(src, rel), "--config", cfg}...); err != nil {
				return fmt.Errorf("delete matched move source %s: %w", rel, err)
			}
		}
	}
	if originalMode == "sync" {
		for _, rel := range p.DestinationExtras {
			if _, _, err := cr.Run(ctx, []string{"deletefile", joinRemotePath(dst, rel), "--config", cfg}...); err != nil {
				return fmt.Errorf("delete extra sync destination %s: %w", rel, err)
			}
		}
	}
	return nil
}

func listRecursiveFilePaths(cfg, target string) ([]string, error) {
	cr := &adapter.CmdRunner{}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()
	out, _, err := cr.Run(ctx, []string{"lsjson", target, "--config", cfg, "--files-only", "--recursive"}...)
	if err != nil {
		return nil, err
	}
	var arr []map[string]any
	if err := json.Unmarshal([]byte(out), &arr); err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(arr))
	for _, it := range arr {
		p, _ := it["Path"].(string)
		if p == "" {
			p, _ = it["path"].(string)
		}
		p = strings.TrimSpace(strings.ReplaceAll(p, "\\", "/"))
		if p != "" {
			paths = append(paths, p)
		}
	}
	return paths, nil
}

func isCASPath(p string) bool {
	return strings.HasSuffix(strings.ToLower(strings.TrimSpace(p)), ".cas")
}

func trimCASSuffix(p string) string {
	if !isCASPath(p) {
		return p
	}
	return p[:len(p)-4]
}

func isCASCompatibleNotFound(path, msg string, openlistCASCompatible bool) bool {
	if !openlistCASCompatible {
		return false
	}
	p := strings.TrimSpace(strings.ReplaceAll(path, "\\", "/"))
	if p == "" || isCASPath(p) {
		return false
	}
	low := strings.ToLower(strings.TrimSpace(msg))
	if low == "" {
		return false
	}
	if !(strings.Contains(low, "not found") || strings.Contains(low, "no such file") || strings.Contains(low, "object not found")) {
		return false
	}
	return true
}

func defaultCASFileExists(cfg, dst, rel string) (bool, error) {
	casRel := strings.TrimPrefix(strings.ReplaceAll(rel, "\\", "/"), "/") + ".cas"
	if strings.TrimSpace(casRel) == ".cas" {
		return false, nil
	}
	cr := &adapter.CmdRunner{}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	out, _, err := cr.Run(ctx, []string{"lsjson", joinRemotePath(dst, casRel), "--config", cfg, "--files-only"}...)
	if err != nil {
		return false, nil
	}
	var arr []map[string]any
	if err := json.Unmarshal([]byte(out), &arr); err != nil {
		return false, err
	}
	for _, it := range arr {
		p, _ := it["Path"].(string)
		if p == "" {
			p, _ = it["path"].(string)
		}
		p = strings.TrimSpace(strings.ReplaceAll(p, "\\", "/"))
		if p == "" {
			continue
		}
		if p == casRel || strings.HasSuffix(p, "/"+casRel) || strings.HasSuffix(p, "/"+filepath.Base(casRel)) || p == filepath.Base(casRel) {
			return true, nil
		}
	}
	return false, nil
}

func (r *Runner) confirmCASMatch(cfg, dst, rel string) bool {
	if r == nil || r.casVerifier == nil {
		return false
	}
	delays := r.casVerifyDelays
	if len(delays) == 0 {
		delays = []time.Duration{0}
	}
	for _, d := range delays {
		if d > 0 {
			time.Sleep(d)
		}
		ok, err := r.casVerifier(cfg, dst, rel)
		if err == nil && ok {
			return true
		}
	}
	return false
}

func (r *Runner) appendCASExclude(excludeFrom, rel string) {
	if r == nil || strings.TrimSpace(excludeFrom) == "" || strings.TrimSpace(rel) == "" {
		return
	}
	r.casExcludeMu.Lock()
	defer r.casExcludeMu.Unlock()
	f, err := os.OpenFile(excludeFrom, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(strings.TrimSpace(rel) + "\n")
}

type casAttemptAnalysis struct {
	CASMatchedPaths map[string]struct{}
	RealFailures    map[string]string
}

func analyzeCASAttemptLogSegment(path string, startOffset int64, openlistCASCompatible bool) casAttemptAnalysis {
	res := casAttemptAnalysis{CASMatchedPaths: map[string]struct{}{}, RealFailures: map[string]string{}}
	f, err := os.Open(path)
	if err != nil {
		return res
	}
	defer f.Close()
	if startOffset > 0 {
		if _, err := f.Seek(startOffset, io.SeekStart); err != nil {
			return res
		}
	}
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 0, 128*1024), 2*1024*1024)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		if m := fileCASMatchedRe.FindStringSubmatch(line); len(m) > 0 {
			res.CASMatchedPaths[strings.TrimSpace(m[1])] = struct{}{}
			delete(res.RealFailures, strings.TrimSpace(m[1]))
			continue
		}
		var rec map[string]any
		if json.Unmarshal([]byte(line), &rec) == nil {
			level := strings.ToUpper(strings.TrimSpace(anyString(rec["level"])))
			msg := strings.TrimSpace(anyString(rec["msg"]))
			obj := strings.TrimSpace(anyString(rec["object"]))
			if level == "ERROR" {
				if obj != "" {
					res.RealFailures[obj] = msg
				}
			}
			continue
		}
		at, level, p, msg, ok := logutil.ParseLogSegment(line)
		_ = at
		if !ok {
			continue
		}
		if isAttemptObjectNotFoundSummary(p, msg) {
			continue
		}
		if strings.EqualFold(level, "ERROR") {
			if strings.TrimSpace(p) != "" {
				res.RealFailures[p] = msg
			}
		}
	}
	return res
}

func configuredRetryCount(opt map[string]any) int {
	if opt != nil {
		switch v := opt["retries"].(type) {
		case int:
			if v > 0 {
				return v
			}
		case int64:
			if v > 0 {
				return int(v)
			}
		case float64:
			if v > 0 {
				return int(v)
			}
		}
	}
	return 1
}