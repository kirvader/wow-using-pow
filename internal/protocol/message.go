package protocol

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
)

type MessageType int32

const (
	Quit MessageType = iota

	ChallengeRequest
	ChallengeResponse

	ResourceRequest
	ResourceResponse
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
	log.Printf("read bytes: %s.", strMsg)
	var msg Message
	err = json.Unmarshal([]byte(strMsg), &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func Write(writer *bufio.Writer, msg *Message) error {
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
