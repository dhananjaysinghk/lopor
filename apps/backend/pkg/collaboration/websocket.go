package collaboration

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// WebSocketUpgradeMiddleware checks if request is WebSocket upgrade
func WebSocketUpgradeMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}
}

// HandleWebSocketConnection manages live WebSocket events per document
func HandleWebSocketConnection(c *websocket.Conn) {
	docID := c.Params("docId")
	userID := c.Query("user_id", "anon_user")
	userName := c.Query("user_name", "Anonymous Contributor")

	client := &Client{
		Conn:       c,
		UserID:     userID,
		UserName:   userName,
		DocumentID: docID,
		Send:       make(chan []byte, 256),
	}

	GlobalHub.RegisterClient(client)

	// Broadcast user presence join event
	GlobalHub.BroadcastEvent(EventMessage{
		Type:       EventUserPresence,
		DocumentID: docID,
		UserID:     userID,
		UserName:   userName,
		Payload:    map[string]string{"status": "connected"},
	})

	// Reader Routine
	go func() {
		defer func() {
			GlobalHub.UnregisterClient(client)
			c.Close()
		}()

		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				break
			}
			GlobalHub.broadcast <- msg
		}
	}()

	// Writer Routine
	for msg := range client.Send {
		if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Printf("[WebSocket Write Error]: %v", err)
			break
		}
	}
}
