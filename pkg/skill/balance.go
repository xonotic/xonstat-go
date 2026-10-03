package skill

import (
	"fmt"
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
	// DefaultMu is the global new-player skill mean. It is the bottom rung of
	// the SPS ladder and the fallback for players whose score-per-second is not
	// yet meaningful. It is never overwritten with a match average.
	DefaultMu    float64
	DefaultSigma float64

	// ScoreFactor is how much weight to give the score as compared to the skill
	// value. 0 = raw skill only, 100 = 2x raw skill value. It applies to tracked
	// players only: anonymous (untracked) player estimates are driven by
	// score-per-second via the SPS ladder instead.
	ScoreFactor float64

	// ScoreConfidenceExponent controls how conservatively an anon's SPS-based
	// estimate is trusted. Confidence is (matchRampFactor * aliveRampFactor)
	// raised to this exponent, so larger values trust score-per-second more
	// slowly.
	ScoreConfidenceExponent float64

	// ScoreRampSeconds is the amount of play time after which score-per-second
	// is considered fully trustworthy. Both the match elapsed time and the
	// player's own alivetime ramp up to this value.
	ScoreRampSeconds float64

	// StabilityThreshold controls how much balance improvement is needed
	// to justify swapping players between teams. A value of 0.05 means
	// the new partition must improve the team sum difference by more than
	// 5% of total skill to be applied. 0 disables stability (always swap).
	StabilityThreshold float64

	// SigmaRange controls how many standard deviations around Mu the skill
	// sample may be drawn from. A value of 0 always samples Mu (the mean),
	// while larger values sample further into the distribution's tails.
	// Values are clamped to [0, 3]. The default of 2 preserves the current
	// behavior of sampling within two standard deviations.
	SigmaRange float64
}

// BalancePlayer is the output data structure that gets sent to
// the template, which is emitted back to the game server so it
// can reassign the player to the given team.
type BalancePlayer struct {
	Hashkey  string
	PlayerID int
	Nick     string
	// Skill is the effective skill used for balancing: for tracked players it
	// is a sample from their DB distribution plus any ScoreFactor bonus; for
	// anonymous players it is the confidence-gated SPS-ladder estimate.
	Skill          float64
	ScorePerSecond float64
	Team           int
	seed           int64
	hasDBSkill     bool
	dbMu           float64
	dbSigma        float64
	aliveTime      float64
}

// seed creates a per-player, per-match consistent value for seeding the RNG
func seed(hashkey, matchID string) int64 {
	// FNV1A64 non-cryptographic hash
	hash := fnv.New64a()
	hash.Write([]byte(hashkey))
	hash.Write([]byte(matchID))

	return int64(hash.Sum64())
}

