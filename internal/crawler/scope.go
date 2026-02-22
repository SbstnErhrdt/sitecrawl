package crawler

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	"golang.org/x/net/idna"
)

var ErrInvalidDomain = errors.New("invalid domain")

const (
	// ScopeClassAllowed indicates a URL host that is inside the crawl scope.
	ScopeClassAllowed = "allowed"
	// ScopeClassExternal indicates a host outside the base domain namespace.
	ScopeClassExternal = "external"
	// ScopeClassOutOfScope indicates a subdomain of the base domain that is disallowed in V1.
	ScopeClassOutOfScope = "out_of_scope"
)

// Scope defines domain-limiting rules for a crawl run.
type Scope struct {
	BaseDomain   string
	AllowedHosts []string
	allowedSet   map[string]struct{}
}

// NewScope normalizes an input domain and builds the exact allowed host set:
// <domain> and www.<domain>.
func NewScope(input string) (Scope, error) {
	raw := strings.TrimSpace(input)
	if raw == "" {
		return Scope{}, fmt.Errorf("%w: domain is empty", ErrInvalidDomain)
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return Scope{}, fmt.Errorf("%w: %v", ErrInvalidDomain, err)
	}
	host := parsed.Hostname()
	if host == "" {
		return Scope{}, fmt.Errorf("%w: missing host", ErrInvalidDomain)
	}
	host, err = normalizeHost(host)
	if err != nil {
		return Scope{}, fmt.Errorf("%w: %v", ErrInvalidDomain, err)
	}
	base := host
	if strings.HasPrefix(base, "www.") {
		base = strings.TrimPrefix(base, "www.")
	}
	if base == "" {
		return Scope{}, fmt.Errorf("%w: empty base domain", ErrInvalidDomain)
	}
	allowed := []string{base, "www." + base}
	set := map[string]struct{}{}
	for _, h := range allowed {
		set[h] = struct{}{}
	}
	return Scope{
		BaseDomain:   base,
		AllowedHosts: allowed,
		allowedSet:   set,
	}, nil
}

// IsAllowedHost returns true only for exact allowed hosts.
func (s Scope) IsAllowedHost(host string) bool {
	normalized, err := normalizeHost(host)
	if err != nil {
		return false
	}
	_, ok := s.allowedSet[normalized]
	return ok
}

// IsAllowedURL returns true if the URL host is allowed by this scope.
func (s Scope) IsAllowedURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return s.IsAllowedHost(parsed.Hostname())
}

// ClassifyHost labels a host as allowed, out-of-scope subdomain, or external.
func (s Scope) ClassifyHost(host string) string {
	normalized, err := normalizeHost(host)
	if err != nil {
		return ScopeClassExternal
	}
	if s.IsAllowedHost(normalized) {
		return ScopeClassAllowed
	}
	if strings.HasSuffix(normalized, "."+s.BaseDomain) {
		return ScopeClassOutOfScope
	}
	return ScopeClassExternal
}

// normalizeHost canonicalizes a host using lowercase + IDNA lookup conversion.
func normalizeHost(host string) (string, error) {
	trimmed := strings.TrimSpace(strings.ToLower(host))
	trimmed = strings.TrimSuffix(trimmed, ".")
	if trimmed == "" {
		return "", errors.New("empty host")
	}
	if h, p, err := net.SplitHostPort(trimmed); err == nil {
		if p != "" {
			trimmed = h
		}
	}
	trimmed = strings.TrimPrefix(trimmed, "[")
	trimmed = strings.TrimSuffix(trimmed, "]")
	ascii, err := idna.Lookup.ToASCII(trimmed)
	if err != nil {
		return "", err
	}
	return strings.ToLower(strings.TrimSuffix(ascii, ".")), nil
}
