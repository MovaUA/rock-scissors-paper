package server

import (
	"context"
	"time"

	pb "github.com/movaua/rock-paper-scissors/pkg/rps"
)

// Game is the server API for the Game.
// It implements pb.GameServer.
type Game struct {
	pb.UnimplementedGameServer
	ctx          context.Context
	roundTimeout time.Duration
	started      context.Context // started context is done when the game is started.
	// start cancels started context,
	// which causes the game to go into "started" state.
	start                      func()
	players                    map[string]*pb.Player // key is player.Id
	connectRequests            chan connectRequest
	playersRequests            chan playersRequest
	notifyPlayerConnectedChans map[playersRequest]chan *pb.Player
	unsubscribePlayers         chan playersRequest
	// round is the current round of the game when it is started.
	// round is nil until the game is started.ß
	round *Round
}

// NewGame returns new initialized game server.
func NewGame(ctx context.Context, roundTimeout time.Duration) *Game {
	started, start := context.WithCancel(context.Background())

	g := &Game{
		ctx:                        ctx,
		roundTimeout:               roundTimeout,
		started:                    started,
		players:                    make(map[string]*pb.Player, 2),
		connectRequests:            make(chan connectRequest),
		playersRequests:            make(chan playersRequest),
		notifyPlayerConnectedChans: make(map[playersRequest]chan *pb.Player, 2),
		unsubscribePlayers:         make(chan playersRequest),
	}

	g.start = func() {
		g.round = NewRound(g.ctx, g.roundTimeout, g.players)
		start()
	}

	go g.handleRequests()

	return g
}

func (g *Game) handleRequests() {
	for {
		select {
		case r := <-g.connectRequests:
			g.handleConnect(r)

		case r := <-g.playersRequests:
			g.handlePlayers(r)

		case r := <-g.unsubscribePlayers:
			delete(g.notifyPlayerConnectedChans, r)

		case <-g.ctx.Done():
		}
	}
}

// isStarted returns true if the game is in "started" state.
func (g *Game) isStarted() bool {
	return g.started.Err() != nil
}
