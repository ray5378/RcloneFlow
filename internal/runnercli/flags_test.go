package runnercli

import (
	"testing"
)

func TestExistsStr(t *testing.T) {
	if existsStr(nil, "key") {
		t.Error("expected false for nil map")
	}
	m := map[string]any{"key": "value"}
	if !existsStr(m, "key") {
		t.Error("expected true for string value")
	}
	m2 := map[string]any{"key": 123}
	if existsStr(m2, "key") {
		t.Error("expected false for non-string value")
	}
	m3 := map[string]any{"other": "value"}
	if existsStr(m3, "key") {
		t.Error("expected false for missing key")
	}
}

func TestExistsBool(t *testing.T) {
	if existsBool(nil, "key") {
		t.Error("expected false for nil map")
	}
	m := map[string]any{"key": true}
	if !existsBool(m, "key") {
		t.Error("expected true for bool value")
	}
	m2 := map[string]any{"key": "true"}
	if existsBool(m2, "key") {
		t.Error("expected false for non-bool value")
	}
	m3 := map[string]any{"other": true}
	if existsBool(m3, "key") {
		t.Error("expected false for missing key")
	}
}

func TestEff(t *testing.T) {
	if eff(nil) != nil {
		t.Error("expected nil for nil map")
	}
	m := map[string]any{"effectiveOptions": map[string]any{"key": "value"}}
	result := eff(m)
	if result == nil || result["key"] != "value" {
		t.Errorf("expected effectiveOptions, got %#v", result)
	}
	m2 := map[string]any{"effectiveOptions": "not-a-map"}
	if eff(m2) != nil {
		t.Error("expected nil for non-map effectiveOptions")
	}
	m3 := map[string]any{"other": "value"}
	if eff(m3) != nil {
		t.Error("expected nil when no effectiveOptions key")
	}
}

func TestBuildFlagsFromOptions_Numeric(t *testing.T) {
	opts := map[string]any{
		"transfers":       float64(8),
		"checkers":        int64(4),
		"retries":         3,
		"lowLevelRetries": "2",
	}
	flags := buildFlagsFromOptions(opts)
	if !containsFlag(flags, "--transfers", "8") {
		t.Errorf("missing --transfers 8 in %v", flags)
	}
	if !containsFlag(flags, "--checkers", "4") {
		t.Errorf("missing --checkers 4 in %v", flags)
	}
	if !containsFlag(flags, "--retries", "3") {
		t.Errorf("missing --retries 3 in %v", flags)
	}
	if !containsFlag(flags, "--low-level-retries", "2") {
		t.Errorf("missing --low-level-retries 2 in %v", flags)
	}
}

func TestBuildFlagsFromOptions_BufferSize(t *testing.T) {
	// float64 → auto M
	opts1 := map[string]any{"bufferSize": float64(16)}
	flags1 := buildFlagsFromOptions(opts1)
	if !containsFlag(flags1, "--buffer-size", "16M") {
		t.Errorf("missing --buffer-size 16M in %v", flags1)
	}

	// int64 → auto M
	opts2 := map[string]any{"bufferSize": int64(32)}
	flags2 := buildFlagsFromOptions(opts2)
	if !containsFlag(flags2, "--buffer-size", "32M") {
		t.Errorf("missing --buffer-size 32M in %v", flags2)
	}

	// string pure number → auto M
	opts3 := map[string]any{"bufferSize": "64"}
	flags3 := buildFlagsFromOptions(opts3)
	if !containsFlag(flags3, "--buffer-size", "64M") {
		t.Errorf("missing --buffer-size 64M in %v", flags3)
	}

	// string with unit → keep as-is
	opts4 := map[string]any{"bufferSize": "1G"}
	flags4 := buildFlagsFromOptions(opts4)
	if !containsFlag(flags4, "--buffer-size", "1G") {
		t.Errorf("missing --buffer-size 1G in %v", flags4)
	}
}

func TestBuildFlagsFromOptions_Booleans(t *testing.T) {
	opts := map[string]any{
		"ignoreExisting":    true,
		"checksum":          true,
		"sizeOnly":          true,
		"useServerModtime":  true,
		"disableHttp2":      true,
		"ignoreExisting2":   false, // should not appear
	}
	flags := buildFlagsFromOptions(opts)
	if !containsFlagBool(flags, "--ignoreexisting") {
		t.Errorf("missing --ignoreexisting in %v", flags)
	}
	if !containsFlagBool(flags, "--checksum") {
		t.Errorf("missing --checksum in %v", flags)
	}
	if !containsFlagBool(flags, "--size-only") {
		t.Errorf("missing --size-only in %v", flags)
	}
	if !containsFlagBool(flags, "--use-server-modtime") {
		t.Errorf("missing --use-server-modtime in %v", flags)
	}
	if !containsFlagBool(flags, "--disable-http2") {
		t.Errorf("missing --disable-http2 in %v", flags)
	}
}

