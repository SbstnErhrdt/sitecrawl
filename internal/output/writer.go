package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sbstn/sitecrawl/internal/crawler"
)

type reportPage struct {
	URL         string   `json:"url"`
	FinalURL    string   `json:"final_url"`
	Depth       int      `json:"depth"`
	Status      string   `json:"status"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	OutPath     string   `json:"out_path,omitempty"`
	LinksCount  int      `json:"links_count"`
	Error       string   `json:"error,omitempty"`
	Score       *float64 `json:"score,omitempty"`
}

type reportTotals struct {
	Visited           int `json:"visited"`
	Errors            int `json:"errors"`
	SkippedExternal   int `json:"skipped_external"`
	SkippedOutOfScope int `json:"skipped_out_of_scope"`
}

type report struct {
	Domain                 string       `json:"domain"`
	AllowedHosts           []string     `json:"allowed_hosts"`
	StartedAt              time.Time    `json:"started_at"`
	FinishedAt             time.Time    `json:"finished_at"`
	Strategy               string       `json:"strategy"`
	MaxPages               int          `json:"max_pages"`
	MaxDepth               int          `json:"max_depth"`
	Clean                  bool         `json:"clean"`
	Headful                bool         `json:"headful"`
	PageRankImplementation string       `json:"pagerank_implementation,omitempty"`
	Pages                  []reportPage `json:"pages"`
	Totals                 reportTotals `json:"totals"`
}

// Write serializes page outputs and writes report.json into outDir.
func Write(result *crawler.CrawlResult, outDir string, format Format) error {
	if result == nil {
		return fmt.Errorf("nil crawl result")
	}
	if outDir == "" {
		return fmt.Errorf("output directory is empty")
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	mapper := NewFilenameMapper(format)
	for _, page := range result.Pages {
		if page.Status != crawler.StatusOK {
			continue
		}
		targetURL := page.FinalURL
		if targetURL == "" {
			targetURL = page.URL
		}
		filename := mapper.FilenameForURL(targetURL)
		content, err := renderPage(page, format, result.Clean)
		if err != nil {
			return err
		}
		fullPath := filepath.Join(outDir, filename)
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			return err
		}
		page.OutPath = filename
	}

	rep := buildReport(result)
	reportJSON, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outDir, "report.json"), reportJSON, 0o644)
}

func renderPage(page *crawler.Page, format Format, clean bool) (string, error) {
	switch format {
	case FormatMarkdown:
		var builder strings.Builder
		if page.Title != "" {
			builder.WriteString("# ")
			builder.WriteString(page.Title)
			builder.WriteString("\n\n")
		}
		builder.WriteString("Source: ")
		if page.FinalURL != "" {
			builder.WriteString(page.FinalURL)
		} else {
			builder.WriteString(page.URL)
		}
		builder.WriteString("\n\n")
		if page.Description != "" {
			builder.WriteString("Description: ")
			builder.WriteString(page.Description)
			builder.WriteString("\n\n")
		}
		if clean {
			builder.WriteString(page.MainText)
			builder.WriteString("\n")
		} else {
			builder.WriteString("```html\n")
			builder.WriteString(page.BodyHTML)
			builder.WriteString("\n```\n")
		}
		return builder.String(), nil
	case FormatHTML:
		if clean {
			title := page.Title
			if title == "" {
				title = "Untitled"
			}
			return "<!doctype html>\n<html><head><meta charset=\"utf-8\"><title>" +
				escapeHTML(title) +
				"</title></head><body>" +
				page.MainHTML +
				"</body></html>\n", nil
		}
		if page.RawHTML != "" {
			return page.RawHTML, nil
		}
		return page.BodyHTML, nil
	case FormatJSON:
		payload := map[string]any{
			"url":          page.URL,
			"final_url":    page.FinalURL,
			"title":        page.Title,
			"description":  page.Description,
			"links":        page.Links,
			"links_count":  len(page.Links),
			"clean":        clean,
			"content":      page.MainText,
			"content_html": page.MainHTML,
		}
		if !clean {
			payload["content"] = page.BodyHTML
			payload["content_html"] = page.BodyHTML
		}
		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data) + "\n", nil
	default:
		return "", fmt.Errorf("unsupported output format: %s", format)
	}
}

func buildReport(result *crawler.CrawlResult) report {
	pages := make([]reportPage, 0, len(result.Pages))
	for _, page := range result.Pages {
		pages = append(pages, reportPage{
			URL:         page.URL,
			FinalURL:    page.FinalURL,
			Depth:       page.Depth,
			Status:      page.Status,
			Title:       page.Title,
			Description: page.Description,
			OutPath:     page.OutPath,
			LinksCount:  len(page.Links),
			Error:       page.Error,
			Score:       page.Score,
		})
	}
	if result.Strategy == crawler.StrategyPageRank {
		sort.SliceStable(pages, func(i, j int) bool {
			iScore := -1.0
			jScore := -1.0
			if pages[i].Score != nil {
				iScore = *pages[i].Score
			}
			if pages[j].Score != nil {
				jScore = *pages[j].Score
			}
			if iScore != jScore {
				return iScore > jScore
			}
			return pages[i].URL < pages[j].URL
		})
	}

	return report{
		Domain:                 result.Domain,
		AllowedHosts:           append([]string(nil), result.AllowedHosts...),
		StartedAt:              result.StartedAt,
		FinishedAt:             result.FinishedAt,
		Strategy:               string(result.Strategy),
		MaxPages:               result.MaxPages,
		MaxDepth:               result.MaxDepth,
		Clean:                  result.Clean,
		Headful:                result.Headful,
		PageRankImplementation: result.PageRankImplementation,
		Pages:                  pages,
		Totals: reportTotals{
			Visited:           result.Totals.Visited,
			Errors:            result.Totals.Errors,
			SkippedExternal:   result.Totals.SkippedExternal,
			SkippedOutOfScope: result.Totals.SkippedOutOfScope,
		},
	}
}

func escapeHTML(value string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&#39;",
	)
	return replacer.Replace(value)
}
