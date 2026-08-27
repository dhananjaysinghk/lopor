package scraper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type WebScraper struct {
	client *http.Client
}

func NewWebScraper() *WebScraper {
	return &WebScraper{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// ScrapeURL fetches HTML from a URL and strips tags to return clean plain text for RAG indexing
func (s *WebScraper) ScrapeURL(ctx context.Context, targetURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	req.Header.Set("User-Agent", "LoporBot/1.0 (+https://lopor.ai)")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP status %d received", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Clean HTML tags and script/style blocks using regex
	cleanText := extractPlainText(string(body))
	return cleanText, nil
}

func extractPlainText(htmlContent string) string {
	// Remove script and style elements
	reScripts := regexp.MustCompile(`(?i)<(script|style)[^>]*>[\s\S]*?</\1>`)
	text := reScripts.ReplaceAllString(htmlContent, "")

	// Strip all HTML tags
	reTags := regexp.MustCompile(`<[^>]+>`)
	text = reTags.ReplaceAllString(text, " ")

	// Replace multiple whitespaces/newlines with single space
	reSpaces := regexp.MustCompile(`\s+`)
	text = reSpaces.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}
