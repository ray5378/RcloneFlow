package active_transfer

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"rcloneflow/internal/adapter"
)

func BuildCandidateFiles(ctx context.Context, cfg, src, dst string, opts *adapter.TaskOptions) ([]TransferCandidateFile, error) {
	cr := &adapter.CmdRunner{}
	out, _, err := cr.Run(ctx, []string{"lsjson", src, "--config", cfg, "--files-only", "--recursive"}...)
	if err != nil {
		return nil, err
	}
	var arr []map[string]any
	if err := json.Unmarshal([]byte(out), &arr); err != nil {
		return nil, err
	}
	matcher, err := newFilterMatcher(opts)
	if err != nil {
		return nil, err
	}
	rawPaths := make([]string, 0, len(arr))
	for _, it := range arr {
		path := strings.TrimSpace(anyString(it["Path"]))
		if path == "" {
			path = strings.TrimSpace(anyString(it["Name"]))
		}
		if path == "" {
			continue
		}
		rawPaths = append(rawPaths, normalizePath(path))
	}
	matcher.Prepare(rawPaths)

	existingEntries := map[string]fileFact{}
	casCompatible := opts != nil && opts.OpenlistCasCompatible
	if shouldCheckExisting(opts, dst) {
		targets := []string{}
		if opts != nil {
			if strings.TrimSpace(opts.CompareDest) != "" {
				targets = append(targets, strings.TrimSpace(opts.CompareDest))
			}
			if strings.TrimSpace(opts.CopyDest) != "" {
				targets = append(targets, strings.TrimSpace(opts.CopyDest))
			}
		}
		if strings.TrimSpace(dst) != "" {
			targets = append([]string{strings.TrimSpace(dst)}, targets...)
		}
		existingEntries = listExistingPathSet(ctx, cr, cfg, targets, matcher.ignoreCase, casCompatible)
	}

	items := make([]TransferCandidateFile, 0, len(arr))
	for _, it := range arr {
		path := strings.TrimSpace(anyString(it["Path"]))
		if path == "" {
			path = strings.TrimSpace(anyString(it["Name"]))
		}
		if path == "" {
			continue
		}
		path = normalizePath(path)
		size := anyInt64(it["Size"])
		modTime := anyTime(it["ModTime"])
		if !matcher.Allow(path, size, modTime) {
			continue
		}
		if shouldSkipByExisting(path, size, modTime, existingEntries, matcher.ignoreCase, opts) {
			continue
		}
		items = append(items, TransferCandidateFile{
			Path:      path,
			Name:      baseName(path),
			SizeBytes: size,
		})
	}
	return items, nil
}

type filterMatcher struct {
	ignoreCase       bool
	includes         []string
	excludes         []string
	filterPlus       []string
	filterMinus      []string
	filesOnly        map[string]struct{}
	excludeIfPresent []string
	blockedDirs      map[string]struct{}
	minSize          int64
	maxSize          int64
	minAge           time.Duration
	maxAge           time.Duration
	ageNow           time.Time
}

func newFilterMatcher(opts *adapter.TaskOptions) (*filterMatcher, error) {
	m := &filterMatcher{filesOnly: map[string]struct{}{}, blockedDirs: map[string]struct{}{}, ageNow: time.Now()}
	if opts == nil {
		return m, nil
	}
	m.ignoreCase = opts.IgnoreCase || opts.IgnoreCaseSync
	m.includes = normalizeRules(append(append([]string{}, opts.Include...), opts.IncludeFrom...), m.ignoreCase)
	m.excludes = normalizeRules(append(append([]string{}, opts.Exclude...), opts.ExcludeFrom...), m.ignoreCase)
	m.excludeIfPresent = normalizeRules(opts.ExcludeIfPresent, m.ignoreCase)
	m.minSize = parseSizeLoose(opts.MinSize)
	m.maxSize = parseSizeLoose(opts.MaxSize)
	m.minAge = parseAgeLoose(opts.MinAge)
	m.maxAge = parseAgeLoose(opts.MaxAge)
	for _, rule := range append(append([]string{}, opts.Filter...), opts.FilterFrom...) {
		r := strings.TrimSpace(rule)
		if r == "" {
			continue
		}
		prefix := ""
		body := r
		if strings.HasPrefix(r, "+") || strings.HasPrefix(r, "-") {
			prefix = r[:1]
			body = strings.TrimSpace(r[1:])
		}
		body = normalizeRule(body, m.ignoreCase)
		if body == "" {
			continue
		}
		if prefix == "+" {
			m.filterPlus = append(m.filterPlus, body)
		} else if prefix == "-" {
			m.filterMinus = append(m.filterMinus, body)
		}
	}
	for _, p := range append(append([]string{}, opts.FilesFrom...), opts.FilesFromRaw...) {
		n := normalizeRule(strings.TrimSpace(p), m.ignoreCase)
		if n != "" {
			m.filesOnly[n] = struct{}{}
		}
	}
	return m, nil
}

