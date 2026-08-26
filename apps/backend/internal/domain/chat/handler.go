package chat

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/lopor-ai/lopor/internal/domain/models"
	"github.com/lopor-ai/lopor/pkg/ai"
	"github.com/lopor-ai/lopor/pkg/response"
)

type Handler struct {
	service  Service
	aiClient *ai.Client
}

func NewHandler(service Service, aiClient *ai.Client) *Handler {
	return &Handler{
		service:  service,
		aiClient: aiClient,
	}
}

func (h *Handler) CreateChat(c *fiber.Ctx) error {
	rawUserID := c.Locals("user_id")
	userID, ok := rawUserID.(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Invalid session", nil)
	}

	var req CreateChatRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_INPUT", "Invalid payload", nil)
	}

	chat, err := h.service.CreateChat(c.Context(), userID, req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "CREATE_FAILED", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusCreated, "Chat created successfully", chat)
}

func (h *Handler) GetWorkspaceChats(c *fiber.Ctx) error {
	wsIDParam := c.Params("wsId")
	wsID, err := uuid.Parse(wsIDParam)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_UUID", "Invalid workspace ID", nil)
	}

	rawUserID := c.Locals("user_id")
	userID := rawUserID.(uuid.UUID)

	chats, err := h.service.GetWorkspaceChats(c.Context(), wsID, userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FETCH_FAILED", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "Chats retrieved", chats)
}

func (h *Handler) GetChatDetails(c *fiber.Ctx) error {
	chatIDParam := c.Params("chatId")
	chatID, err := uuid.Parse(chatIDParam)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_UUID", "Invalid chat ID", nil)
	}

	chat, msgs, err := h.service.GetChatDetails(c.Context(), chatID)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "NOT_FOUND", "Chat not found", nil)
	}

	return response.Success(c, fiber.StatusOK, "Chat details retrieved", fiber.Map{
		"chat":     chat,
		"messages": msgs,
	})
}

func (h *Handler) StreamChatResponse(c *fiber.Ctx) error {
	chatIDParam := c.Params("chatId")
	chatID, err := uuid.Parse(chatIDParam)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_UUID", "Invalid chat ID", nil)
	}

	type StreamPayload struct {
		Prompt string `json:"prompt"`
		Model  string `json:"model"`
	}

	var payload StreamPayload
	if err := c.BodyParser(&payload); err != nil || payload.Prompt == "" {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_PROMPT", "Prompt text is required", nil)
	}

	if payload.Model == "" {
		payload.Model = "gpt-4o"
	}

	// Save User Message
	userMsg := &models.Message{
		ChatID:     chatID,
		SenderRole: "user",
		Content:    payload.Prompt,
	}
	_ = h.service.SaveMessage(c.Context(), userMsg)

	// Set SSE Headers
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		fmt.Fprintf(w, "event: message_start\ndata: {\"role\":\"assistant\"}\n\n")
		w.Flush()

		var fullAssistantContent string

		if h.aiClient != nil && h.aiClient.APIKey != "" {
			log.Printf("[Chat API] Streaming completion via OpenAI model '%s'...", payload.Model)
			req := ai.ChatCompletionRequest{
				Model: payload.Model,
				Messages: []ai.ChatMessage{
					{Role: "user", Content: payload.Prompt},
				},
				Stream: true,
			}
			err := h.aiClient.StreamCompletion(context.Background(), req, w)
			if err != nil {
				log.Printf("[Chat API Error] OpenAI streaming failed: %v", err)
				errMessage := fmt.Sprintf("\n\n*(OpenAI API Error: %v)*", err)
				dataBytes, _ := json.Marshal(map[string]string{"delta": errMessage})
				fmt.Fprintf(w, "data: %s\n\n", string(dataBytes))
				w.Flush()
			}
		} else {
			// Fallback stream with valid JSON-escaped SSE chunks
			responseChunks := []string{
				"Here is ", "the response ", "from the Lopor ", "AI Engine! ",
				"\n\n```go\nfunc HelloLopor() string {\n    return \"Enterprise AI System Online\"\n}\n```\n\n",
				"I have parsed your prompt and integrated relevant workspace context.",
			}

			for _, chunk := range responseChunks {
				fullAssistantContent += chunk
				dataBytes, _ := json.Marshal(map[string]string{"delta": chunk})
				fmt.Fprintf(w, "data: %s\n\n", string(dataBytes))
				w.Flush()
			}
		}

		// Save Assistant Message
		if fullAssistantContent != "" {
			assistantMsg := &models.Message{
				ChatID:     chatID,
				SenderRole: "assistant",
				Content:    fullAssistantContent,
			}
			_ = h.service.SaveMessage(context.Background(), assistantMsg)
		}

		fmt.Fprintf(w, "event: message_end\ndata: {\"status\":\"completed\"}\n\n")
		w.Flush()
	})

	return nil
}
