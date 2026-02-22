package crawler

import "testing"

func TestScopeAllowsOnlyDomainAndWWW(t *testing.T) {
	scope, err := NewScope("www.Example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if scope.BaseDomain != "example.com" {
		t.Fatalf("expected base domain example.com, got %s", scope.BaseDomain)
	}

	allowedHosts := []string{
		"example.com",
		"www.example.com",
		"EXAMPLE.COM",
		"www.example.com:443",
	}
	for _, host := range allowedHosts {
		if !scope.IsAllowedHost(host) {
			t.Fatalf("expected host to be allowed: %s", host)
		}
	}

	disallowedHosts := []string{
		"blog.example.com",
		"api.example.com",
		"example.org",
		"www2.example.com",
	}
	for _, host := range disallowedHosts {
		if scope.IsAllowedHost(host) {
			t.Fatalf("expected host to be disallowed: %s", host)
		}
	}
}

func TestClassifyHost(t *testing.T) {
	scope, err := NewScope("example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := scope.ClassifyHost("www.example.com"); got != ScopeClassAllowed {
		t.Fatalf("expected allowed, got %s", got)
	}
	if got := scope.ClassifyHost("blog.example.com"); got != ScopeClassOutOfScope {
		t.Fatalf("expected out_of_scope, got %s", got)
	}
	if got := scope.ClassifyHost("example.org"); got != ScopeClassExternal {
		t.Fatalf("expected external, got %s", got)
	}
}
