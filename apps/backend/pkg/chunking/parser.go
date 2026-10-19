package chunking

import (
	"regexp"
	"strings"
)

// DocumentSection represents a structurally coherent block of text.
type DocumentSection struct {
	Title          string
	BreadcrumbPath string
	Content        string
	Format         ContentFormat
	HasCodeBlock   bool
	HasTable       bool
	StartLine      int
	EndLine        int
}

// Parser detects formats and divides documents along structural boundaries.
type Parser struct{}

// NewParser creates a new structural parser.
func NewParser() *Parser {
	return &Parser{}
}

// DetectFormat classifies the predominant content type of a document.
func (p *Parser) DetectFormat(content string) ContentFormat {
	trimmed := strings.TrimSpace(content)

	// Check 1: Code presence
	codeRegex := regexp.MustCompile(`(?m)^(package\s+[a-z0-9_]+|import\s+[\("]|func\s+[A-Za-z0-9_]+\(|def\s+[a-z0-9_]+\(|class\s+[A-Za-z0-9_]+|export\s+(default\s+)?(function|const|class)|public\s+class)`)
	if codeRegex.MatchString(trimmed) {
		return FormatCode
	}

	// Check 2: Table structure
	tableRegex := regexp.MustCompile(`(?m)\|[\s\-:]+\|[\s\-:]+\|`)
	if tableRegex.MatchString(trimmed) {
		return FormatTable
	}

	// Check 3: Markdown structure
	mdRegex := regexp.MustCompile(`(?m)^(#{1,6}\s+|\*\s+|\-\s+|>\s+)`)
	if mdRegex.MatchString(trimmed) || strings.Contains(trimmed, "```") {
		return FormatMarkdown
	}

	return FormatProse
}

// ParseMarkdownSections partitions markdown along heading boundaries and preserves code/table integrity.
func (p *Parser) ParseMarkdownSections(documentTitle, content string) []DocumentSection {
	lines := strings.Split(content, "\n")
	var sections []DocumentSection

	type headingItem struct {
		level int
		title string
	}
	var headingStack []headingItem

	var currentContent strings.Builder
	currentTitle := documentTitle
	if currentTitle == "" {
		currentTitle = "Overview"
	}
	hasCode := false
	hasTable := false
	inCodeBlock := false
	startLine := 0

	headingRegex := regexp.MustCompile(`^(#{1,6})\s+(.+)$`)

	buildBreadcrumb := func() string {
		if len(headingStack) == 0 {
			if documentTitle != "" {
				return documentTitle
			}
			return "Overview"
		}
		var parts []string
		if documentTitle != "" {
			parts = append(parts, documentTitle)
		}
		for _, h := range headingStack {
			parts = append(parts, h.title)
		}
		return strings.Join(parts, " > ")
	}

	flushSection := func(endLine int) {
		c := strings.TrimSpace(currentContent.String())
		if c != "" {
			sections = append(sections, DocumentSection{
				Title:          currentTitle,
				BreadcrumbPath: buildBreadcrumb(),
				Content:        c,
				Format:         FormatMarkdown,
				HasCodeBlock:   hasCode,
				HasTable:       hasTable,
				StartLine:      startLine,
				EndLine:        endLine,
			})
		}
		currentContent.Reset()
		hasCode = false
		hasTable = false
		startLine = endLine + 1
	}

	for idx, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Code fence toggle
		if strings.HasPrefix(trimmedLine, "```") {
			inCodeBlock = !inCodeBlock
			hasCode = true
			currentContent.WriteString(line + "\n")
			continue
		}

		if inCodeBlock {
			currentContent.WriteString(line + "\n")
			continue
		}

		// Table line check
		if strings.HasPrefix(trimmedLine, "|") && strings.HasSuffix(trimmedLine, "|") {
			hasTable = true
		}

		// Heading match
		if match := headingRegex.FindStringSubmatch(line); len(match) == 3 {
			level := len(match[1])
			headingText := strings.TrimSpace(match[2])

			// Flush prior section if it contains text
			if currentContent.Len() > 0 {
				flushSection(idx - 1)
			}

			// Pop deeper or equal headings from stack
			for len(headingStack) > 0 && headingStack[len(headingStack)-1].level >= level {
				headingStack = headingStack[:len(headingStack)-1]
			}
			headingStack = append(headingStack, headingItem{level: level, title: headingText})
			currentTitle = headingText

			currentContent.WriteString(line + "\n")
			continue
		}

		currentContent.WriteString(line + "\n")
	}

	// Flush remaining content
	if currentContent.Len() > 0 {
		flushSection(len(lines) - 1)
	}

	return sections
}

// ParseCodeSections partitions code into functional blocks (functions, classes, types).
func (p *Parser) ParseCodeSections(content string) []DocumentSection {
	lines := strings.Split(content, "\n")
	var sections []DocumentSection

	var current strings.Builder
	currentTitle := "Module Declarations"
	startLine := 0

	declRegex := regexp.MustCompile(`^(func\s+[A-Za-z0-9_]+|type\s+[A-Za-z0-9_]+|def\s+[a-z0-9_]+|class\s+[A-Za-z0-9_]+|export\s+(function|const|class))`)

	for idx, line := range lines {
		if declRegex.MatchString(strings.TrimSpace(line)) && current.Len() > 0 {
			sections = append(sections, DocumentSection{
				Title:          currentTitle,
				BreadcrumbPath: "Code > " + currentTitle,
				Content:        strings.TrimSpace(current.String()),
				Format:         FormatCode,
				HasCodeBlock:   true,
				StartLine:      startLine,
				EndLine:        idx - 1,
			})
			current.Reset()
			currentTitle = strings.TrimSpace(line)
			startLine = idx
		}
		current.WriteString(line + "\n")
	}

	if current.Len() > 0 {
		sections = append(sections, DocumentSection{
			Title:          currentTitle,
			BreadcrumbPath: "Code > " + currentTitle,
			Content:        strings.TrimSpace(current.String()),
			Format:         FormatCode,
			HasCodeBlock:   true,
			StartLine:      startLine,
			EndLine:        len(lines) - 1,
		})
	}

	return sections
}
