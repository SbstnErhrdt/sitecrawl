#!/usr/bin/env bash
set -euo pipefail

if command -v sitecrawl >/dev/null 2>&1; then
  exec sitecrawl crawl "$@"
fi

exec go run ./cmd/sitecrawl crawl "$@"
