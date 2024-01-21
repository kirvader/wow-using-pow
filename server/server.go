package server

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/kirvader/wow-using-pow/pkg"
)

func Run(ctx context.Context, address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer listener.Close()
	fmt.Println("server listening on ", listener.Addr())
	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("error accept connection: %w", err)
		}
		go handleConnection(ctx, conn)
	}
}

func handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		var requestBytes []byte
		_, err := reader.Read(requestBytes)
		if err != nil {
			fmt.Println("err read connection:", err)
			return
		}
		msg, err := ProcessRequest(ctx, requestBytes, conn.RemoteAddr().String())
		if err != nil {
			fmt.Println("err process request:", err)
			return
		}
		if msg != nil {
			err := sendMessageToClient(*msg, conn)
			if err != nil {
				fmt.Println("err send message:", err)
			}
		}
	}
}

func ProcessRequest(ctx context.Context, reqBytes []byte, clientInfo string) (*Message, error) {
	msg, err := ParseMessage(reqBytes)
	if err != nil {
		return nil, err
	}

	switch msg.MessageType {
	case Quit:
		return nil, errors.New("client requests to close connection")
	case RequestChallenge:
		fmt.Printf("client %s requests challenge\n", clientInfo)
		conf := ctx.Value(ConfigContextKey{}).(*Config)
		date := time.Now()

		// TODO add redis storage to check
		randValue := rand.Intn(100000)

		hashcash := pkg.HashcashHeader{
			Version:    1,
			ZerosCount: conf.HashcashZerosCount,
			Date:       &date,
			Resource:   clientInfo,
			Rand:       base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", randValue))),
			Counter:    0,
		}
		hashcashMarshaled, err := json.Marshal(hashcash)
		if err != nil {
			return nil, fmt.Errorf("err marshal hashcash: %v", err)
		}
		msg := Message{
			MessageType: ResponseChallenge,
			Payload:     string(hashcashMarshaled),
		}
		return &msg, nil
	case RequestResource:
		fmt.Printf("client %s requests resource with payload %s\n", clientInfo, msg.Payload)
		var hashcash pkg.HashcashHeader
		err := json.Unmarshal([]byte(msg.Payload), &hashcash)
		if err != nil {
			return nil, fmt.Errorf("err unmarshal hashcash: %w", err)
		}
		if hashcash.Resource != clientInfo {
			return nil, fmt.Errorf("invalid hashcash resource")
		}
		conf := ctx.Value("config").(*Config)
		if time.Since(*hashcash.Date) > conf.HashcashChallengeDuration {
			return nil, fmt.Errorf("challenge expired")
		}
		maxIter := hashcash.Counter
		_, err = hashcash.ComputeHashcash(maxIter)
		if err != nil {
			return nil, fmt.Errorf("invalid hashcash: %v", err)
		}
		fmt.Print("Success. Sending a word of wisdom.")
		msg := Message{
			MessageType: ResponseResource,
			Payload:     WOWQuotes[rand.Intn(len(WOWQuotes))],
		}
		return &msg, nil
	default:
		return nil, errors.New("invalid message type")
	}
}

func sendMessageToClient(msg Message, conn net.Conn) error {
	marshalledMsg, err := msg.Marshal()
	if err != nil {
		return err
	}
	_, err = conn.Write(marshalledMsg)
	return err
}
