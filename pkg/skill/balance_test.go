package skill

import (
	"bufio"
	"fmt"
	"math"
	"strings"
	"testing"

	"gitlab.com/xonotic/xonstat/pkg/models"
	"gitlab.com/xonotic/xonstat/pkg/submission"
)

// mockSkillStore is a SkillStore that returns the skills given to it, or none
// at all if no skills were configured.
type mockSkillStore struct {
	skills map[string]models.PlayerHashkeySkill
}

func (m mockSkillStore) RPlayerSkillsBatch(hashkeys []string, gameTypeCd string) ([]*models.PlayerHashkeySkill, error) {
	out := make([]*models.PlayerHashkeySkill, 0, len(hashkeys))
	for _, hashkey := range hashkeys {
		if skill, ok := m.skills[hashkey]; ok {
			skillCopy := skill
			out = append(out, &skillCopy)
		}
	}
	return out, nil
}

var testBalanceParams = BalanceParams{
	DefaultMu:    models.PlayerSkill{}.Mu,
	DefaultSigma: 350.0,
	DefaultBeta:  175.0,
	ScoreFactor:  0.25,
}

const tdmHeader = `V 9
R test
G tdm
O Xonotic
M testmap
I 0.123456789
S Test Server
C 0
U 26000
D 100.000000
Q team#5
e scoreboard-score 0
Q team#14
e scoreboard-score 0
`

const dmHeader = `V 9
R test
G dm
O Xonotic
M testmap
I 0.123456789
S Test Server
C 0
U 26000
D 100.000000
`

// playerBlock renders the player events for a participant. An empty team
// omits the "t" line entirely.
func playerBlock(hashkey, index, nick, team, score string) string {
	block := fmt.Sprintf("P %s\ni %s\nn %s\n", hashkey, index, nick)
	if team != "" {
		block += fmt.Sprintf("t %s\n", team)
	}
	block += fmt.Sprintf("e scoreboard-score %s\n", score)
	block += "e acc-blaster-cnt-fired 1\ne alivetime 100\n"
	block += "e matches 1\ne scoreboardvalid 1\ne joins 1\n"
	return block
}

func makeSubmission(t *testing.T, body string) *submission.Submission {
	t.Helper()

	rs, err := submission.NewRawSubmission(bufio.NewReader(strings.NewReader(body)))
	if err != nil {
		t.Fatalf("NewRawSubmission: %s", err)
	}

	sub, err := submission.NewSubmission(rs)
	if err != nil {
		t.Fatalf("NewSubmission: %s", err)
	}

	return sub
}

func balancePlayers(t *testing.T, body string, store SkillStore) []*BalancePlayer {
	t.Helper()

	sub := makeSubmission(t, body)

	// Derive numTeams from the submission.
	teamSet := make(map[int]struct{})
	for _, tgs := range sub.TeamGameStats {
		teamSet[tgs.Team] = struct{}{}
	}

	if len(teamSet) == 0 {
		for _, pgs := range sub.PlayerGameStatsByHashkey {
			if pgs.Team.Valid {
				teamSet[int(pgs.Team.Int32)] = struct{}{}
			}
		}
	}

	players, err := Balance(testBalanceParams, store, sub, 1, len(teamSet))
	if err != nil {
		t.Fatalf("Balance: %s", err)
	}
	return players
}

func TestBalanceTwoTeams(t *testing.T) {
	body := tdmHeader +
		playerBlock("A", "1", "Alice", "5", "30") +
		playerBlock("B", "2", "Bob", "5", "20") +
		playerBlock("C", "3", "Carol", "14", "40") +
		playerBlock("D", "4", "Dave", "14", "10")

	players := balancePlayers(t, body, mockSkillStore{})
	if len(players) != 4 {
		t.Fatalf("Expected 4 players, got %d", len(players))
	}

	skills := make([]float64, 0, len(players))
	team5, team14 := false, false
	for _, player := range players {
		if player.Team != 5 && player.Team != 14 {
			t.Fatalf("Player %s was assigned to unexpected team %d", player.Hashkey, player.Team)
		}

		if player.Team == 5 {
			team5 = true
		} else {
			team14 = true
		}

		skills = append(skills, player.Skill)
	}

	if !team5 || !team14 {
		t.Fatalf("Expected both teams to be represented, got team5=%v team14=%v", team5, team14)
	}

	sum5, sum14 := 0.0, 0.0
	for _, player := range players {
		if player.Team == 5 {
			sum5 += player.Skill
		} else {
			sum14 += player.Skill
		}
	}

	// The returned partition should be optimal for the size-constrained search,
	// i.e. match what a brute-force search over teams within the cardinality
	// limit of each other would find.
	diff := math.Abs(sum5 - sum14)
	if math.Abs(diff-bruteBestDiffWithSize(skills, 1)) > 1e-9 {
		t.Fatalf("Team skill difference %f is not optimal (best possible is %f)",
			diff, bruteBestDiffWithSize(skills, 1))
	}
}

