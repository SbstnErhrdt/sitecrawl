package crawler

import "testing"

func TestParseStrategy(t *testing.T) {
	tests := []struct {
		input   string
		want    Strategy
		wantErr bool
	}{
		{"pagerank", StrategyPageRank, false},
		{"LIMIT", StrategyLimit, false},
		{" depth ", StrategyDepth, false},
		{"random", "", true},
	}

	for _, tt := range tests {
		got, err := ParseStrategy(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Fatalf("expected error for input %q", tt.input)
			}
			continue
		}
		if err != nil {
			t.Fatalf("unexpected error for input %q: %v", tt.input, err)
		}
		if got != tt.want {
			t.Fatalf("expected %q, got %q for input %q", tt.want, got, tt.input)
		}
	}
}
