package adapter

import (
	"encoding/json"
	"testing"
)

func TestParseTaskOptionsCompat_LegacyShapes(t *testing.T) {
	raw := []byte(`{"enableStreaming":true,"exclude":"","include":"*.cas","ignoreCase":true,"multiThreadStreams":true}`)
	opts, err := ParseTaskOptionsCompat(raw)
	if err != nil {
		t.Fatalf("ParseTaskOptionsCompat err: %v", err)
	}
	if opts == nil {
		t.Fatalf("opts is nil")
	}
	if len(opts.Include) != 1 || opts.Include[0] != "*.cas" {
		t.Fatalf("include=%v", opts.Include)
	}
	if len(opts.Exclude) != 0 {
		t.Fatalf("exclude=%v", opts.Exclude)
	}
	if !opts.IgnoreCase {
		t.Fatalf("ignoreCase should be true")
	}
	if opts.MultiThreadStreams != 4 {
		t.Fatalf("multiThreadStreams=%d want 4", opts.MultiThreadStreams)
	}
}

func TestParseTaskOptionsCompat_ArrayShapesStayIntact(t *testing.T) {
	raw := []byte(`{"exclude":["*.cache"],"include":[],"multiThreadStreams":3}`)
	opts, err := ParseTaskOptionsCompat(raw)
	if err != nil {
		t.Fatalf("ParseTaskOptionsCompat err: %v", err)
	}
	if len(opts.Exclude) != 1 || opts.Exclude[0] != "*.cache" {
		t.Fatalf("exclude=%v", opts.Exclude)
	}
	if len(opts.Include) != 0 {
		t.Fatalf("include=%v", opts.Include)
	}
	if opts.MultiThreadStreams != 3 {
		t.Fatalf("multiThreadStreams=%d want 3", opts.MultiThreadStreams)
	}
}

func TestParseTaskOptionsCompat_RealTask1Shape_ExcludeCachePreserved(t *testing.T) {
	raw := []byte(`{"bwLimit":"3M","enableStreaming":true,"exclude":["*.cache"],"excludeFrom":[],"excludeIfPresent":[],"filesFrom":[],"filesFromRaw":[],"filter":[],"filterFrom":[],"include":[],"includeFrom":[],"multiThreadStreams":true,"singletonMode":true,"transfers":3,"useJsonLog":true}`)
	opts, err := ParseTaskOptionsCompat(raw)
	if err != nil {
		t.Fatalf("ParseTaskOptionsCompat err: %v", err)
	}
	if len(opts.Exclude) != 1 || opts.Exclude[0] != "*.cache" {
		t.Fatalf("exclude=%v", opts.Exclude)
	}
	if opts.MultiThreadStreams != 4 {
		t.Fatalf("multiThreadStreams=%d want 4", opts.MultiThreadStreams)
	}
	if opts.Transfers != 3 {
		t.Fatalf("transfers=%d want 3", opts.Transfers)
	}
}

func TestParseTaskOptionsCompat_RealTask10And11Shapes_IncludeCASPreserved(t *testing.T) {
	cases := [][]byte{
		[]byte(`{"enableStreaming":true,"exclude":"","include":"*.cas","ignoreCase":true}`),
		[]byte(`{"enableStreaming":true,"multiThreadStreams":true,"transfers":3,"exclude":"","include":"*.cas","ignoreExisting":false,"sizeOnly":false}`),
	}
	for i, raw := range cases {
		opts, err := ParseTaskOptionsCompat(raw)
		if err != nil {
			t.Fatalf("case %d ParseTaskOptionsCompat err: %v", i, err)
		}
		if len(opts.Include) != 1 || opts.Include[0] != "*.cas" {
			t.Fatalf("case %d include=%v", i, opts.Include)
		}
		if len(opts.Exclude) != 0 {
			t.Fatalf("case %d exclude=%v", i, opts.Exclude)
		}
	}
}

func TestParseTaskOptionsCompat_ResultStillMarshalsIntoTaskOptionsJSONShape(t *testing.T) {
	raw := []byte(`{"enableStreaming":true,"exclude":"","include":"*.cas","ignoreCase":true,"multiThreadStreams":true}`)
	opts, err := ParseTaskOptionsCompat(raw)
	if err != nil {
		t.Fatalf("ParseTaskOptionsCompat err: %v", err)
	}
	bs, err := json.Marshal(opts)
	if err != nil {
		t.Fatalf("marshal err: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(bs, &decoded); err != nil {
		t.Fatalf("unmarshal marshaled opts err: %v", err)
	}
	if _, ok := decoded["include"].([]any); !ok {
		t.Fatalf("include should marshal as array, got %#v", decoded["include"])
	}
}

func TestMergeTaskOptions_PreservesOpenlistCASCompatible(t *testing.T) {
	merged := MergeTaskOptions(&TaskOptions{OpenlistCasCompatible: true})
	if merged == nil {
		t.Fatalf("merged is nil")
	}
	if !merged.OpenlistCasCompatible {
		t.Fatalf("expected OpenlistCasCompatible to be preserved")
	}
}
