GO ?= go
BINARY ?= sitecrawl

.PHONY: all fmt test build tidy lint ci smoke clean release-snapshot

all: fmt test build

fmt:
	$(GO) fmt ./...

test:
	$(GO) test ./...

build:
	$(GO) build -o $(BINARY) ./cmd/sitecrawl

tidy:
	$(GO) mod tidy

lint:
	$(GO) vet ./...

ci: tidy fmt lint test build

smoke: build
	./$(BINARY) crawl --domain example.com --out ./tmp/smoke --format md --max-pages 1 --strategy limit --delay-ms 0 --page-timeout 10s

clean:
	rm -f $(BINARY)
	rm -rf ./tmp

release-snapshot:
	goreleaser release --snapshot --clean
