package server

import (
	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/movaua/rock-paper-scissors/pkg/rps"
)

// GetPlayers streams all connected players
// and then every new one.
func (g *game) GetPlayers(e *empty.Empty, stream pb.Gamer_GetPlayersServer) error {
	playerCh := make(chan *pb.Player, 1)
	errCh := make(chan error)

	go func() {
		for {
			select {
			case p := <-playerCh:
				if err := stream.Send(p); err != nil {
					errCh <- err
					return
				}
			case <-g.ctx.Done():
				errCh <- g.ctx.Err()
				return
			}
		}
	}()

	r := getPlayersRequest{playerCh: playerCh}

	g.getPlayersRequests <- r

	err := <-errCh

	g.unsubscribeGetPlayers <- r

	return err
}

type getPlayersRequest struct {
	playerCh chan *pb.Player
}

func (g *game) handleGetPlayers(r getPlayersRequest) {
	for _, p := range g.players {
		r.playerCh <- p
	}
	g.notifyPlayerConnectedChans[r] = r.playerCh
}
