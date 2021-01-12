package skill

import (
	"testing"
)

// test that errors are thrown if a submission is for a blank game
func TestP1Wins(t *testing.T) {
	p1Result := PlayerResult{
		PlayerID: 1,
		Score: 10.0,
	}

	p2Result := PlayerResult{
		PlayerID: 2,
		Score: 5.0,
	}

	result := MatchResult{
		MatchID: 1,
		PlayerResults: []PlayerResult{p1Result, p2Result},
	}

	p1Skill := Rating{
		Mu: MU,
		Sigma: SIGMA,
	}

	p2Skill := Rating{
		Mu: MU,
		Sigma: SIGMA,
	}

	skills := []Rating{p1Skill, p2Skill}

	newSkills, err := WengLinBT(result, skills)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if newSkills[0].Mu != 27.63523138347365 || newSkills[0].Sigma != 8.065506316323548 {
		t.Fatal("P1 skill calculation is not correct")
	}

	if newSkills[1].Mu != 22.36476861652635 || newSkills[1].Sigma != 8.065506316323548 {
		t.Fatal("P2 skill calculation is not correct")
	}
}
