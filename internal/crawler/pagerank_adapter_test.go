package crawler

import (
	"strings"
	"testing"
)

func TestPageRankIntegrationProducesScoresAndOrdering(t *testing.T) {
	graph := NewLinkGraph()
	a := "https://example.com/a"
	b := "https://example.com/b"
	c := "https://example.com/c"

	graph.AddEdge(a, b)
	graph.AddEdge(b, c)
	graph.AddEdge(c, b)

	result := &CrawlResult{
		Strategy: StrategyPageRank,
		Pages: []*Page{
			{URL: a, FinalURL: a, Status: StatusOK},
			{URL: b, FinalURL: b, Status: StatusOK},
			{URL: c, FinalURL: c, Status: StatusOK},
		},
	}

	ApplyPageRankScores(result, graph)

	if !strings.Contains(result.PageRankImplementation, "pkg/pagerank") {
		t.Fatalf("expected pagerank implementation metadata to reference pkg/pagerank, got %q", result.PageRankImplementation)
	}

	for _, page := range result.Pages {
		if page.Score == nil {
			t.Fatalf("expected score for page %s", page.URL)
		}
	}

	if len(result.Pages) == 0 || result.Pages[0].Score == nil {
		t.Fatalf("expected ordered pages with scores")
	}
	if result.Pages[0].FinalURL != b {
		t.Fatalf("expected page %s to be ranked first, got %s", b, result.Pages[0].FinalURL)
	}
}
