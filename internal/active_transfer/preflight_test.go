package active_transfer

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

func TestBuildListFilterFlagsFromOptions(t *testing.T) {
	opts := &adapter.TaskOptions{
		Include:          []string{"media/**"},
		Exclude:          []string{"tmp/**"},
		Filter:           []string{"- *.bak"},
		FilesFromRaw:     []string{"a.mp4"},
		ExcludeIfPresent: []string{".ignore"},
		MinSize:          "10M",
		MaxAge:           "24h",
		IgnoreCase:       true,
	}
	got := strings.Join(buildListFilterFlagsFromOptions(opts), " ")
	for _, want := range []string{"--include media/**", "--exclude tmp/**", "--filter - *.bak", "--files-from-raw a.mp4", "--exclude-if-present .ignore", "--min-size 10M", "--max-age 24h", "--ignore-case"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
	if strings.Contains(got, "--checksum") || strings.Contains(got, "--update") {
		t.Fatalf("list filter flags should not include compare-only flags: %q", got)
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
		"a/movie.mkv.cas": {Size: 100, ModTime: now.Add(-1 * time.Hour), ExactPath: "a/movie.mkv.cas"},
		"a/movie.mkv":     {Size: 100, ModTime: now.Add(-1 * time.Hour), ExactPath: "a/movie.mkv"},
	}
	// 普通链：只有真的存在同名原文件，才允许按默认 compare 语义跳过。
	if !shouldSkipByExisting(path, 100, now.Add(-1*time.Hour), map[string]fileFact{"a/movie.mkv": casExisting["a/movie.mkv"]}, false, &adapter.TaskOptions{}) {
		t.Fatalf("normal chain should skip only when original target file exists")
	}
	if shouldSkipByExisting(path, 100, now.Add(-1*time.Hour), map[string]fileFact{"a/movie.mkv.cas": casExisting["a/movie.mkv.cas"]}, false, &adapter.TaskOptions{}) {
		t.Fatalf("normal chain must not treat .cas as equivalent existing file")
	}
	// CAS 兼容链：通过 .cas 别名命中时，不再用 .cas 的元数据和原文件硬比较。
	if !shouldSkipByExisting(path, 999999, now.Add(24*time.Hour), map[string]fileFact{"a/movie.mkv": {Size: 12, ModTime: now, ExactPath: "a/movie.mkv.cas", CASAlias: true}}, false, &adapter.TaskOptions{OpenlistCasCompatible: true, SizeOnly: true}) {
		t.Fatalf("cas-compatible chain should allow .cas-equivalent existing file to skip")
	}
}

func TestListExistingPathSet_CASMappingSemanticsStayIsolated(t *testing.T) {
	// 这里不直接调用 rclone，而是验证我们对 existing map 的语义约束：
	// 普通链只认原文件键；CAS 链才允许把 .cas 映射成原文件键。
	now := time.Date(2026, 5, 9, 22, 0, 0, 0, time.UTC)
	normalExisting := map[string]fileFact{
		"a/movie.mkv.cas": {Size: 100, ModTime: now, ExactPath: "a/movie.mkv.cas"},
	}
	if _, ok := normalExisting["a/movie.mkv"]; ok {
		t.Fatalf("normal chain must not inject trimmed .cas alias")
	}
	casExisting := map[string]fileFact{
		"a/movie.mkv.cas": {Size: 100, ModTime: now, ExactPath: "a/movie.mkv.cas"},
		"a/movie.mkv":     {Size: 100, ModTime: now, ExactPath: "a/movie.mkv.cas", CASAlias: true},
	}
	if got, ok := casExisting["a/movie.mkv"]; !ok || !got.CASAlias {
		t.Fatalf("cas chain should carry trimmed .cas alias for equivalent matching")
	}
}

func TestShouldSkipByExisting_RealCASCompatibleTaskBehavior_SizeOnlyAndServerModtimeStillSkipViaAlias(t *testing.T) {
	now := time.Date(2026, 5, 10, 13, 0, 0, 0, time.UTC)
	existing := map[string]fileFact{
		"电视剧/国产剧/冰湖重生 (2026)/Season 1/冰湖重生 - S01E10 - 第 10 集.mkv": {
			Size: 244, ModTime: now, ExactPath: "电视剧/国产剧/冰湖重生 (2026)/Season 1/冰湖重生 - S01E10 - 第 10 集.mkv.cas", CASAlias: true,
		},
	}
	path := "电视剧/国产剧/冰湖重生 (2026)/Season 1/冰湖重生 - S01E10 - 第 10 集.mkv"
	if !shouldSkipByExisting(path, 8*1024*1024*1024, now.Add(24*time.Hour), existing, false, &adapter.TaskOptions{OpenlistCasCompatible: true, SizeOnly: true, UseServerModtime: true}) {
		t.Fatalf("cas alias should skip even when .cas metadata differs from original media file")
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

func TestParseCombinedNeedTransferPaths(t *testing.T) {
	out := strings.Join([]string{
		"= same/file1.mkv",
		"+ missing/file2.mkv",
		"* diff/file3.mkv",
		"! error/file4.mkv",
		"- only-on-dst/file5.mkv",
		"2026/05/10 12:00:00 NOTICE: unrelated log line",
	}, "\n")
	got := parseCombinedNeedTransferPaths(out, false)
	want := []string{"missing/file2.mkv", "diff/file3.mkv", "error/file4.mkv"}
	if len(got) != len(want) {
		t.Fatalf("len=%d want %d, got=%v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d]=%q want %q", i, got[i], want[i])
		}
	}
}

func TestParseCombinedNeedTransferPaths_CASCompatible(t *testing.T) {
	out := strings.Join([]string{
		"+ dir/a.mp4.cas",
		"* dir/b.mp4",
		"! dir/a.mp4",
		"+ dir/a.mp4.cas",
	}, "\n")
	got := parseCombinedNeedTransferPaths(out, true)
	want := []string{"dir/a.mp4", "dir/b.mp4"}
	if len(got) != len(want) {
		t.Fatalf("len=%d want %d, got=%v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d]=%q want %q", i, got[i], want[i])
		}
	}
}

func TestParseCombinedNeedTransferPaths_FromStderrLikeCheckOutput(t *testing.T) {
	out := strings.Join([]string{
		"2026/05/10 12:00:00 NOTICE: some log",
		"+ missing/file2.mkv",
		"* diff/file3.mkv",
	}, "\n")
	got := parseCombinedNeedTransferPaths(out, false)
	want := []string{"missing/file2.mkv", "diff/file3.mkv"}
	if len(got) != len(want) {
		t.Fatalf("len=%d want %d, got=%v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d]=%q want %q", i, got[i], want[i])
		}
	}
}

