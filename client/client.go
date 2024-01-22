package client

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"

	protocol "github.com/kirvader/wow-using-pow/internal/protocol"
)

type Client struct {
	conn net.Conn

	hashCashMaxIterationsAmount int32
}

func NewClient(serverAddress string, hashCashMaxIterationsAmount int32) (*Client, error) {
	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		return nil, err
	}
	log.Printf("connected to %s", serverAddress)

	return &Client{
		conn:                        conn,
		hashCashMaxIterationsAmount: hashCashMaxIterationsAmount,
	}, nil
}

func (client *Client) Close() error {
	return client.conn.Close()
}

func (client *Client) HandleConnection(ctx context.Context) error {
	connBufReader := bufio.NewReader(client.conn)
	connBufWriter := bufio.NewWriter(client.conn)

	// request challenge
	if err := protocol.RequestChallenge(connBufWriter); err != nil {
		return fmt.Errorf("sending request for challenge failed: %v", err)
	}

	// receive challenge
	powPuzzle, err := protocol.ReceiveChallenge(connBufReader)
	if err != nil {
		return fmt.Errorf("reading challenge msg failed: %v", err)
	}
	log.Printf("solving puzzle: %v", powPuzzle)

	// solve challenge
	err = powPuzzle.Solve(client.hashCashMaxIterationsAmount)
	if err != nil {
		return fmt.Errorf("puzzle solving failed: %v", err)
	}
	log.Printf("puzzle solved: %v", powPuzzle)

	// send challenge solution
	err = protocol.SendChallengeSolution(connBufWriter, powPuzzle)
	if err != nil {
		return fmt.Errorf("sending request failed: %v", err)
	}
	log.Print("challenge solution sent to server")

	// receive resource
	resource, err := protocol.ReceiveResource(connBufReader)
	if err != nil {
		return fmt.Errorf("reading resource failed: %v", err)
	}
	log.Printf("Received quote: %s", resource)

	return nil
}
