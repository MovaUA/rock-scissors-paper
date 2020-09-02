package server

import (
	"context"

	"github.com/google/uuid"
	pb "github.com/movaua/rock-paper-scissors/pkg/rps"
)

// Connect connects a player to the game.
// Request Player must have Name set (Id is ignored).
// Response Player is assigned Id.
func (g *Game) Connect(ctx context.Context, p *pb.Player) (*pb.Player, error) {
	r := connectRequest{player: p, res: make(chan connectResponse)}
	g.connectRequests <- r
	res := <-r.res
	return res.player, res.err
}

type connectRequest struct {
	player *pb.Player
	res    chan connectResponse
}

type connectResponse struct {
	player *pb.Player
	err    error
}

func (g *Game) handleConnect(r connectRequest) {
	defer close(r.res)

	if r.player.GetName() == "" {
		r.res <- connectResponse{err: errEmptyName}
		return
	}
	if _, ok := g.playerNames[r.player.GetName()]; ok {
		r.res <- connectResponse{err: errConnected(r.player.GetName())}
		return
	}
	if g.isStarted() {
		r.res <- connectResponse{err: errStarted}
		return
	}

	player := &pb.Player{
		Id:   uuid.New().String(),
		Name: r.player.GetName(),
	}
	g.players[player.Id] = player
	g.playerNames[player.Name] = struct{}{}
	r.res <- connectResponse{player: player}

	for _, notifyPlayerConnected := range g.notifyPlayerConnectedChans {
		notifyPlayerConnected <- player
	}

	if len(g.players) > 1 {
		g.start()
	}
}
