# Architecture

## High-Level Flow

1. CLI parses flags and validates inputs.
2. Crawler initializes scope, robots cache, and Chrome browser context.
3. Start URL is selected (`https://` first, `http://` fallback).
4. URLs are crawled with strategy constraints (`pagerank`, `limit`, `depth`).
5. Each visited page is normalized, extracted, and linked into the graph.
6. When strategy is `pagerank`, scores are computed using `pkg/pagerank`.
7. Output writer persists page files and `report.json`.

## Package Layout

- `cmd/sitecrawl`: command-line interface
- `internal/crawler`:
  - scope and host gating
  - URL normalization
  - robots checks
  - chromedp navigation and extraction
  - strategy execution and PageRank adaptation
- `internal/output`:
  - deterministic file naming
  - format rendering
  - report generation
- `pkg/pagerank`:
  - local directed graph + PageRank implementation

## Design Choices

- **Strict host scope** avoids unintended subdomain/external crawling.
- **Browser-based extraction** captures rendered pages and dynamic content.
- **Deterministic output naming** makes repeated runs and diffs stable.
- **Report-first contract** supports downstream agent workflows.
- **Local PageRank package** avoids external service dependencies.
