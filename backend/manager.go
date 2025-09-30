package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	webSocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type Manager struct {
	clients ClientList
	sync.RWMutex

	handlers map[string]EventHandler
}

func NewManager() *Manager {
	m := &Manager{
		clients:  make(ClientList),
		handlers: make(map[string]EventHandler),
	}

	m.setupEventHandlers()

	return m
}

func (m *Manager) setupEventHandlers() {
	m.handlers[EventSendMessage] = SendMessage
}

func SendMessage(event Event, c *Client) error {
	fmt.Printf("Sending event: %v\n", event)
	return nil
}

func (m *Manager) routeEvent(event Event, c *Client) error {
	if handler, ok := m.handlers[event.Type]; ok {
		if err := handler(event, c); err != nil {
			return err
		}
		return nil
	} else {
		fmt.Println("No handler for event type:", event.Type)
	}
	return fmt.Errorf("no handler for event type: %s", event.Type)
}

// ServeWS handles WebSocket connections.
func (m *Manager) ServeWS(w http.ResponseWriter, r *http.Request) {
	fmt.Println("New WebSocket connection")

	conn, err := webSocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed to upgrade to WebSocket:", err)
		return
	}
	//defer conn.Close()

	client := NewClient(conn, m)
	m.AddClient(client)

	// Start Client Processes
	go client.readMessages()
	go client.writeMessages()
}

func (m *Manager) AddClient(c *Client) {
	m.Lock()
	defer m.Unlock()
	m.clients[c] = true
	fmt.Println("Client added. Total clients:", len(m.clients))
}

func (m *Manager) RemoveClient(c *Client) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.clients[c]; ok {
		delete(m.clients, c)
		fmt.Println("Client removed. Total clients:", len(m.clients))
	}
}
