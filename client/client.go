package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/kirvader/wow-using-pow/pkg"
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
	fmt.Println("connected to", serverAddress)

	return &Client{
		conn:                        conn,
		hashCashMaxIterationsAmount: hashCashMaxIterationsAmount,
	}, nil
}

func (client *Client) Run(ctx context.Context) error {
	message, err := client.HandleConnection(ctx)
	if err != nil {
		return err
	}
	log.Printf("quote result: %s", message)
	return nil
}

func (client *Client) Close() error {
	return client.conn.Close()
}

func (client *Client) HandleConnection(ctx context.Context) (string, error) {
	connBufReader := bufio.NewReader(client.conn)
	connBufWriter := bufio.NewWriter(client.conn)

	// 1. requesting challenge
	err := pkg.SendMsg(connBufWriter, &pkg.Message{
		MessageType: pkg.RequestChallenge,
	})
	if err != nil {
		return "", fmt.Errorf("sending request failed: %w", err)
	}
	msg, err := pkg.ReadMsg(connBufReader)
	if err != nil {
		return "", fmt.Errorf("reading msg failed: %w", err)
	}
	var hashcash *pkg.HashcashHeader
	err = json.Unmarshal([]byte(msg.Payload), hashcash)
	if err != nil {
		return "", fmt.Errorf("hashcash unmarshal failed: %w", err)
	}
	log.Printf("got hashcash: %v", hashcash)

	// 2. got challenge, compute hashcash
	hashcash, err = pkg.ComputeHashcash(hashcash, client.hashCashMaxIterationsAmount)
	if err != nil {
		return "", fmt.Errorf("hashcash computing failed: %w", err)
	}
	log.Printf("hashcash computed: %v", hashcash)
	byteData, err := json.Marshal(hashcash)
	if err != nil {
		return "", fmt.Errorf("hashcash marshalling failed: %w", err)
	}

	// 3. send challenge solution back to server
	err = pkg.SendMsg(connBufWriter, &pkg.Message{
		MessageType: pkg.RequestResource,
		Payload:     string(byteData),
	})
	if err != nil {
		return "", fmt.Errorf("sending request failed: %w", err)
	}
	log.Print("challenge solution sent to server")

	// 4. get result quote from server
	resultMsg, err := pkg.ReadMsg(connBufReader)
	if err != nil {
		return "", fmt.Errorf("reading msg failed: %w", err)
	}
	return resultMsg.Payload, nil
}
