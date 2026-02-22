// Package crawler implements the crawl runtime, including:
//   - strict host scoping (<domain> + www.<domain>)
//   - URL normalization and deduplication
//   - robots.txt compliance
//   - crawl strategy execution (pagerank, limit, depth)
//   - PageRank integration through the local pkg/pagerank adapter
package crawler