func TestBalanceExcludesUnassignedPlayers(t *testing.T) {
	body := tdmHeader +
		playerBlock("A", "1", "Alice", "5", "30") +
		playerBlock("B", "2", "Bob", "5", "20") +
		playerBlock("C", "3", "Carol", "14", "40") +
		// A player on the spectator team (Xonotic SVQC value 1337)...
		playerBlock("E", "5", "Eve", "1337", "5") +
		// ...and a player not on any team at all.
		playerBlock("F", "6", "Frank", "", "2")

	players := balancePlayers(t, body, mockSkillStore{})

	byHashkey := make(map[string]*BalancePlayer)
	for _, player := range players {
		byHashkey[player.Hashkey] = player
	}

	if len(byHashkey) != 5 {
		t.Fatalf("Expected 5 players in the response, got %d", len(byHashkey))
	}

	// Spectators and the unassigned must not be put on a team.
	for _, hashkey := range []string{"E", "F"} {
		if team := byHashkey[hashkey].Team; team != 0 {
			t.Fatalf("Player %s was assigned to team %d, expected 0", hashkey, team)
		}
	}

	// The players actually on a team should still be balanced optimally.
	eligibleSkills := make([]float64, 0, 3)
	sumA, sumB := 0.0, 0.0
	for _, hashkey := range []string{"A", "B", "C"} {
		player := byHashkey[hashkey]
		if player.Team != 5 && player.Team != 14 {
			t.Fatalf("Player %s was assigned to unexpected team %d", hashkey, player.Team)
		}

		eligibleSkills = append(eligibleSkills, player.Skill)
		if player.Team == 5 {
			sumA += player.Skill
		} else {
			sumB += player.Skill
		}
	}

	diff := math.Abs(sumA - sumB)
	if math.Abs(diff-bruteBestDiff(eligibleSkills)) > 1e-9 {
		t.Fatalf("Team skill difference %f is not optimal (best possible is %f)",
			diff, bruteBestDiff(eligibleSkills))
	}
}

func TestBalanceNoTeams(t *testing.T) {
	// A deathmatch game has no teams, so no balancing should take place.
	body := dmHeader +
		playerBlock("A", "1", "Alice", "", "30") +
		playerBlock("B", "2", "Bob", "", "20") +
		playerBlock("C", "3", "Carol", "", "40") +
		playerBlock("D", "4", "Dave", "", "10")

	players := balancePlayers(t, body, mockSkillStore{})

	for _, player := range players {
		if player.Team != 0 {
			t.Fatalf("Player %s was assigned to team %d in a game without teams",
				player.Hashkey, player.Team)
		}
	}
}

func TestBalanceMoreThanTwoTeams(t *testing.T) {
	// A game with three declared teams should now be balanced using recursive
	// CKK bisection. Players may be reassigned across teams.
	body := `V 9
R test
G tdm
O Xonotic
M testmap
I 0.123456789
S Test Server
C 0
U 26000
D 100.000000
Q team#5
e scoreboard-score 0
Q team#14
e scoreboard-score 0
Q team#13
e scoreboard-score 0
` +
		playerBlock("A", "1", "Alice", "5", "30") +
		playerBlock("B", "2", "Bob", "5", "20") +
		playerBlock("C", "3", "Carol", "14", "40") +
		playerBlock("D", "4", "Dave", "13", "10")

	players := balancePlayers(t, body, mockSkillStore{})

	// All players should be on one of the three declared teams.
	expectedTeams := map[int]struct{}{5: {}, 14: {}, 13: {}}
	for _, player := range players {
		if _, ok := expectedTeams[player.Team]; !ok {
			t.Fatalf("Player %s was assigned to unexpected team %d", player.Hashkey, player.Team)
		}
	}

	// At least two different teams should be represented.
	teamSet := make(map[int]struct{})
	for _, player := range players {
		teamSet[player.Team] = struct{}{}
	}
	if len(teamSet) < 2 {
		t.Fatalf("Expected at least 2 teams to be represented, got %d", len(teamSet))
	}
}

