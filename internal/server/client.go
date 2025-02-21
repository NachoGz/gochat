package server

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	// ws connection
	conn *websocket.Conn
	// buffered channel of outbound messages
	send chan Message
	// User info
	username string
}

func (c *Client) readPump(s *Server) {
	for client := range s.clients {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			log.Printf("error reading message: %v", err)
			return
		}
		log.Printf("message received: %s", message)
	}
}

func (c *Client) writePump(s *Server) {
	for client := range s.clients {
		for elem := range client.send {
			err := c.conn.WriteMessage(websocket.TextMessage, []byte(elem.Content))
			if err != nil {
				log.Printf("error writing message: %v", err)
				return
			}
		}
	}
}