// When debug is true, a summary of the input parameters, the effective skill
// ladder, and one line per eligible player are written to the standard logger.
func Balance(params BalanceParams, db SkillStore, sub *submission.Submission, maxDifference int, numTeams int, debug bool) ([]*BalancePlayer, error) {
	if debug {
		log.Printf("[balance] params: defaultMu=%g defaultSigma=%g scoreFactor=%g scoreConfidenceExponent=%g scoreRampSeconds=%g stabilityThreshold=%g sigmaRange=%g maxDifference=%d numTeams=%d",
			params.DefaultMu, params.DefaultSigma, params.ScoreFactor, params.ScoreConfidenceExponent,
			params.ScoreRampSeconds, params.StabilityThreshold, params.SigmaRange, maxDifference, numTeams)
	}

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
			bp.aliveTime = alivetimesecs
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

		for _, skill := range skills {
			i, ok := idx[skill.Hashkey]
			if !ok {
				continue
			}

			bp := players[i]
			bp.hasDBSkill = true
			bp.dbMu = skill.Mu
			bp.dbSigma = skill.Sigma

			// Take a sample from this player's Gaussian distribution to
			// get a single "skill" number. Use a consistent seed per match,
			// clamped to ±SigmaRange standard deviations.
			rng := rand.New(rand.NewSource(bp.seed))
			noise := math.Max(-params.SigmaRange, math.Min(params.SigmaRange, rng.NormFloat64()))
			bp.Skill = noise*skill.Sigma + skill.Mu
		}
	}

	// Score-per-second is scaled against this; avoid division by zero.
	if maxScorePerSecond == 0.0 {
		maxScorePerSecond = 1.0
	}

	// Tracked players receive up to ScoreFactor fraction of their own skill as
	// a bonus for in-match performance. Anonymous players do not: their
	// skill is derived via the SPS ladder below, and it already has the
	// scorefactor bits applied.
	for _, bp := range players {
		if bp.hasDBSkill {
			bp.Skill += bp.Skill * bp.ScorePerSecond / maxScorePerSecond * params.ScoreFactor
		}
	}

	// Build the SPS→skill ladder from the tracked players' effective skills
	// (sample + ScoreFactor), plus the unedited new-player default as the
	// bottom rung. Anonymous players are interpolated against this ladder.
	ladder := buildLadder(players, params.DefaultMu)

	if debug {
		rungs := make([]string, len(ladder))
		for i, rung := range ladder {
			rungs[i] = fmt.Sprintf("%.4f->%.2f", rung.sps, rung.skill)
		}
		log.Printf("[balance] ladder: %s", strings.Join(rungs, " "))
	}

	// The zero-reference fallback (no tracked players) differentiates anons
	// around the default using a normalised SPS offset.
	meanNormSPS := meanNormalizedSPS(players, teamSet, maxScorePerSecond)

	elapsed := 0.0
	if sub.Game.Duration != nil {
		elapsed = sub.Game.Duration.Seconds()
	}

	// Final pass: estimate anonymous players' skill from the SPS ladder,
	// gated by a confidence curve so early-match / freshly-joined players
	// stay near the default.
	for _, bp := range players {
		if bp.hasDBSkill {
			continue
		}

		var pairedSkill float64
		if len(ladder) > 1 {
			// If we have ladder values, we'll interpolate off of it.
			pairedSkill = targetNeighborInterp(ladder, bp.ScorePerSecond)
		} else {
			// If we do NOT have usable ladder values, we'll use the default Mu
			// and tweak the Sigma based on score-per-second performance.
			pairedSkill = zeroRefPairedSkill(
				bp.ScorePerSecond, maxScorePerSecond, meanNormSPS,
				params.DefaultMu, params.DefaultSigma,
			)
		}

		confidence := confidenceFromAlivetime(
			bp.aliveTime, elapsed, params.ScoreRampSeconds, params.ScoreConfidenceExponent,
		)

		bp.Skill = (1-confidence)*params.DefaultMu + confidence*pairedSkill
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

		if debug {
			for _, bp := range eligible {
				if bp.hasDBSkill {
					log.Printf("[balance] player: nick=%q team=%d alivetime=%.1f sps=%.4f mu=%g sigma=%g skill=%.2f",
						bp.Nick, bp.Team, bp.aliveTime, bp.ScorePerSecond, bp.dbMu, bp.dbSigma, bp.Skill)
				} else {
					log.Printf("[balance] player: nick=%q team=%d alivetime=%.1f sps=%.4f skill=%.2f",
						bp.Nick, bp.Team, bp.aliveTime, bp.ScorePerSecond, bp.Skill)
				}
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

			oldDiff := computeDiffFromTeams(eligible, prevTeamIDs)
			newDiff := computeDiffFromTeams(eligible, newTeamIDs)

			if shouldApplyNewPartition(oldDiff, newDiff, totalSkill, params.StabilityThreshold) {
				for i, bp := range eligible {
					bp.Team = newTeamIDs[i]
				}
			}
			// else: keep previous teams (already set on bp.Team)
		}
	}

	return players, nil
}

// skillLadder is a single reference point on the SPS→skill ladder. The skill
// value is an "effective skill": a tracked player's DB sample (noise*sigma+mu)
// plus any ScoreFactor bonus, i.e. exactly the weight that player carries into
// the partition.
type skillLadder struct {
	sps   float64
	skill float64
}

// buildLadder assembles the SPS→skill ladder: the unedited new-player default
// at SPS 0, plus one rung per tracked player. Rungs are sorted by SPS and any
// duplicate-SPS rungs are collapsed by averaging their skill.
func buildLadder(players []*BalancePlayer, defaultMu float64) []skillLadder {
	ladder := make([]skillLadder, 0, len(players)+1)
	ladder = append(ladder, skillLadder{sps: 0, skill: defaultMu})

	for _, bp := range players {
		if !bp.hasDBSkill {
			continue
		}
		ladder = append(ladder, skillLadder{sps: bp.ScorePerSecond, skill: bp.Skill})
	}

	sort.SliceStable(ladder, func(i, j int) bool { return ladder[i].sps < ladder[j].sps })

	// Collapse rungs that share an SPS. SPS is more volatile and ultimately
	// less reliable than the DB skill values, so when one SPS maps to several
	// DB skills we could bias down (pessimistic), bias up (optimistic), or
	// average. We average here for simplicity. Collapsing also keeps
	// targetNeighborInterp's interpolation line non-degenerate.
	collapsed := make([]skillLadder, 0, len(ladder))
	counts := make([]int, 0, len(ladder))
	for _, rung := range ladder {
		n := len(collapsed)
		if n > 0 && rung.sps == collapsed[n-1].sps {
			collapsed[n-1].skill = (collapsed[n-1].skill*float64(counts[n-1]) + rung.skill) / float64(counts[n-1]+1)
			counts[n-1]++
			continue
		}
		collapsed = append(collapsed, rung)
		counts = append(counts, 1)
	}

	return collapsed
}

