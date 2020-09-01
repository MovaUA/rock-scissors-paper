package server

import (
	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/movaua/rock-paper-scissors/pkg/rps"
)

// Players streams all connected players,
// and all players which connect later.
func (g *game) Players(_ *empty.Empty, stream pb.Game_PlayersServer) error {
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

	r := playersRequest{playerCh: playerCh}

	g.playersRequests <- r

	err := <-errCh

	g.unsubscribePlayers <- r

	return err
}

type playersRequest struct {
	playerCh chan *pb.Player
}

func (g *game) handlePlayers(r playersRequest) {
	for _, p := range g.players {
		r.playerCh <- p
	}
	g.notifyPlayerConnectedChans[r] = r.playerCh
}
