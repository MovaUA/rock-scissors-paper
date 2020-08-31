package server

import (
	"context"
	"time"

	pb "github.com/movaua/rock-paper-scissors/pkg/rps"
)

// game is the server API for the game.
type game struct {
	pb.UnimplementedGamerServer
	ctx          context.Context
	roundTimeout time.Duration
	started      context.Context // started context is done when the game is started.
	// start cancels started context,
	// which causes the game to go into "started" state.
	start                      func()
	players                    map[string]*pb.Player // key is player.Id
	authRequests               chan authRequest
	getPlayersRequests         chan getPlayersRequest
	notifyPlayerConnectedChans map[getPlayersRequest]chan *pb.Player
	unsubscribeGetPlayers      chan getPlayersRequest
}

// New returns new initialized game server.
func New(ctx context.Context, roundTimeout time.Duration) pb.GamerServer {
	started, start := context.WithCancel(context.Background())
	g := &game{
		ctx:                        ctx,
		roundTimeout:               roundTimeout,
		started:                    started,
		start:                      start,
		players:                    make(map[string]*pb.Player, 2),
		authRequests:               make(chan authRequest),
		getPlayersRequests:         make(chan getPlayersRequest),
		notifyPlayerConnectedChans: make(map[getPlayersRequest]chan *pb.Player, 2),
	}
	go g.handleRequests()
	return g
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

// isStarted returns true if the game is in "started" state.
func (g *game) isStarted() bool {
	return g.started.Err() != nil
}
