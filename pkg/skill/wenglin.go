package skill

import (
	"fmt"
	"log"
	"math"
)

// Params represent the default values used in calculating Weng-Lin. 
type Params struct {
    // DefaultMU is the mean skill value for a brand new player.
	DefaultMu float64

	// DefaultSigma is the standard deviation for skill of a brand new player.
	DefaultSigma float64

	// DefaultBeta is a component used in calculating Weng-Lin values.
	DefaultBeta float64
}

// DefaultParams are the param values used for Weng-Lin when a player hasn't been seen before.
var DefaultParams = Params{
	DefaultMu: 1500.0,
	DefaultSigma: 350.0,
	DefaultBeta: 350.0/2.0,
}

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

	// K factor to scale skill deltas for those who haven't participated fully, played at a disadvantage, etc.
	KFactor float32
}

// MatchResult is the results of a match, and is used to calculate skill updates.
type MatchResult struct {
	// A unique identifier for the match.
	MatchID int

	// The player performances that are to be evaluated.
	PlayerResults []PlayerResult
}

func minKFactor(a, b float32) float32 {
	if a <= b {
		return a
	} else {
		return b
	}
}

// WengLinBT calculates the updates to the player skills using the Weng-Lin
// Bradley-Terry full pair algorithm.
// Original code here: http://www.csie.ntu.edu.tw/~cjlin/papers/online_ranking/
func WengLinBT(params *Params, result MatchResult, skills []Rating) ([]Rating, error) {
	// If we're not provided with parameters, we'll use the defaults.
	if params == nil {
		params = &DefaultParams
	}

	// We expect the caller to provide the list of starting skills such that each
	// player result in the match has a corresponding skill at the same index.
	if len(result.PlayerResults) != len(skills) {
		log.Printf("Mismatched data: %d players in game, %d skills provided.",
			len(result.PlayerResults), len(skills))

		return nil, fmt.Errorf("number of players and skills do not match")
	}

	// Omega is the amount added to Mu to determine its new value.
	omega := make([]float64, len(result.PlayerResults))

	// Delta is the amount multiplied with Sigma to determine its new value.
	delta := make([]float64, len(result.PlayerResults))

	for p1index, p1 := range result.PlayerResults {
		for p2index := p1index + 1; p2index < len(result.PlayerResults); p2index++ {
			p2 := result.PlayerResults[p2index]
			if p2.PlayerID == p1.PlayerID {
				continue
			}

			p1SigmaSquared := math.Pow(skills[p1index].Sigma, 2.0)
			p2SigmaSquared := math.Pow(skills[p2index].Sigma, 2.0)

			betaSquared := math.Pow(params.DefaultBeta, 2.0)
			ciq := math.Sqrt(p1SigmaSquared + p2SigmaSquared + (2 * betaSquared))

			p1MuDiff := skills[p2index].Mu - skills[p1index].Mu
			p2MuDiff := skills[p1index].Mu - skills[p2index].Mu

			p1piq := 1. / (1. + math.Exp(p1MuDiff/ciq))
			p2piq := 1. / (1. + math.Exp(p2MuDiff/ciq))

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

			// Scaling is done based on the smallest K factor, thus the biggest reduction to the Mu/Sigma delta.
			k := minKFactor(p1.KFactor, p2.KFactor)

			omega[p1index] += ((p1SigmaSquared / ciq) * (s - p1piq)) * float64(k)
			omega[p2index] += ((p2SigmaSquared / ciq) * ((1 - s) - p2piq)) * float64(k)

			delta[p1index] += ((skills[p1index].Sigma / ciq) * (p1SigmaSquared / ciq) / ciq * p1piq * (1 - p1piq)) * float64(k)
			delta[p2index] += ((skills[p2index].Sigma / ciq) * (p2SigmaSquared / ciq) / ciq * p2piq * (1 - p2piq)) * float64(k)

			/* Debugging...
			fmt.Printf("Comparing player %d and player %d\n", p1.PlayerID, p2.PlayerID)
			fmt.Printf("player %d MuDiff: %f\n", p1.PlayerID, p1MuDiff)
			fmt.Printf("player %d piq: %f\n", p1.PlayerID, p1piq)
			fmt.Printf("player %d omega += %f\n", p1.PlayerID, (p1SigmaSquared/ciq)*(s-p1piq))
			fmt.Printf("player %d delta += %f\n", p1.PlayerID, (skills[p1index].Sigma/ciq)*(p1SigmaSquared/ciq)/ciq*p1piq*(1-p1piq))
			fmt.Printf("player %d MuDiff: %f\n", p2.PlayerID, p2MuDiff)
			fmt.Printf("player %d piq: %f\n", p2.PlayerID, p2piq)
			fmt.Printf("player %d omega += %f\n", p2.PlayerID, (p2SigmaSquared/ciq)*((1-s)-p2piq))
			fmt.Printf("player %d delta += %f\n\n", p2.PlayerID, (skills[p2index].Sigma/ciq)*(p2SigmaSquared/ciq)/ciq*p2piq*(1-p2piq))
			*/
		}
	}

	newSkills := make([]Rating, len(skills))
	for i := range result.PlayerResults {
		// Clamp the factor by which sigma changes.
		d := 1 - delta[i]
		if d < 0.0001 {
			d = 0.0001
		}

		// The floor for Mu is 0.0.
		newMu := skills[i].Mu + omega[i]
		if newMu < 0.0 {
			newMu = 0.0
		}

		newSkills[i] = Rating{
			Mu:    newMu,
			Sigma: skills[i].Sigma * math.Sqrt(d),
		}
	}

	return newSkills, nil
}
