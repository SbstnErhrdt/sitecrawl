package output

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

type FilenameMapper struct {
	format Format
	used   map[string]string
}

// NewFilenameMapper creates a deterministic URL->filename mapper for one run.
func NewFilenameMapper(format Format) *FilenameMapper {
	return &FilenameMapper{
		format: format,
		used:   map[string]string{},
	}
}

// FilenameForURL maps a normalized URL to a stable file name and resolves collisions.
func (m *FilenameMapper) FilenameForURL(rawURL string) string {
	name := sanitizeURLPathToName(rawURL)
	ext := string(m.format)
	base := name + "." + ext
	if existing, ok := m.used[base]; !ok {
		m.used[base] = rawURL
		return base
	} else if existing == rawURL {
		return base
	}

	hash := shortHash(rawURL)
	candidate := fmt.Sprintf("%s_%s.%s", name, hash, ext)
	if existing, ok := m.used[candidate]; !ok || existing == rawURL {
		m.used[candidate] = rawURL
		return candidate
	}

	for i := 2; ; i++ {
		next := fmt.Sprintf("%s_%s_%d.%s", name, hash, i, ext)
		if existing, ok := m.used[next]; !ok || existing == rawURL {
			m.used[next] = rawURL
			return next
		}
	}
}

func sanitizeURLPathToName(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "index"
	}
	path := strings.TrimSpace(parsed.Path)
	if path == "" || path == "/" {
		return "index"
	}
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		return "index"
	}
	parts := strings.Split(path, "/")
	sanitized := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		if unescaped, err := url.PathUnescape(part); err == nil {
			part = unescaped
		}
		safe := sanitizeSegment(part)
		if safe != "" {
			sanitized = append(sanitized, safe)
		}
	}
	if len(sanitized) == 0 {
		return "index"
	}
	return strings.Join(sanitized, "_")
}

func sanitizeSegment(segment string) string {
	var builder strings.Builder
	lastUnderscore := false
	for _, r := range segment {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(unicode.ToLower(r))
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			builder.WriteRune('_')
			lastUnderscore = true
		}
	}
	result := strings.Trim(builder.String(), "_")
	return result
}

func shortHash(value string) string {
	sum := sha1.Sum([]byte(value))
	return hex.EncodeToString(sum[:4])
}
