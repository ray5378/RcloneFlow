package runnercli

import "testing"

func TestParseOneLineProgress_AggregateWithXfrAndETA(t *testing.T) {
	line := `2026/04/17 14:34:08 INFO : 121.377 MiB / 335.968 MiB, 36%, 2.474 MiB/s, ETA 1m26s (xfr#18/53)`
	prog, ok := parseOneLineProgress(line)
	if !ok {
		t.Fatalf("expected aggregate line to parse")
	}
	if got := int(prog["completedFiles"].(float64)); got != 18 {
		t.Fatalf("completedFiles=%d, want 18", got)
	}
	if got := int(prog["plannedFiles"].(float64)); got != 53 {
		t.Fatalf("plannedFiles=%d, want 53", got)
	}
	if got := int(prog["eta"].(float64)); got != 86 {
		t.Fatalf("eta=%d, want 86", got)
	}
	if got := int(prog["percentage"].(float64)); got != 36 {
		t.Fatalf("percentage=%d, want 36", got)
	}
}

func TestParseOneLineProgress_AggregateWithETAOnly(t *testing.T) {
	line := `2026/04/17 14:34:09 INFO : 123.377 MiB / 335.968 MiB, 37%, 2.434 MiB/s, ETA 1m27s`
	prog, ok := parseOneLineProgress(line)
	if !ok {
		t.Fatalf("expected aggregate ETA-only line to parse")
	}
	if got := int(prog["eta"].(float64)); got != 87 {
		t.Fatalf("eta=%d, want 87", got)
	}
	if got := int(prog["percentage"].(float64)); got != 37 {
		t.Fatalf("percentage=%d, want 37", got)
	}
	if got := int(prog["bytes"].(float64)); got <= 1024*1024 {
		t.Fatalf("bytes=%d looks wrong; likely matched timestamp instead of MiB payload", got)
	}
	if got := int(prog["totalBytes"].(float64)); got <= 1024*1024 {
		t.Fatalf("totalBytes=%d looks wrong; likely matched timestamp instead of MiB payload", got)
	}
}

func TestParseOneLineProgress_IgnoreFileLevelProgress(t *testing.T) {
	line := `2026/04/17 14:34:08 INFO : 20260417/20260417135617-61000.mp4: 10.000 MiB / 100.000 MiB, 10%, 2.474 MiB/s`
	if prog, ok := parseOneLineProgress(line); ok || prog != nil {
		t.Fatalf("expected file-level progress line to be ignored, got %#v", prog)
	}
}

func TestParseOneLineProgress_IgnoreFileCopied(t *testing.T) {
	line := `2026/04/17 14:34:07 INFO : 20260417/20260417135617-61000.mp4: Copied (new)`
	if prog, ok := parseOneLineProgress(line); ok || prog != nil {
		t.Fatalf("expected copied line to be ignored, got %#v", prog)
	}
}

func TestParseOneLineProgress_IgnoreDeleted(t *testing.T) {
	line := `2026/04/17 14:34:07 INFO : 20260417/20260417135617-61000.mp4: Deleted`
	if prog, ok := parseOneLineProgress(line); ok || prog != nil {
		t.Fatalf("expected deleted line to be ignored, got %#v", prog)
	}
}

func TestParseOneLineProgress_IgnoreNonAggregateWithoutETAOrXfr(t *testing.T) {
	line := `2026/04/17 14:34:10 NOTICE: something happened`
	if prog, ok := parseOneLineProgress(line); ok || prog != nil {
		t.Fatalf("expected unrelated line to be ignored, got %#v", prog)
	}
}
