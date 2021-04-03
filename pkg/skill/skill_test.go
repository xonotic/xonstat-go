package skill

import (
	"testing"
)

// The values here are from other implementations. We use them here for cross-verification.
var params = Params{
	DefaultMu: 25.0,
	DefaultSigma: 25.0/ 3,
	DefaultBeta: (25.0/3.0)/2.0,
}

func TestMissingData(t *testing.T) {
	result := MatchResult{
		MatchID: 1,
		PlayerResults: []PlayerResult{
			{
				PlayerID: 1,
				Score:    5.0,
				KFactor:  1.0,
			},
			{
				PlayerID: 2,
				Score:    5.0,
				KFactor:  1.0,
			},
		},
	}

	skills := []Rating{
		{
			Mu:    params.DefaultMu,
			Sigma: params.DefaultSigma,
		},
	}

	_, err := WengLinBT(&params, result, skills)
	if err == nil {
		t.Fatal("Expected mismatch error, got nil!")
	}
}

func TestDefaultSkills(t *testing.T) {
	result := MatchResult{
		MatchID: 1,
		PlayerResults: []PlayerResult{
			{
				PlayerID: 1,
				Score:    10.0,
				KFactor:  1.0,
			},
			{
				PlayerID: 2,
				Score:    5.0,
				KFactor:  1.0,
			},
		},
	}

	skills := []Rating{
		{
			Mu:    params.DefaultMu,
			Sigma: params.DefaultSigma,
		},
		{
			Mu:    params.DefaultMu,
			Sigma: params.DefaultSigma,
		},
	}

	newSkills, err := WengLinBT(&params, result, skills)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if newSkills[0].Mu != 27.63523138347365 || newSkills[0].Sigma != 8.065506316323548 {
		t.Fatalf("P1 skill calculation is not correct: %+v", newSkills[0])
	}

	if newSkills[1].Mu != 22.36476861652635 || newSkills[1].Sigma != 8.065506316323548 {
		t.Fatalf("P2 skill calculation is not correct: %+v", newSkills[1])
	}
}

func TestDefaultSkillsKFactor(t *testing.T) {
	// Same as the above, but with an added K factor as if
	// one person only played 50% of the match.
	result := MatchResult{
		MatchID: 1,
		PlayerResults: []PlayerResult{
			{
				PlayerID: 1,
				Score:    10.0,
				KFactor:  1.0,
			},
			{
				PlayerID: 2,
				Score:    5.0,
				KFactor:  .5,
			},
		},
	}

	skills := []Rating{
		{
			Mu:    params.DefaultMu,
			Sigma: params.DefaultSigma,
		},
		{
			Mu:    params.DefaultMu,
			Sigma: params.DefaultSigma,
		},
	}

	newSkills, err := WengLinBT(&params, result, skills)
	if err != nil {
		t.Fatalf("%s", err)
	}

	// Since we have a 50% K factor, we expect that the winner should not gain as much Mu and the loser not
	// lose as much. Both the Sigmas should stay bigger (the distribution won't get as "narrow").
	if newSkills[0].Mu >= 27.63523138347365 || newSkills[0].Sigma <= 8.065506316323548 {
		t.Fatalf("P1 skill calculation is not correct %+v", newSkills[0])
	}

	if newSkills[1].Mu <= 22.36476861652635 || newSkills[1].Sigma <= 8.065506316323548 {
		t.Fatalf("P2 skill calculation is not correct %+v", newSkills[1])
	}
}

func TestTieDefaultSkills(t *testing.T) {
	result := MatchResult{
		MatchID: 1,
		PlayerResults: []PlayerResult{
			{
				PlayerID: 1,
				Score:    5.0,
				KFactor:  1.0,
			},
			{
				PlayerID: 2,
				Score:    5.0,
				KFactor:  1.0,
			},
		},
	}

	skills := []Rating{
		{
			Mu:    params.DefaultMu,
			Sigma: params.DefaultSigma,
		},
		{
			Mu:    params.DefaultMu,
			Sigma: params.DefaultSigma,
		},
	}

	newSkills, err := WengLinBT(&params, result, skills)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if newSkills[0].Sigma != 8.065506316323548 || (newSkills[0].Sigma != newSkills[1].Sigma) {
		t.Fatalf("We expect sigma to shrink at the same rate for both players in ties.")
	}
}

func TestThreePlayerGame(t *testing.T) {
	result := MatchResult{
		MatchID: 1,
		PlayerResults: []PlayerResult{
			{
				PlayerID: 1,
				Score:    10.0,
				KFactor:  1.0,
			},
			{
				PlayerID: 2,
				Score:    5.0,
				KFactor:  1.0,
			},
			{
				PlayerID: 3,
				Score:    2.0,
				KFactor:  1.0,
			},
		},
	}

	skills := []Rating{
		{
			Mu:    23.4084,
			Sigma: 0.788529,
		},
		{
			Mu:    23.0283,
			Sigma: 0.788645,
		},
		{
			Mu:    21.4084,
			Sigma: 0.768529,
		},
	}

	newSkills, err := WengLinBT(&params, result, skills)
	if err != nil {
		t.Fatalf("%s", err)
	}

	// Check each skill update matches.
	if newSkills[0].Mu != 23.501886897287523 || newSkills[0].Sigma != 0.7880868387194325 {
		t.Fatalf("Player 1's rating is incorrect: %+v", newSkills[0])
	}

	if newSkills[1].Mu != 23.02299821039129 || newSkills[1].Sigma != 0.78820049488184 {
		t.Fatalf("Player 2's rating is incorrect: %+v", newSkills[1])
	}

	if newSkills[2].Mu != 21.32463007828219 || newSkills[2].Sigma != 0.7681332152849163 {
		t.Fatalf("Player 3's rating is incorrect: %+v", newSkills[2])
	}
}