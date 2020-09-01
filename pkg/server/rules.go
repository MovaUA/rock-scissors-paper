package server

import pb "github.com/movaua/rock-paper-scissors/pkg/rps"

// these are the game rules
var (
	// key: choice,
	// value: rule
	rules = map[pb.EnumChoice]rule{
		pb.EnumChoice_Rock:     {strongTo: pb.EnumChoice_Scissors, weakTo: pb.EnumChoice_Paper},
		pb.EnumChoice_Paper:    {strongTo: pb.EnumChoice_Rock, weakTo: pb.EnumChoice_Scissors},
		pb.EnumChoice_Scissors: {strongTo: pb.EnumChoice_Paper, weakTo: pb.EnumChoice_Rock},
	}

	// key: status,
	// value: score
	statusScores = map[pb.EnumStatus]int32{
		pb.EnumStatus_UnknownStatus: 0,
		pb.EnumStatus_Looser:        0,
		pb.EnumStatus_Draw:          1,
		pb.EnumStatus_Winner:        2,
	}
)

type rule struct {
	strongTo pb.EnumChoice
	weakTo   pb.EnumChoice
}

// GetStatus compares player choice agaist other player choice.
func GetStatus(player pb.EnumChoice, other pb.EnumChoice) pb.EnumStatus {
	if player == pb.EnumChoice_UnknownChoice || other == pb.EnumChoice_UnknownChoice {
		return pb.EnumStatus_UnknownStatus
	}
	if rules[player].strongTo == other {
		return pb.EnumStatus_Winner
	}
	if rules[player].weakTo == other {
		return pb.EnumStatus_Looser
	}
	return pb.EnumStatus_Draw
}

// SetRoundStatuses sets statuses to each player based on rank of player score.
func SetRoundStatuses(r []*pb.RoundResult) {
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
			if r[i].Choice != pb.EnumChoice_UnknownChoice {
				r[i].Status = pb.EnumStatus_Looser
			}
		}
		return
	}

	drawScore := r[len(r)-1].Score
	for i := 0; i < len(r); i++ {
		switch {
		case r[i].Score == 0 || r[i].Score < drawScore:
			if r[i].Choice != pb.EnumChoice_UnknownChoice {
				r[i].Status = pb.EnumStatus_Looser
			}
		case r[i].Score == drawScore:
			r[i].Status = pb.EnumStatus_Draw
		}
	}
}
