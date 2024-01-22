package server

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	protocol "github.com/kirvader/wow-using-pow/internal/protocol"
	"github.com/kirvader/wow-using-pow/internal/utils"
	"github.com/kirvader/wow-using-pow/pkg/clock"
	pow "github.com/kirvader/wow-using-pow/pkg/proof_of_work"
	"github.com/kirvader/wow-using-pow/pkg/storage"
)

type hashcashConfig struct {
	zerosCount        int32
	challengeDuration *time.Duration
}

type Server struct {
	listener net.Listener

	storageSet storage.StorageSet
	clock      clock.Clock

	hashcashConfig *hashcashConfig
}

func NewServer(ctx context.Context, serverAddress, redisAddress string, hashcashZerosCount int32, hashcashChallengeDuration *time.Duration) (*Server, error) {
	listener, err := net.Listen("tcp", serverAddress)
	if err != nil {
		return nil, err
	}

	redisClient, err := storage.NewRedisCache(ctx, redisAddress)
	if err != nil {
		return nil, fmt.Errorf("error init cache: %v", err)
	}

	rand.Seed(time.Now().UnixNano())

	return &Server{
		storageSet: redisClient,
		listener:   listener,

		clock: &clock.SystemClock{},
		hashcashConfig: &hashcashConfig{
			zerosCount:        hashcashZerosCount,
			challengeDuration: hashcashChallengeDuration,
		},
	}, nil
}

func (srv *Server) Close(ctx context.Context) error {
	if err := srv.storageSet.Close(ctx); err != nil {
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
		msg, err := protocol.Read(reader)
		if err != nil {
			log.Printf("reading message from connection failed: %v", err)
			return
		}
		if err = srv.processRequest(ctx, msg, conn); err != nil {
			log.Printf("process request error: %v", err)
			return
		}
	}
}

func (srv *Server) processRequest(ctx context.Context, msg *protocol.Message, conn net.Conn) error {
	clientInfo := conn.RemoteAddr().String()
	currentTime := srv.clock.Now()
	connBufWriter := bufio.NewWriter(conn)

	switch msg.MessageType {
	case protocol.Quit:
		log.Print("client requests to close connection")
		return nil

	case protocol.ChallengeRequest:
		log.Printf("client %s requests challenge\n", clientInfo)

		randVal := utils.GenerateRandomString()

		// to check that server actually used this randVal
		err := srv.storageSet.Add(ctx, randVal, *srv.hashcashConfig.challengeDuration)
		if err != nil {
			return fmt.Errorf("err add rand to cache: %w", err)
		}

		hashcash := pow.Hashcash{
			Version:    1,
			ZerosCount: srv.hashcashConfig.zerosCount,
			Date:       &currentTime,
			Resource:   clientInfo,
			Rand:       randVal,
			Counter:    0,
		}

		return protocol.SendChallenge(connBufWriter, &hashcash)
	case protocol.ResourceRequest:
		powSolution := new(pow.Hashcash)
		if err := powSolution.FromJSON(msg.Payload); err != nil {
			return err
		}

		log.Printf("client solved challenge and requests resource. client: %s, payload %s\n", clientInfo, msg.Payload)

		// check if resource of puzzle is client's
		if powSolution.Resource != clientInfo {
			return errors.New("invalid hashcash resource")
		}

		// verify that puzzle's rand value was generated on server
		exists, err := srv.storageSet.Exists(ctx, powSolution.Rand)
		if err != nil {
			return fmt.Errorf("err get rand from cache: %w", err)
		}
		if !exists {
			return errors.New("challenge expired")
		}

		// as a rand value is generated for each challenge - we want to make sure it won't be used to verify duplicates
		if err := srv.storageSet.Delete(ctx, powSolution.Rand); err != nil {
			return fmt.Errorf("err get rand from cache: %w", err)
		}

		// solution shouldn't take more than setup challengeDuration
		if currentTime.Sub(*powSolution.Date) > *srv.hashcashConfig.challengeDuration {
			return errors.New("challenge expired")
		}

		// verify according to POW puzzle
		valid, err := powSolution.Verify()
		if err != nil {
			return fmt.Errorf("verifying hashcash failed: %v", err)
		}
		if !valid {
			return errors.New("invalid hashcash")
		}
		log.Print("Solution verified. Sending a word of wisdom.")

		return protocol.SendResource(connBufWriter)
	default:
		return errors.New("invalid message type")
	}
}
