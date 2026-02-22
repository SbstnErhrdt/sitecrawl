package pagerank

import "testing"

func TestNodeIDString(t *testing.T) {
	id := NodeID("https://example.com")
	if got := id.String(); got != "https://example.com" {
		t.Fatalf("expected string value, got %q", got)
	}
}
