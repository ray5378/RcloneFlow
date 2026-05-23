package runnercli

import (
	"testing"
)

func TestDefaultCASFileExists_EmptyRel(t *testing.T) {
	ok, err := defaultCASFileExists("", "dst:/path", "")
	if ok || err != nil {
		t.Fatalf("expected (false, nil) for empty rel, got (%v, %v)", ok, err)
	}
}

func TestDefaultCASFileExists_WhitespaceRel(t *testing.T) {
	ok, err := defaultCASFileExists("", "dst:/path", "  ")
	if ok || err != nil {
		t.Fatalf("expected (false, nil) for whitespace rel, got (%v, %v)", ok, err)
	}
}

func TestDefaultCASFileExists_JustCasExt(t *testing.T) {
	ok, err := defaultCASFileExists("", "dst:/path", ".cas")
	if ok || err != nil {
		t.Fatalf("expected (false, nil) for .cas rel, got (%v, %v)", ok, err)
	}
}
