package server

import (
	pb "github.com/movaua/rock-paper-scissors/pkg/rps"
)

// Start starts the game.
func (g *Game) Start(stream pb.Game_StartServer) error {
	scoreCh := make(chan *pb.Score)
	disconnect := make(chan struct{})
	defer close(disconnect)

	g.startRequestsCh <- startRequest{
		scoreCh:    scoreCh,
		disconnect: disconnect,
	}

	// wait for the game is started
	<-g.started.Done()

	recvErr := make(chan error)
	sendErr := make(chan error)

	go func() {
		for {
			c, err := stream.Recv()
			if err != nil {
				recvErr <- err
				close(recvErr)
				recvErr = nil
				return
			}
			g.round.MakeChoise(c)
		}
	}()

	go func() {
		for {
			select {
			case s := <-scoreCh:
				err := stream.Send(s)
				if err != nil {
					sendErr <- err
					close(sendErr)
					sendErr <- nil
					return
				}
			}
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

type startRequest struct {
	scoreCh    chan<- *pb.Score
	disconnect <-chan struct{}
}

func (g *Game) handleStartRequests(r startRequest) {
	g.startRequests[r] = r.scoreCh
	go func() {
		<-r.disconnect
		g.cancelStartRequestsCh <- r
	}()
}
