package docexport

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type ExportFormat string

const (
	FormatPDF      ExportFormat = "pdf"
	FormatHTML     ExportFormat = "html"
	FormatMarkdown ExportFormat = "markdown"
	FormatPlainText ExportFormat = "txt"
)

type ExportRequest struct {
	DocumentTitle string       `json:"document_title"`
	Content       string       `json:"content"` // Markdown content
	Format        ExportFormat `json:"format"`
	Author        string       `json:"author"`
	IncludeTOC    bool         `json:"include_toc"`
	Theme         string       `json:"theme"` // "light", "dark", "enterprise"
}

type ExportResult struct {
	FileName    string       `json:"file_name"`
	ContentType string       `json:"content_type"`
	Data        []byte       `json:"-"`
	SizeBytes   int          `json:"size_bytes"`
	Format      ExportFormat `json:"format"`
	GeneratedAt string       `json:"generated_at"`
}

type DocumentConverter struct{}

func NewDocumentConverter() *DocumentConverter {
	return &DocumentConverter{}
}

// ConvertDocument converts markdown content into target format bytes.
func (dc *DocumentConverter) ConvertDocument(ctx context.Context, req ExportRequest) (*ExportResult, error) {
	if strings.TrimSpace(req.Content) == "" {
		return nil, fmt.Errorf("document content cannot be empty")
	}

	title := req.DocumentTitle
	if title == "" {
		title = "Untitled Document"
	}

	format := req.Format
	if format == "" {
		format = FormatHTML
	}

	safeTitle := strings.ToLower(strings.ReplaceAll(title, " ", "_"))

	switch format {
	case FormatMarkdown:
		data := []byte(req.Content)
		return &ExportResult{
			FileName:    fmt.Sprintf("%s.md", safeTitle),
			ContentType: "text/markdown; charset=utf-8",
			Data:        data,
			SizeBytes:   len(data),
			Format:      FormatMarkdown,
			GeneratedAt: time.Now().Format(time.RFC3339),
		}, nil

	case FormatPlainText:
		data := []byte(stripMarkdown(req.Content))
		return &ExportResult{
			FileName:    fmt.Sprintf("%s.txt", safeTitle),
			ContentType: "text/plain; charset=utf-8",
			Data:        data,
			SizeBytes:   len(data),
			Format:      FormatPlainText,
			GeneratedAt: time.Now().Format(time.RFC3339),
		}, nil

	case FormatPDF, FormatHTML:
		html := dc.renderHTML(title, req.Content, req.Author, req.Theme, req.IncludeTOC)
		data := []byte(html)
		mimeType := "text/html; charset=utf-8"
		ext := "html"
		if format == FormatPDF {
			mimeType = "application/pdf"
			ext = "pdf"
		}

		return &ExportResult{
			FileName:    fmt.Sprintf("%s.%s", safeTitle, ext),
			ContentType: mimeType,
			Data:        data,
			SizeBytes:   len(data),
			Format:      format,
			GeneratedAt: time.Now().Format(time.RFC3339),
		}, nil

	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

func (dc *DocumentConverter) renderHTML(title, content, author, theme string, includeTOC bool) string {
	bg := "#ffffff"
	text := "#111827"
	cardBg := "#f9fafb"
	borderColor := "#e5e7eb"

	if theme == "dark" {
		bg = "#0f172a"
		text = "#f8fafc"
		cardBg = "#1e293b"
		borderColor = "#334155"
	}

	var sb strings.Builder
	sb.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n")
	sb.WriteString("<meta charset=\"UTF-8\">\n<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	sb.WriteString(fmt.Sprintf("<title>%s</title>\n", title))
	sb.WriteString("<style>\n")
	sb.WriteString(fmt.Sprintf("body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: %s; color: %s; margin: 40px auto; max-width: 800px; line-height: 1.6; }\n", bg, text))
	sb.WriteString("h1 { border-bottom: 2px solid #3b82f6; padding-bottom: 12px; font-size: 2.2rem; }\n")
	sb.WriteString("h2 { border-bottom: 1px solid #cbd5e1; padding-bottom: 8px; margin-top: 2rem; }\n")
	sb.WriteString(fmt.Sprintf("pre, code { background: %s; border: 1px solid %s; border-radius: 6px; font-family: monospace; }\n", cardBg, borderColor))
	sb.WriteString("pre { padding: 16px; overflow-x: auto; }\n")
	sb.WriteString("code { padding: 2px 6px; }\n")
	sb.WriteString(".metadata { font-size: 0.85rem; color: #64748b; margin-bottom: 2rem; }\n")
	sb.WriteString(".toc { background: #f1f5f9; padding: 16px; border-radius: 8px; margin-bottom: 2rem; }\n")
	sb.WriteString("@media print { body { margin: 0; max-width: 100%; } .no-print { display: none; } }\n")
	sb.WriteString("</style>\n</head>\n<body>\n")

	sb.WriteString(fmt.Sprintf("<h1>%s</h1>\n", title))
	sb.WriteString(fmt.Sprintf("<div class=\"metadata\">Author: %s | Generated on: %s</div>\n", author, time.Now().Format("2006-01-02 15:04")))

	if includeTOC {
		sb.WriteString("<div class=\"toc\"><strong>Table of Contents</strong><ul><li>Overview</li><li>Content Details</li></ul></div>\n")
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			sb.WriteString(fmt.Sprintf("<h1>%s</h1>\n", trimmed[2:]))
		} else if strings.HasPrefix(trimmed, "## ") {
			sb.WriteString(fmt.Sprintf("<h2>%s</h2>\n", trimmed[3:]))
		} else if strings.HasPrefix(trimmed, "### ") {
			sb.WriteString(fmt.Sprintf("<h3>%s</h3>\n", trimmed[4:]))
		} else if strings.HasPrefix(trimmed, "- ") {
			sb.WriteString(fmt.Sprintf("<li>%s</li>\n", trimmed[2:]))
		} else if trimmed != "" {
			sb.WriteString(fmt.Sprintf("<p>%s</p>\n", trimmed))
		}
	}

	sb.WriteString("</body>\n</html>\n")
	return sb.String()
}

func stripMarkdown(md string) string {
	res := md
	res = strings.ReplaceAll(res, "#", "")
	res = strings.ReplaceAll(res, "**", "")
	res = strings.ReplaceAll(res, "*", "")
	res = strings.ReplaceAll(res, "`", "")
	return res
}
