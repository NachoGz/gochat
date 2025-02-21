package server

import (
	"encoding/json"
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
	defer func() {
		s.unregister <- c
		c.conn.Close()
	}()

	for {
		// read message from current client connection
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading message: %v", err)
			}
			return
		}

		// Parse message and broadcast
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("error marshaling message: %v", err)
			continue
		}

		s.broadcast <- msg
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		// read from send channel
		message, ok := <-c.send
		if !ok {
			// channel is closed
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		// write message to the current client connection
		err := c.conn.WriteJSON(message)
		if err != nil {
			log.Printf("error writing message: %v", err)
			return
		}
	}
}
