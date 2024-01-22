package server

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"time"

	"github.com/kirvader/wow-using-pow/pkg"
)

type hashcashConfig struct {
	zerosCount        int32
	challengeDuration *time.Duration
}

type Server struct {
	redisCache *RedisClient
	listener   net.Listener

	clock Clock

	hashcashConfig *hashcashConfig
}

func NewServer(ctx context.Context, serverAddress, redisAddress string, hashcashZerosCount int32, hashcashChallengeDuration *time.Duration) (*Server, error) {
	listener, err := net.Listen("tcp", serverAddress)
	if err != nil {
		return nil, err
	}

	redisClient, err := InitRedisCache(ctx, redisAddress)
	if err != nil {
		return nil, fmt.Errorf("error init cache: %v", err)
	}

	rand.Seed(time.Now().UnixNano())

	return &Server{
		redisCache: redisClient,
		listener:   listener,

		clock: &SystemClock{},
		hashcashConfig: &hashcashConfig{
			zerosCount:        hashcashZerosCount,
			challengeDuration: hashcashChallengeDuration,
		},
	}, nil
}

func (srv *Server) Close(ctx context.Context) error {
	if err := srv.redisCache.Close(ctx); err != nil {
		return err
	}
	if err := srv.listener.Close(); err != nil {
		return err
	}
	return nil
}

func (srv *Server) Run(ctx context.Context) error {
	for {
		conn, err := srv.listener.Accept()
		if err != nil {
			return fmt.Errorf("error accepting connection: %w", err)
		}
		go srv.handleConnection(ctx, conn)
	}
}

func (srv *Server) handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	log.Printf("handling connection: %s", conn.RemoteAddr().String())
	reader := bufio.NewReader(conn)
	for {
		msg, err := pkg.ReadMsg(reader)
		if err != nil {
			log.Printf("reading message from connection failed: %v", err)
		}
		if err = srv.processRequest(ctx, msg, conn); err != nil {
			log.Printf("process request error: %v", err)
			return
		}
	}
}

func (srv *Server) processRequest(ctx context.Context, msg *pkg.Message, conn net.Conn) error {
	clientInfo := conn.RemoteAddr().String()
	currentTime := srv.clock.Now()

	switch msg.MessageType {
	case pkg.Quit:
		return errors.New("client requests to close connection")
	case pkg.RequestChallenge:
		log.Printf("client %s requests challenge\n", clientInfo)

		randVal := rand.Intn(100000)
		err := srv.redisCache.Add(ctx, randVal, *srv.hashcashConfig.challengeDuration)
		if err != nil {
			return fmt.Errorf("err add rand to cache: %w", err)
		}

		hashcash := pkg.HashcashHeader{
			Version:    1,
			ZerosCount: srv.hashcashConfig.zerosCount,
			Date:       &currentTime,
			Resource:   clientInfo,
			Rand:       base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(randVal))),
			Counter:    0,
		}
		hashcashMarshaled, err := json.Marshal(hashcash)
		if err != nil {
			return fmt.Errorf("err marshal hashcash: %v", err)
		}
		msg := &pkg.Message{
			MessageType: pkg.ResponseChallenge,
			Payload:     string(hashcashMarshaled),
		}
		return sendMessageToClient(msg, conn)
	case pkg.RequestResource:
		var got pkg.HashcashHeader
		err := json.Unmarshal([]byte(msg.Payload), &got)
		if err != nil {
			return fmt.Errorf("err unmarshal hashcash: %w", err)
		}
		log.Printf("client solved challenge and requests resource. client: %s, payload %s\n", clientInfo, msg.Payload)
		if got.Resource != clientInfo {
			return errors.New("invalid hashcash resource")
		}

		randValBytes, err := base64.StdEncoding.DecodeString(got.Rand)
		if err != nil {
			return fmt.Errorf("err decode rand: %w", err)
		}
		randVal, err := strconv.Atoi(string(randValBytes))
		if err != nil {
			return fmt.Errorf("err decode rand: %w", err)
		}

		exists, err := srv.redisCache.Exists(ctx, randVal)
		if err != nil {
			return fmt.Errorf("err get rand from cache: %w", err)
		}
		if !exists {
			return fmt.Errorf("challenge expired or not sent")
		}

		if currentTime.Sub(*got.Date) > *srv.hashcashConfig.challengeDuration {
			return errors.New("challenge expired")
		}
		maxIter := got.Counter
		_, err = pkg.ComputeHashcash(&got, maxIter)
		if err != nil {
			return fmt.Errorf("invalid hashcash: %v", err)
		}
		log.Print("Success. Sending a word of wisdom.")
		msg := &pkg.Message{
			MessageType: pkg.ResponseResource,
			Payload:     WOWQuotes[rand.Intn(len(WOWQuotes))],
		}

		// todo check where do i need to delete data for better stability
		srv.redisCache.Delete(ctx, randVal)
		return sendMessageToClient(msg, conn)
	default:
		return errors.New("invalid message type")
	}
}

func sendMessageToClient(msg *pkg.Message, conn net.Conn) error {
	connBufWriter := bufio.NewWriter(conn)
	return pkg.SendMsg(connBufWriter, msg)
}
