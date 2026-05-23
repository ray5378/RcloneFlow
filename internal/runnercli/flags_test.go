package runnercli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildFlagsFromOptions_Empty(t *testing.T) {
	flags := buildFlagsFromOptions(map[string]any{})
	assert.Empty(t, flags)
}

func TestBuildFlagsFromOptions_Nil(t *testing.T) {
	flags := buildFlagsFromOptions(nil)
	assert.Empty(t, flags)
}

func TestBuildFlagsFromOptions_IntTypes(t *testing.T) {
	opt := map[string]any{
		"transfers":      8,
		"checkers":      16,
		"retries":       5,
		"lowLevelRetries": 3,
	}
	flags := buildFlagsFromOptions(opt)
	assert.Contains(t, flags, "--transfers")
	assert.Contains(t, flags, "8")
	assert.Contains(t, flags, "--checkers")
	assert.Contains(t, flags, "16")
	assert.Contains(t, flags, "--retries")
	assert.Contains(t, flags, "5")
	assert.Contains(t, flags, "--low-level-retries")
	assert.Contains(t, flags, "3")
}

func TestBuildFlagsFromOptions_Float64Types(t *testing.T) {
	opt := map[string]any{
		"transfers": float64(4),
		"checkers":  float64(8),
	}
	flags := buildFlagsFromOptions(opt)
	assert.Contains(t, flags, "--transfers")
	assert.Contains(t, flags, "4")
	assert.Contains(t, flags, "--checkers")
	assert.Contains(t, flags, "8")
}

func TestBuildFlagsFromOptions_Int64Types(t *testing.T) {
	opt := map[string]any{
		"transfers": int64(12),
	}
	flags := buildFlagsFromOptions(opt)
	assert.Contains(t, flags, "12")
}

func TestBuildFlagsFromOptions_StringInt(t *testing.T) {
	opt := map[string]any{
		"transfers": "  10  ",
	}
	flags := buildFlagsFromOptions(opt)
	assert.Contains(t, flags, "10")
}

func TestBuildFlagsFromOptions_EmptyString(t *testing.T) {
	opt := map[string]any{
		"transfers": "   ",
	}
	flags := buildFlagsFromOptions(opt)
	assert.NotContains(t, flags, "transfers")
}

func TestBuildFlagsFromOptions_BufferSize(t *testing.T) {
	opt := map[string]any{
		"bufferSize": float64(256),
	}
	flags := buildFlagsFromOptions(opt)
	assert.Contains(t, flags, "--buffer-size")
	assert.Contains(t, flags, "256M")
}

func TestBuildFlagsFromOptions_BufferSizeInt64(t *testing.T) {
	opt := map[string]any{
		"bufferSize": int64(512),
	}
	flags := buildFlagsFromOptions(opt)
	assert.Contains(t, flags, "512M")
}

func TestBuildFlagsFromOptions_BufferSizeStringPureNumber(t *testing.T) {
	opt := map[string]any{
		"bufferSize": "128",
	}
	flags := buildFlagsFromOptions(opt)
	assert.Contains(t, flags, "128M")
}

func TestBuildFlagsFromOptions_BufferSizeStringWithUnit(t *testing.T) {
	opt := map[string]any{
		"bufferSize": "256M",
	}
	flags := buildFlagsFromOptions(opt)
	assert.Contains(t, flags, "256M")
}

func TestBuildFlagsFromOptions_BufferSizeStringWithK(t *testing.T) {
	opt := map[string]any{
		"bufferSize": "64K",
	}
	flags := buildFlagsFromOptions(opt)
	assert.Contains(t, flags, "64K")
}

func TestBuildFlagsFromOptions_BwLimit(t *testing.T) {
	opt := map[string]any{
		"bwLimit": "10M",
	}
	flags := buildFlagsFromOptions(opt)
	assert.Contains(t, flags, "--bwlimit")
	assert.Contains(t, flags, "10M")
}

func TestBuildFlagsFromOptions_BoolFlags(t *testing.T) {
	opt := map[string]any{
		"checksum":       true,
		"dryRun":         true,
		"ignoreExisting": true,
	}
	flags := buildFlagsFromOptions(opt)
	assert.Contains(t, flags, "--checksum")
	assert.Contains(t, flags, "--dryrun")
	assert.Contains(t, flags, "--ignoreexisting")
}

