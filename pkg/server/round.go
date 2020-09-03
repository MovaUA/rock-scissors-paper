package server

import (
	"context"
	"sort"
	"time"

	pb "github.com/movaua/rock-paper-scissors/pkg/rps"
)

// Round is an API of a single round of the game.
type Round struct {
	ctx       context.Context
	cancel    func()
	players   map[string]*pb.Player // key is player.Id
	choicesCh chan *pb.Choice
	choices   map[string]*pb.Choice // key is player.Id
	resultsCh chan []*pb.RoundResult
}

// NewRound return new started round of the game.
func NewRound(
	ctx context.Context,
	timeout time.Duration,
	players map[string]*pb.Player,
) *Round {
	ctx, cancel := context.WithTimeout(ctx, timeout)

	r := &Round{
		ctx:       ctx,
		cancel:    cancel,
		players:   players,
		choicesCh: make(chan *pb.Choice, len(players)),
		choices:   make(map[string]*pb.Choice, len(players)),
		resultsCh: make(chan []*pb.RoundResult, 1),
	}

	go r.start()

	return r
}

// MakeChoise accepts a player's choise.
func (r *Round) MakeChoise(c *pb.Choice) {
	r.choicesCh <- c
}

// Result returns a chan with round results.
func (r *Round) Result() <-chan []*pb.RoundResult {
	return r.resultsCh
}

// start waits for all players made their choises or round is timed out.
// Then it reports the result back to game.
func (r *Round) start() {
	r.handleChoises()
	r.cancel()
	r.resultsCh <- r.getResults()
	close(r.resultsCh)
}

// handleChoises returns when all players have made their choises
// or when timeout is expired.
func (r *Round) handleChoises() {
	for {
		select {

		// handle player's choice:
		case choise := <-r.choicesCh:
			// reject not connected players:
			if _, connected := r.players[choise.GetPlayerId()]; !connected {
				break
			}

			r.choices[choise.GetPlayerId()] = choise

			// complete the round when all players made their choises:
			if len(r.choices) == len(r.players) {
				return
			}

			// compelete the round when timeout expired or game server is closed:
		case <-r.ctx.Done():
			return
		}
	}
}

func (r *Round) getResults() []*pb.RoundResult {
	// calculate score
	results := r.scoreResults()

	// sort ascending by score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score < results[j].Score
	})

	// calculate status
	SetRoundStatuses(results)

	return results
}

// scoreResults calculates a round score for each player.
func (r *Round) scoreResults() []*pb.RoundResult {
	scores := make([]*pb.RoundResult, len(r.players))

	for _, player := range r.players {
		choice := r.choices[player.GetId()].GetChoice()
		score := int32(0)

		for _, otherPlayer := range r.players {
			if otherPlayer.GetId() == player.GetId() {
				continue
			}

			otherPlayerChoice := r.choices[otherPlayer.GetId()].GetChoice()
			status := GetStatus(choice, otherPlayerChoice)

			score += statusScores[status]
		}

		scores = append(scores,
			&pb.RoundResult{
				Player: player,
				Choice: choice,
				Score:  score,
			})
	}

	return scores
}