func TestBuildFlagsFromOptions_StringValues(t *testing.T) {
	opts := map[string]any{
		"bwLimit":     "10M",
		"compareDest": "/compare",
		"copyDest":    "/copy",
		"backupDir":   "/backup",
		"logFile":     "/tmp/log.txt",
	}
	flags := buildFlagsFromOptions(opts)
	if !containsFlag(flags, "--bwlimit", "10M") {
		t.Errorf("missing --bwlimit 10M in %v", flags)
	}
	if !containsFlag(flags, "--compare-dest", "/compare") {
		t.Errorf("missing --compare-dest in %v", flags)
	}
	if !containsFlag(flags, "--copy-dest", "/copy") {
		t.Errorf("missing --copy-dest in %v", flags)
	}
	if !containsFlag(flags, "--backup-dir", "/backup") {
		t.Errorf("missing --backup-dir in %v", flags)
	}
	if !containsFlag(flags, "--log-file", "/tmp/log.txt") {
		t.Errorf("missing --log-file in %v", flags)
	}
}

func TestBuildFlagsFromOptions_IncludeExclude(t *testing.T) {
	opts := map[string]any{
		"include": []any{"*.jpg", "*.png"},
		"exclude": "*.tmp",
	}
	flags := buildFlagsFromOptions(opts)
	if !containsFlag(flags, "--include", "*.jpg") {
		t.Errorf("missing --include *.jpg in %v", flags)
	}
	if !containsFlag(flags, "--include", "*.png") {
		t.Errorf("missing --include *.png in %v", flags)
	}
	if !containsFlag(flags, "--exclude", "*.tmp") {
		t.Errorf("missing --exclude *.tmp in %v", flags)
	}
}

func TestBuildFlagsFromOptions_IncludeExcludeDedup(t *testing.T) {
	opts := map[string]any{
		"include": []any{"*.jpg", "*.jpg"}, // duplicate
	}
	flags := buildFlagsFromOptions(opts)
	count := 0
	for _, f := range flags {
		if f == "*.jpg" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 --include *.jpg after dedup, got %d in %v", count, flags)
	}
}

func TestBuildFlagsFromOptions_IncludeExcludeNormalizeDotPrefix(t *testing.T) {
	opts := map[string]any{
		"include": ".git",
	}
	flags := buildFlagsFromOptions(opts)
	if !containsFlag(flags, "--include", "*.git") {
		t.Errorf("expected .git normalized to *.git in %v", flags)
	}
}

func TestBuildFlagsFromOptions_Timeouts(t *testing.T) {
	opts := map[string]any{
		"timeout":               float64(60),
		"connTimeout":           "30",
		"expectContinueTimeout": 15,
	}
	flags := buildFlagsFromOptions(opts)
	if !containsFlag(flags, "--timeout", "60s") {
		t.Errorf("missing --timeout 60s in %v", flags)
	}
	if !containsFlag(flags, "--contimeout", "30s") {
		t.Errorf("missing --contimeout 30s in %v", flags)
	}
	if !containsFlag(flags, "--expect-continue-timeout", "15s") {
		t.Errorf("missing --expect-continue-timeout 15s in %v", flags)
	}
}

func TestBuildFlagsFromOptions_MaxTransferAndDuration(t *testing.T) {
	opts := map[string]any{
		"maxTransfer":  float64(1073741824),
		"maxDuration":  3600,
	}
	flags := buildFlagsFromOptions(opts)
	if !containsFlag(flags, "--max-transfer", "1073741824") {
		t.Errorf("missing --max-transfer in %v", flags)
	}
	if !containsFlag(flags, "--max-duration", "3600s") {
		t.Errorf("missing --max-duration in %v", flags)
	}
}

func TestBuildFlagsFromOptions_Empty(t *testing.T) {
	flags := buildFlagsFromOptions(nil)
	if len(flags) != 0 {
		t.Errorf("expected empty flags for nil opts, got %v", flags)
	}
	flags2 := buildFlagsFromOptions(map[string]any{})
	if len(flags2) != 0 {
		t.Errorf("expected empty flags for empty opts, got %v", flags2)
	}
}

func TestBuildFlagsFromOptions_BwlimitLowercase(t *testing.T) {
	opts := map[string]any{"bwlimit": "5M"}
	flags := buildFlagsFromOptions(opts)
	if !containsFlag(flags, "--bwlimit", "5M") {
		t.Errorf("missing --bwlimit 5M in %v", flags)
	}
}

func TestToKebab(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"useServerModtime", "use-server-modtime"},
		{"noCheckDest", "no-check-dest"},
		{"noTraverse", "no-traverse"},
		{"sizeOnly", "size-only"},
		{"ignoreSize", "ignore-size"},
		{"ignoreTimes", "ignore-times"},
		{"checkFirst", "check-first"},
		{"deleteBefore", "delete-before"},
		{"deleteDuring", "delete-during"},
		{"deleteAfter", "delete-after"},
		{"trackRenames", "track-renames"},
		{"ignoreErrors", "ignore-errors"},
		{"bufferSize", "buffer-size"},
		{"serverSideAcrossConfigs", "server-side-across-configs"},
		{"unknownKey", "unknownkey"},
	}
	for _, tt := range tests {
		got := toKebab(tt.input)
		if got != tt.expected {
			t.Errorf("toKebab(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func containsFlag(flags []string, key, value string) bool {
	for i, f := range flags {
		if f == key && i+1 < len(flags) && flags[i+1] == value {
			return true
		}
	}
	return false
}

func containsFlagBool(flags []string, key string) bool {
	for _, f := range flags {
		if f == key {
			return true
		}
	}
	return false
}
