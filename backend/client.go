package main

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type ClientList map[*Client]bool

type Client struct {
	connection *websocket.Conn
	manager    *Manager

	// egress is used to avoid concurrent writes to the WebSocket connection.
	egress chan []byte
}

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		connection: conn,
		manager:    manager,
		egress:     make(chan []byte),
	}
}

func (c *Client) readMessages() {
	defer func() {
		// clean up connection
		c.manager.RemoveClient(c)
	}()
	for {
		messageType, payload, err := c.connection.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("Error: %v\n", err)
			}
			break
		}

		for wsClient := range c.manager.clients {
			wsClient.egress <- payload
		}

		fmt.Printf("MessageType %d\n", messageType)
		fmt.Printf("Payload %s\n", string(payload))
	}
}

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
			if err := c.connection.WriteMessage(websocket.TextMessage, message); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			fmt.Printf("Message sent: %s\n", string(message))
		}
	}
}
