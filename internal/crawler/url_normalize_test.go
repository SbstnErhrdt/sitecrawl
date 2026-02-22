package crawler

import "testing"

func TestNormalizeURLRemovesFragmentsAndTrackingParamsWhenClean(t *testing.T) {
	got, err := NormalizeURL("https://Example.com/docs/?utm_source=x&b=2&gclid=abc#section", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "https://example.com/docs?b=2"
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestNormalizeURLKeepsTrackingParamsWhenNotClean(t *testing.T) {
	got, err := NormalizeURL("https://Example.com/docs/?utm_source=x&b=2#section", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "https://example.com/docs?b=2&utm_source=x"
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestResolveAndNormalize(t *testing.T) {
	got, err := ResolveAndNormalize("https://example.com/docs/", "../a/?fbclid=123", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "https://example.com/a"
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}
