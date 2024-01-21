package pkg

import (
	"bufio"
	"encoding/json"
)

type MessageType int32

const (
	Quit MessageType = iota

	RequestChallenge
	ResponseChallenge

	RequestResource
	ResponseResource
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

func ReadMsg(reader *bufio.Reader) (*Message, error) {
	var msgBytes []byte
	if _, err := reader.Read(msgBytes); err != nil {
		return nil, err
	}
	msg, err := ParseMessage(msgBytes)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func SendMsg(writer *bufio.Writer, msg *Message) error {
	msgBytes, err := msg.Marshal()
	if err != nil {
		return err
	}

	_, err = writer.Write(msgBytes)
	return err
}
