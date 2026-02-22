# Contributing

## Prerequisites

- Go 1.25+
- Chrome or Chromium installed (for runtime testing)

## Local Checks

```sh
make ci
```

## Coding Standards

- Keep new code covered by tests where practical.
- Keep dependencies minimal.
- Preserve scope guarantees and robots compliance behavior.
- Update documentation for user-facing behavior changes.

## Pull Requests

Include:

- problem statement
- approach summary
- test evidence (`go test ./...` output)
