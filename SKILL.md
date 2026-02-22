---
name: sitecrawl
description: Use this skill when you need to crawl a public website domain and produce agent-ready content files plus a structured report with URL/title/description metadata and optional PageRank scoring.
---

# sitecrawl Skill

Use this skill to collect high-quality website content for downstream agent workflows.

## When To Use

- You need a reproducible crawl of a single domain.
- You need content files (`md`, `html`, or `json`) plus a machine-readable report.
- You need URL metadata (`url`, `title`, `description`) for triage, ranking, and routing.
- You need scoped crawling limited to `<domain>` + `www.<domain>`.

## Inputs

Required:

- `domain` (example: `example.com`)
- `out` directory
- `format` (`md`, `html`, `json`)

Optional:

- `strategy` (`pagerank`, `limit`, `depth`)
- `max-pages`
- `max-depth`
- `clean`
- `headful`
- `delay-ms`
- `page-timeout`
- `user-agent`
- `log`

## Recommended Invocation

```sh
sitecrawl crawl --domain <domain> --format md --out <out_dir> --strategy pagerank --max-pages 100
```

Or through bundled helper script:

```sh
./scripts/run_crawl.sh --domain <domain> --format md --out <out_dir> --strategy pagerank --max-pages 100
```

## Workflow

1. Validate inputs and normalize domain scope.
2. Run crawl with strict host filtering and robots compliance.
3. Write per-page outputs.
4. Read `report.json` as the primary index for downstream steps.
5. Prioritize pages with `status=ok` and highest `score` (pagerank strategy).

For architecture and release details, read:

- `docs/ARCHITECTURE.md`
- `docs/RELEASES.md`

## Output Contract

Expect these outputs in `<out_dir>`:

- per-page files (`*.md`, `*.html`, or `*.json`)
- `report.json`

`report.json` contains:

- run metadata (`domain`, strategy, timings, options)
- page entries:
  - `url`
  - `title`
  - `description`
  - `final_url`
  - `status`
  - `out_path`
  - `links_count`
  - `score` (if strategy is `pagerank`)
- totals (`visited`, `errors`, skipped counters)

## Quality Checks

After crawl:

- ensure `errors` in `report.json` are acceptable
- verify high-value pages exist in `pages[]`
- confirm metadata completeness (`url`, `title`, `description`)
- if needed, rerun with larger `max-pages` or markdown format for cleaner content
