package main

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"server/rpc"

	"google.golang.org/grpc"
)

// Used to pause execution of the main goroutine.
var wg sync.WaitGroup

func main() {
	// Seed the random number generator.
	rand.Seed(time.Now().UTC().UnixNano())

	// Start listening on :9212.
	listener, err := net.Listen("tcp", ":9212")
	if err != nil {
		fmt.Printf("failed to start listener: %s\n", err.Error())
	}

	// Start the gRPC server and register the blackjack service.
	server := grpc.NewServer()
	service := &BlackjackService{
		Players:          []Player{},
		IncomingMessages: make(chan rpc.ClientAction, 100),
	}
	rpc.RegisterBlackjackServer(server, service)

	// We start the blackjack service in its own goroutine seperate from main.
	go func() {
		if err := server.Serve(listener); err != nil {
			fmt.Printf("gRPC server failed: %s\n", err.Error())
			wg.Done()
		}
	}()
	wg.Add(1)

	// The blackjack game is managed in its own goroutine.
	go PlayBlackjack(service)
	wg.Add(1)

	// Block the main goroutine.
	wg.Wait()
}
