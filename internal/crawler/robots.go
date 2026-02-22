package crawler

import (
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/temoto/robotstxt"
)

type robotsEntry struct {
	loaded bool
	group  *robotstxt.Group
}

type robotsCache struct {
	scope     Scope
	userAgent string
	logger    *slog.Logger
	client    *http.Client
	mu        sync.Mutex
	entries   map[string]*robotsEntry
}

func newRobotsCache(scope Scope, userAgent string, logger *slog.Logger) *robotsCache {
	return &robotsCache{
		scope:     scope,
		userAgent: userAgent,
		logger:    logger,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		entries: map[string]*robotsEntry{},
	}
}

func (rc *robotsCache) Allowed(rawURL string) (bool, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false, err
	}
	host := parsed.Hostname()
	if !rc.scope.IsAllowedHost(host) {
		return false, nil
	}

	entry, err := rc.getOrLoad(host)
	if err != nil {
		return true, err
	}
	if entry == nil || entry.group == nil {
		return true, nil
	}
	path := parsed.EscapedPath()
	if path == "" {
		path = "/"
	}
	if parsed.RawQuery != "" {
		path += "?" + parsed.RawQuery
	}
	return entry.group.Test(path), nil
}

func (rc *robotsCache) getOrLoad(host string) (*robotsEntry, error) {
	normalizedHost, err := normalizeHost(host)
	if err != nil {
		return nil, err
	}

	rc.mu.Lock()
	entry, ok := rc.entries[normalizedHost]
	if ok && entry.loaded {
		rc.mu.Unlock()
		return entry, nil
	}
	rc.mu.Unlock()

	group, loadErr := rc.loadGroup(normalizedHost)
	if loadErr != nil {
		rc.logger.Warn("robots.txt fetch failed; allowing crawl for host", "host", normalizedHost, "error", loadErr)
	}

	rc.mu.Lock()
	defer rc.mu.Unlock()
	updated := &robotsEntry{
		loaded: true,
		group:  group,
	}
	rc.entries[normalizedHost] = updated
	return updated, loadErr
}

func (rc *robotsCache) loadGroup(host string) (*robotstxt.Group, error) {
	var lastErr error
	for _, scheme := range []string{"https", "http"} {
		robotsURL := scheme + "://" + host + "/robots.txt"
		req, err := http.NewRequest(http.MethodGet, robotsURL, nil)
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("User-Agent", rc.userAgent)
		resp, err := rc.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		data, err := robotstxt.FromResponse(resp)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}
		return data.FindGroup(rc.userAgent), nil
	}
	if lastErr == nil {
		lastErr = io.EOF
	}
	return nil, lastErr
}
