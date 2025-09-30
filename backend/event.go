package main

import "encoding/json"

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type EventHandler func(event Event, client *Client) error

const (
	EventSendMessage = "send_message"
	NewMessage       = "new_message"
)

type SendMessageEvent struct {
	Message string `json:"message"`
	From    string `json:"from"`
}
