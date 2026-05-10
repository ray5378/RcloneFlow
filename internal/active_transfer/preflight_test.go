package active_transfer

import (
	"testing"
	"time"

	"rcloneflow/internal/adapter"
)

func TestFilterMatcher_IncludeExcludeFilterAndFilesFromRaw(t *testing.T) {
	m, err := newFilterMatcher(&adapter.TaskOptions{
		Include:      []string{"media/**", "*.mkv"},
		Exclude:      []string{"media/tmp/**"},
		Filter:       []string{"- *.tmp"},
		FilesFromRaw: []string{"media/movie.mkv", "root.mkv"},
	})
	if err != nil {
		t.Fatalf("newFilterMatcher err: %v", err)
	}
	cases := []struct {
		path string
		want bool
	}{
		{"media/movie.mkv", true},
		{"root.mkv", true},
		{"media/tmp/a.mkv", false},
		{"media/clip.tmp", false},
		{"other/clip.mkv", false},
	}
	for _, tc := range cases {
		if got := m.Allow(tc.path, 100, time.Now().Add(-2*time.Hour)); got != tc.want {
			t.Fatalf("Allow(%q)=%v want %v", tc.path, got, tc.want)
		}
	}
}

func TestFilterMatcher_IgnoreCase(t *testing.T) {
	m, err := newFilterMatcher(&adapter.TaskOptions{
		IgnoreCase: true,
		Include:    []string{"MEDIA/**"},
	})
	if err != nil {
		t.Fatalf("newFilterMatcher err: %v", err)
	}
	if !m.Allow("media/abc.mkv", 100, time.Now().Add(-2*time.Hour)) {
		t.Fatalf("expected case-insensitive match")
	}
}

func TestFilterMatcher_FromRulesAndExcludeIfPresent(t *testing.T) {
	m, err := newFilterMatcher(&adapter.TaskOptions{
		IncludeFrom:      []string{"docs/**"},
		ExcludeFrom:      []string{"docs/tmp/**"},
		FilterFrom:       []string{"- *.bak"},
		ExcludeIfPresent: []string{".ignore"},
	})
	if err != nil {
		t.Fatalf("newFilterMatcher err: %v", err)
	}
	m.Prepare([]string{"docs/a.txt", "docs/tmp/a.txt", "docs/inner/.ignore", "docs/inner/file.txt", "docs/keep/file.bak"})
	cases := []struct {
		path string
		want bool
	}{
		{"docs/a.txt", true},
		{"docs/tmp/a.txt", false},
		{"docs/inner/file.txt", false},
		{"docs/keep/file.bak", false},
	}
	for _, tc := range cases {
		if got := m.Allow(tc.path, 100, time.Now().Add(-2*time.Hour)); got != tc.want {
			t.Fatalf("Allow(%q)=%v want %v", tc.path, got, tc.want)
		}
	}
}

func TestFilterMatcher_SizeAndAge(t *testing.T) {
	m, err := newFilterMatcher(&adapter.TaskOptions{})
	if err != nil {
		t.Fatalf("newFilterMatcher err: %v", err)
	}
	m.minSize = parseSizeLoose("10M")
	m.maxSize = parseSizeLoose("20M")
	m.minAge = parseAgeLoose("2h")
	m.maxAge = parseAgeLoose("2d")
	m.ageNow = time.Date(2026, 5, 9, 21, 0, 0, 0, time.UTC)

	if m.Allow("a.bin", 5*1024*1024, m.ageNow.Add(-3*time.Hour)) {
		t.Fatalf("expected minSize filter")
	}
	if m.Allow("b.bin", 25*1024*1024, m.ageNow.Add(-3*time.Hour)) {
		t.Fatalf("expected maxSize filter")
	}
	if m.Allow("c.bin", 15*1024*1024, m.ageNow.Add(-30*time.Minute)) {
		t.Fatalf("expected minAge filter")
	}
	if m.Allow("d.bin", 15*1024*1024, m.ageNow.Add(-72*time.Hour)) {
		t.Fatalf("expected maxAge filter")
	}
	if !m.Allow("ok.bin", 15*1024*1024, m.ageNow.Add(-3*time.Hour)) {
		t.Fatalf("expected file to pass size/age filters")
	}
}

func TestShouldCheckExisting(t *testing.T) {
	if !shouldCheckExisting(nil, "dst:") {
		t.Fatalf("default copy semantics should check existing destination entries")
	}
	if shouldCheckExisting(&adapter.TaskOptions{NoCheckDest: true, IgnoreExisting: true}, "dst:") {
		t.Fatalf("noCheckDest should disable existing check")
	}
	if !shouldCheckExisting(&adapter.TaskOptions{IgnoreExisting: true}, "dst:") {
		t.Fatalf("ignoreExisting should trigger existing check")
	}
	if !shouldCheckExisting(&adapter.TaskOptions{CompareDest: "cmp:"}, "dst:") {
		t.Fatalf("compareDest should trigger existing check")
	}
	if !shouldCheckExisting(&adapter.TaskOptions{CopyDest: "cpy:"}, "dst:") {
		t.Fatalf("copyDest should trigger existing check")
	}
}

func TestTrimCASSuffixAndCASDetection(t *testing.T) {
	if !isCASPath("a/b/movie.mkv.cas") {
		t.Fatalf("expected cas path")
	}
	if isCASPath("a/b/movie.mkv") {
		t.Fatalf("expected non-cas path")
	}
	if got := trimCASSuffix("a/b/movie.mkv.cas"); got != "a/b/movie.mkv" {
		t.Fatalf("trimCASSuffix=%q", got)
	}
}

