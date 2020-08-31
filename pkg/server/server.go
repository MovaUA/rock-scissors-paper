package server

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	pb "github.com/movaua/rock-paper-scissors/pkg/rps"
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

// NewGame returns new initialized game server.
func NewGame(ctx context.Context, roundTimeout time.Duration) pb.GamerServer {
	g := &game{
		ctx:          ctx,
		roundTimeout: roundTimeout,
		authRequests: make(chan authRequest),
		players:      make(map[string]*pb.Player, 2),
	}
	go g.processRequests()
	return g
}

// game is the server API for the game.
type game struct {
	pb.UnimplementedGamerServer
	ctx          context.Context
	roundTimeout time.Duration
	started      bool
	authRequests chan authRequest
	players      map[string]*pb.Player // key is player.Id
}

// Auth authenticates a player in the game.
// Request Player must have Name set (and Id is ignored).
// Response Player have the same Name as request Player and assigned Id.
func (g *game) Auth(ctx context.Context, r *pb.Player) (*pb.Player, error) {
	request := authRequest{player: r, response: make(chan authResponse)}
	g.authRequests <- request
	response := <-request.response
	return response.player, response.err
}

// Play starts the game.
func (g *game) Play(s pb.Gamer_PlayServer) error {
	return nil
}

type authRequest struct {
	player   *pb.Player
	response chan authResponse
}
type authResponse struct {
	player *pb.Player
	err    error
}

func (g *game) processRequests() {
	for {
		select {
		case r := <-g.authRequests:
			if r.player.GetName() == "" {
				r.response <- authResponse{err: errEmptyName}
				close(r.response)
				break
			}
			if player := g.findPlayerByName(r.player.GetName()); player != nil {
				r.response <- authResponse{err: errConnected(r.player.GetName())}
				close(r.response)
				break
			}
			if g.started {
				r.response <- authResponse{err: errStarted}
				close(r.response)
				break
			}

			player := &pb.Player{
				Id:   uuid.New().String(),
				Name: r.player.GetName(),
			}
			g.players[player.Id] = player
			r.response <- authResponse{player: player}
			close(r.response)

		case <-g.ctx.Done():
		}
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
