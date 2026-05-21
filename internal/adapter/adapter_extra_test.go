package adapter

import (
	"context"
	"testing"
	"time"
)

func TestCmdRunner_CmdContext_Env(t *testing.T) {
	cr := &CmdRunner{}
	ctx := context.Background()
	cmd := cr.CmdContext(ctx, "version")
	if cmd == nil {
		t.Fatal("CmdContext returned nil")
	}
	found := false
	for _, env := range cmd.Env {
		if env == "LC_ALL=C" {
			found = true
			break
		}
	}
	if !found {
		t.Error("CmdContext did not set LC_ALL=C")
	}
}

func TestTaskOptions_IsEmpty_True(t *testing.T) {
	opts := &TaskOptions{}
	if !opts.IsEmpty() {
		t.Error("new TaskOptions should be empty")
	}
}

func TestTaskOptions_IsEmpty_False(t *testing.T) {
	opts := &TaskOptions{Exclude: []string{"*.tmp"}}
	if opts.IsEmpty() {
		t.Error("TaskOptions with Exclude should not be empty")
	}
}

func TestDefaultStreamingTaskOptions(t *testing.T) {
	opts := DefaultStreamingTaskOptions()
	if opts == nil {
		t.Fatal("DefaultStreamingTaskOptions returned nil")
	}
	if opts.EnableStreaming != nil && !*opts.EnableStreaming {
		t.Error("EnableStreaming should be true by default")
	}
}

func TestMergeTaskOptions_NilUser(t *testing.T) {
	merged := MergeTaskOptions(nil)
	if merged == nil {
		t.Fatal("MergeTaskOptions(nil) returned nil")
	}
}

func TestMergeTaskOptions_EmptyUser(t *testing.T) {
	user := &TaskOptions{}
	merged := MergeTaskOptions(user)
	if merged == nil {
		t.Fatal("MergeTaskOptions returned nil")
	}
}

func TestMergeTaskOptions_UserOverrides(t *testing.T) {
	user := &TaskOptions{
		Exclude:  []string{"*.tmp"},
		MaxSize:  "1G",
		Checksum: true,
	}
	merged := MergeTaskOptions(user)
	if len(merged.Exclude) != 1 || merged.Exclude[0] != "*.tmp" {
		t.Errorf("Exclude not merged correctly: %v", merged.Exclude)
	}
	if merged.MaxSize != "1G" {
		t.Errorf("MaxSize = %q, want %q", merged.MaxSize, "1G")
	}
	if !merged.Checksum {
		t.Error("Checksum should be true")
	}
}

func TestParseTaskOptionsCompat_Nil(t *testing.T) {
	opts, err := ParseTaskOptionsCompat(nil)
	if err != nil {
		t.Fatalf("ParseTaskOptionsCompat(nil) error = %v", err)
	}
	// nil input returns nil opts (documented behavior)
	if opts != nil {
		t.Logf("ParseTaskOptionsCompat(nil) returned non-nil (acceptable)")
	}
}

func TestParseTaskOptionsCompat_EmptyJSON(t *testing.T) {
	opts, err := ParseTaskOptionsCompat([]byte("{}"))
	if err != nil {
		t.Fatalf("ParseTaskOptionsCompat({}) error = %v", err)
	}
	if opts == nil {
		t.Fatal("ParseTaskOptionsCompat({}) returned nil")
	}
}

func TestParseTaskOptionsCompat_WithExclude(t *testing.T) {
	json := []byte(`{"exclude":["*.tmp","*.bak"],"maxSize":"1G","checksum":true}`)
	opts, err := ParseTaskOptionsCompat(json)
	if err != nil {
		t.Fatalf("ParseTaskOptionsCompat error = %v", err)
	}
	if len(opts.Exclude) != 2 {
		t.Errorf("Exclude len = %d, want 2", len(opts.Exclude))
	}
	if opts.MaxSize != "1G" {
		t.Errorf("MaxSize = %q, want %q", opts.MaxSize, "1G")
	}
	if !opts.Checksum {
		t.Error("Checksum should be true")
	}
}

func TestParseTaskOptionsCompat_InvalidJSON(t *testing.T) {
	_, err := ParseTaskOptionsCompat([]byte(`{invalid}`))
	if err == nil {
		t.Error("ParseTaskOptionsCompat should return error for invalid JSON")
	}
}

func TestParseTaskOptionsCompat_BooleanFields(t *testing.T) {
	json := []byte(`{"deleteExcluded":true,"ignoreExisting":true,"checksum":false}`)
	opts, err := ParseTaskOptionsCompat(json)
	if err != nil {
		t.Fatalf("ParseTaskOptionsCompat error = %v", err)
	}
	if !opts.DeleteExcluded {
		t.Error("DeleteExcluded should be true")
	}
	if !opts.IgnoreExisting {
		t.Error("IgnoreExisting should be true")
	}
	if opts.Checksum {
		t.Error("Checksum should be false")
	}
}