func TestBuildFlagsFromOptions_BoolFalse(t *testing.T) {
	opt := map[string]any{
		"checksum": false,
		"dryRun":   false,
	}
	flags := buildFlagsFromOptions(opt)
	assert.NotContains(t, flags, "--checksum")
	assert.NotContains(t, flags, "--dry-run")
}

func TestBuildFlagsFromOptions_StringFlags(t *testing.T) {
	opt := map[string]any{
		"compareDest": "remote:compare",
		"copyDest":    "remote:backup",
		"backupDir":   "remote:backup-dir",
	}
	flags := buildFlagsFromOptions(opt)
	assert.Contains(t, flags, "--compare-dest")
	assert.Contains(t, flags, "remote:compare")
	assert.Contains(t, flags, "--copy-dest")
	assert.Contains(t, flags, "remote:backup")
	assert.Contains(t, flags, "--backup-dir")
	assert.Contains(t, flags, "remote:backup-dir")
}

func TestBuildFlagsFromOptions_ExcludeArray(t *testing.T) {
	opt := map[string]any{
		"exclude": []string{"*.tmp", "*.log", "*.bak"},
	}
	flags := buildFlagsFromOptions(opt)
	assert.Contains(t, flags, "--exclude")
	assert.Contains(t, flags, "*.tmp")
	assert.Contains(t, flags, "*.log")
	assert.Contains(t, flags, "*.bak")
}

func TestBuildFlagsFromOptions_ExcludeAnyArray(t *testing.T) {
	opt := map[string]any{
		"exclude": []any{"*.tmp", "*.log"},
	}
	flags := buildFlagsFromOptions(opt)
	assert.Contains(t, flags, "*.tmp")
	assert.Contains(t, flags, "*.log")
}

func TestBuildFlagsFromOptions_ExcludeCommaString(t *testing.T) {
	opt := map[string]any{
		"exclude": "*.tmp, *.log, *.bak",
	}
	flags := buildFlagsFromOptions(opt)
	assert.Contains(t, flags, "*.tmp")
	assert.Contains(t, flags, "*.log")
	assert.Contains(t, flags, "*.bak")
}

func TestBuildFlagsFromOptions_ExcludeNewlineString(t *testing.T) {
	opt := map[string]any{
		"exclude": "*.tmp\n*.log\n.git/*",
	}
	flags := buildFlagsFromOptions(opt)
	assert.Contains(t, flags, "*.tmp")
	assert.Contains(t, flags, "*.log")
	assert.Contains(t, flags, ".git/*")
}

func TestBuildFlagsFromOptions_ExcludeCarriageReturn(t *testing.T) {
	opt := map[string]any{
		"exclude": "*.tmp\r\n*.log",
	}
	flags := buildFlagsFromOptions(opt)
	assert.Contains(t, flags, "*.tmp")
	assert.Contains(t, flags, "*.log")
}

func TestBuildFlagsFromOptions_IncludeArray(t *testing.T) {
	opt := map[string]any{
		"include": []string{"*.mp4", "*.avi"},
	}
	flags := buildFlagsFromOptions(opt)
	assert.Contains(t, flags, "--include")
	assert.Contains(t, flags, "*.mp4")
	assert.Contains(t, flags, "*.avi")
}

func TestBuildFlagsFromOptions_Deduplication(t *testing.T) {
	opt := map[string]any{
		"exclude": []string{"*.tmp", "*.tmp", "*.log"},
	}
	flags := buildFlagsFromOptions(opt)
	count := 0
	for _, f := range flags {
		if f == "*.tmp" {
			count++
		}
	}
	assert.Equal(t, 1, count)
}

func TestBuildFlagsFromOptions_AutoGlob(t *testing.T) {
	opt := map[string]any{
		"exclude": ".tmp",
	}
	flags := buildFlagsFromOptions(opt)
	assert.Contains(t, flags, "*.tmp")
}

func TestBuildFlagsFromOptions_NoAutoGlob(t *testing.T) {
	opt := map[string]any{
		"exclude": "*.tmp",
	}
	flags := buildFlagsFromOptions(opt)
	assert.Contains(t, flags, "*.tmp")
	assert.NotContains(t, flags, "**/*.tmp")
}

func TestBuildFlagsFromOptions_Timeout(t *testing.T) {
	opt := map[string]any{
		"timeout": 300,
	}
	flags := buildFlagsFromOptions(opt)
	assert.Contains(t, flags, "--timeout")
	assert.Contains(t, flags, "300s")
}