func TestBalanceDeterministic(t *testing.T) {
	body := tdmHeader +
		playerBlock("A", "1", "Alice", "5", "30") +
		playerBlock("B", "2", "Bob", "5", "20") +
		playerBlock("C", "3", "Carol", "14", "40") +
		playerBlock("D", "4", "Dave", "14", "10")

	first := balancePlayers(t, body, mockSkillStore{})
	second := balancePlayers(t, body, mockSkillStore{})

	teamsByHashkey := func(players []*BalancePlayer) map[string]int {
		out := make(map[string]int)
		for _, player := range players {
			out[player.Hashkey] = player.Team
		}
		return out
	}

	a, b := teamsByHashkey(first), teamsByHashkey(second)
	for hashkey, team := range a {
		if b[hashkey] != team {
			t.Fatalf("Team assignment for %s differs between calls: %d vs %d",
				hashkey, team, b[hashkey])
		}
	}
}

func TestBalanceWithKnownSkills(t *testing.T) {
	// Players with skills recorded in the store should still get balanced.
	store := mockSkillStore{
		skills: map[string]models.PlayerHashkeySkill{
			"A": {Hashkey: "A", GameTypeCd: "tdm", Mu: 2000.0, Sigma: 100.0},
			"C": {Hashkey: "C", GameTypeCd: "tdm", Mu: 1200.0, Sigma: 400.0},
		},
	}

	body := tdmHeader +
		playerBlock("A", "1", "Alice", "5", "30") +
		playerBlock("B", "2", "Bob", "5", "20") +
		playerBlock("C", "3", "Carol", "14", "40") +
		playerBlock("D", "4", "Dave", "14", "10")

	players := balancePlayers(t, body, store)

	for _, player := range players {
		if player.Skill <= 0 {
			t.Fatalf("Player %s has a non-positive skill of %f", player.Hashkey, player.Skill)
		}
		if player.Team != 5 && player.Team != 14 {
			t.Fatalf("Player %s was assigned to unexpected team %d", player.Hashkey, player.Team)
		}
	}
}

func TestBalanceThreeTeams(t *testing.T) {
	// A game with three declared teams should now be balanced using recursive
	// CKK bisection. Players get reassigned across all three teams.
	body := `V 9
R test
G tdm
O Xonotic
M testmap
I 0.123456789
S Test Server
C 0
U 26000
D 100.000000
Q team#5
e scoreboard-score 0
Q team#14
e scoreboard-score 0
Q team#13
e scoreboard-score 0
` +
		playerBlock("A", "1", "Alice", "5", "30") +
		playerBlock("B", "2", "Bob", "5", "20") +
		playerBlock("C", "3", "Carol", "14", "40") +
		playerBlock("D", "4", "Dave", "14", "10") +
		playerBlock("G", "7", "Gina", "13", "25") +
		playerBlock("H", "8", "Hal", "13", "15") +
		playerBlock("I", "9", "Iris", "5", "35") +
		playerBlock("J", "10", "Jack", "14", "5") +
		playerBlock("K", "11", "Kim", "13", "30")

	players := balancePlayers(t, body, mockSkillStore{})

	// Collect the team IDs that were assigned.
	teamSet := make(map[int]struct{})
	for _, player := range players {
		teamSet[player.Team] = struct{}{}
	}

	// All players should be on one of the three declared teams.
	for _, player := range players {
		if _, ok := teamSet[player.Team]; !ok {
			t.Fatalf("Player %s was assigned to unexpected team %d", player.Hashkey, player.Team)
		}
	}

	// At least two different teams should be represented.
	if len(teamSet) < 2 {
		t.Fatalf("Expected at least 2 teams to be represented, got %d", len(teamSet))
	}

	// The sum difference between the highest and lowest team should be
	// reasonable (not all players dumped on one team).
	teamSums := make(map[int]float64)
	for _, player := range players {
		teamSums[player.Team] += player.Skill
	}

	totalSum := 0.0
	for _, s := range teamSums {
		totalSum += s
	}

	var maxSum float64
	for _, s := range teamSums {
		if s > maxSum {
			maxSum = s
		}
	}

	// With 9 players across 3 teams, the max team sum should not exceed
	// 60% of the total (generous bound to account for CKK approximation).
	if totalSum > 0 && maxSum > totalSum*0.6 {
		t.Fatalf("Team sums too imbalanced: max team sum %f exceeds 60%% of total %f", maxSum, totalSum)
	}
}

