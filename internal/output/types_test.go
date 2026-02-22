package output

import "testing"

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input   string
		want    Format
		wantErr bool
	}{
		{"md", FormatMarkdown, false},
		{"HTML", FormatHTML, false},
		{" json ", FormatJSON, false},
		{"pdf", "", true},
	}

	for _, tt := range tests {
		got, err := ParseFormat(tt.input)
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