func TestBuildFlagsFromOptions_ConnTimeout(t *testing.T) {
	opt := map[string]any{
		"connTimeout": 60,
	}
	flags := buildFlagsFromOptions(opt)
	assert.Contains(t, flags, "--contimeout")
	assert.Contains(t, flags, "60s")
}

func TestBuildFlagsFromOptions_MaxTransfer(t *testing.T) {
	opt := map[string]any{
		"maxTransfer": 1073741824,
	}
	flags := buildFlagsFromOptions(opt)
	assert.Contains(t, flags, "--max-transfer")
	assert.Contains(t, flags, "1073741824")
}

func TestBuildFlagsFromOptions_MaxDuration(t *testing.T) {
	opt := map[string]any{
		"maxDuration": 3600,
	}
	flags := buildFlagsFromOptions(opt)
	assert.Contains(t, flags, "--max-duration")
	assert.Contains(t, flags, "3600s")
}

func TestBuildFlagsFromOptions_LogFile(t *testing.T) {
	opt := map[string]any{
		"logFile": "/path/to/log",
	}
	flags := buildFlagsFromOptions(opt)
	assert.Contains(t, flags, "--log-file")
	assert.Contains(t, flags, "/path/to/log")
}

func TestBuildFlagsFromOptions_DisableHttp2(t *testing.T) {
	opt := map[string]any{
		"disableHttp2": true,
	}
	flags := buildFlagsFromOptions(opt)
	assert.Contains(t, flags, "--disable-http2")
}

func TestBuildFlagsFromOptions_FullOptions(t *testing.T) {
	opt := map[string]any{
		"transfers":        float64(8),
		"checkers":         float64(16),
		"retries":          float64(3),
		"lowLevelRetries":  float64(5),
		"bufferSize":       "256M",
		"bwLimit":          "10M",
		"checksum":         true,
		"dryRun":           true,
		"exclude":          []string{"*.tmp", "*.log"},
		"include":          []string{"*.mp4"},
	}
	flags := buildFlagsFromOptions(opt)
	assert.Contains(t, flags, "--transfers")
	assert.Contains(t, flags, "--checkers")
	assert.Contains(t, flags, "--buffer-size")
	assert.Contains(t, flags, "--bwlimit")
	assert.Contains(t, flags, "--checksum")
	assert.Contains(t, flags, "--dryrun")
	assert.Contains(t, flags, "*.tmp")
	assert.Contains(t, flags, "*.mp4")
}

func TestToKebab(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"useServerModtime", "use-server-modtime"},
		{"noCheckDest", "no-check-dest"},
		{"checksum", "checksum"},
		{"ignoreErrors", "ignore-errors"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, toKebab(tt.input))
		})
	}
}

func TestAddFilterFlags_Nil(t *testing.T) {
	args := []string{"rclone", "copy"}
	result := addFilterFlags(args, nil, true)
	assert.Equal(t, args, result)
}

func TestAddFilterFlags_Empty(t *testing.T) {
	args := []string{"rclone", "copy"}
	result := addFilterFlags(args, map[string]any{}, true)
	assert.Equal(t, args, result)
}

func TestAddFilterFlags_Include(t *testing.T) {
	args := []string{"rclone", "copy"}
	opts := map[string]any{"include": "*.mp4"}
	result := addFilterFlags(args, opts, true)
	assert.Contains(t, result, "--include")
	assert.Contains(t, result, "*.mp4")
}

func TestAddFilterFlags_Exclude(t *testing.T) {
	args := []string{"rclone", "copy"}
	opts := map[string]any{"exclude": "*.tmp"}
	result := addFilterFlags(args, opts, true)
	assert.Contains(t, result, "--exclude")
	assert.Contains(t, result, "*.tmp")
}

func TestAddFilterFlags_Filter(t *testing.T) {
	args := []string{"rclone", "copy"}
	opts := map[string]any{"filter": "- *.tmp"}
	result := addFilterFlags(args, opts, true)
	assert.Contains(t, result, "--filter")
	assert.Contains(t, result, "- *.tmp")
}

