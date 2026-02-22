package crawler

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"path"
	"sort"
	"strings"
)

var ErrUnsupportedLinkScheme = errors.New("unsupported link scheme")

// NormalizeURL normalizes URL scheme/host/path/query to a canonical crawl key.
func NormalizeURL(raw string, clean bool) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", err
	}
	return normalizeParsedURL(parsed, clean)
}

// ResolveAndNormalize resolves href against baseURL and returns a normalized URL.
func ResolveAndNormalize(baseURL, href string, clean bool) (string, error) {
	trimmedHref := strings.TrimSpace(href)
	if trimmedHref == "" {
		return "", errors.New("empty link")
	}
	lowerHref := strings.ToLower(trimmedHref)
	if strings.HasPrefix(lowerHref, "javascript:") ||
		strings.HasPrefix(lowerHref, "mailto:") ||
		strings.HasPrefix(lowerHref, "tel:") ||
		strings.HasPrefix(lowerHref, "data:") {
		return "", ErrUnsupportedLinkScheme
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	ref, err := url.Parse(trimmedHref)
	if err != nil {
		return "", err
	}
	resolved := base.ResolveReference(ref)
	return normalizeParsedURL(resolved, clean)
}

func normalizeParsedURL(parsed *url.URL, clean bool) (string, error) {
	if parsed == nil {
		return "", errors.New("nil url")
	}
	if parsed.Scheme == "" {
		return "", fmt.Errorf("missing scheme in url %q", parsed.String())
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", ErrUnsupportedLinkScheme
	}
	host, err := normalizeHost(parsed.Hostname())
	if err != nil {
		return "", err
	}
	port := parsed.Port()
	if (parsed.Scheme == "http" && port == "80") || (parsed.Scheme == "https" && port == "443") {
		port = ""
	}
	if port != "" {
		parsed.Host = net.JoinHostPort(host, port)
	} else {
		parsed.Host = host
	}
	parsed.Fragment = ""
	parsed.User = nil

	cleanPath := path.Clean(parsed.EscapedPath())
	if cleanPath == "." || cleanPath == "" {
		cleanPath = "/"
	}
	if !strings.HasPrefix(cleanPath, "/") {
		cleanPath = "/" + cleanPath
	}
	if cleanPath != "/" {
		cleanPath = strings.TrimSuffix(cleanPath, "/")
	}
	parsed.Path = cleanPath
	parsed.RawPath = ""

	queryValues := parsed.Query()
	if clean {
		for key := range queryValues {
			lowerKey := strings.ToLower(key)
			if strings.HasPrefix(lowerKey, "utm_") || lowerKey == "gclid" || lowerKey == "fbclid" {
				queryValues.Del(key)
			}
		}
	}
	if len(queryValues) == 0 {
		parsed.RawQuery = ""
	} else {
		parsed.RawQuery = encodeSortedQuery(queryValues)
	}

	return parsed.String(), nil
}

func encodeSortedQuery(values url.Values) string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		encodedKey := url.QueryEscape(key)
		vals := append([]string(nil), values[key]...)
		sort.Strings(vals)
		for _, value := range vals {
			parts = append(parts, encodedKey+"="+url.QueryEscape(value))
		}
		if len(vals) == 0 {
			parts = append(parts, encodedKey+"=")
		}
	}
	return strings.Join(parts, "&")
}
