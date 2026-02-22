package crawler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

type fetchedPage struct {
	FinalURL    string
	Title       string
	Description string
	Links       []string
	BodyHTML    string
	MainHTML    string
	MainText    string
	RawHTML     string
}

func newBrowserContext(parent context.Context, cfg Config) (context.Context, func()) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", !cfg.Headful),
		chromedp.UserAgent(cfg.UserAgent),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
	)
	allocatorCtx, cancelAllocator := chromedp.NewExecAllocator(parent, opts...)
	browserCtx, cancelBrowser := chromedp.NewContext(allocatorCtx)
	cleanup := func() {
		cancelBrowser()
		cancelAllocator()
	}
	return browserCtx, cleanup
}

func fetchPageWithRetry(
	ctx context.Context,
	browserCtx context.Context,
	targetURL string,
	clean bool,
	pageTimeout time.Duration,
	retries int,
	logger *slog.Logger,
) (fetchedPage, error) {
	var lastErr error
	attempts := retries + 1
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(attempt) * 500 * time.Millisecond
			if !sleepWithContext(ctx, backoff) {
				return fetchedPage{}, ctx.Err()
			}
			logger.Warn("retrying page navigation", "url", targetURL, "attempt", attempt+1)
		}
		page, err := fetchPageOnce(ctx, browserCtx, targetURL, clean, pageTimeout)
		if err == nil {
			return page, nil
		}
		lastErr = err
		if !isTransientNavigationError(err) || attempt == attempts-1 {
			return fetchedPage{}, err
		}
	}
	if lastErr == nil {
		lastErr = errors.New("navigation failed")
	}
	return fetchedPage{}, lastErr
}

func fetchPageOnce(
	ctx context.Context,
	browserCtx context.Context,
	targetURL string,
	clean bool,
	pageTimeout time.Duration,
) (fetchedPage, error) {
	// Use a fresh target context for each fetch. Some sites close or poison the
	// current target after navigation, which can cascade "context canceled"
	// errors for all subsequent URLs when reusing a single tab.
	tabCtx, cancelTab := chromedp.NewContext(browserCtx)
	defer cancelTab()

	pageCtx, cancel := context.WithTimeout(tabCtx, pageTimeout)
	defer cancel()

	var html string
	var finalURL string
	var extracted struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Links       []string `json:"links"`
		BodyHTML    string   `json:"bodyHTML"`
		MainHTML    string   `json:"mainHTML"`
		MainText    string   `json:"mainText"`
	}

	err := chromedp.Run(pageCtx,
		chromedp.Navigate(targetURL),
		chromedp.WaitReady("html", chromedp.ByQuery),
		chromedp.ActionFunc(waitForReadyStateComplete),
		chromedp.Sleep(350*time.Millisecond),
		chromedp.OuterHTML("html", &html, chromedp.ByQuery),
		chromedp.Evaluate(extractionScript(clean), &extracted),
		chromedp.Location(&finalURL),
	)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return fetchedPage{}, err
		}
		return fetchedPage{}, fmt.Errorf("navigate %s: %w", targetURL, err)
	}
	if finalURL == "" {
		finalURL = targetURL
	}
	return fetchedPage{
		FinalURL:    finalURL,
		Title:       strings.TrimSpace(extracted.Title),
		Description: strings.TrimSpace(extracted.Description),
		Links:       extracted.Links,
		BodyHTML:    extracted.BodyHTML,
		MainHTML:    extracted.MainHTML,
		MainText:    strings.TrimSpace(extracted.MainText),
		RawHTML:     html,
	}, nil
}

func extractionScript(clean bool) string {
	cleanJS := strconv.FormatBool(clean)
	return `(function () {
		const clean = ` + cleanJS + `;
		const sourceRoot = document.documentElement;
		const cloneRoot = sourceRoot.cloneNode(true);
		if (clean) {
			cloneRoot.querySelectorAll('script,style,noscript').forEach(el => el.remove());
			cloneRoot.querySelectorAll('[style]').forEach(el => el.removeAttribute('style'));
		}
		const body = cloneRoot.querySelector('body') || cloneRoot;
		let main = cloneRoot.querySelector('main') ||
			cloneRoot.querySelector('article') ||
			cloneRoot.querySelector('[role="main"]') ||
			body;

		const normalize = (v) => (v || '').replace(/\s+/g, ' ').trim();
		let mainText = '';
		if (clean) {
			const blocks = Array.from(main.querySelectorAll('h1,h2,h3,h4,h5,h6,p,li,blockquote,pre'))
				.map(el => normalize(el.textContent))
				.filter(Boolean);
			mainText = blocks.length > 0 ? blocks.join('\n\n') : normalize(main.textContent);
		} else {
			main = body;
			mainText = normalize(body.textContent);
		}

		return {
			title: normalize((document.querySelector('title') || {}).textContent || ''),
			description: normalize((document.querySelector('meta[name="description"]') || {}).content || ''),
			links: Array.from(document.querySelectorAll('a[href]'))
				.map(a => normalize(a.getAttribute('href')))
				.filter(Boolean),
			bodyHTML: body.innerHTML || '',
			mainHTML: main.innerHTML || '',
			mainText: mainText
		};
	})()`
}

func waitForReadyStateComplete(ctx context.Context) error {
	const maxWait = 5 * time.Second
	deadline := time.Now().Add(maxWait)
	for {
		var state string
		if err := chromedp.Evaluate(`document.readyState`, &state).Do(ctx); err != nil {
			return err
		}
		if state == "complete" {
			return nil
		}
		if time.Now().After(deadline) {
			return nil
		}
		if !sleepWithContext(ctx, 100*time.Millisecond) {
			return ctx.Err()
		}
	}
}

func sleepWithJitter(ctx context.Context, delay time.Duration) bool {
	if delay <= 0 {
		return true
	}
	jitterBound := delay / 4
	jitter := time.Duration(0)
	if jitterBound > 0 {
		jitter = time.Duration(rand.Int63n(int64(jitterBound)))
	}
	return sleepWithContext(ctx, delay+jitter)
}

func sleepWithContext(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func isTransientNavigationError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}
	lowerErr := strings.ToLower(err.Error())
	needles := []string{
		"timeout",
		"net::err_",
		"connection reset",
		"connection refused",
		"temporary",
		"no such host",
	}
	for _, needle := range needles {
		if strings.Contains(lowerErr, needle) {
			return true
		}
	}
	return false
}

func shouldFallbackToHTTP(err error) bool {
	if err == nil {
		return false
	}
	lowerErr := strings.ToLower(err.Error())
	needles := []string{
		"net::err_ssl",
		"net::err_cert",
		"tls",
		"x509",
		"unsupported protocol scheme",
		"net::err_connection_refused",
		"net::err_empty_response",
	}
	for _, needle := range needles {
		if strings.Contains(lowerErr, needle) {
			return true
		}
	}
	return false
}
