package server

import (
	"context"
	"sort"
	"time"

	pb "github.com/movaua/rock-paper-scissors/pkg/rps"
)

// Round is an API of a single round of the game.
type Round struct {
	parentCtx context.Context
	ctx       context.Context
	timeout   time.Duration
	players   map[string]*pb.Player // key is player.Id
	choicesCh chan *pb.Choice
	choices   map[string]*pb.Choice // key is player.Id
	results   chan []*pb.RoundResult
}

// NewRound return new started round of the game.
func NewRound(
	ctx context.Context,
	timeout time.Duration,
	players map[string]*pb.Player,
	results chan []*pb.RoundResult,
) *Round {
	r := &Round{
		parentCtx: ctx,
		timeout:   timeout,
		players:   players,
		results:   results,
		choicesCh: make(chan *pb.Choice, len(players)),
		choices:   make(map[string]*pb.Choice, len(players)),
	}

	go r.start()

	return r
}

// MakeChoise accepts a player's choise.
func (r *Round) MakeChoise(c *pb.Choice) {
	r.choicesCh <- c
}

// start waits for all players made their choises or round is timed out.
// Then it repots round results to channel.
func (r *Round) start() {
	for {
		for _, choice := range r.choices {
			choice.Choice = pb.EnumChoice_UnknownChoice
		}

		ctx, cancel := context.WithTimeout(r.parentCtx, r.timeout)

		r.handleChoises(ctx)
		cancel()

		r.results <- r.getResults()
	}
}

// handleChoises returns when all players have made their choises
// or when timeout is expired.
func (r *Round) handleChoises(ctx context.Context) {
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
		case <-ctx.Done():
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
		choice := r.choices[player.Id].GetChoice()
		score := int32(0)

		for _, otherPlayer := range r.players {
			if otherPlayer.Id == player.Id {
				continue
			}

			otherPlayerChoice := r.choices[otherPlayer.Id].GetChoice()
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
