package collaboration

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
)

type EventType string

const (
	EventCursorMove   EventType = "cursor_move"
	EventUserPresence EventType = "user_presence"
	EventDocChange    EventType = "doc_change"
)

type EventMessage struct {
	Type       EventType   `json:"type"`
	DocumentID string      `json:"document_id"`
	UserID     string      `json:"user_id"`
	UserName   string      `json:"user_name"`
	Payload    interface{} `json:"payload"`
}

type Client struct {
	Conn       *websocket.Conn
	UserID     string
	UserName   string
	DocumentID string
	Send       chan []byte
}

type Hub struct {
	mu         sync.RWMutex
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

var GlobalHub = NewHub()

func NewHub() *Hub {
	h := &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	go h.run()
	return h
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("[WebSocket Hub] Client %s joined document %s", client.UserName, client.DocumentID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				log.Printf("[WebSocket Hub] Client %s disconnected", client.UserName)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) RegisterClient(c *Client) {
	h.register <- c
}

func (h *Hub) UnregisterClient(c *Client) {
	h.unregister <- c
}

func (h *Hub) BroadcastEvent(event EventMessage) {
	bytes, err := json.Marshal(event)
	if err == nil {
		h.broadcast <- bytes
	}
}
