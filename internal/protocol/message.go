package protocol

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
)

type MessageType string

const (
	ForceQuit MessageType = "Force quit"

	ChallengeRequest  MessageType = "Challenge request"
	ChallengeResponse MessageType = "Challenge response"

	ResourceRequest  MessageType = "Resource request"
	ResourceResponse MessageType = "Resource response"
)

type Message struct {
	MessageType MessageType `json:"type"`
	Payload     string      `json:"payload,omitempty"`
}

func Read(reader *bufio.Reader) (*Message, error) {
	strMsg, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	var msg Message
	err = json.Unmarshal([]byte(strMsg), &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func Write(writer *bufio.Writer, msg *Message) error {
	if msg == nil {
		return errors.New("message is nil")
	}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = writer.WriteString(fmt.Sprintf("%s\n", string(msgBytes)))
	if err != nil {
		return err
	}
	return writer.Flush()
}
