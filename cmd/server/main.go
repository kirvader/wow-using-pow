package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/kirvader/wow-using-pow/server"
)

const (
	SERVER_HOST = "SERVER_HOST"
	SERVER_PORT = "SERVER_PORT"

	REDIS_HOST = "REDIS_HOST"
	REDIS_PORT = "REDIS_PORT"

	HASHCASH_ZEROS_COUNT        = "HASHCASH_ZEROS_COUNT"
	HASHCASH_CHALLENGE_LIFETIME = "HASHCASH_CHALLENGE_LIFETIME"
)

func getHashcashZerosCountValueFromEnv() (int32, error) {
	stringHashcashZerosCount, ok := os.LookupEnv(HASHCASH_ZEROS_COUNT)
	if !ok {
		return 0, errors.New("HASHCASH_ZEROS_COUNT variable is not set")
	}
	hashcashZerosCount, err := strconv.Atoi(stringHashcashZerosCount)
	if err != nil {
		return 0, errors.New("HASHCASH_ZEROS_COUNT variable is invalid")
	}
	return int32(hashcashZerosCount), nil
}

func getHashcashChallengeLifetimeValueFromEnv() (*time.Duration, error) {
	stringHashcashChallengeDuration, ok := os.LookupEnv(HASHCASH_CHALLENGE_LIFETIME)
	if !ok {
		return nil, errors.New("HASHCASH_CHALLENGE_LIFETIME variable is not set")
	}
	hashcashChallengeDuration, err := time.ParseDuration(stringHashcashChallengeDuration)
	if err != nil {
		return nil, errors.New("HASHCASH_CHALLENGE_LIFETIME variable is invalid")
	}
	return &hashcashChallengeDuration, nil
}

func getServerAddressFromEnv() (string, error) {
	serverHost, ok := os.LookupEnv(SERVER_HOST)
	if !ok {
		return "", errors.New("SERVER_HOST variable is not set")
	}
	stringServerPort, ok := os.LookupEnv(SERVER_PORT)
	if !ok {
		return "", errors.New("SERVER_PORT variable is not set")
	}
	serverPort, err := strconv.Atoi(stringServerPort)
	if err != nil {
		return "", errors.New("SERVER_PORT variable is invalid")
	}
	return fmt.Sprintf("%s:%d", serverHost, serverPort), nil
}

func getRedisAddressFromEnv() (string, error) {
	redisHost, ok := os.LookupEnv(REDIS_HOST)
	if !ok {
		panic(errors.New("REDIS_HOST variable is not set"))
	}
	stringRedisPort, ok := os.LookupEnv(REDIS_PORT)
	if !ok {
		panic(errors.New("REDIS_PORT variable is not set"))
	}
	redisPort, err := strconv.Atoi(stringRedisPort)
	if err != nil {
		panic(errors.New("REDIS_PORT variable is invalid"))
	}
	return fmt.Sprintf("%s:%d", redisHost, redisPort), nil
}

func main() {
	log.Println("starting server...")

	ctx := context.Background()
	hashcashZerosCount, err := getHashcashZerosCountValueFromEnv()
	if err != nil {
		panic(err)
	}

	hashcashChallengeLifetime, err := getHashcashChallengeLifetimeValueFromEnv()
	if err != nil {
		panic(err)
	}

	serverAddress, err := getServerAddressFromEnv()
	if err != nil {
		panic(err)
	}
	redisAddress, err := getRedisAddressFromEnv()
	if err != nil {
		panic(err)
	}
	server, err := server.NewServer(ctx, serverAddress, redisAddress, hashcashZerosCount, hashcashChallengeLifetime)
	if err != nil {
		panic(fmt.Errorf("error when creating a server: %v", err))
	}

	defer func() {
		if err := server.Close(ctx); err != nil {
			panic(err)
		}
	}()

	if err = server.Run(ctx); err != nil {
		log.Printf("server error: %v", err)
	}
}
