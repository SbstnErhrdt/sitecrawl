package output

import "testing"

func TestFilenameMapping(t *testing.T) {
	mapper := NewFilenameMapper(FormatMarkdown)

	cases := map[string]string{
		"https://example.com/":       "index.md",
		"https://example.com/docs":   "docs.md",
		"https://example.com/docs/a": "docs_a.md",
	}

	for input, expected := range cases {
		got := mapper.FilenameForURL(input)
		if got != expected {
			t.Fatalf("expected %s for %s, got %s", expected, input, got)
		}
	}
}

func TestFilenameMappingCollisionUsesStableHash(t *testing.T) {
	mapper := NewFilenameMapper(FormatJSON)

	first := mapper.FilenameForURL("https://example.com/docs")
	second := mapper.FilenameForURL("https://example.com/docs?x=1")

	if first != "docs.json" {
		t.Fatalf("unexpected first filename: %s", first)
	}
	if second == first {
		t.Fatalf("expected collision resolution for second filename")
	}
	if len(second) <= len("docs_.json") || second[:5] != "docs_" {
		t.Fatalf("expected hashed suffix in %s", second)
	}
}