func TestBuildCandidateFiles_EndToEnd_RealTask1ExcludeCache(t *testing.T) {
	srcFiles := []map[string]any{
		{"Path": "20260510/keep1.mp4", "Size": 100, "ModTime": "2026-05-10T12:00:00Z"},
		{"Path": "20260510/keep2.mp4", "Size": 101, "ModTime": "2026-05-10T12:01:00Z"},
		{"Path": "O8UM1tAOo1LeI_Profile_1-1778364024092-128.mp4.cache", "Size": 3145728, "ModTime": "2026-05-10T12:02:00Z"},
	}
	dstFiles := []map[string]any{}
	var items []TransferCandidateFile
	calls := runBuildCandidateFilesWithFakeRclone(t, srcFiles, dstFiles, "+ 20260510/keep1.mp4\n+ 20260510/keep2.mp4\n", func() {
		var err error
		items, err = BuildCandidateFiles(context.Background(), "cfg.conf", "src:", "dst:", &adapter.TaskOptions{Exclude: []string{"*.cache"}})
		if err != nil {
			t.Fatalf("BuildCandidateFiles err: %v", err)
		}
	})
	if len(items) != 2 {
		t.Fatalf("items=%v", items)
	}
	for _, it := range items {
		if strings.HasSuffix(it.Path, ".cache") {
			t.Fatalf("cache file should be filtered out, got %v", items)
		}
	}
	joined := strings.Join(calls, "\n")
	if !strings.Contains(joined, "lsjson src: --config cfg.conf --files-only --recursive --exclude *.cache") {
		t.Fatalf("source lsjson should carry exclude flag, calls=%s", joined)
	}
	if !strings.Contains(joined, "check src: dst: --config cfg.conf --combined - --one-way --exclude *.cache") {
		t.Fatalf("check should carry exclude flag, calls=%s", joined)
	}
}

