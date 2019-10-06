package main

import (
	"fmt"
	"server/rpc"
	"time"
)

// Gamestate is an enumeration over all the different gamestates.
type Gamestate int

const (
	// StateInitial represents the gamestate where the dealer is dealing the cards.
	StateInitial Gamestate = 0
	// StatePlayersPlaying represents the gamestate where players are playing.
	StatePlayersPlaying Gamestate = 1
	// StateDealerPlaying represents the gamestate where the dealer plays.
	StateDealerPlaying Gamestate = 2
	// StateGameOver ends the game and determines winners/losers.
	StateGameOver Gamestate = 3
)

// PlayBlackjack coordinates the blackjack game loop.
func PlayBlackjack(s *BlackjackService) {
	for {
	GAME_START:
		// Perform setup for a round.
		numPlayers := len(s.Players)

		// We don't initiate any round of blackjack if there are no players in the room.
		if numPlayers == 0 {
			time.Sleep(time.Second)
			continue
		}

		// Start the game of blackjack.
		DispatchMessage(s, "Starting a new game of Blackjack with %d player(s).", numPlayers)

		// Generate the dealers hand.
		dealer := GenerateNewHand()
		DispatchMessage(s, "Dealers hand is dealt, showing a %s.", dealer.ReturnFirstCardString())

		// Check if the dealer got blackjack.
		if dealer.GetTotalValue() == 21 {
			DispatchMessage(s, "Dealer got blackjack, everyone lost!")
			goto GAME_START
		}

		// Deal the players hands.
		DealPlayerHands(s)

		// For each player, get them to play their hand.
		var i int
		for i < numPlayers {
			DispatchMessage(s, "It is %s's turn, they have a: %s", s.Players[i].Username, s.Players[i].CurrentHand.GetStringRepresentation())

		GET_MESSAGE:
			// Tell the player to perform an action.
			NotifyPlayerToDoAction(s, i)

		WAIT_FOR_MESSAGE:
			// Wait for an incoming message.
			action := <-s.IncomingMessages

			// If any player sends a leave message, remove them and restart the game.
			if action.GetType() == rpc.ClientAction_LEAVE {
				s.DeletePlayerWithUsername(action.GetUsername())
				DispatchMessage(s, "%s left the game, restarting the round!", action.GetUsername())
				goto GAME_START
			}

			// We need to ensure we recieve the message from the proper player.
			if action.GetUsername() != s.Players[i].Username {
				goto WAIT_FOR_MESSAGE
			}

			// Depending on the type, we will hit or stand.
			switch action.GetType() {
			case rpc.ClientAction_HIT:
				s.Players[i].Hit()
				DispatchMessage(s, "%s just hit! They now have: %s", s.Players[i].Username, s.Players[i].CurrentHand.GetStringRepresentation())
				if s.Players[i].CurrentHand.IsOver() {
					DispatchMessage(s, "%s busted!", s.Players[i].Username)
					break
				} else {
					goto GET_MESSAGE
				}
			case rpc.ClientAction_STAND:
				DispatchMessage(s, "%s just stood on: %s", s.Players[i].Username, s.Players[i].CurrentHand.GetStringRepresentation())
				break
			}

			// Once we have reached here, the players turn is over and we can move onto the next player.
			i++
		}

		// The dealer must play now.
		DispatchMessage(s, "Everyone is finished playing, it is now the dealers turn.")
		DispatchMessage(s, "The dealer has: %s", dealer.GetStringRepresentation())
		for dealer.GetTotalValue() < 17 && !dealer.IsOver() {
			dealer = append(dealer, GetRandomCard())
			DispatchMessage(s, "The dealer took another card, they now have: %s", dealer.GetStringRepresentation())
		}
		DispatchMessage(s, "The dealer ended up with: %s", dealer.GetStringRepresentation())

		// If the dealer busted, everyone wins.
		if dealer.IsOver() {
			DispatchMessage(s, "The dealer busted, everyone wins!")
			goto GAME_START
		}

		// We now must check who won.
		for _, player := range s.Players {
			// Ties are counted as wins in our simplifed version of blackjack.
			if player.CurrentHand.GetTotalValue() >= dealer.GetTotalValue() && !player.CurrentHand.IsOver() {
				DispatchMessage(s, "%s won!", player.Username)
			} else {
				DispatchMessage(s, "%s lost.", player.Username)
			}
		}

		DispatchMessage(s, "The game is now over, restarting...")
	}
}

// DealPlayerHands deals the hand for all players.
func DealPlayerHands(s *BlackjackService) {
	for i := 0; i < len(s.Players); i++ {
		s.Players[i].CurrentHand = GenerateNewHand()
		DispatchMessage(s, "%s got dealt %s.", s.Players[i].Username, s.Players[i].CurrentHand.GetStringRepresentation())
	}
}

// NotifyPlayerToDoAction sends a message to the player asking them to perform an action.
func NotifyPlayerToDoAction(s *BlackjackService, idx int) {
	s.Players[idx].MessageQueue <- rpc.ServerResponse{Type: rpc.ServerResponse_NEEDS_INPUT}
}

// Hit adds a random card to a players hand.
func (p *Player) Hit() {
	p.CurrentHand = append(p.CurrentHand, GetRandomCard())
}

// DispatchMessage to all current game players.
func DispatchMessage(s *BlackjackService, msg string, args ...interface{}) {
	for _, player := range s.Players {
		player.MessageQueue <- rpc.ServerResponse{
			Type:           rpc.ServerResponse_MESSAGE,
			OptionalString: fmt.Sprintf(msg, args...),
		}
	}
	// We introduce an artifical delay to give the user time to read the text.
	time.Sleep(time.Millisecond * 600)
}