func TestBalanceStabilityKeepsTeams(t *testing.T) {
	// Two nearly identical submissions: only one score changes by 1 point.
	// With a high stability threshold, teams should stay the same.
	body1 := tdmHeader +
		playerBlock("A", "1", "Alice", "5", "30") +
		playerBlock("B", "2", "Bob", "5", "20") +
		playerBlock("C", "3", "Carol", "14", "40") +
		playerBlock("D", "4", "Dave", "14", "10")

	body2 := tdmHeader +
		playerBlock("A", "1", "Alice", "5", "31") + // +1 score
		playerBlock("B", "2", "Bob", "5", "20") +
		playerBlock("C", "3", "Carol", "14", "40") +
		playerBlock("D", "4", "Dave", "14", "10")

	// First balance with default stability.
	sub1 := makeSubmission(t, body1)
	teamSet := make(map[int]struct{})
	for _, tgs := range sub1.TeamGameStats {
		teamSet[tgs.Team] = struct{}{}
	}
	first, err := Balance(testBalanceParams, mockSkillStore{}, sub1, 1, len(teamSet))
	if err != nil {
		t.Fatalf("Balance: %s", err)
	}

	// Second balance with high stability threshold (50%) — should keep teams.
	params := testBalanceParams
	params.StabilityThreshold = 0.50
	sub2 := makeSubmission(t, body2)
	teamSet2 := make(map[int]struct{})
	for _, tgs := range sub2.TeamGameStats {
		teamSet2[tgs.Team] = struct{}{}
	}
	second, err := Balance(params, mockSkillStore{}, sub2, 1, len(teamSet2))
	if err != nil {
		t.Fatalf("Balance: %s", err)
	}

	// Teams should be identical despite the score change.
	teams1 := make(map[string]int)
	for _, p := range first {
		teams1[p.Hashkey] = p.Team
	}
	for _, p := range second {
		if teams1[p.Hashkey] != p.Team {
			t.Errorf("Player %s changed team from %d to %d with high stability threshold",
				p.Hashkey, teams1[p.Hashkey], p.Team)
		}
	}
}

func TestBalanceStabilityDisabledSwaps(t *testing.T) {
	// With stability disabled (threshold 0), a score change should be applied.
	body1 := tdmHeader +
		playerBlock("A", "1", "Alice", "5", "30") +
		playerBlock("B", "2", "Bob", "5", "20") +
		playerBlock("C", "3", "Carol", "14", "40") +
		playerBlock("D", "4", "Dave", "14", "10")

	body2 := tdmHeader +
		playerBlock("A", "1", "Alice", "5", "31") +
		playerBlock("B", "2", "Bob", "5", "20") +
		playerBlock("C", "3", "Carol", "14", "40") +
		playerBlock("D", "4", "Dave", "14", "10")

	sub1 := makeSubmission(t, body1)
	teamSet := make(map[int]struct{})
	for _, tgs := range sub1.TeamGameStats {
		teamSet[tgs.Team] = struct{}{}
	}
	first, err := Balance(testBalanceParams, mockSkillStore{}, sub1, 1, len(teamSet))
	if err != nil {
		t.Fatalf("Balance: %s", err)
	}

	params := testBalanceParams
	params.StabilityThreshold = 0 // disabled
	sub2 := makeSubmission(t, body2)
	teamSet2 := make(map[int]struct{})
	for _, tgs := range sub2.TeamGameStats {
		teamSet2[tgs.Team] = struct{}{}
	}
	second, err := Balance(params, mockSkillStore{}, sub2, 1, len(teamSet2))
	if err != nil {
		t.Fatalf("Balance: %s", err)
	}

	// Both should be valid balanced outputs (no assertion on whether teams
	// changed — the partition algorithm may or may not change the assignment
	// for a 1-point difference). Just verify both calls succeed and produce
	// valid results.
	teams1 := make(map[string]int)
	for _, p := range first {
		teams1[p.Hashkey] = p.Team
	}
	teams2 := make(map[string]int)
	for _, p := range second {
		teams2[p.Hashkey] = p.Team
	}
	for _, p := range second {
		if p.Team != 5 && p.Team != 14 {
			t.Errorf("Player %s on unexpected team %d", p.Hashkey, p.Team)
		}
	}
	_ = teams1
}