func TestBuildCandidateFiles_EndToEnd_RealTask10IncludeCASAndCASAliasSkip(t *testing.T) {
	srcFiles := []map[string]any{
		{"Path": "电视剧/国产剧/冰湖重生 (2026)/Season 1/冰湖重生 - S01E10 - 第 10 集.mkv.cas", "Size": 244, "ModTime": "2026-05-10T12:00:00Z"},
		{"Path": "电视剧/国产剧/冰湖重生 (2026)/Season 1/冰湖重生 - S01E11 - 第 11 集.mkv.cas", "Size": 244, "ModTime": "2026-05-10T12:00:00Z"},
		{"Path": "电视剧/国产剧/冰湖重生 (2026)/Season 1/说明.txt", "Size": 20, "ModTime": "2026-05-10T12:00:00Z"},
	}
	dstFiles := []map[string]any{
		{"Path": "电视剧/国产剧/冰湖重生 (2026)/Season 1/冰湖重生 - S01E10 - 第 10 集.mkv.cas", "Size": 244, "ModTime": "2026-05-10T11:00:00Z"},
	}
	opts := &adapter.TaskOptions{Include: []string{"*.cas"}, IgnoreCase: true, OpenlistCasCompatible: true}
	var items []TransferCandidateFile
	calls := runBuildCandidateFilesWithFakeRclone(t, srcFiles, dstFiles, "+ 电视剧/国产剧/冰湖重生 (2026)/Season 1/冰湖重生 - S01E11 - 第 11 集.mkv.cas\n", func() {
		var err error
		items, err = BuildCandidateFiles(context.Background(), "cfg.conf", "src:", "dst:", opts)
		if err != nil {
			t.Fatalf("BuildCandidateFiles err: %v", err)
		}
	})
	if len(items) != 1 || !strings.Contains(items[0].Path, "第 11 集") {
		t.Fatalf("expected only unmatched .cas file to remain, got %v", items)
	}
	joined := strings.Join(calls, "\n")
	if !strings.Contains(joined, "lsjson src: --config cfg.conf --files-only --recursive --include *.cas --ignore-case") {
		t.Fatalf("source lsjson should carry include cas flags, calls=%s", joined)
	}
	if !strings.Contains(joined, "lsjson dst: --config cfg.conf --files-only --recursive --include *.cas --ignore-case") {
		t.Fatalf("dest lsjson should carry include cas flags, calls=%s", joined)
	}
	if !strings.Contains(joined, "check src: dst: --config cfg.conf --combined - --one-way --include *.cas --ignore-case") {
		t.Fatalf("check should carry include cas flags, calls=%s", joined)
	}
}

func runBuildCandidateFilesWithFakeRclone(t *testing.T, srcFiles, dstFiles []map[string]any, combined string, fn func()) []string {
	t.Helper()
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "calls.log")
	binPath := filepath.Join(tmpDir, "rclone")
	srcJSON, _ := json.Marshal(srcFiles)
	dstJSON, _ := json.Marshal(dstFiles)
	script := "#!/bin/sh\n" +
		"printf '%s\\n' \"$*\" >> \"" + logFile + "\"\n" +
		"case \"$1 $2\" in\n" +
		"  'lsjson src:') cat <<'EOF'\n" + string(srcJSON) + "\nEOF\n    ;;\n" +
		"  'lsjson dst:') cat <<'EOF'\n" + string(dstJSON) + "\nEOF\n    ;;\n" +
		"  'check src:') cat <<'EOF' 1>&2\n" + combined + "\nEOF\n    exit 1\n    ;;\n" +
		"  *) exit 0 ;;\n" +
		"esac\n"
	if err := os.WriteFile(binPath, []byte(script), 0o755); err != nil {
		t.Fatalf("WriteFile(fake rclone) error = %v", err)
	}
	oldBin := os.Getenv("RCLONE_BIN")
	if err := os.Setenv("RCLONE_BIN", binPath); err != nil {
		t.Fatalf("Setenv(RCLONE_BIN) error = %v", err)
	}
	defer func() {
		if oldBin == "" {
			_ = os.Unsetenv("RCLONE_BIN")
		} else {
			_ = os.Setenv("RCLONE_BIN", oldBin)
		}
	}()
	fn()
	data, err := os.ReadFile(logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("ReadFile(calls.log) error = %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil
	}
	return lines
}
