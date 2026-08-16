package skill

import (
	"hash/fnv"
	"log"
	"math"
	"math/rand"
	"sort"
	"strings"

	"gitlab.com/xonotic/xonstat/pkg/models"
	"gitlab.com/xonotic/xonstat/pkg/submission"
)

// SkillStore is the subset of the datastore that the balance calculation
// needs in order to look up player skills.
type SkillStore interface {
	RPlayerSkillsBatch(hashkeys []string, gameTypeCd string) ([]*models.PlayerHashkeySkill, error)
}

type BalanceParams struct {
	// These are what we'll use as default skill values if we can't calculate 
	// an average from the players in the match. 
	DefaultMu float64
	DefaultSigma float64
	DefaultBeta float64

	// How much weight to give the score as compared to the skill value. 
	ScoreFactor float64
}

// BalancePlayer is the output data structure that gets sent to 
// the template, which is emitted back to the game server so it
// can reassign the player to the given team.
type BalancePlayer struct {
	Hashkey        string
	PlayerID       int
	Nick           string
	Skill          float64
	ScorePerSecond float64
	Team int
}

// seed creates a per-player, per-match consistent value for seeding the RNG
func seed(hashkey, matchID string) int64 {
	// FNV1A64 non-cryptographic hash
	hash := fnv.New64a()
	hash.Write([]byte(hashkey + matchID))

	return int64(hash.Sum64())
}

func Balance(params BalanceParams, db SkillStore, sub *submission.Submission) ([]*BalancePlayer, error) {
	// The destination for the balanced (sorted) players.
	players := make([]*BalancePlayer, 0)

	// Use hashkeys to make updates to entries in the above slice.
	hashkeysToPlayers := make(map[string]*BalancePlayer, 0)

	// Per-match random seeds (hashkey + match ID)
	hashkeysToSeeds := make(map[string]int64, 0)

	// Players that actually have skills/ratings in the DB.
	hashkeysWithSkills := make(map[string]struct{})

	// Hashkeys for tracked players
	tracked := make([]string, 0)

	// Keep track of the maxes, for use in scaling later.
	var maxScorePerSecond float64 = 0.0
	var maxSkill float64 = 1.0

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

		// The player's current team (the "t" record), or zero if they aren't
		// on a team (a spectator or otherwise unassigned player).
		team := 0
		if gamestat.Team.Valid {
			team = int(gamestat.Team.Int32)
		}

		player := BalancePlayer{
			Hashkey:        hashkey,
			PlayerID:       player.PlayerID,
			Nick:           player.Nick.String,
			ScorePerSecond: sps,
			Team:           team,
		}

		players = append(players, &player)

		hashkeysToPlayers[hashkey] = &player

		if !strings.HasPrefix(hashkey, "player#") {
			// This is a tracked player, and might have a skill record. Save these to query the DB.
			tracked = append(tracked, hashkey)
		}
	}

	// Use a deterministic order for the rest of the calculation so the team
	// assignment doesn't depend on map iteration order.
	sort.Slice(players, func(i, j int) bool { return players[i].PlayerID < players[j].PlayerID })

	// Those that are tracked are the ones that might have skill values in the DB.
	// We query for them in batch to decrease round-trips. 
	if len(tracked) > 0 {
		skills, err := db.RPlayerSkillsBatch(tracked, sub.Game.GameTypeCd)
		if err != nil {
			log.Printf("Error: %s", err)
			return nil, err
		}

		var sumMu, sumSigma float64

		for _, skill := range skills {
			hashkeysWithSkills[skill.Hashkey] = struct{}{}

			sumMu += skill.Mu
			sumSigma += skill.Sigma

			player := hashkeysToPlayers[skill.Hashkey]

			rng := rand.New(rand.NewSource(hashkeysToSeeds[skill.Hashkey]))
			player.Skill = math.Exp((rng.NormFloat64()*(skill.Sigma)+skill.Mu)/params.DefaultBeta)

			if player.Skill > maxSkill {
				maxSkill = player.Skill
			}
		}

		if len(skills) > 1 {
			// We DO have at least two data points to establish an average skill.
			// Update our parameters to reflect that.
			params.DefaultMu = sumMu / float64(len(skills))
			params.DefaultSigma = sumSigma / float64(len(skills))
		}
	}

	// Final pass to factor in the score
	if maxScorePerSecond == 0.0 {
		maxScorePerSecond = 1.0
	}

	scorePerSecondScale := maxSkill / float64(maxScorePerSecond)
	for hashkey, player := range hashkeysToPlayers {
		if _, ok := hashkeysWithSkills[hashkey]; !ok {
			// This player doesn't yet have a skill. Derive one from the default.
			// Use a consistent per-match seed
			rng := rand.New(rand.NewSource(hashkeysToSeeds[hashkey]))
			player.Skill = math.Exp((rng.NormFloat64()*params.DefaultSigma+params.DefaultMu)/params.DefaultBeta)
		}

		// Finally, add score to the linearized value
		scoreComponent := scorePerSecondScale * player.ScorePerSecond * params.ScoreFactor
		player.Skill += scoreComponent
	}

	// Determine which teams exist in the match. The teams are the ones declared
	// in the submission's team records (the "Q" lines); if none were declared,
	// fall back to the distinct teams the players are assigned to.
	teamSet := make(map[int]struct{})
	for _, tgs := range sub.TeamGameStats {
		teamSet[tgs.Team] = struct{}{}
	}

	if len(teamSet) == 0 {
		for _, player := range players {
			if player.Team > 0 {
				teamSet[player.Team] = struct{}{}
			}
		}
	}

	// Balancing only for two-team games. Single team games like
	// DM don't need sorting, and games with more than two teams are rare
	// enough that we defer them to a later implementation.
	if len(teamSet) == 2 {
		teamIDs := make([]int, 0, 2)
		for teamID := range teamSet {
			teamIDs = append(teamIDs, teamID)
		}
		sort.Ints(teamIDs)

		// Only balance the players who are already on one of the two teams.
		// Players not on a team (spectators, the unassigned) don't participate
		// and are reported with a team of zero.
		eligible := make([]*BalancePlayer, 0, len(players))
		for _, player := range players {
			if _, ok := teamSet[player.Team]; ok {
				eligible = append(eligible, player)
			} else {
				player.Team = 0
			}
		}

		if len(eligible) > 0 {
			items := make([]float64, len(eligible))
			for i, player := range eligible {
				items[i] = player.Skill
			}

			inTeamA := Partition(items)
			for i, player := range eligible {
				if inTeamA[i] {
					player.Team = teamIDs[0]
				} else {
					player.Team = teamIDs[1]
				}
			}
		}
	}

	return players, nil
}
