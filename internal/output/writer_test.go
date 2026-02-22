package output

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sbstn/sitecrawl/internal/crawler"
)

func TestReportIncludesPageRankScores(t *testing.T) {
	tmpDir := t.TempDir()

	scoreHigh := 0.9
	scoreLow := 0.1
	result := &crawler.CrawlResult{
		Domain:       "example.com",
		AllowedHosts: []string{"example.com", "www.example.com"},
		Strategy:     crawler.StrategyPageRank,
		MaxPages:     25,
		MaxDepth:     2,
		Clean:        true,
		Pages: []*crawler.Page{
			{
				URL:      "https://example.com/a",
				FinalURL: "https://example.com/a",
				Status:   crawler.StatusOK,
				Score:    &scoreLow,
				MainText: "a",
			},
			{
				URL:      "https://example.com/b",
				FinalURL: "https://example.com/b",
				Status:   crawler.StatusOK,
				Score:    &scoreHigh,
				MainText: "b",
			},
		},
	}

	if err := Write(result, tmpDir, FormatJSON); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	reportBytes, err := os.ReadFile(filepath.Join(tmpDir, "report.json"))
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}

	var parsed struct {
		Pages []struct {
			URL   string   `json:"url"`
			Score *float64 `json:"score"`
		} `json:"pages"`
	}
	if err := json.Unmarshal(reportBytes, &parsed); err != nil {
		t.Fatalf("unexpected json error: %v", err)
	}

	if len(parsed.Pages) != 2 {
		t.Fatalf("expected 2 pages in report, got %d", len(parsed.Pages))
	}
	if parsed.Pages[0].Score == nil || parsed.Pages[1].Score == nil {
		t.Fatalf("expected score fields in report pages")
	}
	if parsed.Pages[0].URL != "https://example.com/b" {
		t.Fatalf("expected highest score first, got %s", parsed.Pages[0].URL)
	}
}

func TestReportIncludesAgentMetadataFields(t *testing.T) {
	tmpDir := t.TempDir()

	result := &crawler.CrawlResult{
		Domain:       "example.com",
		AllowedHosts: []string{"example.com", "www.example.com"},
		Strategy:     crawler.StrategyLimit,
		MaxPages:     1,
		MaxDepth:     1,
		Clean:        true,
		Pages: []*crawler.Page{
			{
				URL:         "https://example.com/",
				FinalURL:    "https://example.com/",
				Status:      crawler.StatusOK,
				Title:       "Example Domain",
				Description: "Example description",
				MainText:    "Example body",
			},
		},
	}

	if err := Write(result, tmpDir, FormatMarkdown); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	reportBytes, err := os.ReadFile(filepath.Join(tmpDir, "report.json"))
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}

	var parsed struct {
		Pages []struct {
			URL         string `json:"url"`
			Title       string `json:"title"`
			Description string `json:"description"`
		} `json:"pages"`
	}
	if err := json.Unmarshal(reportBytes, &parsed); err != nil {
		t.Fatalf("unexpected json error: %v", err)
	}

	if len(parsed.Pages) != 1 {
		t.Fatalf("expected 1 page in report, got %d", len(parsed.Pages))
	}
	if parsed.Pages[0].URL == "" || parsed.Pages[0].Title == "" || parsed.Pages[0].Description == "" {
		t.Fatalf("expected url/title/description metadata in report page entry")
	}
}
