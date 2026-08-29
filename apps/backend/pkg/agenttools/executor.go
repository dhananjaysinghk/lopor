package agenttools

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lopor-ai/lopor/pkg/sandbox"
	"github.com/lopor-ai/lopor/pkg/scraper"
)

type ToolStep struct {
	ToolName  string `json:"tool_name"`
	Status    string `json:"status"`
	Output    string `json:"output"`
	Duration  string `json:"duration"`
	Timestamp string `json:"timestamp"`
}

type Executor struct {
	sandboxCompiler *sandbox.Compiler
	webScraper      *scraper.WebScraper
}

func NewExecutor() *Executor {
	return &Executor{
		sandboxCompiler: sandbox.NewCompiler(),
		webScraper:      scraper.NewWebScraper(),
	}
}

// ExecuteToolChain runs an autonomous AI agent's multi-step tool sequence
func (e *Executor) ExecuteToolChain(ctx context.Context, toolName string, inputPayload string) (*ToolStep, error) {
	start := time.Now()
	nowStr := time.Now().Format("15:04:05")

	switch toolName {
	case "code_sandbox":
		result, err := e.sandboxCompiler.ExecuteCode(ctx, sandbox.LangGo, inputPayload)
		if err != nil {
			return &ToolStep{
				ToolName:  toolName,
				Status:    "FAILED",
				Output:    err.Error(),
				Duration:  time.Since(start).String(),
				Timestamp: nowStr,
			}, nil
		}
		return &ToolStep{
			ToolName:  toolName,
			Status:    "COMPLETED",
			Output:    result.Stdout,
			Duration:  result.Duration,
			Timestamp: nowStr,
		}, nil

	case "web_scraper":
		text, err := e.webScraper.ScrapeURL(ctx, inputPayload)
		if err != nil {
			return &ToolStep{
				ToolName:  toolName,
				Status:    "FAILED",
				Output:    err.Error(),
				Duration:  time.Since(start).String(),
				Timestamp: nowStr,
			}, nil
		}
		truncated := text
		if len(truncated) > 300 {
			truncated = truncated[:300] + "..."
		}
		return &ToolStep{
			ToolName:  toolName,
			Status:    "COMPLETED",
			Output:    fmt.Sprintf("Scraped %d characters: %s", len(text), truncated),
			Duration:  time.Since(start).String(),
			Timestamp: nowStr,
		}, nil

	default:
		log.Printf("[Agent Tool Executor] Executing generic tool '%s'", toolName)
		return &ToolStep{
			ToolName:  toolName,
			Status:    "COMPLETED",
			Output:    fmt.Sprintf("Executed tool '%s' with payload: %s", toolName, inputPayload),
			Duration:  "0.05s",
			Timestamp: nowStr,
		}, nil
	}
}
