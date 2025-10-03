package handler

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for development
		// In production, check the origin properly
		return true
	},
}

// EventHub manages WebSocket connections for a specific event type
type EventHub struct {
	name       string
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mu         sync.RWMutex
}

func NewEventHub(name string) *EventHub {
	return &EventHub{
		name:       name,
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

func (h *EventHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("[%s] WebSocket client connected, total: %d", h.name, len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mu.Unlock()
			log.Printf("[%s] WebSocket client disconnected, total: %d", h.name, len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Printf("[%s] Error broadcasting to client: %v", h.name, err)
					client.Close()
					h.mu.RUnlock()
					h.mu.Lock()
					delete(h.clients, client)
					h.mu.Unlock()
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all connected clients for this event
func (h *EventHub) Broadcast(message []byte) {
	h.broadcast <- message
}

// HandleWebSocket upgrades the HTTP connection and manages the WebSocket for this specific event
func (h *EventHub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[%s] WebSocket upgrade error: %v", h.name, err)
		return
	}

	h.register <- conn

	// Keep connection alive and handle incoming messages
	go func() {
		defer func() {
			h.unregister <- conn
		}()

		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("[%s] WebSocket error: %v", h.name, err)
				}
				break
			}

			// Echo the message back to this client (or handle as needed)
			if messageType == websocket.TextMessage {
				log.Printf("[%s] Received message: %s", h.name, message)
				// You can handle incoming messages here if needed
				// For now, we just keep the connection alive for broadcasting
			}
		}
	}()
}
