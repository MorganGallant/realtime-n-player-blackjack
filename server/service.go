package main

import (
	"context"
	"errors"
	"fmt"
	"server/rpc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Player stores information about a current blackjack player.
type Player struct {
	Username     string
	MessageQueue chan rpc.ServerResponse // Used to send messages to the player.
	CurrentHand  Hand
}

// BlackjackService handles RPC requests to the Blackjack Service.
type BlackjackService struct {
	Players          []Player
	IncomingMessages chan rpc.ClientAction // All incoming messages are fed into this queue.
}

// GetIndexOfPlayerWithUsername returns the index of a player with a given username.
func (s *BlackjackService) GetIndexOfPlayerWithUsername(username string) (int, error) {
	for i, player := range s.Players {
		if player.Username == username {
			return i, nil
		}
	}
	return 0, errors.New("player doesn't exist")
}

// HasPlayerWithUsername checks if a player with a given username exists.
func (s *BlackjackService) HasPlayerWithUsername(username string) bool {
	_, err := s.GetIndexOfPlayerWithUsername(username)
	return err == nil
}

// DeletePlayerWithUsername removes a player from the players array with a given username.
func (s *BlackjackService) DeletePlayerWithUsername(username string) {
	idx, err := s.GetIndexOfPlayerWithUsername(username)
	if err != nil {
		// The user doesn't exist, no need to delete anything.
		return
	}

	// The new players array contains every player except for the player at index idx.
	s.Players = append(s.Players[:idx], s.Players[idx+1:]...)
}

// SimulatePlayerLeaveRequest simulates to the blackjack goroutine that a player
// has just left. This is used when the player uses CTRL-C to exit the game
// instead of properly leaving, and fixes the problem of the blackjack being
// blocked waiting for user input for a user that doesn't exist.
func (s *BlackjackService) SimulatePlayerLeaveRequest(username string) {
	s.IncomingMessages <- rpc.ClientAction{
		Type:     rpc.ClientAction_LEAVE,
		Username: username,
	}
}

// Subscribe allows a client to subscribe to the game of blackjack, recieving a stream of updates.
// Each update contains either a string message which is to be printed and shown to the user, or an indicator
// that the client should ask the player for their choice of action.
func (s *BlackjackService) Subscribe(req *rpc.SubscribeRequest, stream rpc.Blackjack_SubscribeServer) error {
	// Check for duplicates.
	if s.HasPlayerWithUsername(req.GetUsername()) {
		return status.Error(codes.AlreadyExists, "username already exists")
	}

	// Create the new player.
	s.Players = append(s.Players, Player{
		Username:     req.GetUsername(),
		MessageQueue: make(chan rpc.ServerResponse, 100),
	})

	fmt.Println("Registered Player:", req.GetUsername())

	// Consume messages from the players message queue and send them to the client.
	for {
		idx, err := s.GetIndexOfPlayerWithUsername(req.GetUsername())
		if err != nil {
			return err
		}

		// This blocks until a message is recieved in the channel.
		msg := <-s.Players[idx].MessageQueue
		if err := stream.SendMsg(&msg); err != nil {
			status, _ := status.FromError(err)
			// If the status is an Unavailible, it means the client disconnected.
			// When this happens we have to do two things:
			// 1. Remove the player from the Players array.
			// 2. Notify the Blackjack Game that the player left.
			if status.Code() == codes.Unavailable {
				s.DeletePlayerWithUsername(req.GetUsername())
				s.SimulatePlayerLeaveRequest(req.GetUsername())
				fmt.Println("Unregistered Player:", req.GetUsername())
				return nil
			}

			// Otherwise, we have to report the error.
			fmt.Printf("Unhandled message exception: %s\n", err.Error())
			return err
		}
	}
}

// PerformAction is called by the client when requested to do so. It allows the client to pass a users
// choice of action to the server, such as a blackjack HIT or STAND.
func (s *BlackjackService) PerformAction(ctx context.Context, req *rpc.ClientAction) (*rpc.Empty, error) {
	s.IncomingMessages <- *req
	return &rpc.Empty{}, nil
}
