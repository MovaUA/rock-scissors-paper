package server

import (
	pb "github.com/movaua/rock-paper-scissors/pkg/rps"
)

// Play starts the game.
func (g *Game) Start(stream pb.Game_StartServer) error {
	// wait for the game is started
	<-g.started.Done()

	recvErr := make(chan error)
	sendErr := make(chan error)

	go func() {
		for {
			c, err := stream.Recv()
			if err != nil {
				recvErr <- err
				return
			}
			g.round.MakeChoise(c)
		}
	}()

	go func() {
		for {

		}
	}()

	select {
	case err := <-recvErr:
		return err
	case err := <-sendErr:
		return err
	case <-g.ctx.Done():
		return g.ctx.Err()
	}
}
