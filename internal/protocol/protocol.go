package protocol

import (
	"bufio"
	"fmt"
	"math/rand"

	pow "github.com/kirvader/wow-using-pow/pkg/proof_of_work"
)

func RequestChallenge(bufWriter *bufio.Writer) error {
	return Write(bufWriter, &Message{
		MessageType: ChallengeRequest,
	})
}

func SendChallenge(bufWriter *bufio.Writer, puzzle pow.POWPuzzle) error {
	payload, err := puzzle.ToJSON()
	if err != nil {
		return fmt.Errorf("marshalling puzzle failed: %v", err)
	}
	msg := &Message{
		MessageType: ChallengeResponse,
		Payload:     payload,
	}
	return Write(bufWriter, msg)
}

func ReceiveChallenge(bufReader *bufio.Reader) (pow.POWPuzzle, error) {
	msg, err := Read(bufReader)
	if err != nil {
		return nil, fmt.Errorf("reading challenge msg failed: %w", err)
	}
	hashcash := new(pow.Hashcash)
	if err = hashcash.FromJSON(msg.Payload); err != nil {
		return nil, fmt.Errorf("hashcash unmarshal failed: %w", err)
	}
	return hashcash, nil
}

func SendChallengeSolution(bufWriter *bufio.Writer, solution pow.POWPuzzle) error {
	payload, err := solution.ToJSON()
	if err != nil {
		return err
	}
	err = Write(bufWriter, &Message{
		MessageType: ResourceRequest,
		Payload:     payload,
	})
	if err != nil {
		return fmt.Errorf("sending solution failed: %v", err)
	}
	return nil
}

func SendResource(bufWriter *bufio.Writer) error {
	msg := &Message{
		MessageType: ResourceResponse,
		Payload:     WOWQuotes[rand.Intn(len(WOWQuotes))],
	}

	return Write(bufWriter, msg)
}

func ReceiveResource(bufReader *bufio.Reader) (string, error) {
	msgWithResource, err := Read(bufReader)
	if err != nil {
		return "", fmt.Errorf("receiving resource failed: %v", err)
	}
	return msgWithResource.Payload, nil
}
