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
	DefaultMu    float64
	DefaultSigma float64
	DefaultBeta  float64

	// How much weight to give the score as compared to the skill value.
	// 0 = raw skill only, 100 = 2x raw skill value
	ScoreFactor float64

	// StabilityThreshold controls how much balance improvement is needed
	// to justify swapping players between teams. A value of 0.05 means
	// the new partition must improve the team sum difference by more than
	// 5% of total skill to be applied. 0 disables stability (always swap).
	StabilityThreshold float64
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
	Team           int
	seed           int64
	hasDBSkill     bool
}

// seed creates a per-player, per-match consistent value for seeding the RNG
func seed(hashkey, matchID string) int64 {
	// FNV1A64 non-cryptographic hash
	hash := fnv.New64a()
	hash.Write([]byte(hashkey))
	hash.Write([]byte(matchID))

	return int64(hash.Sum64())
}

func Balance(params BalanceParams, db SkillStore, sub *submission.Submission, maxDifference int, numTeams int) ([]*BalancePlayer, error) {
	// The destination for the balanced (sorted) players.
	players := make([]*BalancePlayer, 0)

	// Hashkeys for tracked players that might have skill records in the DB.
	trackedHashkeys := make([]string, 0)

	// Keep track of the max score-per-second, for use in scaling later.
	var maxScorePerSecond float64

	// Determine which teams exist in the match. The teams are the ones declared
	// in the submission's team records (the "Q" lines); if none were declared,
	// we'll fall back to the distinct teams the players are assigned to.
	teamSet := make(map[int]struct{})
	for _, tgs := range sub.TeamGameStats {
		teamSet[tgs.Team] = struct{}{}
	}

	// First pass: build player records, compute SPS, and track DB lookups.
	// Skip bots, players without stats, and ineligible (off-team) players.
	for hashkey, player := range sub.PlayersByHashkey {
		gamestat, hasGameStat := sub.PlayerGameStatsByHashkey[hashkey]
		if strings.HasPrefix(hashkey, "bot#") || !hasGameStat {
			continue
		}

		team := 0
		if gamestat.Team.Valid {
			team = int(gamestat.Team.Int32)
		}

		// Build the BalancePlayer entry for every human, but only compute
		// stats for players who are actually on a declared team.
		bp := &BalancePlayer{
			Hashkey:  hashkey,
			PlayerID: player.PlayerID,
			Nick:     player.Nick.String,
			Team:     team,
			seed:     seed(hashkey, sub.Game.MatchID.String),
		}

		_, onTeam := teamSet[team]
		if len(teamSet) == 0 || onTeam {
			score := gamestat.Score.Int32

			alivetimesecs := 0.0
			if gamestat.AliveTime != nil {
				alivetimesecs = gamestat.AliveTime.Seconds()
			}

			sps := 0.0
			if alivetimesecs > 0 {
				sps = float64(score) / alivetimesecs
			}

			bp.ScorePerSecond = sps
			if sps > maxScorePerSecond {
				maxScorePerSecond = sps
			}

			if !strings.HasPrefix(hashkey, "player#") {
				trackedHashkeys = append(trackedHashkeys, hashkey)
			}
		}

		players = append(players, bp)
	}

	// Use a deterministic order (by PlayerID) for the rest of the calculation so the team
	// assignment doesn't depend on map iteration order.
	sort.Slice(players, func(i, j int) bool { return players[i].PlayerID < players[j].PlayerID })

	// Those that are tracked are the ones that might have skill values in the DB.
	// We query for them in batch to decrease round-trips.
	if len(trackedHashkeys) > 0 {
		skills, err := db.RPlayerSkillsBatch(trackedHashkeys, sub.Game.GameTypeCd)
		if err != nil {
			log.Printf("Error: %s", err)
			return nil, err
		}

		// Build a hashkey→index map for O(1) skill lookup.
		idx := make(map[string]int, len(players))
		for i, p := range players {
			idx[p.Hashkey] = i
		}

		var sumMu, sumSigma float64

		for _, skill := range skills {
			i, ok := idx[skill.Hashkey]
			if !ok {
				continue
			}

			bp := players[i]
			bp.hasDBSkill = true

			sumMu += skill.Mu
			sumSigma += skill.Sigma

			// Take a sample from this player's Gaussian distribution to
			// get a single "skill" number. Use a consistent seed per match.
			rng := rand.New(rand.NewSource(bp.seed))
			noise := math.Max(-2.0, math.Min(2.0, rng.NormFloat64()))
			bp.Skill = noise*skill.Sigma + skill.Mu
		}

		if len(skills) > 1 {
			// We DO have at least two data points to establish an average skill.
			// Update our parameters to reflect that.
			params.DefaultMu = sumMu / float64(len(skills))
			params.DefaultSigma = sumSigma / float64(len(skills))
		}
	}

	// Final pass: assign default skills and factor in the score.
	if maxScorePerSecond == 0.0 {
		maxScorePerSecond = 1.0
	}

	for _, bp := range players {
		if !bp.hasDBSkill {
			// Players not having skill values are sampled from the
			// default parameters, which represent the average Mu and
			// Sigma values we've seen from the other players in the match.
			rng := rand.New(rand.NewSource(bp.seed))
			noise := math.Max(-2.0, math.Min(2.0, rng.NormFloat64()))
			bp.Skill = noise*params.DefaultSigma + params.DefaultMu
		}

		// Each player receives up to ScoreFactor fraction of their own skill as
		// a bonus for in-match performance.
		bp.Skill += bp.Skill * bp.ScorePerSecond / maxScorePerSecond * params.ScoreFactor
	}

	// Fallback team detection: if no Q lines declared teams, derive them
	// from the distinct teams the players are assigned to.
	if len(teamSet) == 0 {
		for _, bp := range players {
			if bp.Team > 0 {
				teamSet[bp.Team] = struct{}{}
			}
		}
	}

	// Balancing is for team games only.
	if len(teamSet) >= 2 {
		teamIDs := make([]int, 0, len(teamSet))
		for teamID := range teamSet {
			teamIDs = append(teamIDs, teamID)
		}
		sort.Ints(teamIDs)

		// Only balance the players who are already on one of the declared teams.
		// Players not on a team (spectators, the unassigned) don't participate
		// and are reported with a team of zero.
		eligible := make([]*BalancePlayer, 0, len(players))
		for _, bp := range players {
			if _, ok := teamSet[bp.Team]; ok {
				eligible = append(eligible, bp)
			} else {
				bp.Team = 0
			}
		}

		if len(eligible) > 0 {
			items := make([]float64, len(eligible))
			for i, bp := range eligible {
				items[i] = bp.Skill
			}

			actualTeams := len(teamIDs)
			if numTeams > 0 && numTeams < actualTeams {
				actualTeams = numTeams
			}

			// Save each player's current team before overwriting.
			prevTeamIDs := make([]int, len(eligible))
			for i, bp := range eligible {
				prevTeamIDs[i] = bp.Team
			}

			// Assign teams using one of the three partitioning algorithms, which depend
			// upon the number of players and the number of teams.
			teamAssignments := Partition(items, actualTeams, maxDifference)

			// Relabel: find the bijection from partition indices to game team IDs
			// that minimises the number of players who change teams.
			newTeamIDs := minimizeSwaps(teamAssignments, teamIDs, prevTeamIDs)

			// Stability check: only apply the new assignment if it improves
			// balance by more than the threshold fraction of total skill.
			totalSkill := 0.0
			for _, bp := range eligible {
				totalSkill += math.Abs(bp.Skill)
			}

			threshold := params.StabilityThreshold
			if threshold == 0 {
				threshold = defaultStabilityThreshold
			}

			oldDiff := computeDiffFromTeams(eligible, prevTeamIDs)
			newDiff := computeDiffFromTeams(eligible, newTeamIDs)

			if shouldApplyNewPartition(oldDiff, newDiff, totalSkill, threshold) {
				for i, bp := range eligible {
					bp.Team = newTeamIDs[i]
				}
			}
			// else: keep previous teams (already set on bp.Team)
		}
	}

	return players, nil
}
