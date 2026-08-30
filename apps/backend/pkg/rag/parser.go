package rag

import (
	"regexp"
	"strings"
)

type ParsedDocument struct {
	Title     string   `json:"title"`
	Format    string   `json:"format"` // "markdown", "pdf", "code", "text"
	Sections  []string `json:"sections"`
	CharCount int      `json:"char_count"`
}

type SmartParser struct{}

func NewSmartParser() *SmartParser {
	return &SmartParser{}
}

// ParseDocumentText intelligently parses Markdown, Code, and Text into semantic sections
func (p *SmartParser) ParseDocumentText(content string, filename string) *ParsedDocument {
	format := detectFormat(filename, content)
	var sections []string

	if format == "markdown" {
		// Split by Markdown headers (#, ##, ###)
		reHeader := regexp.MustCompile(`(?m)^#{1,3}\s+`)
		rawSections := reHeader.Split(content, -1)
		for _, s := range rawSections {
			trimmed := strings.TrimSpace(s)
			if len(trimmed) > 0 {
				sections = append(sections, trimmed)
			}
		}
	}

	if len(sections) == 0 {
		// Fallback split by double newlines (paragraphs)
		rawSections := strings.Split(content, "\n\n")
		for _, s := range rawSections {
			trimmed := strings.TrimSpace(s)
			if len(trimmed) > 0 {
				sections = append(sections, trimmed)
			}
		}
	}

	return &ParsedDocument{
		Title:     filename,
		Format:    format,
		Sections:  sections,
		CharCount: len(content),
	}
}

func detectFormat(filename string, content string) string {
	lowerName := strings.ToLower(filename)
	if strings.HasSuffix(lowerName, ".md") || strings.Contains(content, "# ") {
		return "markdown"
	}
	if strings.HasSuffix(lowerName, ".go") || strings.HasSuffix(lowerName, ".py") || strings.HasSuffix(lowerName, ".js") {
		return "code"
	}
	if strings.HasSuffix(lowerName, ".pdf") {
		return "pdf"
	}
	return "text"
}