func (m *filterMatcher) Prepare(paths []string) {
	if m == nil || len(m.excludeIfPresent) == 0 {
		return
	}
	for _, path := range paths {
		p := normalizeRule(path, m.ignoreCase)
		base := baseName(p)
		for _, marker := range m.excludeIfPresent {
			if base == marker {
				dir := filepath.Dir(p)
				if dir == "." {
					dir = ""
				}
				m.blockedDirs[dir] = struct{}{}
			}
		}
	}
}

func (m *filterMatcher) Allow(path string, size int64, modTime time.Time) bool {
	if m == nil {
		return true
	}
	p := normalizeRule(path, m.ignoreCase)
	if m.minSize > 0 && size > 0 && size < m.minSize {
		return false
	}
	if m.maxSize > 0 && size > 0 && size > m.maxSize {
		return false
	}
	if !modTime.IsZero() {
		age := m.ageNow.Sub(modTime)
		if m.minAge > 0 && age < m.minAge {
			return false
		}
		if m.maxAge > 0 && age > m.maxAge {
			return false
		}
	}
	for dir := range m.blockedDirs {
		if dir == "" || p == dir || strings.HasPrefix(p, dir+"/") {
			return false
		}
	}
	if len(m.filesOnly) > 0 {
		if _, ok := m.filesOnly[p]; !ok {
			return false
		}
	}
	if len(m.filterPlus) > 0 {
		matched := false
		for _, rule := range m.filterPlus {
			if matchRule(rule, p) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	for _, rule := range m.filterMinus {
		if matchRule(rule, p) {
			return false
		}
	}
	if len(m.includes) > 0 {
		matched := false
		for _, rule := range m.includes {
			if matchRule(rule, p) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	for _, rule := range m.excludes {
		if matchRule(rule, p) {
			return false
		}
	}
	return true
}

func shouldCheckExisting(opts *adapter.TaskOptions, dst string) bool {
	if strings.TrimSpace(dst) == "" {
		return false
	}
	// rclone 默认就会检查目标端是否已存在，只有显式 no-check-dest 时才应跳过。
	// 这里的候选文件/未传输列表应尽量贴近“实际会传哪些文件”，而不是把整个源目录都算进 pending。
	if opts == nil {
		return true
	}
	if opts.NoCheckDest {
		return false
	}
	return true
}

type fileFact struct {
	Size    int64
	ModTime time.Time
}

func parseModifyWindowLoose(s string) time.Duration {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	if d, err := time.ParseDuration(s); err == nil && d >= 0 {
		return d
	}
	return 0
}

func withinModifyWindow(a, b time.Time, window time.Duration) bool {
	if a.IsZero() || b.IsZero() {
		return false
	}
	delta := a.Sub(b)
	if delta < 0 {
		delta = -delta
	}
	return delta <= window
}

func listExistingPathSet(ctx context.Context, cr *adapter.CmdRunner, cfg string, targets []string, ignoreCase bool, casCompatible bool) map[string]fileFact {
	set := map[string]fileFact{}
	seen := map[string]struct{}{}
	for _, target := range targets {
		target = strings.TrimSpace(target)
		if target == "" {
			continue
		}
		if _, ok := seen[target]; ok {
			continue
		}
		seen[target] = struct{}{}
		out, _, err := cr.Run(ctx, []string{"lsjson", target, "--config", cfg, "--files-only", "--recursive"}...)
		if err != nil {
			continue
		}
		var arr []map[string]any
		if json.Unmarshal([]byte(out), &arr) != nil {
			continue
		}
		for _, it := range arr {
			path := strings.TrimSpace(anyString(it["Path"]))
			if path == "" {
				path = strings.TrimSpace(anyString(it["Name"]))
			}
			path = normalizePath(path)
			if path != "" {
				norm := normalizeRule(path, ignoreCase)
				fact := fileFact{Size: anyInt64(it["Size"]), ModTime: anyTime(it["ModTime"])}
				set[norm] = fact
				if casCompatible && isCASPath(norm) {
					set[trimCASSuffix(norm)] = fact
				}
			}
		}
	}
	return set
}

func shouldSkipByExisting(path string, size int64, modTime time.Time, existing map[string]fileFact, ignoreCase bool, opts *adapter.TaskOptions) bool {
	if len(existing) == 0 {
		return false
	}
	fact, ok := existing[normalizeRule(path, ignoreCase)]
	if !ok {
		return false
	}
	if opts == nil {
		return true
	}
	modifyWindow := parseModifyWindowLoose(opts.ModifyWindow)
	if opts.IgnoreSize {
		if opts.Update && !modTime.IsZero() && !fact.ModTime.IsZero() {
			if withinModifyWindow(modTime, fact.ModTime, modifyWindow) {
				return true
			}
			return !modTime.After(fact.ModTime)
		}
		if opts.IgnoreTimes {
			return true
		}
		return true
	}
	if opts.SizeOnly {
		return size > 0 && fact.Size > 0 && size == fact.Size
	}
	if opts.Update && !modTime.IsZero() && !fact.ModTime.IsZero() {
		if size > 0 && fact.Size > 0 && size != fact.Size {
			return false
		}
		if withinModifyWindow(modTime, fact.ModTime, modifyWindow) {
			return true
		}
		return !modTime.After(fact.ModTime)
	}
	if opts.IgnoreTimes {
		return size > 0 && fact.Size > 0 && size == fact.Size
	}
	if size > 0 && fact.Size > 0 && size != fact.Size {
		return false
	}
	if !modTime.IsZero() && !fact.ModTime.IsZero() {
		if withinModifyWindow(modTime, fact.ModTime, modifyWindow) {
			return true
		}
		if opts.UseServerModtime {
			return !modTime.After(fact.ModTime)
		}
		return modTime.Equal(fact.ModTime)
	}
	return true
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

func normalizeRules(in []string, ignoreCase bool) []string {
	out := make([]string, 0, len(in))
	for _, v := range in {
		n := normalizeRule(v, ignoreCase)
		if n != "" {
			out = append(out, n)
		}
	}
	return out
}

func normalizeRule(s string, ignoreCase bool) string {
	s = normalizePath(strings.TrimSpace(s))
	if ignoreCase {
		s = strings.ToLower(s)
	}
	return s
}

func matchRule(rule, path string) bool {
	if rule == "" {
		return false
	}
	if ok, _ := filepath.Match(rule, path); ok {
		return true
	}
	if !strings.Contains(rule, "/") {
		if ok, _ := filepath.Match(rule, baseName(path)); ok {
			return true
		}
	}
	if strings.HasSuffix(rule, "/**") {
		prefix := strings.TrimSuffix(rule, "/**")
		return path == prefix || strings.HasPrefix(path, prefix+"/")
	}
	return false
}

func anyString(v any) string {
	s, _ := v.(string)
	return s
}

func anyInt64(v any) int64 {
	switch x := v.(type) {
	case float64:
		return int64(x)
	case int64:
		return x
	case int:
		return int64(x)
	default:
		return 0
	}
}

func anyTime(v any) time.Time {
	s, _ := v.(string)
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	return time.Time{}
}

func parseSizeLoose(s string) int64 {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" {
		return 0
	}
	mul := int64(1)
	switch {
	case strings.HasSuffix(s, "KB"):
		mul, s = 1024, strings.TrimSuffix(s, "KB")
	case strings.HasSuffix(s, "MB"):
		mul, s = 1024*1024, strings.TrimSuffix(s, "MB")
	case strings.HasSuffix(s, "GB"):
		mul, s = 1024*1024*1024, strings.TrimSuffix(s, "GB")
	case strings.HasSuffix(s, "TB"):
		mul, s = 1024*1024*1024*1024, strings.TrimSuffix(s, "TB")
	case strings.HasSuffix(s, "K"):
		mul, s = 1024, strings.TrimSuffix(s, "K")
	case strings.HasSuffix(s, "M"):
		mul, s = 1024*1024, strings.TrimSuffix(s, "M")
	case strings.HasSuffix(s, "G"):
		mul, s = 1024*1024*1024, strings.TrimSuffix(s, "G")
	case strings.HasSuffix(s, "T"):
		mul, s = 1024*1024*1024*1024, strings.TrimSuffix(s, "T")
	}
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil || f <= 0 {
		return 0
	}
	return int64(f * float64(mul))
}

func parseAgeLoose(s string) time.Duration {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return 0
	}
	if d, err := time.ParseDuration(s); err == nil {
		return d
	}
	if strings.HasSuffix(s, "d") {
		f, err := strconv.ParseFloat(strings.TrimSuffix(s, "d"), 64)
		if err == nil && f > 0 {
			return time.Duration(f * float64(24*time.Hour))
		}
	}
	return 0
}
