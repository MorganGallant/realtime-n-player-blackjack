package main

import (
	"client/rpc"
	"context"
	"fmt"
	"os"
	"strings"
)

// SubscribeToStream subscribes to stream messages from the Blackjack server.
func SubscribeToStream(username string, client rpc.BlackjackClient, stream rpc.Blackjack_SubscribeClient) {
	for {
		var msg rpc.ServerResponse
		// Blocks until message is recieved.
		err := stream.RecvMsg(&msg)
		if err != nil {
			panic(err)
		}

		// Depending on the type of message, we execute different responses.
		switch msg.GetType() {
		case rpc.ServerResponse_MESSAGE:
			PrintServerMessage(msg.GetOptionalString())
			break
		case rpc.ServerResponse_NEEDS_INPUT:
			GetUserBlackjackAction(username, client)
			break
		}
	}
}

// PrintServerMessage is used to print an incoming server message to the console.
func PrintServerMessage(msg string) {
	fmt.Printf(">> %s\n", msg)
}

// GetUserBlackjackAction prompts the user to enter an action (either hit, stand or leave).
func GetUserBlackjackAction(username string, client rpc.BlackjackClient) {
	for {
		fmt.Println("")
		fmt.Println("What would you like to do? ('hit', 'stand' or 'leave')")
		in := strings.ToUpper(GetUserInput())
		fmt.Println("")

		// We use the proto-generate map to get an action from the string.
		if action, ok := rpc.ClientAction_ClientActionType_value[in]; ok {
			clientAction := rpc.ClientAction_ClientActionType(action)
			// Send the users choice to the server.
			_, err := client.PerformAction(context.Background(), &rpc.ClientAction{
				Username: username,
				Type:     clientAction,
			})
			// If there was an error, panic.
			if err != nil {
				panic(err)
			}

			// If the action was leave, exit the application.
			if clientAction == rpc.ClientAction_LEAVE {
				fmt.Println("Thanks for playing!")
				os.Exit(0)
			}

			// Since we have done a proper request, we break the loop.
			break
		} else {
			// If the user entered invalid input, show an error message and allow them to restart.
			fmt.Println("Invalid input! Try again.")
			fmt.Println("")
		}
	}
}
