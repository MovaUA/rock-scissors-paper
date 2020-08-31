package server

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
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

// New returns new initialized game server.
func New(ctx context.Context, roundTimeout time.Duration) pb.GamerServer {
	g := &game{
		ctx:                        ctx,
		roundTimeout:               roundTimeout,
		players:                    make(map[string]*pb.Player, 2),
		authRequests:               make(chan authRequest),
		getPlayersRequests:         make(chan getPlayersRequest),
		notifyPlayerConnectedChans: make(map[getPlayersRequest]chan *pb.Player, 2),
	}
	go g.handleRequests()
	return g
}

// game is the server API for the game.
type game struct {
	pb.UnimplementedGamerServer
	ctx                        context.Context
	roundTimeout               time.Duration
	started                    bool
	players                    map[string]*pb.Player // key is player.Id
	authRequests               chan authRequest
	getPlayersRequests         chan getPlayersRequest
	notifyPlayerConnectedChans map[getPlayersRequest]chan *pb.Player
	unsubscribeGetPlayers      chan getPlayersRequest
}

func (g *game) handleRequests() {
	for {
		select {
		case r := <-g.authRequests:
			g.handleAuth(r)

		case r := <-g.getPlayersRequests:
			g.handleGetPlayers(r)

		case r := <-g.unsubscribeGetPlayers:
			delete(g.notifyPlayerConnectedChans, r)

		case <-g.ctx.Done():
		}
	}
}

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

// Play starts the game.
func (g *game) Play(s pb.Gamer_PlayServer) error {
	return nil
}
