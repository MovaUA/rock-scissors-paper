package server

import (
	"sort"

	pb "github.com/movaua/rock-paper-scissors/pkg/rps"
)

func (g *Game) handleRoundResults(roundResults []*pb.RoundResult) {
	for _, roundResult := range roundResults {
		gameResult := g.gameResults[roundResult.Player.Id]
		gameResult.Rounds++
		gameResult.Score += roundResult.Score
	}

	gameResults := make([]*pb.GameResult, 0, len(g.gameResults))
	for _, gameResult := range g.gameResults {
		gameResults = append(gameResults, gameResult)
	}

	sort.Slice(gameResults, func(i, j int) bool {
		return gameResults[i].Score < gameResults[j].Score
	})

	SetGameStatuses(gameResults)

	score := &pb.Score{
		RoundResults: roundResults,
		GameResults:  gameResults,
	}

	for _, scoreCh := range g.startRequests {
		scoreCh <- score
	}
}

// SetGameStatuses sets statuses to each player based on their game score.
func SetGameStatuses(r []*pb.GameResult) {
	if len(r) == 2 {
		if r[1].Score > r[0].Score {
			r[1].Status = pb.EnumStatus_Winner
			r[0].Status = pb.EnumStatus_Looser
		}
		if r[1].Score == r[0].Score && r[1].Score > 0 {
			r[1].Status = pb.EnumStatus_Draw
			r[0].Status = pb.EnumStatus_Draw
		}
		return
	}

	if len(r) < 2 {
		return
	}

	if hasWinner := r[len(r)-1].Score > r[len(r)-2].Score; hasWinner {
		r[len(r)-1].Status = pb.EnumStatus_Winner
		for i := 0; i < len(r)-1; i++ {
			r[i].Status = pb.EnumStatus_Looser
		}
		return
	}

	drawScore := r[len(r)-1].Score
	for i := 0; i < len(r); i++ {
		switch {
		case r[i].Score < drawScore:
			r[i].Status = pb.EnumStatus_Looser
		default:
			r[i].Status = pb.EnumStatus_Draw
		}
	}
}
