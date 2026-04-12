package scheduler

import (
	"testing"
	"time"
)

func TestParseSpec(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		wantOK  bool
		wantDur time.Duration
	}{
		{
			name:    "valid 5m",
			spec:    "@every 5m",
			wantOK:  true,
			wantDur: 5 * time.Minute,
		},
		{
			name:    "valid 1h",
			spec:    "@every 1h",
			wantOK:  true,
			wantDur: 1 * time.Hour,
		},
		{
			name:    "valid 30s",
			spec:    "30s",
			wantOK:  true,
			wantDur: 30 * time.Second,
		},
		{
			name:    "valid 2h30m",
			spec:    "2h30m",
			wantOK:  true,
			wantDur: 2*time.Hour + 30*time.Minute,
		},
		{
			name:    "invalid empty",
			spec:    "",
			wantOK:  false,
			wantDur: 0,
		},
		{
			name:    "invalid negative",
			spec:    "-5m",
			wantOK:  false,
			wantDur: 0,
		},
		{
			name:    "invalid zero",
			spec:    "0m",
			wantOK:  false,
			wantDur: 0,
		},
		{
			name:    "invalid random",
			spec:    "abc",
			wantOK:  false,
			wantDur: 0,
		},
		{
			name:    "with spaces",
			spec:    "  @every 10m  ",
			wantOK:  true,
			wantDur: 10 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dur, ok := parseSpec(tt.spec)
			if ok != tt.wantOK {
				t.Errorf("parseSpec(%q) ok = %v, want %v", tt.spec, ok, tt.wantOK)
			}
			if ok && dur != tt.wantDur {
				t.Errorf("parseSpec(%q) = %v, want %v", tt.spec, dur, tt.wantDur)
			}
		})
	}
}
