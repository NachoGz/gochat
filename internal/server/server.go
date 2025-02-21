package server

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Server struct {
	// Keep track of active clients
	clients map[*Client]bool
	// channel for broadcasting messages to clients
	broadcast chan Message
	// Register new clients
	register chan *Client
	// Unregister clients
	unregister chan *Client
	// Thread-safe operations
	mu sync.RWMutex
}

type Message struct {
	Type     string `json:"type"` // "chat", "user_joined", "user_left"
	Content  string `json:"content"`
	Username string `json:"username"`
	Time     string `json:"time"`
}

// Create a new chat server
func NewServer() *Server {
	return &Server{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Start the chat server
func (s *Server) Start() {
	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client] = true
			s.mu.Unlock()

		case client := <-s.unregister:
			s.mu.Lock()
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client.send)
			}
			s.mu.Unlock()

		case message := <-s.broadcast:
			s.mu.RLock()
			for client := range s.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(s.clients, client)
				}
			}
			s.mu.RUnlock()
		}
	}
}

// handle ws connections
func (s *Server) HandleWS(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Websocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	client := &Client{
		conn:     conn,
		send:     make(chan Message, 256),
		username: r.URL.Query().Get("username"),
	}

	s.register <- client

	// Start goroutins for reading and writing messages
	go client.writePump(s)
	go client.readPump(s)
}
