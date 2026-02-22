package crawler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"sort"
	"time"
)

type queueItem struct {
	URL   string
	Depth int
}

// Crawl executes one crawl run using the configured strategy and scope rules.
func Crawl(ctx context.Context, cfg Config, logger *slog.Logger) (*CrawlResult, error) {
	if logger == nil {
		logger = slog.Default()
	}
	scope, err := NewScope(cfg.Domain)
	if err != nil {
		return nil, err
	}
	if cfg.MaxPages <= 0 {
		cfg.MaxPages = 25
	}
	if cfg.MaxDepth < 0 {
		cfg.MaxDepth = 0
	}
	if cfg.PageTimeout <= 0 {
		cfg.PageTimeout = 20 * time.Second
	}
	if cfg.Delay < 0 {
		cfg.Delay = 0
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = DefaultUserAgent
	}

	result := &CrawlResult{
		Domain:       scope.BaseDomain,
		AllowedHosts: append([]string(nil), scope.AllowedHosts...),
		StartedAt:    time.Now().UTC(),
		Strategy:     cfg.Strategy,
		MaxPages:     cfg.MaxPages,
		MaxDepth:     cfg.MaxDepth,
		Clean:        cfg.Clean,
		Headful:      cfg.Headful,
		Pages:        []*Page{},
	}

	browserCtx, cleanup := newBrowserContext(ctx, cfg)
	defer cleanup()

	startHTTPS := fmt.Sprintf("https://%s/", scope.BaseDomain)
	startHTTP := fmt.Sprintf("http://%s/", scope.BaseDomain)
	startURL := startHTTPS
	startFetch, err := fetchPageWithRetry(ctx, browserCtx, startHTTPS, cfg.Clean, cfg.PageTimeout, 1, logger)
	if err != nil && shouldFallbackToHTTP(err) {
		logger.Warn("https start failed, trying http", "url", startHTTPS, "error", err)
		startFetch, err = fetchPageWithRetry(ctx, browserCtx, startHTTP, cfg.Clean, cfg.PageTimeout, 1, logger)
		if err != nil {
			result.FinishedAt = time.Now().UTC()
			return result, err
		}
		startURL = startHTTP
	}
	if err != nil {
		result.FinishedAt = time.Now().UTC()
		return result, err
	}
	logger.Info("using start URL", "url", startURL)

	seedPages := map[string]fetchedPage{
		startURL: startFetch,
	}
	robots := newRobotsCache(scope, cfg.UserAgent, logger)
	graph := NewLinkGraph()

	queue := []queueItem{{URL: startURL, Depth: 0}}
	enqueued := map[string]struct{}{startURL: {}}
	processed := map[string]struct{}{}
	firstNavigation := true

	for len(queue) > 0 && result.Totals.Visited < cfg.MaxPages {
		if ctx.Err() != nil {
			break
		}
		current := queue[0]
		queue = queue[1:]

		if _, seen := processed[current.URL]; seen {
			continue
		}
		processed[current.URL] = struct{}{}

		if cfg.Strategy == StrategyDepth && current.Depth > cfg.MaxDepth {
			continue
		}

		allowedByRobots, robotsErr := robots.Allowed(current.URL)
		if robotsErr != nil {
			logger.Warn("robots check failed, allowing crawl", "url", current.URL, "error", robotsErr)
		}
		if !allowedByRobots {
			result.Pages = append(result.Pages, &Page{
				URL:      current.URL,
				FinalURL: current.URL,
				Depth:    current.Depth,
				Status:   StatusSkippedRobots,
				Error:    "disallowed by robots.txt",
			})
			continue
		}

		var fetched fetchedPage
		if seed, ok := seedPages[current.URL]; ok {
			fetched = seed
			delete(seedPages, current.URL)
		} else {
			if !firstNavigation && !sleepWithJitter(ctx, cfg.Delay) {
				break
			}
			firstNavigation = false

			var fetchErr error
			fetched, fetchErr = fetchPageWithRetry(ctx, browserCtx, current.URL, cfg.Clean, cfg.PageTimeout, 1, logger)
			if fetchErr != nil {
				result.Totals.Errors++
				result.Totals.Visited++
				result.Pages = append(result.Pages, &Page{
					URL:      current.URL,
					FinalURL: current.URL,
					Depth:    current.Depth,
					Status:   StatusError,
					Error:    fetchErr.Error(),
				})
				continue
			}
		}
		if firstNavigation {
			firstNavigation = false
		}

		normalizedFinal, normalizeErr := NormalizeURL(fetched.FinalURL, cfg.Clean)
		if normalizeErr != nil {
			normalizedFinal = current.URL
		}
		if !scope.IsAllowedURL(normalizedFinal) {
			result.Totals.SkippedOutOfScope++
			result.Pages = append(result.Pages, &Page{
				URL:      current.URL,
				FinalURL: normalizedFinal,
				Depth:    current.Depth,
				Status:   StatusSkippedOutOfHost,
				Error:    "redirected out of allowed host scope",
			})
			continue
		}

		internalLinks := make([]string, 0, len(fetched.Links))
		linkSet := map[string]struct{}{}
		for _, href := range fetched.Links {
			normalizedLink, linkErr := ResolveAndNormalize(normalizedFinal, href, cfg.Clean)
			if linkErr != nil {
				continue
			}
			parsedLink, parseErr := url.Parse(normalizedLink)
			if parseErr != nil {
				continue
			}
			switch scope.ClassifyHost(parsedLink.Hostname()) {
			case ScopeClassAllowed:
				if _, exists := linkSet[normalizedLink]; exists {
					continue
				}
				linkSet[normalizedLink] = struct{}{}
				internalLinks = append(internalLinks, normalizedLink)
			case ScopeClassOutOfScope:
				result.Totals.SkippedOutOfScope++
			default:
				result.Totals.SkippedExternal++
			}
		}
		sort.Strings(internalLinks)

		graph.AddNode(normalizedFinal)
		for _, link := range internalLinks {
			graph.AddEdge(normalizedFinal, link)
		}

		for _, nextURL := range internalLinks {
			if _, exists := enqueued[nextURL]; exists {
				continue
			}
			nextDepth := current.Depth + 1
			if cfg.Strategy == StrategyDepth && nextDepth > cfg.MaxDepth {
				continue
			}
			enqueued[nextURL] = struct{}{}
			queue = append(queue, queueItem{
				URL:   nextURL,
				Depth: nextDepth,
			})
		}

		result.Pages = append(result.Pages, &Page{
			URL:         current.URL,
			FinalURL:    normalizedFinal,
			Depth:       current.Depth,
			Status:      StatusOK,
			Title:       fetched.Title,
			Description: fetched.Description,
			Links:       internalLinks,
			MainText:    fetched.MainText,
			MainHTML:    fetched.MainHTML,
			BodyHTML:    fetched.BodyHTML,
			RawHTML:     fetched.RawHTML,
		})
		result.Totals.Visited++
	}

	if cfg.Strategy == StrategyPageRank {
		ApplyPageRankScores(result, graph)
	}

	result.FinishedAt = time.Now().UTC()
	if ctx.Err() != nil && !errors.Is(ctx.Err(), context.Canceled) {
		return result, ctx.Err()
	}
	return result, ctx.Err()
}
