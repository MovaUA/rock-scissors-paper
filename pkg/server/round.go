package server

import (
	"context"
	"sort"
	"time"

	pb "github.com/movaua/rock-paper-scissors/pkg/rps"
)

// round is an API of a single round of the game.
type round struct {
	ctx       context.Context
	cancel    func()
	players   map[string]*pb.Player // key is player.Id
	choicesCh chan *pb.Choice
	choices   map[string]*pb.Choice // key is player.Id
	resultsCh chan []*pb.RoundResult
}

// newRound return new started round of the game.
func newRound(
	ctx context.Context,
	timeout time.Duration,
	players map[string]*pb.Player,
) *round {
	ctx, cancel := context.WithTimeout(ctx, timeout)

	r := &round{
		ctx:       ctx,
		cancel:    cancel,
		players:   players,
		choicesCh: make(chan *pb.Choice, len(players)),
		choices:   make(map[string]*pb.Choice, len(players)),
		resultsCh: make(chan []*pb.RoundResult),
	}

	go r.start()

	return r
}

// MakeChoise accepts a player's choise.
func (r *round) MakeChoise(c *pb.Choice) {
	r.choicesCh <- c
}

// start waits for all players made their choises or round is timed out.
// Then it reports the result back to game.
func (r *round) start() {
	defer r.cancel()

	r.handleChoises()

	r.resultsCh <- r.getResults()
	close(r.resultsCh)
}

// handleChoises returns when all players have made their choises
// or when timeout is expired.
func (r *round) handleChoises() {
	for {
		select {

		// player made their choise:
		case choise := <-r.choicesCh:
			// if choise is send by connected player:
			if _, ok := r.players[choise.GetPlayerId()]; ok {
				r.choices[choise.GetPlayerId()] = choise
				// if all players made their choises
				if len(r.choices) == len(r.players) {
					return
				}
			}

			// timeout expired (or game server is closed):
		case <-r.ctx.Done():
			return
		}
	}
}

func (r *round) getResults() []*pb.RoundResult {
	results := make([]*pb.RoundResult, len(r.players))

	playerScores := make([]playerScore, len(r.players))

	for _, player := range r.players {
		playerChoice := r.choices[player.GetId()].GetChoice()
		score := 0

		for _, otherPlayer := range r.players {
			if otherPlayer.GetId() == player.GetId() {
				continue
			}

			otherPlayerChoice := r.choices[otherPlayer.GetId()].GetChoice()
			status := compareChoises(playerChoice, otherPlayerChoice)

			score += statusScore[status]
		}

		playerScores = append(playerScores,
			playerScore{
				player: player,
				choice: playerChoice,
				score:  score,
			})
	}

	sort.Slice(playerScores, func(i, j int) bool {
		return playerScores[i].score < playerScores[j].score
	})

	return results
}

type playerScore struct {
	player *pb.Player
	choice pb.EnumChoice
	score  int
}

var (
	gameItems = map[pb.EnumChoice]gameItem{
		pb.EnumChoice_Rock:     {strongTo: pb.EnumChoice_Scissors, weakTo: pb.EnumChoice_Paper},
		pb.EnumChoice_Paper:    {strongTo: pb.EnumChoice_Rock, weakTo: pb.EnumChoice_Scissors},
		pb.EnumChoice_Scissors: {strongTo: pb.EnumChoice_Paper, weakTo: pb.EnumChoice_Rock},
	}

	statusScore = map[pb.EnumStatus]int{
		pb.EnumStatus_UnknownStatus: 0,
		pb.EnumStatus_Looser:        0,
		pb.EnumStatus_Draw:          1,
		pb.EnumStatus_Winner:        2,
	}
)

type gameItem struct {
	strongTo pb.EnumChoice
	weakTo   pb.EnumChoice
}

func compareChoises(playerChoice pb.EnumChoice, otherPlayerChoice pb.EnumChoice) pb.EnumStatus {
	if playerChoice == pb.EnumChoice_UnknownChoice || otherPlayerChoice == pb.EnumChoice_UnknownChoice {
		return pb.EnumStatus_UnknownStatus
	}
	if gameItems[playerChoice].strongTo == otherPlayerChoice {
		return pb.EnumStatus_Winner
	}
	if gameItems[playerChoice].weakTo == otherPlayerChoice {
		return pb.EnumStatus_Looser
	}
	return pb.EnumStatus_Draw
}

// Result returns a chan with round results.
func (r *round) Result() <-chan []*pb.RoundResult {
	return r.resultsCh
}
