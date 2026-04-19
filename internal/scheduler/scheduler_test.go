package scheduler

import "testing"

func TestParseSpecToCron(t *testing.T) {
	tests := []struct {
		name     string
		spec     string
		wantOK   bool
		wantCron string
	}{
		{
			name:     "valid wildcard spec",
			spec:     "*|*|*|*|*",
			wantOK:   true,
			wantCron: "0 * * * * *",
		},
		{
			name:     "valid minute hour spec",
			spec:     "04,03,06|17,19|*|*|*",
			wantOK:   true,
			wantCron: "0 04,03,06 17,19 * * *",
		},
		{
			name:   "invalid empty",
			spec:   "",
			wantOK: false,
		},
		{
			name:   "invalid parts length",
			spec:   "*|*|*",
			wantOK: false,
		},
		{
			name:   "invalid minute",
			spec:   "61|*|*|*|*",
			wantOK: false,
		},
		{
			name:   "invalid hour",
			spec:   "*|24|*|*|*",
			wantOK: false,
		},
		{
			name:   "invalid day",
			spec:   "*|*|0|*|*",
			wantOK: false,
		},
		{
			name:   "invalid month",
			spec:   "*|*|*|13|*",
			wantOK: false,
		},
		{
			name:   "invalid week",
			spec:   "*|*|*|*|7",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cronExpr, ok := ParseSpecToCron(tt.spec)
			if ok != tt.wantOK {
				t.Errorf("ParseSpecToCron(%q) ok = %v, want %v", tt.spec, ok, tt.wantOK)
			}
			if ok && cronExpr != tt.wantCron {
				t.Errorf("ParseSpecToCron(%q) = %q, want %q", tt.spec, cronExpr, tt.wantCron)
			}
		})
	}
}
