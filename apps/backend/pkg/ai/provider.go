package ai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
)

type ProviderType string

const (
	ProviderOpenAI    ProviderType = "openai"
	ProviderAnthropic ProviderType = "anthropic"
	ProviderOllama    ProviderType = "ollama"
	ProviderDeepSeek  ProviderType = "deepseek"
)

type ModelInfo struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Provider    ProviderType `json:"provider"`
	ContextSize int          `json:"context_size"`
	IsLocal     bool         `json:"is_local"`
	IsHealthy   bool         `json:"is_healthy"`
}

type ModelRouter struct {
	providers map[ProviderType]*Client
	fallback  []ProviderType
}

func NewModelRouter() *ModelRouter {
	mr := &ModelRouter{
		providers: make(map[ProviderType]*Client),
		fallback:  []ProviderType{ProviderOpenAI, ProviderAnthropic, ProviderOllama},
	}
	mr.providers[ProviderOpenAI] = NewClient("", "https://api.openai.com/v1")
	mr.providers[ProviderOllama] = NewClient("", "http://localhost:11434/v1")
	return mr
}

// GetAvailableModels returns list of all active AI models across providers
func (mr *ModelRouter) GetAvailableModels() []ModelInfo {
	return []ModelInfo{
		{ID: "gpt-4o", Name: "OpenAI GPT-4o", Provider: ProviderOpenAI, ContextSize: 128000, IsLocal: false, IsHealthy: true},
		{ID: "gpt-4o-mini", Name: "OpenAI GPT-4o Mini", Provider: ProviderOpenAI, ContextSize: 128000, IsLocal: false, IsHealthy: true},
		{ID: "claude-3-5-sonnet", Name: "Anthropic Claude 3.5 Sonnet", Provider: ProviderAnthropic, ContextSize: 200000, IsLocal: false, IsHealthy: true},
		{ID: "llama3-8b", Name: "Ollama Llama 3 (Local GPU)", Provider: ProviderOllama, ContextSize: 8192, IsLocal: true, IsHealthy: true},
		{ID: "deepseek-coder", Name: "DeepSeek Coder V2", Provider: ProviderDeepSeek, ContextSize: 64000, IsLocal: false, IsHealthy: true},
	}
}

// RouteStream attempts to stream completion with automatic provider fallback
func (mr *ModelRouter) RouteStream(ctx context.Context, req ChatCompletionRequest, writer io.Writer) error {
	var lastErr error

	for _, provider := range mr.fallback {
		client, exists := mr.providers[provider]
		if !exists || client == nil {
			continue
		}

		log.Printf("[AI Model Router] Attempting completion stream via provider '%s' for model '%s'...", provider, req.Model)
		err := client.StreamCompletion(ctx, req, writer)
		if err == nil {
			return nil
		}

		log.Printf("[AI Model Router] Provider '%s' failed: %v. Initiating fallback...", provider, err)
		lastErr = err
	}

	if lastErr == nil {
		lastErr = errors.New("no valid AI providers available in router stack")
	}

	return fmt.Errorf("all AI provider fallbacks failed: %w", lastErr)
}
