package server

import (
	"fmt"
	"testing"

	pb "github.com/movaua/rock-paper-scissors/pkg/rps"
)

func TestGetStatus(t *testing.T) {
	tests := []struct {
		player   pb.EnumChoice
		other    pb.EnumChoice
		expected pb.EnumStatus
	}{
		{pb.EnumChoice_UnknownChoice, pb.EnumChoice_UnknownChoice, pb.EnumStatus_UnknownStatus},
		{pb.EnumChoice_UnknownChoice, pb.EnumChoice_Rock, pb.EnumStatus_UnknownStatus},
		{pb.EnumChoice_UnknownChoice, pb.EnumChoice_Paper, pb.EnumStatus_UnknownStatus},
		{pb.EnumChoice_UnknownChoice, pb.EnumChoice_Scissors, pb.EnumStatus_UnknownStatus},

		{pb.EnumChoice_Rock, pb.EnumChoice_UnknownChoice, pb.EnumStatus_UnknownStatus},
		{pb.EnumChoice_Paper, pb.EnumChoice_UnknownChoice, pb.EnumStatus_UnknownStatus},
		{pb.EnumChoice_Scissors, pb.EnumChoice_UnknownChoice, pb.EnumStatus_UnknownStatus},

		{pb.EnumChoice_Rock, pb.EnumChoice_Rock, pb.EnumStatus_Draw},
		{pb.EnumChoice_Rock, pb.EnumChoice_Paper, pb.EnumStatus_Looser},
		{pb.EnumChoice_Rock, pb.EnumChoice_Scissors, pb.EnumStatus_Winner},

		{pb.EnumChoice_Paper, pb.EnumChoice_Rock, pb.EnumStatus_Winner},
		{pb.EnumChoice_Paper, pb.EnumChoice_Paper, pb.EnumStatus_Draw},
		{pb.EnumChoice_Paper, pb.EnumChoice_Scissors, pb.EnumStatus_Looser},

		{pb.EnumChoice_Scissors, pb.EnumChoice_Rock, pb.EnumStatus_Looser},
		{pb.EnumChoice_Scissors, pb.EnumChoice_Paper, pb.EnumStatus_Winner},
		{pb.EnumChoice_Scissors, pb.EnumChoice_Scissors, pb.EnumStatus_Draw},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("%s:%s:%s", tt.player, tt.other, tt.expected), func(t *testing.T) {

		})
	}
}

func TestSetRoundStatuses(t *testing.T) {
	tests := []struct {
		name     string
		in       []*pb.RoundResult
		expected []pb.EnumStatus
	}{
		{
			name: "01 two players: winner and looser",
			in: []*pb.RoundResult{
				{Score: 0},
				{Score: 1},
			},
			expected: []pb.EnumStatus{
				pb.EnumStatus_Looser,
				pb.EnumStatus_Winner,
			},
		},
		{
			name: "02 two players: draw",
			in: []*pb.RoundResult{
				{Score: 1},
				{Score: 1},
			},
			expected: []pb.EnumStatus{
				pb.EnumStatus_Draw,
				pb.EnumStatus_Draw,
			},
		},
		{
			name: "03 tree players: unknown looser winner",
			in: []*pb.RoundResult{
				{Score: 0, Choice: pb.EnumChoice_UnknownChoice},
				{Score: 1, Choice: pb.EnumChoice_Scissors},
				{Score: 2},
			},
			expected: []pb.EnumStatus{
				pb.EnumStatus_UnknownStatus,
				pb.EnumStatus_Looser,
				pb.EnumStatus_Winner,
			},
		},
		{
			name: "04 complex with winner",
			in: []*pb.RoundResult{
				{Score: 0, Choice: pb.EnumChoice_UnknownChoice},
				{Score: 0, Choice: pb.EnumChoice_Rock},
				{Score: 0, Choice: pb.EnumChoice_Paper},
				{Score: 0, Choice: pb.EnumChoice_Scissors},
				{Score: 1, Choice: pb.EnumChoice_Rock},
				{Score: 2, Choice: pb.EnumChoice_Paper},
				{Score: 3, Choice: pb.EnumChoice_Scissors},
				{Score: 4},
			},
			expected: []pb.EnumStatus{
				pb.EnumStatus_UnknownStatus,
				pb.EnumStatus_Looser,
				pb.EnumStatus_Looser,
				pb.EnumStatus_Looser,
				pb.EnumStatus_Looser,
				pb.EnumStatus_Looser,
				pb.EnumStatus_Looser,
				pb.EnumStatus_Winner,
			},
		},
		{
			name: "05 complex with draw",
			in: []*pb.RoundResult{
				{Score: 0, Choice: pb.EnumChoice_UnknownChoice},
				{Score: 0, Choice: pb.EnumChoice_Rock},
				{Score: 0, Choice: pb.EnumChoice_Paper},
				{Score: 0, Choice: pb.EnumChoice_Scissors},
				{Score: 1, Choice: pb.EnumChoice_Rock},
				{Score: 2, Choice: pb.EnumChoice_Paper},
				{Score: 3, Choice: pb.EnumChoice_Scissors},
				{Score: 3, Choice: pb.EnumChoice_Scissors},
			},
			expected: []pb.EnumStatus{
				pb.EnumStatus_UnknownStatus,
				pb.EnumStatus_Looser,
				pb.EnumStatus_Looser,
				pb.EnumStatus_Looser,
				pb.EnumStatus_Looser,
				pb.EnumStatus_Looser,
				pb.EnumStatus_Draw,
				pb.EnumStatus_Draw,
			},
		},
		{
			name: "06 tree zeros",
			in: []*pb.RoundResult{
				{Score: 0},
				{Score: 0},
				{Score: 0, Choice: pb.EnumChoice_Rock},
			},
			expected: []pb.EnumStatus{
				pb.EnumStatus_UnknownStatus,
				pb.EnumStatus_UnknownStatus,
				pb.EnumStatus_Looser,
			},
		},
		{
			name: "07 one player",
			in: []*pb.RoundResult{
				{Score: 100, Choice: pb.EnumChoice_Rock},
			},
			expected: []pb.EnumStatus{
				pb.EnumStatus_UnknownStatus,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.in) != len(tt.expected) {
				t.Fatal("Invalid test case: slices are of different length")
			}

			SetRoundStatuses(tt.in)

			for i := 0; i < len(tt.in); i++ {
				if tt.in[i].Status != tt.expected[i] {
					t.Errorf("result %d, expected status %s, got %s", i, tt.expected[i], tt.in[i].Status)
				}
			}
		})
	}
}
