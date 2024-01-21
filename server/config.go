package server

import (
	"errors"
	"os"
	"strconv"
	"time"
)

const (
	// env variables
	SERVER_HOST                 = "SERVER_HOST"
	SERVER_PORT                 = "SERVER_PORT"
	HASHCASH_ZEROS_COUNT        = "HASHCASH_ZEROS_COUNT"
	HASHCASH_CHALLENGE_LIFETIME = "HASHCASH_CHALLENGE_LIFETIME"
)

type ConfigContextKey struct{}

type Config struct {
	ServerHost                string
	ServerPort                int32
	HashcashZerosCount        int32
	HashcashChallengeDuration time.Duration
}

func LoadConfig() *Config {
	serverHost, ok := os.LookupEnv(SERVER_HOST)
	if !ok {
		panic(errors.New("SERVER_HOST variable is not set"))
	}
	stringServerPort, ok := os.LookupEnv(SERVER_PORT)
	if !ok {
		panic(errors.New("SERVER_PORT variable is not set"))
	}
	serverPort, err := strconv.Atoi(stringServerPort)
	if err != nil {
		panic(errors.New("SERVER_PORT variable is invalid"))
	}

	stringHashcashZerosCount, ok := os.LookupEnv(HASHCASH_ZEROS_COUNT)
	if !ok {
		panic(errors.New("HASHCASH_ZEROS_COUNT variable is not set"))
	}
	hashcashZerosCount, err := strconv.Atoi(stringHashcashZerosCount)
	if err != nil {
		panic(errors.New("HASHCASH_ZEROS_COUNT variable is invalid"))
	}

	stringHashcashChallengeDuration, ok := os.LookupEnv(HASHCASH_CHALLENGE_LIFETIME)
	if !ok {
		panic(errors.New("HASHCASH_CHALLENGE_LIFETIME variable is not set"))
	}
	hashcashChallengeDuration, err := time.ParseDuration(stringHashcashChallengeDuration)
	if err != nil {
		panic(errors.New("HASHCASH_CHALLENGE_LIFETIME variable is invalid"))
	}

	return &Config{
		ServerHost:                serverHost,
		ServerPort:                int32(serverPort),
		HashcashZerosCount:        int32(hashcashZerosCount),
		HashcashChallengeDuration: hashcashChallengeDuration,
	}
}
