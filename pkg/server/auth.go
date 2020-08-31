package server

import (
	"context"

	"github.com/google/uuid"
	pb "github.com/movaua/rock-paper-scissors/pkg/rps"
)

// Auth authenticates a player in the game.
// Request Player must have Name set (and Id is ignored).
// Response Player have the same Name as request Player and assigned Id.
func (g *game) Auth(ctx context.Context, p *pb.Player) (*pb.Player, error) {
	r := authRequest{player: p, res: make(chan authResponse)}
	g.authRequests <- r
	res := <-r.res
	return res.player, res.err
}

type authRequest struct {
	player *pb.Player
	res    chan authResponse
}

type authResponse struct {
	player *pb.Player
	err    error
}

func (g *game) handleAuth(r authRequest) {
	defer close(r.res)

	if r.player.GetName() == "" {
		r.res <- authResponse{err: errEmptyName}
		return
	}
	if player := g.findPlayerByName(r.player.GetName()); player != nil {
		r.res <- authResponse{err: errConnected(r.player.GetName())}
		return
	}
	if g.started {
		r.res <- authResponse{err: errStarted}
		return
	}

	player := &pb.Player{
		Id:   uuid.New().String(),
		Name: r.player.GetName(),
	}
	g.players[player.Id] = player
	r.res <- authResponse{player: player}

	for _, notifyPlayerConnected := range g.notifyPlayerConnectedChans {
		notifyPlayerConnected <- player
	}
}

func (g *game) findPlayerByName(name string) *pb.Player {
	for _, player := range g.players {
		if player.GetName() == name {
			return player
		}
	}
	return nil
}
