# sitecrawl

`sitecrawl` is a production-focused domain crawler for agent workflows.
It uses Chrome (`chromedp`) to fetch real rendered pages, extracts content,
and emits deterministic page files plus a machine-readable crawl report.

## Key Features

- Strict crawl scope:
  - `<domain>`
  - `www.<domain>`
- Crawl strategies:
  - `pagerank` (default, backed by local `pkg/pagerank`)
  - `limit`
  - `depth`
- Robots compliance enabled by default (`robots.txt`)
- Clean content mode for agent-ready text output
- Deterministic per-page file naming
- Structured `report.json` with URL/title/description metadata + scores
- Graceful shutdown with partial output preservation

## Installation

### Homebrew (pipeline-managed)

```sh
brew tap <owner>/<tap-repo>
brew install sitecrawl
```

### From source

```sh
go build -o sitecrawl ./cmd/sitecrawl
```

## Quick Start

```sh
sitecrawl crawl --domain example.com --format md --out ./out
sitecrawl crawl --domain example.com --strategy limit --max-pages 50 --format md --out ./out
sitecrawl crawl --domain example.com --strategy depth --max-depth 2 --format json --out ./out --headful
```

## CLI Reference

Required:

- `--domain <string>`
- `--out <string>`
- `--format <md|html|json>`

Optional:

- `--strategy pagerank|limit|depth` (default: `pagerank`)
- `--max-pages <int>` (default: `25`)
- `--max-depth <int>` (default: `2`)
- `--clean` (default: `true`)
- `--headful` (default: `false`)
- `--delay-ms <int>` (default: `750`)
- `--page-timeout <duration>` (default: `20s`)
- `--user-agent <string>`
- `--log debug|info|warn|error` (default: `info`)

## Output Contract

Each run writes:

1. One file per page (`.md`, `.html`, or `.json`)
2. `report.json` with:
   - crawl metadata (`domain`, `strategy`, times, options)
   - per-page metadata:
     - `url`
     - `title`
     - `description`
     - `final_url`
     - `status`
     - `out_path`
     - `links_count`
     - `score` (when `strategy=pagerank`)
   - totals (`visited`, `errors`, `skipped_external`, `skipped_out_of_scope`)

## Agent Skill

This repository includes an agent skill definition at:

- `SKILL.md`
- `agents/openai.yaml`

Use the skill when you need repeatable website crawling with clean content and report-driven downstream automation.

For Codex/local agent environments, point the skill loader at this repository root
or copy `SKILL.md` (+ optional `agents/openai.yaml`) into your skills directory.

## Development

```sh
go test ./...
go build ./cmd/sitecrawl
```

Useful local targets are documented in `Makefile`.

## CI / Releases

- CI matrix runs on:
  - Ubuntu
  - macOS Intel
  - macOS Apple Silicon
  - Windows
- Release pipeline builds binaries for:
  - `darwin/amd64`, `darwin/arm64`
  - `linux/amd64`, `linux/arm64`
  - `windows/amd64`, `windows/arm64`
- Homebrew formula publishing is automated through GoReleaser workflow.

See:

- `.github/workflows/ci.yml`
- `.github/workflows/release.yml`
- `.goreleaser.yaml`