func TestAddFilterFlags_FilterFrom(t *testing.T) {
	args := []string{"rclone", "copy"}
	opts := map[string]any{"filterFrom": "/path/to/filter.txt"}
	result := addFilterFlags(args, opts, true)
	assert.Contains(t, result, "--filter-from")
	assert.Contains(t, result, "/path/to/filter.txt")
}

func TestAddFilterFlags_IncludeFrom(t *testing.T) {
	args := []string{"rclone", "copy"}
	opts := map[string]any{"includeFrom": "/path/to/include.txt"}
	result := addFilterFlags(args, opts, true)
	assert.Contains(t, result, "--include-from")
	assert.Contains(t, result, "/path/to/include.txt")
}

func TestAddFilterFlags_ExcludeFrom(t *testing.T) {
	args := []string{"rclone", "copy"}
	opts := map[string]any{"excludeFrom": "/path/to/exclude.txt"}
	result := addFilterFlags(args, opts, true)
	assert.Contains(t, result, "--exclude-from")
	assert.Contains(t, result, "/path/to/exclude.txt")
}

func TestAddFilterFlags_FilesFrom(t *testing.T) {
	args := []string{"rclone", "copy"}
	opts := map[string]any{"filesFrom": "/path/to/files.txt"}
	result := addFilterFlags(args, opts, true)
	assert.Contains(t, result, "--files-from")
	assert.Contains(t, result, "/path/to/files.txt")
}

func TestAddFilterFlags_MinSize(t *testing.T) {
	args := []string{"rclone", "copy"}
	opts := map[string]any{"minSize": "1M"}
	result := addFilterFlags(args, opts, true)
	assert.Contains(t, result, "--min-size")
	assert.Contains(t, result, "1M")
}

func TestAddFilterFlags_MaxSize(t *testing.T) {
	args := []string{"rclone", "copy"}
	opts := map[string]any{"maxSize": "100M"}
	result := addFilterFlags(args, opts, true)
	assert.Contains(t, result, "--max-size")
	assert.Contains(t, result, "100M")
}

func TestAddFilterFlags_MinAge(t *testing.T) {
	args := []string{"rclone", "copy"}
	opts := map[string]any{"minAge": "1d"}
	result := addFilterFlags(args, opts, true)
	assert.Contains(t, result, "--min-age")
	assert.Contains(t, result, "1d")
}

func TestAddFilterFlags_MaxAge(t *testing.T) {
	args := []string{"rclone", "copy"}
	opts := map[string]any{"maxAge": "30d"}
	result := addFilterFlags(args, opts, true)
	assert.Contains(t, result, "--max-age")
	assert.Contains(t, result, "30d")
}

func TestAddFilterFlags_FastList_True(t *testing.T) {
	args := []string{"rclone", "copy"}
	opts := map[string]any{"fastList": true}
	result := addFilterFlags(args, opts, true)
	assert.Contains(t, result, "--fast-list")
}

func TestAddFilterFlags_FastList_StringTrue(t *testing.T) {
	args := []string{"rclone", "copy"}
	opts := map[string]any{"fastList": "true"}
	result := addFilterFlags(args, opts, true)
	assert.Contains(t, result, "--fast-list")
}

func TestAddFilterFlags_FastList_String1(t *testing.T) {
	args := []string{"rclone", "copy"}
	opts := map[string]any{"fastList": "1"}
	result := addFilterFlags(args, opts, true)
	assert.Contains(t, result, "--fast-list")
}

func TestAddFilterFlags_FastList_False(t *testing.T) {
	args := []string{"rclone", "copy"}
	opts := map[string]any{"fastList": false}
	result := addFilterFlags(args, opts, true)
	assert.NotContains(t, result, "--fast-list")
}

func TestAddFilterFlags_FastList_Disabled(t *testing.T) {
	args := []string{"rclone", "copy"}
	opts := map[string]any{"fastList": true}
	result := addFilterFlags(args, opts, false)
	assert.NotContains(t, result, "--fast-list")
}

func TestAddFilterFlags_Multiple(t *testing.T) {
	args := []string{"rclone", "copy"}
	opts := map[string]any{
		"exclude":  "*.tmp",
		"include":  "*.mp4",
		"minSize": "1M",
		"maxAge":  "30d",
	}
	result := addFilterFlags(args, opts, true)
	assert.Contains(t, result, "--exclude")
	assert.Contains(t, result, "--include")
	assert.Contains(t, result, "--min-size")
	assert.Contains(t, result, "--max-age")
}