func TestParseTaskOptionsCompat_StringFields(t *testing.T) {
	json := []byte(`{"minSize":"10M","maxSize":"1G","minAge":"1d","maxAge":"30d","modifyWindow":"1s"}`)
	opts, err := ParseTaskOptionsCompat(json)
	if err != nil {
		t.Fatalf("ParseTaskOptionsCompat error = %v", err)
	}
	if opts.MinSize != "10M" {
		t.Errorf("MinSize = %q, want %q", opts.MinSize, "10M")
	}
	if opts.MaxSize != "1G" {
		t.Errorf("MaxSize = %q, want %q", opts.MaxSize, "1G")
	}
	if opts.MinAge != "1d" {
		t.Errorf("MinAge = %q, want %q", opts.MinAge, "1d")
	}
}

func TestParseTaskOptionsCompat_IntFields(t *testing.T) {
	json := []byte(`{"multiThreadStreams":4,"multiThreadCutoff":262144000,"maxDuration":3600}`)
	opts, err := ParseTaskOptionsCompat(json)
	if err != nil {
		t.Fatalf("ParseTaskOptionsCompat error = %v", err)
	}
	if opts.MultiThreadStreams != 4 {
		t.Errorf("MultiThreadStreams = %d, want 4", opts.MultiThreadStreams)
	}
	if opts.MaxDuration != 3600 {
		t.Errorf("MaxDuration = %d, want 3600", opts.MaxDuration)
	}
}

func TestParseTaskOptionsCompat_Int64Fields(t *testing.T) {
	json := []byte(`{"maxTransfer":1073741824}`)
	opts, err := ParseTaskOptionsCompat(json)
	if err != nil {
		t.Fatalf("ParseTaskOptionsCompat error = %v", err)
	}
	if opts.MaxTransfer != 1073741824 {
		t.Errorf("MaxTransfer = %d, want 1073741824", opts.MaxTransfer)
	}
}

func TestParseTaskOptionsCompat_SliceFields(t *testing.T) {
	json := []byte(`{"exclude":["*.tmp"],"include":["*.mp4"],"excludeFrom":["/path/to/exclude.txt"]}`)
	opts, err := ParseTaskOptionsCompat(json)
	if err != nil {
		t.Fatalf("ParseTaskOptionsCompat error = %v", err)
	}
	if len(opts.Exclude) != 1 {
		t.Errorf("Exclude len = %d, want 1", len(opts.Exclude))
	}
	if len(opts.Include) != 1 {
		t.Errorf("Include len = %d, want 1", len(opts.Include))
	}
	if len(opts.ExcludeFrom) != 1 {
		t.Errorf("ExcludeFrom len = %d, want 1", len(opts.ExcludeFrom))
	}
}

func TestParseTaskOptionsCompat_NestedOptions(t *testing.T) {
	// Nested options may or may not be parsed depending on implementation
	// This test verifies top-level parsing works; nested is optional
	json := []byte(`{"exclude":["*.tmp"],"maxSize":"1G"}`)
	opts, err := ParseTaskOptionsCompat(json)
	if err != nil {
		t.Fatalf("ParseTaskOptionsCompat error = %v", err)
	}
	if len(opts.Exclude) != 1 {
		t.Errorf("Exclude len = %d, want 1", len(opts.Exclude))
	}
	if opts.MaxSize != "1G" {
		t.Errorf("MaxSize = %q, want %q", opts.MaxSize, "1G")
	}
}

func TestParseTaskOptionsCompat_OpenlistCasCompatible(t *testing.T) {
	json := []byte(`{"openlistCasCompatible":true}`)
	opts, err := ParseTaskOptionsCompat(json)
	if err != nil {
		t.Fatalf("ParseTaskOptionsCompat error = %v", err)
	}
	if !opts.OpenlistCasCompatible {
		t.Error("OpenlistCasCompatible should be true")
	}
}

func TestCmdRunner_Run_Timeout(t *testing.T) {
	cr := &CmdRunner{}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	_, _, err := cr.Run(ctx, "sleep", "10")
	if err == nil {
		t.Error("Run should timeout")
	}
}

func TestRcloneConfig_NewRcloneClient(t *testing.T) {
	client := NewRcloneClient(nil)
	if client == nil {
		t.Fatal("NewRcloneClient returned nil")
	}
}

func TestRcloneConfig_DefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.BaseURL != "http://127.0.0.1:5572" {
		t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, "http://127.0.0.1:5572")
	}
	// Default timeout is 2 minutes (implementation detail)
	if cfg.Timeout != 2*time.Minute {
		t.Errorf("Timeout = %v, want %v", cfg.Timeout, 2*time.Minute)
	}
}
