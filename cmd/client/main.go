package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/kirvader/wow-using-pow/client"
)

const (
	// to limit client resources
	HashcashMaxIterationsAmount = "HASHCASH_MAX_ITERATIONS_AMOUNT"

	ServerHost = "SERVER_HOST"
	ServerPort = "SERVER_PORT"
)

func getHashcashMaxIterationsAmountFromEnv() (int32, error) {
	strVal, ok := os.LookupEnv(HashcashMaxIterationsAmount)
	if !ok {
		return 0, errors.New("HASHCASH_MAX_ITERATIONS_AMOUNT variable is not set")
	}
	val, err := strconv.Atoi(strVal)
	if err != nil {
		return 0, errors.New("HASHCASH_MAX_ITERATIONS_AMOUNT variable is invalid")
	}
	return int32(val), nil
}

func getServerAddressFromEnv() (string, error) {
	serverHost, ok := os.LookupEnv(ServerHost)
	if !ok {
		return "", errors.New("SERVER_HOST variable is not set")
	}
	stringServerPort, ok := os.LookupEnv(ServerPort)
	if !ok {
		return "", errors.New("SERVER_PORT variable is not set")
	}
	serverPort, err := strconv.Atoi(stringServerPort)
	if err != nil {
		return "", errors.New("SERVER_PORT variable is invalid")
	}
	return fmt.Sprintf("%s:%d", serverHost, serverPort), nil
}

func main() {
	log.Println("starting client...")
	ctx := context.Background()

	hashcashMaxIterationsAmount, err := getHashcashMaxIterationsAmountFromEnv()
	if err != nil {
		panic(err)
	}

	serverAddress, err := getServerAddressFromEnv()
	if err != nil {
		panic(err)
	}

	clientInstance, err := client.NewClient(serverAddress, hashcashMaxIterationsAmount)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := clientInstance.Close(); err != nil {
			panic(err)
		}
	}()

	for {
		log.Println("running client...")
		err = clientInstance.HandleConnection(ctx)
		if err != nil {
			log.Printf("client error: %v", err)
		}

		time.Sleep(10 * time.Second)
	}
}
