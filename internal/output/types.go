package output

import (
	"fmt"
	"strings"
)

type Format string

const (
	// FormatMarkdown writes one markdown document per crawled page.
	FormatMarkdown Format = "md"
	// FormatHTML writes HTML output per crawled page.
	FormatHTML Format = "html"
	// FormatJSON writes structured JSON per crawled page.
	FormatJSON Format = "json"
)

// ParseFormat validates and normalizes the output format flag.
func ParseFormat(raw string) (Format, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(FormatMarkdown):
		return FormatMarkdown, nil
	case string(FormatHTML):
		return FormatHTML, nil
	case string(FormatJSON):
		return FormatJSON, nil
	default:
		return "", fmt.Errorf("invalid format %q (allowed: md, html, json)", raw)
	}
}
