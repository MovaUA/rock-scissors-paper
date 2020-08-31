package server

import (
	pb "github.com/movaua/rock-paper-scissors/pkg/rps"
)

// Play starts the game.
func (g *game) Play(s pb.Gamer_PlayServer) error {
	return nil
}
