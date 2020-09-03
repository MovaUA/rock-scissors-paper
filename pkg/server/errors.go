package server

import (
	"fmt"
)

var (
	// errGameStarted happens when an operation is illegal when the game is already started.
	errGameStarted = fmt.Errorf("the game is already started")

	// errPlayerEmptyName happens when player with empty name tries to connect.
	errPlayerEmptyName = fmt.Errorf("player name is empty")
)

// errPlayerConnected happens when a connected player tries to connect again.
type errPlayerConnected string

func (e errPlayerConnected) Error() string {
	return fmt.Sprintf("a player with name %q is already connected to the game", string(e))
}
