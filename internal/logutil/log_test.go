package logutil

import "testing"

func TestSplitLogSegments_Empty(t *testing.T) {
	segs := SplitLogSegments("")
	if segs != nil {
		t.Errorf("expected nil, got %v", segs)
	}
	segs = SplitLogSegments("   ")
	if segs != nil {
		t.Errorf("expected nil, got %v", segs)
	}
}

func TestSplitLogSegments_Single(t *testing.T) {
	line := "2025/01/01 12:00:00 INFO : path: message"
	segs := SplitLogSegments(line)
	if len(segs) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segs))
	}
	if segs[0] != line {
		t.Errorf("expected %q, got %q", line, segs[0])
	}
}

func TestSplitLogSegments_Multiple(t *testing.T) {
	line := "2025/01/01 12:00:00 INFO : path1: msg1\n2025/01/01 12:01:00 ERROR: path2: msg2"
	segs := SplitLogSegments(line)
	if len(segs) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(segs))
	}
}

func TestSplitLogSegments_NoTimestamp(t *testing.T) {
	line := "some random log line without timestamp"
	segs := SplitLogSegments(line)
	if len(segs) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segs))
	}
	if segs[0] != line {
		t.Errorf("expected %q, got %q", line, segs[0])
	}
}

func TestParseLogSegment_Standard(t *testing.T) {
	at, level, path, msg, ok := ParseLogSegment("2025/01/01 12:00:00 INFO : /path/to/file: test message")
	if !ok {
		t.Fatal("expected ok=true")
	}
	if at != "2025/01/01 12:00:00" {
		t.Errorf("at = %q", at)
	}
	if level != "INFO" {
		t.Errorf("level = %q", level)
	}
	if path != "/path/to/file" {
		t.Errorf("path = %q", path)
	}
	if msg != "test message" {
		t.Errorf("msg = %q", msg)
	}
}

func TestParseLogSegment_WithoutTimestamp(t *testing.T) {
	_, level, path, msg, ok := ParseLogSegment("INFO : /path: simple message")
	if !ok {
		t.Fatal("expected ok=true")
	}
	if level != "INFO" {
		t.Errorf("level = %q", level)
	}
	if path != "/path" {
		t.Errorf("path = %q", path)
	}
	if msg != "simple message" {
		t.Errorf("msg = %q", msg)
	}
}

func TestParseLogSegment_JSON(t *testing.T) {
	at, level, path, msg, ok := ParseLogSegment(`{"time":"2025-01-01T12:00:00Z","level":"error","object":"/path/to/file","msg":"test message"}`)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if level != "ERROR" {
		t.Errorf("level = %q", level)
	}
	if msg != "test message" {
		t.Errorf("msg = %q", msg)
	}
	if path != "/path/to/file" {
		t.Errorf("path = %q", path)
	}
	if at != "2025-01-01T12:00:00Z" {
		t.Errorf("at = %q", at)
	}
}

func TestParseLogSegment_JSON_TimestampField(t *testing.T) {
	at, _, _, msg, ok := ParseLogSegment(`{"timestamp":"2025-01-01T12:00:00Z","level":"info","msg":"test"}`)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if msg != "test" {
		t.Errorf("msg = %q", msg)
	}
	if at != "2025-01-01T12:00:00Z" {
		t.Errorf("at = %q", at)
	}
}

func TestParseLogSegment_Invalid(t *testing.T) {
	_, _, _, _, ok := ParseLogSegment("completely invalid")
	if ok {
		t.Error("expected ok=false")
	}
}

func TestParseLogSegment_JSON_NoMsg(t *testing.T) {
	_, _, _, _, ok := ParseLogSegment(`{"level":"info","time":"now"}`)
	if ok {
		t.Error("expected ok=false for JSON without msg field")
	}
}