// targetNeighborInterp estimates the skill for an SPS value by linearly
// interpolating between the two ladder rungs nearest to it in SPS. Ties are
// broken toward the lower-SPS rung.
func targetNeighborInterp(ladder []skillLadder, sps float64) float64 {
	if len(ladder) == 0 {
		return 0
	}
	if len(ladder) == 1 {
		return ladder[0].skill
	}

	idx := sort.Search(len(ladder), func(i int) bool { return ladder[i].sps >= sps })

	// Nearest rung.
	nearest := 0
	switch {
	case idx == 0:
		nearest = 0
	case idx >= len(ladder):
		nearest = len(ladder) - 1
	case sps-ladder[idx-1].sps <= ladder[idx].sps-sps:
		nearest = idx - 1
	default:
		nearest = idx
	}

	// The second-nearest rung is always adjacent to the nearest one.
	neighbor := 0
	switch {
	case nearest == 0:
		neighbor = 1
	case nearest == len(ladder)-1:
		neighbor = len(ladder) - 2
	case sps-ladder[nearest-1].sps <= ladder[nearest+1].sps-sps:
		neighbor = nearest - 1
	default:
		neighbor = nearest + 1
	}

	// Note: the ladder is built and collapsed such that each distinct
	// sps value has its own rung, so this could be "if neighbor == nearest"
	// but we'll keep this just in case this func ever gets called with
	// a ladder array that does NOT collapse in that fashion.
	a, b := ladder[neighbor], ladder[nearest]
	if a.sps == b.sps {
		return b.skill
	}

	result := a.skill + (b.skill-a.skill)*(sps-a.sps)/(b.sps-a.sps)

	// The interpolation intentionally allows an anon who outperforms every
	// tracked player to ride the line through the top two rungs and exceed the
	// top rung's skill. That is what lets anons performing near the strongest
	// tracked players land high enough to force the partition solver to split
	// the strong players. If estimates above the top rung ever prove
	// undesirable, enable this upper-bound clamp: it caps the estimate at
	// the best tracked player's skill, at the cost of that split behaviour.
	// if sps > ladder[len(ladder)-1].sps { return ladder[len(ladder)-1].skill }

	return result
}

// confidenceFromAlivetime returns how much an anonymous player's SPS-based
// estimate should be trusted. It is the product of two ramp factors, each of
// which reaches 1 once ScoreRampSeconds of the relevant time has elapsed:
//
//   - matchRampFactor: how long the match has been running, and
//   - aliveRampFactor: how much play time the player has personally accrued.
//
// The product is raised to the configured exponent. A non-positive elapsed time
// or ramp yields zero confidence (no SPS trust).
func confidenceFromAlivetime(alivetime, elapsed, rampSeconds, exponent float64) float64 {
	if elapsed <= 0 || rampSeconds <= 0 {
		return 0
	}

	matchRampFactor := math.Max(0, math.Min(1, elapsed/rampSeconds))
	aliveRampFactor := math.Max(0, math.Min(1, alivetime/rampSeconds))

	return math.Pow(matchRampFactor*aliveRampFactor, exponent)
}

// meanNormalizedSPS is the mean of sps/maxSPS across anonymous players that
// take part in balancing, used by the zero-reference fallback.
func meanNormalizedSPS(players []*BalancePlayer, teamSet map[int]struct{}, maxSPS float64) float64 {
	if maxSPS <= 0 {
		return 0
	}

	var sum float64
	var count int
	for _, bp := range players {
		if bp.hasDBSkill {
			continue
		}
		if len(teamSet) > 0 {
			if _, ok := teamSet[bp.Team]; !ok {
				continue
			}
		}
		sum += bp.ScorePerSecond / maxSPS
		count++
	}

	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

// zeroRefPairedSkill differentiates anonymous players around the default when
// there are no tracked players to build a ladder from.
func zeroRefPairedSkill(sps, maxSPS, meanNormSPS, defaultMu, defaultSigma float64) float64 {
	if maxSPS <= 0 {
		return defaultMu
	}
	normSPS := sps / maxSPS
	return defaultMu + (normSPS-meanNormSPS)*defaultSigma
}
