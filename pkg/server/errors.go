package server

import (
	"fmt"
)

var (
	// errStarted happens when an operation is illegal when the game is already started.
	errStarted = fmt.Errorf("the game is already started")

	// errEmptyName happens when player with empty name tries to connect.
	errEmptyName = fmt.Errorf("plaery name is empty")
)

// errConnected happens when a connected player tries to connect again.
type errConnected string

func (e errConnected) Error() string {
	return fmt.Sprintf("a player with name %q is already connected to the game", string(e))
}
