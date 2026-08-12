package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type Client struct {
	APIKey  string
	BaseURL string
}

func NewClient(apiKey, baseURL string) *Client {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &Client{
		APIKey:  apiKey,
		BaseURL: strings.TrimSuffix(baseURL, "/"),
	}
}

// StreamCompletion sends a request to OpenAI / Ollama and pipes raw SSE chunk deltas to an io.Writer
func (c *Client) StreamCompletion(ctx context.Context, req ChatCompletionRequest, writer io.Writer) error {
	req.Stream = true
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/chat/completions", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute ai completion request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ai provider error status %d: %s", resp.StatusCode, string(respBytes))
	}

	// Stream chunks directly to the response writer
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, _ = writer.Write(buf[:n])
			if flusher, ok := writer.(http.Flusher); ok {
				flusher.Flush()
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}

	return nil
}
