package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"client/rpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// The wait group is used to block the main goroutine.
var wg sync.WaitGroup

// Reader is used to read from Stdin.
var Reader = bufio.NewReader(os.Stdin)

// GetUserInput from the user.
func GetUserInput() string {
	str, _ := Reader.ReadString('\n')
	str = strings.TrimSuffix(str, "\n")
	return str
}

func main() {
	// Establish a connection with the gRPC server (on port 9212).
	connection, err := grpc.Dial("0.0.0.0:9212", grpc.WithInsecure())
	if err != nil {
		fmt.Printf("failed to dial grpc server: %s\n", err.Error())
	}
	defer connection.Close()

	// The client will be used to send RPC requests.
	client := rpc.NewBlackjackClient(connection)

	// Print the introduction messages to the user.
	PrintIntroduction()

	// Get the users chosen username and open a stream.
	username, stream := GetChosenUsername(client)

	// Subscribe to the incoming server stream.
	go SubscribeToStream(username, client, stream)
	wg.Add(1)

	// Block the main goroutine.
	wg.Wait()
}

// PrintIntroduction prints an introduction string to the console.
func PrintIntroduction() {
	fmt.Println("Welcome to Realtime N-Player Blackjack.")
	fmt.Println("Created by Morgan Gallant for his KP Fellowship Application.")
	fmt.Println("")
	fmt.Println("This blackjack game isn't meant to be a great blackjack game. And that's ok.")
	fmt.Println("Rather, it attempts to showcase my experience with server communications in the context of blackjack.")
	fmt.Println("")
	fmt.Println("One note, please type 'leave' when it is your turn to leave the game (rather than CTRL-C).")
	fmt.Println("")
	fmt.Println("OK, now that all that is done with, let's get started!")
	fmt.Println("")
}

// GetChosenUsername prompts the user to enter a username. The function also does server checking
// to ensure that the username is valid.
func GetChosenUsername(client rpc.BlackjackClient) (string, rpc.Blackjack_SubscribeClient) {
	var username string
	var stream rpc.Blackjack_SubscribeClient
	var err error

	for username == "" {
		fmt.Println("Which username do you want to use?")
		in := GetUserInput()
		stream, err = client.Subscribe(context.Background(), &rpc.SubscribeRequest{Username: in})
		if err != nil {
			// If the server returns an AlreadyExists, we can ask the user to retry. Otherwise, panic.
			resp, _ := status.FromError(err)
			if resp.Code() == codes.AlreadyExists {
				fmt.Println("That username is already in use! Try again.")
				fmt.Println("")
				continue
			} else {
				panic(err)
			}
		}
		fmt.Println("")
		username = in
	}

	return username, stream
}
