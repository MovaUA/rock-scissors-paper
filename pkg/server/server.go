package server

import (
	"context"
	"time"

	pb "github.com/movaua/rock-paper-scissors/pkg/rps"
)

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
