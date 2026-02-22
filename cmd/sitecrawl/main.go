package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/sbstn/sitecrawl/internal/crawler"
	"github.com/sbstn/sitecrawl/internal/output"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		printRootUsage(os.Stderr)
		return 2
	}

	switch args[0] {
	case "crawl":
		return runCrawl(args[1:])
	case "-h", "--help", "help":
		printRootUsage(os.Stdout)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", args[0])
		printRootUsage(os.Stderr)
		return 2
	}
}

func runCrawl(args []string) int {
	flagSet := flag.NewFlagSet("crawl", flag.ContinueOnError)
	flagSet.SetOutput(os.Stderr)

	var domain string
	var outDir string
	var formatRaw string
	var strategyRaw string
	var maxPages int
	var maxDepth int
	var clean bool
	var headful bool
	var delayMS int
	var pageTimeout time.Duration
	var userAgent string
	var logLevelRaw string

	flagSet.StringVar(&domain, "domain", "", "Domain to crawl (required). Example: example.com")
	flagSet.StringVar(&outDir, "out", "", "Output directory (required)")
	flagSet.StringVar(&formatRaw, "format", "", "Output format (required): md|html|json")
	flagSet.StringVar(&strategyRaw, "strategy", "pagerank", "Crawl strategy: pagerank|limit|depth")
	flagSet.IntVar(&maxPages, "max-pages", 25, "Max pages to crawl (hard cap for all strategies)")
	flagSet.IntVar(&maxDepth, "max-depth", 2, "Max crawl depth for strategy=depth")
	flagSet.BoolVar(&clean, "clean", true, "Strip scripts/styles and output readable/main content")
	flagSet.BoolVar(&headful, "headful", false, "Run Chrome in headful mode")
	flagSet.IntVar(&delayMS, "delay-ms", 750, "Politeness delay between navigations in milliseconds")
	flagSet.DurationVar(&pageTimeout, "page-timeout", 20*time.Second, "Per-page timeout, e.g. 20s")
	flagSet.StringVar(&userAgent, "user-agent", crawler.DefaultUserAgent, "User-Agent string")
	flagSet.StringVar(&logLevelRaw, "log", "info", "Log level: debug|info|warn|error")

	flagSet.Usage = func() {
		fmt.Fprintf(flagSet.Output(), "Usage:\n")
		fmt.Fprintf(flagSet.Output(), "  sitecrawl crawl --domain <domain> --out <dir> --format <md|html|json> [flags]\n\n")
		flagSet.PrintDefaults()
	}

	if err := flagSet.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}

	if domain == "" || outDir == "" || formatRaw == "" {
		fmt.Fprintln(os.Stderr, "error: --domain, --out, and --format are required")
		fmt.Fprintln(os.Stderr)
		flagSet.Usage()
		return 2
	}
	if maxPages <= 0 {
		fmt.Fprintln(os.Stderr, "error: --max-pages must be >= 1")
		return 2
	}
	if maxDepth < 0 {
		fmt.Fprintln(os.Stderr, "error: --max-depth must be >= 0")
		return 2
	}
	if delayMS < 0 {
		fmt.Fprintln(os.Stderr, "error: --delay-ms must be >= 0")
		return 2
	}

	format, err := output.ParseFormat(formatRaw)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 2
	}
	strategy, err := crawler.ParseStrategy(strategyRaw)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 2
	}

	logger := newLogger(logLevelRaw)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg := crawler.Config{
		Domain:      domain,
		Strategy:    strategy,
		MaxPages:    maxPages,
		MaxDepth:    maxDepth,
		Clean:       clean,
		Headful:     headful,
		Delay:       time.Duration(delayMS) * time.Millisecond,
		PageTimeout: pageTimeout,
		UserAgent:   userAgent,
	}

	result, crawlErr := crawler.Crawl(ctx, cfg, logger)
	if result != nil {
		if writeErr := output.Write(result, outDir, format); writeErr != nil {
			logger.Error("failed to write output", "error", writeErr)
			return 1
		}
		logger.Info("report written", "path", filepath.Join(outDir, "report.json"))
	}

	if crawlErr != nil {
		if errors.Is(crawlErr, context.Canceled) {
			logger.Warn("crawl interrupted, partial output written")
			return 130
		}
		logger.Error("crawl failed", "error", crawlErr)
		return 1
	}

	if result != nil {
		logger.Info("crawl finished",
			"visited", result.Totals.Visited,
			"errors", result.Totals.Errors,
			"skipped_external", result.Totals.SkippedExternal,
			"skipped_out_of_scope", result.Totals.SkippedOutOfScope,
		)
	}
	return 0
}

func printRootUsage(out *os.File) {
	fmt.Fprintln(out, "sitecrawl")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Commands:")
	fmt.Fprintln(out, "  crawl   Crawl a domain and write page outputs + report.json")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Examples:")
	fmt.Fprintln(out, "  sitecrawl crawl --domain example.com --format md --out ./out")
	fmt.Fprintln(out, "  sitecrawl crawl --domain example.com --strategy limit --max-pages 50 --format md --out ./out")
	fmt.Fprintln(out, "  sitecrawl crawl --domain example.com --strategy depth --max-depth 2 --format json --out ./out --headful")
}

func newLogger(levelRaw string) *slog.Logger {
	var level slog.Level
	switch strings.ToLower(strings.TrimSpace(levelRaw)) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})
	return slog.New(handler)
}
