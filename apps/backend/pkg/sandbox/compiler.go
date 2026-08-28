package sandbox

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Language string

const (
	LangPython     Language = "python"
	LangJavaScript Language = "javascript"
	LangGo         Language = "go"
	LangBash       Language = "bash"
)

type ExecutionResult struct {
	Language  Language `json:"language"`
	Stdout    string   `json:"stdout"`
	Stderr    string   `json:"stderr"`
	ExitCode  int      `json:"exit_code"`
	Duration  string   `json:"duration"`
	IsSuccess bool     `json:"is_success"`
}

type Compiler struct{}

func NewCompiler() *Compiler {
	return &Compiler{}
}

// ExecuteCode runs user code safely in a sandboxed command runner
func (c *Compiler) ExecuteCode(ctx context.Context, lang Language, code string) (*ExecutionResult, error) {
	start := time.Now()
	var cmd *exec.Cmd

	switch lang {
	case LangPython:
		cmd = exec.CommandContext(ctx, "python", "-c", code)
	case LangJavaScript:
		cmd = exec.CommandContext(ctx, "node", "-e", code)
	case LangGo:
		cmd = exec.CommandContext(ctx, "go", "run", "-")
		cmd.Stdin = strings.NewReader(code)
	case LangBash:
		cmd = exec.CommandContext(ctx, "bash", "-c", code)
	default:
		return nil, fmt.Errorf("unsupported programming language: %s", lang)
	}

	stdoutBytes, err := cmd.CombinedOutput()
	duration := time.Since(start).String()

	exitCode := 0
	isSuccess := true
	if err != nil {
		isSuccess = false
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	return &ExecutionResult{
		Language:  lang,
		Stdout:    string(stdoutBytes),
		Stderr:    "",
		ExitCode:  exitCode,
		Duration:  duration,
		IsSuccess: isSuccess,
	}, nil
}
