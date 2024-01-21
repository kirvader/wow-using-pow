package server

import (
	"encoding/json"
)

type MessageType int32

const (
	Quit              MessageType = iota // on quit each side (server or client) should close connection
	RequestChallenge                     // from client to server - request new challenge from server
	ResponseChallenge                    // from server to client - message with challenge for client
	RequestResource                      // from client to server - message with solved challenge
	ResponseResource                     // from server to client - message with useful info is solution is correct, or with error if not
)

type Message struct {
	MessageType MessageType `json:"type"`
	Payload     string      `json:"payload,omitempty"`
}

func (m *Message) Marshal() ([]byte, error) {
	result, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func ParseMessage(msgBytes []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(msgBytes, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}
