package skill

import (
	"math"
	"time"
)

// Initial parameters for Ratings.
const defaultMu = 25.0
const defaultSigma = defaultMu/3
const defaultBeta = defaultSigma/2

// Rating is the Weng-Lin skill value for a given player. It is basically the two components
// describing a normal distribution: mu (the mean) and sigma (the standard deviation).
type Rating struct {
	Mu float64
	Sigma float64
}

// PlayerResult holds the results scored/obtained by a participant in the match.
type PlayerResult struct {
	// A unique identifier for this participant who owns these results.
	PlayerID int

	// How this player performed in the match. This figure is central to the skill
	// calculations, thus it should be normalized across all participants in the.
	Score float32
}

// MatchResult is the results of a match, and is used to calculate skill updates.
type MatchResult struct {
	// A unique identifier for the match.
	MatchID int

	// The player performances that are to be evaluated.
	PlayerResults []PlayerResult

	// How long the match lasted.
	Duration time.Duration
}

func addSkillIfMissing(pid int, skills map[int]*Rating) {
	if _, ok := skills[pid]; !ok {
		skills[pid] = &Rating{
			Mu: defaultMu,
			Sigma: defaultSigma,
		}
	}
}

// WengLinBT calculates the updates to the player skills using the Weng-Lin 
// Bradley-Terry full pair algorithm.
// Original code here: http://www.csie.ntu.edu.tw/~cjlin/papers/online_ranking/
func WengLinBT(matches []MatchResult, skills map[int]*Rating) {
	for _, matchResult := range matches {
		omega := make(map[int]float64, len(matchResult.PlayerResults))
		delta := make(map[int]float64, len(matchResult.PlayerResults))
		for _, p1 := range matchResult.PlayerResults {
			omega[p1.PlayerID] = 0.0
			delta[p1.PlayerID] = 0.0
			for _, p2 := range matchResult.PlayerResults {
				if p2.PlayerID == p1.PlayerID {
					continue
				}

				// Ensure both players have skill values
				addSkillIfMissing(p1.PlayerID, skills)
				addSkillIfMissing(p2.PlayerID, skills)

				p1SigmaSquared := math.Pow(skills[p1.PlayerID].Sigma, 2.0)
				p2SigmaSquared := math.Pow(skills[p2.PlayerID].Sigma, 2.0)
				betaSquared := math.Pow(defaultBeta, 2.0)
				ciq := math.Sqrt(p1SigmaSquared + p2SigmaSquared + (2*betaSquared))

				muDiff := skills[p2.PlayerID].Mu - skills[p1.PlayerID].Mu
				piq := 1. / (1. + math.Exp(muDiff / ciq))

				// TODO: This is currently winner-take-all. Implement scaling?
				s := 0.0
				if p1.Score > p2.Score {
					// P1 won
					s = 1
				} else if p1.Score == p2.Score {
					// P1 and P2 tied
					s = 0.5
				}

				omega[p1.PlayerID] += (p1SigmaSquared / ciq) * (s - piq)
                gamma := skills[p1.PlayerID].Sigma / ciq
                delta[p1.PlayerID] += gamma * (p1SigmaSquared / ciq) / ciq * piq * (1 - piq)
			}
		}

		for _, player := range matchResult.PlayerResults {
			// This is the resulting skill update after we've processed the match
			skills[player.PlayerID].Mu += omega[player.PlayerID]

			d := 1-delta[player.PlayerID]
			if d < 0.0001 {
				d = 0.0001
			}
			skills[player.PlayerID].Sigma *= math.Sqrt(d)
		}
	}
}