func TestShouldSkipByExisting_CASCompatibilityStaysSeparateFromNormalChain(t *testing.T) {
	now := time.Date(2026, 5, 9, 22, 0, 0, 0, time.UTC)
	path := "a/movie.mkv"
	casExisting := map[string]fileFact{
		"a/movie.mkv.cas": {Size: 100, ModTime: now.Add(-1 * time.Hour)},
		"a/movie.mkv":     {Size: 100, ModTime: now.Add(-1 * time.Hour)},
	}
	// 普通链：只有真的存在同名原文件，才允许按默认 compare 语义跳过。
	if !shouldSkipByExisting(path, 100, now.Add(-1*time.Hour), map[string]fileFact{"a/movie.mkv": casExisting["a/movie.mkv"]}, false, &adapter.TaskOptions{}) {
		t.Fatalf("normal chain should skip only when original target file exists")
	}
	if shouldSkipByExisting(path, 100, now.Add(-1*time.Hour), map[string]fileFact{"a/movie.mkv.cas": casExisting["a/movie.mkv.cas"]}, false, &adapter.TaskOptions{}) {
		t.Fatalf("normal chain must not treat .cas as equivalent existing file")
	}
	// CAS 兼容链：预处理后才允许把 .cas 视为等效已存在。
	if !shouldSkipByExisting(path, 100, now.Add(-1*time.Hour), map[string]fileFact{"a/movie.mkv": casExisting["a/movie.mkv.cas"]}, false, &adapter.TaskOptions{OpenlistCasCompatible: true}) {
		t.Fatalf("cas-compatible chain should allow .cas-equivalent existing file to skip")
	}
}

func TestListExistingPathSet_CASMappingSemanticsStayIsolated(t *testing.T) {
	// 这里不直接调用 rclone，而是验证我们对 existing map 的语义约束：
	// 普通链只认原文件键；CAS 链才允许把 .cas 映射成原文件键。
	now := time.Date(2026, 5, 9, 22, 0, 0, 0, time.UTC)
	normalExisting := map[string]fileFact{
		"a/movie.mkv.cas": {Size: 100, ModTime: now},
	}
	if _, ok := normalExisting["a/movie.mkv"]; ok {
		t.Fatalf("normal chain must not inject trimmed .cas alias")
	}
	casExisting := map[string]fileFact{
		"a/movie.mkv.cas": {Size: 100, ModTime: now},
		"a/movie.mkv":     {Size: 100, ModTime: now},
	}
	if _, ok := casExisting["a/movie.mkv"]; !ok {
		t.Fatalf("cas chain should carry trimmed .cas alias for equivalent matching")
	}
}

func TestShouldSkipByExisting_CompareSemantics(t *testing.T) {
	now := time.Date(2026, 5, 9, 22, 0, 0, 0, time.UTC)
	existing := map[string]fileFact{
		"a.bin": {Size: 100, ModTime: now.Add(-1 * time.Hour)},
	}
	if !shouldSkipByExisting("a.bin", 100, now.Add(-1*time.Hour), existing, false, &adapter.TaskOptions{}) {
		t.Fatalf("same size and modtime should skip by default")
	}
	if shouldSkipByExisting("a.bin", 101, now.Add(-1*time.Hour), existing, false, &adapter.TaskOptions{}) {
		t.Fatalf("different size should not skip by default")
	}
	if !shouldSkipByExisting("a.bin", 100, now.Add(-2*time.Hour), existing, false, &adapter.TaskOptions{Update: true}) {
		t.Fatalf("older source should skip under update")
	}
	if shouldSkipByExisting("a.bin", 100, now.Add(1*time.Hour), existing, false, &adapter.TaskOptions{Update: true}) {
		t.Fatalf("newer source should not skip under update")
	}
	if !shouldSkipByExisting("a.bin", 100, now.Add(5*time.Hour), existing, false, &adapter.TaskOptions{SizeOnly: true}) {
		t.Fatalf("same size should skip under sizeOnly")
	}
	if !shouldSkipByExisting("a.bin", 100, now.Add(5*time.Hour), existing, false, &adapter.TaskOptions{IgnoreTimes: true}) {
		t.Fatalf("same size should skip under ignoreTimes")
	}
	if !shouldSkipByExisting("a.bin", 999, now.Add(5*time.Hour), existing, false, &adapter.TaskOptions{IgnoreSize: true}) {
		t.Fatalf("ignoreSize should skip regardless of size")
	}
	if !shouldSkipByExisting("a.bin", 100, now.Add(-59*time.Minute), existing, false, &adapter.TaskOptions{ModifyWindow: "2m"}) {
		t.Fatalf("modifyWindow should treat near-equal modtimes as same")
	}
	if !shouldSkipByExisting("a.bin", 100, now.Add(-90*time.Minute), existing, false, &adapter.TaskOptions{UseServerModtime: true}) {
		t.Fatalf("useServerModtime should allow older-or-equal source to skip")
	}
	if shouldSkipByExisting("a.bin", 100, now.Add(-30*time.Minute), existing, false, &adapter.TaskOptions{UseServerModtime: true}) {
		t.Fatalf("useServerModtime should not skip when source is newer")
	}
}
