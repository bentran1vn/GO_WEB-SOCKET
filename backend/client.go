package main

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

type ClientList map[*Client]bool

type Client struct {
	connection *websocket.Conn
	manager    *Manager

	// egress is used to avoid concurrent writes to the WebSocket connection.
	egress chan Event
}

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		connection: conn,
		manager:    manager,
		egress:     make(chan Event),
	}
}

// Read means read message from WebSocket connection
func (c *Client) readMessages() {
	defer func() {
		// clean up connection
		c.manager.RemoveClient(c)
	}()
	for {
		_, payload, err := c.connection.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("Error: %v\n", err)
			}
			break
		}

		var request Event

		if err := json.Unmarshal(payload, &request); err != nil {
			fmt.Printf("Error unmarshaling message: %v\n", err)
			continue
		}

		if err := c.manager.routeEvent(request, c); err != nil {
			fmt.Printf("Error routing event: %v\n", err)
		}
	}
}

// Write mean write message to other clients
func (c *Client) writeMessages() {
	defer func() {
		// clean up connection
		c.manager.RemoveClient(c)
	}()
	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				if err := c.connection.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					fmt.Printf("Error: %v\n", err)
				}
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				fmt.Printf("Error marshaling message: %v\n", err)
				continue
			}

			if err := c.connection.WriteMessage(websocket.TextMessage, data); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			fmt.Printf("Message sent")
		}
	}
}
