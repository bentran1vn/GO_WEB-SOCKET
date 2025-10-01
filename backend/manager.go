package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	webSocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     checkOrigin,
	}
)

type Manager struct {
	clients ClientList
	sync.RWMutex

	otps RetentionMap

	handlers map[string]EventHandler
}

func NewManager(ctx context.Context) *Manager {
	m := &Manager{
		clients:  make(ClientList),
		handlers: make(map[string]EventHandler),
		otps:     NewRetentionMap(ctx, 5*time.Minute),
	}

	m.setupEventHandlers()

	return m
}

func (m *Manager) setupEventHandlers() {
	m.handlers[EventSendMessage] = SendMessage
	m.handlers[EventChangeChatRoom] = ChatRoomChangeHandler
}

func ChatRoomChangeHandler(event Event, c *Client) error {
	var chatRoomEvent ChangeChatRoomEvent

	if err := json.Unmarshal(event.Payload, &chatRoomEvent); err != nil {
		return fmt.Errorf("could not unmarshal payload: %w", err)
	}

	c.chatroom = chatRoomEvent.Name
	return nil
}

func SendMessage(event Event, c *Client) error {
	var chatevent SendMessageEvent

	if err := json.Unmarshal(event.Payload, &chatevent); err != nil {
		return fmt.Errorf("could not unmarshal payload: %w", err)
	}

	var broadcastEvent NewMessageEvent

	broadcastEvent.Message = chatevent.Message
	broadcastEvent.From = chatevent.From
	broadcastEvent.Sent = time.Now()

	data, err := json.Marshal(broadcastEvent)
	if err != nil {
		return fmt.Errorf("could not marshal payload: %w", err)
	}

	outgoingEvent := Event{
		Type:    EventNewMessage,
		Payload: data,
	}

	for client := range c.manager.clients {
		if client.chatroom == c.chatroom {
			client.egress <- outgoingEvent
		}
	}

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

	otp := r.URL.Query().Get("otp")

	if otp == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !m.otps.ValidateOTP(otp) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

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

func (m *Manager) loginHandler(w http.ResponseWriter, r *http.Request) {
	type userLoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var req userLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Username == "ben" && req.Password == "123" {
		type response struct {
			OTP string `json:"otp"`
		}

		otp := m.otps.NewOTP()
		resp := response{OTP: otp.Key}

		data, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "Failed to generate response", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	}
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("Invalid credentials"))
	return
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

func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")

	switch origin {
	case "https://localhost:8080":
		return true
	case "https://localhost:9000":
		return true
	default:
		return false
	}
	return true
}
