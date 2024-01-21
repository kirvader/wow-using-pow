package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/kirvader/wow-using-pow/server"
)

func main() {
	fmt.Println("start server")

	serverConfig := server.LoadConfig()

	ctx := context.Background()
	ctx = context.WithValue(ctx, server.ConfigContextKey{}, serverConfig)

	rand.Seed(time.Now().UnixNano())

	serverAddress := fmt.Sprintf("%s:%d", serverConfig.ServerHost, serverConfig.ServerPort)
	err := server.Run(ctx, serverAddress)
	if err != nil {
		fmt.Println("server Run() returned error:", err)
	}
}
