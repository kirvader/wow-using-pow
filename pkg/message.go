package pkg

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
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
	strMsg, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	log.Printf("read bytes: %s.", strMsg)
	msg, err := ParseMessage([]byte(strMsg))
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
	log.Printf("sending %v", *msg)

	_, err = writer.WriteString(fmt.Sprintf("%s\n", string(msgBytes)))
	if err != nil {
		return err
	}
	return writer.Flush()
}
