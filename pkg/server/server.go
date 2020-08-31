package server

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	pb "github.com/movaua/rock-paper-scissors/pkg/rps"
)

var (
	// ErrStarted happens when an operation is illegal when the game is started.
	ErrStarted = fmt.Errorf("the game is already started")
)

// Game is the server API for the game.
type Game struct {
	pb.UnimplementedGamerServer
	ctx          context.Context
	roundTimeout time.Duration
	started      bool
	authRequests chan authRequest
	players      []*pb.Player
}

// NewGame returns new initialized game.
func NewGame(ctx context.Context, roundTimeout time.Duration) *Game {
	g := &Game{
		ctx:          ctx,
		roundTimeout: roundTimeout,
		authRequests: make(chan authRequest),
	}
	go g.processRequests()
	return g
}

// Auth authenticates a player in the game.
func (g *Game) Auth(ctx context.Context, r *pb.AuthRequest) (*pb.Player, error) {
	rq := authRequest{rq: r, rs: make(chan authResponse)}
	g.authRequests <- rq
	rs := <-rq.rs
	return rs.player, rs.err
}

// Play starts the game.
func (g *Game) Play(s pb.Gamer_PlayServer) error {
	return nil
}

type authRequest struct {
	rq *pb.AuthRequest
	rs chan authResponse
}
type authResponse struct {
	player *pb.Player
	err    error
}

func (g *Game) processRequests() {
	for {
		select {
		case r := <-g.authRequests:
			if g.started {
				r.rs <- authResponse{err: ErrStarted}
				close(r.rs)
				break
			}
			if player := g.findPlayerByName(r.rq.GetName()); player != nil {
				r.rs <- authResponse{err: ErrAlreadyConnected(r.rq.GetName())}
				close(r.rs)
				break
			}

			player := &pb.Player{
				Id:   uuid.New().String(),
				Name: r.rq.GetName(),
			}
			g.players = append(g.players, player)
			r.rs <- authResponse{player: player}
			close(r.rs)

		case <-g.ctx.Done():
		}
	}
}

func (g *Game) findPlayerByName(name string) *pb.Player {
	for _, player := range g.players {
		if player.GetName() == name {
			return player
		}
	}
	return nil
}
func (g *Game) findPlayerByID(id string) *pb.Player {
	for _, player := range g.players {
		if player.GetId() == id {
			return player
		}
	}
	return nil
}

// ErrAlreadyConnected happens when a connected player tries to connect again.
type ErrAlreadyConnected string

func (e ErrAlreadyConnected) Error() string {
	return fmt.Sprintf("a player with name %q is already connected to the game", string(e))
}
