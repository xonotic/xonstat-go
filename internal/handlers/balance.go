package handlers

import (
	"bufio"
	"fmt"
	"hash/fnv"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strings"

	wenglin "gitlab.com/xonotic/xonstat/pkg/skill"
	"gitlab.com/xonotic/xonstat/pkg/submission"
)

type balancePlayer struct {
	Hashkey        string
	PlayerID       int
	Nick           string
	Skill          float64
	ScorePerSecond float64
}

type balanceResponse struct {
	Version int
	Release string
	Time    int64
	Players []*balancePlayer
}

// preprocess handles assembling the POST body into a type that we can work with
func preprocess(w http.ResponseWriter, r *http.Request) (*submission.Submission, error) {
	bodyReader := bufio.NewReader(r.Body)

	rawSubmission, err := submission.NewRawSubmission(bodyReader)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("422 %s", http.StatusText(422)), 422)
		return nil, err
	}

	// It doesn't make sense to balance games with only one or two players.
	if rawSubmission.NumHumansPlayed < 3 {
		log.Printf("Error: not enough players to balance")
		http.Error(w, fmt.Sprintf("422 %s", http.StatusText(422)), 422)
		return nil, fmt.Errorf("not enough players to balance")
	}

	sub, err := submission.NewSubmission(rawSubmission)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("422 %s", http.StatusText(422)), 422)
		return nil, err
	}

	return sub, nil
}

// seed creates a per-player, per-match consistent value for seeding the RNG
func seed(hashkey, matchID string) int64 {
	// FNV1A64 non-cryptographic hash
	hash := fnv.New64a()
	hash.Write([]byte(hashkey + matchID))

	return int64(hash.Sum64())
}

// BalanceHandler godoc
// @Summary Best guess ordering of players according to skill and score data.
// @Accept  text/plain
// @Produce  text/plain
// @Success 200 {object} balanceResponse
// @Router /balance [post]
func (ae *AppEnv) BalanceHandler(w http.ResponseWriter, r *http.Request) {
	sub, err := preprocess(w, r)
	if err != nil {
		return
	}

	// The destination for the balanced (sorted) players.
	players := make([]*balancePlayer, 0)

	// Use hashkeys to make updates to entries in the above slice.
	hashkeysToPlayers := make(map[string]*balancePlayer, 0)

	// Per-match random seeds (hashkey + match ID)
	hashkeysToSeeds := make(map[string]int64, 0)

	// Players that actually have skills/ratings in the DB.
	hashkeysWithSkills := make(map[string]struct{})

	// Hashkeys for tracked players
	tracked := make([]string, 0)

	// Keep track of the maxes, for use in scaling later.
	var maxScorePerSecond float64 = 1.0
	var maxSkill float64 = 1.0

	// In case there are NO skill values to establish a "match average",
	// we will use the default ones.
	avgMu := wenglin.DefaultParams.DefaultMu
	avgSigma := wenglin.DefaultParams.DefaultSigma

	// TODO: take this as input
	randomness := 1.0

	// TODO: establish a better value
	scoreFactor := 0.25

	for hashkey, player := range sub.PlayersByHashkey {
		gamestat, hasGameStat := sub.PlayerGameStatsByHashkey[hashkey]
		if strings.HasPrefix(hashkey, "bot#") || !hasGameStat {
			// Bots and players that don't have a stat record aren't considered.
			continue
		}

		// Use a consistent per-match seed
		hashkeysToSeeds[hashkey] = seed(hashkey, sub.Game.MatchID.String)

		score := gamestat.Score.Int32
		alivetimesecs := sub.PlayerGameStatsByHashkey[hashkey].AliveTime.Seconds()
		sps := float64(score) / alivetimesecs

		if sps > maxScorePerSecond {
			maxScorePerSecond = sps
		}

		player := balancePlayer{
			Hashkey:        hashkey,
			PlayerID:       player.PlayerID,
			Nick:           player.Nick.String,
			ScorePerSecond: sps,
		}

		players = append(players, &player)

		hashkeysToPlayers[hashkey] = &player

		if !strings.HasPrefix(hashkey, "player#") {
			// This is a tracked player, and might have a skill record. Save these to query the DB.
			tracked = append(tracked, hashkey)
		}
	}

	// Those that are tracked are the ones that might have skill values in the DB.
	// We query for them in batch to decrease round-trips. 
	if len(tracked) > 0 {
		skills, err := ae.db.RPlayerSkillsBatch(tracked, sub.Game.GameTypeCd)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("422 %s", http.StatusText(422)), 422)
			return
		}

		var sumMu, sumSigma float64

		for _, skill := range skills {
			hashkeysWithSkills[skill.Hashkey] = struct{}{}

			sumMu += skill.Mu
			sumSigma += skill.Sigma

			player := hashkeysToPlayers[skill.Hashkey]

			rng := rand.New(rand.NewSource(hashkeysToSeeds[skill.Hashkey]))
			player.Skill = math.Exp((rng.NormFloat64()*(skill.Sigma*randomness)+skill.Mu)/wenglin.DefaultParams.DefaultBeta)

			if player.Skill > maxSkill {
				maxSkill = player.Skill
			}
		}

		if len(skills) > 1 {
			// We DO have at least two data points to establish an average skill.
			avgMu = sumMu / float64(len(skills))
			avgSigma = sumSigma / float64(len(skills))
		}
	}

	// Final pass to factor in the score
	scorePerSecondScale := maxSkill / float64(maxScorePerSecond)

	for hashkey, player := range hashkeysToPlayers {
		if _, ok := hashkeysWithSkills[hashkey]; !ok {
			// This player doesn't yet have a skill. Derive one from the default.
			// Use a consistent per-match seed
			rng := rand.New(rand.NewSource(hashkeysToSeeds[hashkey]))
			player.Skill = math.Exp((rng.NormFloat64()*(avgSigma*randomness)+avgMu)/wenglin.DefaultParams.DefaultBeta)
		}

		// Finally, add score to the linearized value
		player.Skill += scorePerSecondScale * player.ScorePerSecond * scoreFactor
	}

	response := balanceResponse{
		Version: 1,
		Release: "XonStat/1.0",
		Time:    sub.CreateDt.Unix(),
		Players: players,
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "text/plain")

	err = ae.templates["balance.page.html"].Execute(w, response)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}
}
