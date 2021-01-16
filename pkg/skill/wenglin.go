package skill

import (
	"fmt"
	"log"
	"math"
)

// MU is the mean skill value for a brand new player.
const MU = 25.0

// SIGMA is the standard deviation for skill of a brand new player.
const SIGMA = MU / 3

// BETA is a component used in calculating Weng-Lin values.
const BETA = SIGMA / 2

// Rating is the Weng-Lin skill value for a given player. It is basically the two components
// describing a normal distribution: mu (the mean) and sigma (the standard deviation).
type Rating struct {
	Mu    float64
	Sigma float64
}

// PlayerResult holds the results scored/obtained by a participant in the match.
type PlayerResult struct {
	// A unique identifier for this participant who owns these results.
	PlayerID int

	// How this player performed in the match. This figure is central to the skill
	// calculations, thus it should be normalized across all participants in the.
	Score float32

	// TODO: K factor to scale skill deltas for those who haven't participated fully, played at
	// a disadvantage, etc?
}

// MatchResult is the results of a match, and is used to calculate skill updates.
type MatchResult struct {
	// A unique identifier for the match.
	MatchID int

	// The player performances that are to be evaluated.
	PlayerResults []PlayerResult
}

// WengLinBT calculates the updates to the player skills using the Weng-Lin
// Bradley-Terry full pair algorithm.
// Original code here: http://www.csie.ntu.edu.tw/~cjlin/papers/online_ranking/
func WengLinBT(result MatchResult, skills []Rating) ([]Rating, error) {
	// We expect the caller to provide the list of starting skills such that each
	// player result in the match has a corresponding skill at the same index.
	if len(result.PlayerResults) != len(skills) {
		log.Printf("Mismatched data: %d players in game, %d skills provided.",
			len(result.PlayerResults), len(skills))

		return nil, fmt.Errorf("number of players and skills do not match")
	}

	omega := make(map[int]float64, len(result.PlayerResults))
	delta := make(map[int]float64, len(result.PlayerResults))
	for p1index, p1 := range result.PlayerResults {
		// Omega is the amount added to Mu to determine its new value.
		omega[p1.PlayerID] = 0.0

		// Delta is the amount multiplied with Sigma to determine its new value.
		delta[p1.PlayerID] = 0.0

		for p2index, p2 := range result.PlayerResults {
			if p2.PlayerID == p1.PlayerID {
				continue
			}

			fmt.Printf("Comparing %d and %d\n", p1index, p2index)

			p1SigmaSquared := math.Pow(skills[p1index].Sigma, 2.0)
			p2SigmaSquared := math.Pow(skills[p2index].Sigma, 2.0)
			betaSquared := math.Pow(BETA, 2.0)
			ciq := math.Sqrt(p1SigmaSquared + p2SigmaSquared + (2 * betaSquared))

			muDiff := skills[p2index].Mu - skills[p1index].Mu
			piq := 1. / (1. + math.Exp(muDiff/ciq))

			fmt.Printf("p1MuDiff: %f\n", muDiff)
			fmt.Printf("p1piq: %f\n", piq)

			// TODO: This is currently winner-take-all. Implement scaling?
			// If we implement scaling, we also need to normalize the scores with
			// offsets so they can be compared.
			s := 0.0
			if p1.Score > p2.Score {
				// P1 won
				s = 1
			} else if p1.Score == p2.Score {
				// P1 and P2 tied
				s = 0.5
			}

			omega[p1.PlayerID] += (p1SigmaSquared / ciq) * (s - piq)
			gamma := skills[p1index].Sigma / ciq
			delta[p1.PlayerID] += gamma * (p1SigmaSquared / ciq) / ciq * piq * (1 - piq)

			fmt.Printf("p1 omega += %f\n", (p1SigmaSquared/ciq)*(s-piq))
			fmt.Printf("p1 delta += %f\n\n", gamma*(p1SigmaSquared/ciq)/ciq*piq*(1-piq))
		}
	}

	newSkills := make([]Rating, len(skills))
	for i, player := range result.PlayerResults {
		// Clamp the factor by which sigma changes.
		d := 1 - delta[player.PlayerID]
		if d < 0.0001 {
			d = 0.0001
		}

		newSkills[i] = Rating{
			Mu:    skills[i].Mu + omega[player.PlayerID],
			Sigma: skills[i].Sigma * math.Sqrt(d),
		}
	}

	return newSkills, nil
}
