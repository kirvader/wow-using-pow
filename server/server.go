package server

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"time"

	protocol "github.com/kirvader/wow-using-pow/internal/protocol"
	"github.com/kirvader/wow-using-pow/pkg/clock"
	pow "github.com/kirvader/wow-using-pow/pkg/proof_of_work"
	"github.com/kirvader/wow-using-pow/pkg/storage"
)

type Server struct {
	listener         net.Listener
	storageSet       storage.StorageSet
	clock            clock.Clock
	challengeBuilder pow.POWChallengeBuilder
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
		challengeBuilder: &pow.HashcashBuilder{
			ZerosCount:        hashcashZerosCount,
			ChallengeDuration: hashcashChallengeDuration,
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
		if err = srv.processRequest(ctx, conn, conn.RemoteAddr().String(), msg); err != nil {
			log.Printf("process request error: %v", err)
			return
		}
	}
}

func (srv *Server) sendChallengeToClient(ctx context.Context, connWriter io.Writer, clientId string) error {
	currentTime := srv.clock.Now()
	connBufWriter := bufio.NewWriter(connWriter)

	challenge, randVal := srv.challengeBuilder.GenerateRandomChallenge(&currentTime, clientId)
	// to check in the future that server actually used this randVal
	err := srv.storageSet.InsertClientToken(ctx, clientId, randVal, *srv.challengeBuilder.GetChallengeDuration())
	if err != nil {
		return fmt.Errorf("inserting into storage set failed: %w", err)
	}

	return protocol.SendChallenge(connBufWriter, challenge)
}

func (srv *Server) checkClientChallengeSolution(ctx context.Context, clientId string, solution pow.POWChallenge) (bool, error) {
	if solution.GetResourse() != clientId {
		return false, errors.New("invalid resource")
	}

	// verify that challenge's rand value was generated on server
	clientToken, err := srv.storageSet.GetClientToken(ctx, solution.GetResourse())
	if err != nil {
		return false, fmt.Errorf("client token doesn't exist: %w", err)
	}
	if clientToken != solution.GetRand() {
		return false, errors.New("wrong client token")
	}

	// as a rand value is generated for each challenge - we want to make sure it won't be used to verify duplicates
	if err := srv.storageSet.Delete(ctx, solution.GetResourse()); err != nil {
		return false, fmt.Errorf("deleting from storage set failed: %w", err)
	}

	// solution shouldn't take more than setup challengeDuration
	if srv.clock.Now().Sub(*solution.GetDate()) > *srv.challengeBuilder.GetChallengeDuration() {
		return false, errors.New("challenge expired")
	}

	// verify according to POW puzzle
	valid, err := solution.Verify()
	if err != nil {
		return false, fmt.Errorf("verifying challenge solution failed: %w", err)
	}
	return valid, err
}

func (srv *Server) processRequest(ctx context.Context, connWriter io.Writer, clientId string, msg *protocol.Message) error {
	switch msg.MessageType {
	case protocol.ForceQuit:
		log.Print("client requests to close connection")
		return nil

	case protocol.ChallengeRequest:
		log.Printf("client %s requests challenge\n", clientId)
		return srv.sendChallengeToClient(ctx, connWriter, clientId)
	case protocol.ResourceRequest:
		powSolution := new(pow.Hashcash)
		if err := powSolution.FromJSON(msg.Payload); err != nil {
			return err
		}

		log.Printf("client solved challenge and requests resource. client: %s, payload %s\n", clientId, msg.Payload)

		valid, err := srv.checkClientChallengeSolution(ctx, clientId, powSolution)
		if err != nil {
			return fmt.Errorf("verifying challenge solution failed: %v", err)
		}
		if !valid {
			return errors.New("invalid solution")
		}

		log.Print("Solution verified. Sending a word of wisdom.")

		return protocol.SendResource(bufio.NewWriter(connWriter))
	default:
		return errors.New("invalid message type")
	}
}
