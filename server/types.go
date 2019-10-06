package main

import (
	"fmt"
	"math/rand"
)

// Card is a card in the game of blackjack.
type Card struct {
	Value int
}

// GetRandomCard returns a random card.
func GetRandomCard() Card {
	return Card{
		Value: rand.Intn(13),
	}
}

// GetValue returns the numerical value of this card.
func (c Card) GetValue() int {
	val := c.Value + 1
	if val > 10 {
		val = 10
	}
	return val
}

// GetStringRepresentation returns the card string for a given value.
func (c Card) GetStringRepresentation() string {
	return fmt.Sprintf("[%d]", c.GetValue())
}

// Hand is composed of an array of cards.
type Hand []Card

// GenerateNewHand is used to generate a new hand of two cards.
func GenerateNewHand() Hand {
	return []Card{GetRandomCard(), GetRandomCard()}
}

// GetTotalValue of a hand.
func (h Hand) GetTotalValue() int {
	val := 0
	for _, card := range h {
		val += card.GetValue()
	}
	return val
}

// IsOver checks if the hand is over the limit, aka busted.
func (h Hand) IsOver() bool {
	return h.GetTotalValue() > 21
}

// ReturnFirstCardString returns a string representation of the first card in the hand.
func (h Hand) ReturnFirstCardString() string {
	if len(h) == 0 {
		return ""
	}
	return h[0].GetStringRepresentation()
}

// GetStringRepresentation returns a string representation of a hand.
func (h Hand) GetStringRepresentation() string {
	if len(h) == 0 {
		return ""
	}
	str := h[0].GetStringRepresentation()
	for i := 1; i < len(h); i++ {
		str += " " + h[i].GetStringRepresentation()
	}

	// We also append the hands value as a helper to the player.
	return str + fmt.Sprintf(" (Value: %d)", h.GetTotalValue())
}
