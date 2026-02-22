package crawler

import (
	"fmt"
	"strings"
	"time"
)

// DefaultUserAgent is the user-agent value used when callers do not provide one.
const DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"

// PageRankImplementation identifies the rank engine used in report metadata.
const PageRankImplementation = "pkg/pagerank.NewPageRank + (*PageRank).CalcPageRank"

// Strategy controls URL selection behavior during crawling.
type Strategy string

const (
	// StrategyPageRank crawls up to the hard cap and ranks pages by PageRank.
	StrategyPageRank Strategy = "pagerank"
	// StrategyLimit uses breadth-first order until MaxPages is reached.
	StrategyLimit Strategy = "limit"
	// StrategyDepth uses breadth-first order constrained by MaxDepth and MaxPages.
	StrategyDepth Strategy = "depth"
)

// ParseStrategy validates and normalizes a strategy flag value.
func ParseStrategy(raw string) (Strategy, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(StrategyPageRank):
		return StrategyPageRank, nil
	case string(StrategyLimit):
		return StrategyLimit, nil
	case string(StrategyDepth):
		return StrategyDepth, nil
	default:
		return "", fmt.Errorf("invalid strategy %q (allowed: pagerank, limit, depth)", raw)
	}
}

// Config contains all crawl runtime configuration.
type Config struct {
	Domain      string
	Strategy    Strategy
	MaxPages    int
	MaxDepth    int
	Clean       bool
	Headful     bool
	Delay       time.Duration
	PageTimeout time.Duration
	UserAgent   string
}

// Totals tracks crawl counters for report generation.
type Totals struct {
	Visited           int
	Errors            int
	SkippedExternal   int
	SkippedOutOfScope int
}

// Page stores extracted and output metadata for a crawled URL.
type Page struct {
	URL         string
	FinalURL    string
	Depth       int
	Status      string
	Title       string
	Description string
	Links       []string
	MainText    string
	MainHTML    string
	BodyHTML    string
	RawHTML     string
	Error       string
	OutPath     string
	Score       *float64
}

// CrawlResult is the complete crawl output before serialization.
type CrawlResult struct {
	Domain                 string
	AllowedHosts           []string
	StartedAt              time.Time
	FinishedAt             time.Time
	Strategy               Strategy
	MaxPages               int
	MaxDepth               int
	Clean                  bool
	Headful                bool
	PageRankImplementation string
	Pages                  []*Page
	Totals                 Totals
}

const (
	// StatusOK indicates a successfully crawled page.
	StatusOK = "ok"
	// StatusError indicates a crawl failure.
	StatusError = "error"
	// StatusSkippedRobots indicates robots.txt denied this URL.
	StatusSkippedRobots = "skipped_robots"
	// StatusSkippedOutOfHost indicates redirection or resolution escaped scope.
	StatusSkippedOutOfHost = "skipped_out_of_scope"
)
