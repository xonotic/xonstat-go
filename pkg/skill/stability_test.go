package skill

import "testing"

func TestMinimizeSwapsIdentical(t *testing.T) {
	// Players already on the optimal teams: zero swaps needed.
	assignments := []int{0, 0, 1, 1}
	teamIDs := []int{5, 14}
	prevTeamIDs := []int{5, 5, 14, 14}

	result := minimizeSwaps(assignments, teamIDs, prevTeamIDs)

	for i, expected := range prevTeamIDs {
		if result[i] != expected {
			t.Errorf("player %d: expected team %d, got %d", i, expected, result[i])
		}
	}
}

func TestMinimizeSwapsRelabel(t *testing.T) {
	// Partition assigns group 0 to players on team 14, group 1 to team 5.
	// Without relabeling, all 4 players swap. With relabeling, zero swaps.
	assignments := []int{1, 1, 0, 0} // group 0 → first player pair, group 1 → second
	teamIDs := []int{5, 14}          // group 0 → 5, group 1 → 14
	prevTeamIDs := []int{14, 14, 5, 5}

	result := minimizeSwaps(assignments, teamIDs, prevTeamIDs)

	for i, expected := range prevTeamIDs {
		if result[i] != expected {
			t.Errorf("player %d: expected team %d, got %d", i, expected, result[i])
		}
	}
}

func TestMinimizeSwapsPartialSwaps(t *testing.T) {
	// 2 teams: partition groups player 2 with team 10, but player 2 was on team 20.
	// Best relabeling keeps players 0,1 on their teams; player 2 must swap.
	assignments := []int{0, 1, 0}
	teamIDs := []int{10, 20}
	prevTeamIDs := []int{10, 20, 20}

	result := minimizeSwaps(assignments, teamIDs, prevTeamIDs)

	swaps := 0
	for i := range result {
		if result[i] != prevTeamIDs[i] {
			swaps++
		}
	}
	if swaps != 1 {
		t.Errorf("expected 1 swap, got %d (result: %v)", swaps, result)
	}
}

func TestMinimizeSwapsSingleTeam(t *testing.T) {
	assignments := []int{0, 0, 0}
	teamIDs := []int{5}
	prevTeamIDs := []int{5, 5, 5}

	result := minimizeSwaps(assignments, teamIDs, prevTeamIDs)
	for i, expected := range prevTeamIDs {
		if result[i] != expected {
			t.Errorf("player %d: expected team %d, got %d", i, expected, result[i])
		}
	}
}

func TestComputeDiffFromTeams(t *testing.T) {
	players := []*BalancePlayer{
		{Skill: 10},
		{Skill: 20},
		{Skill: 30},
		{Skill: 40},
	}
	// Team 5: 10+40=50, Team 14: 20+30=50 → diff = 0
	teamIDs := []int{5, 14, 14, 5}
	diff := computeDiffFromTeams(players, teamIDs)
	if diff != 0 {
		t.Errorf("expected diff 0, got %f", diff)
	}

	// Team 5: 10+20=30, Team 14: 30+40=70 → diff = 40
	teamIDs2 := []int{5, 5, 14, 14}
	diff2 := computeDiffFromTeams(players, teamIDs2)
	if diff2 != 40 {
		t.Errorf("expected diff 40, got %f", diff2)
	}
}

func TestComputeDiffFromTeamsThreeTeams(t *testing.T) {
	players := []*BalancePlayer{
		{Skill: 10},
		{Skill: 20},
		{Skill: 30},
		{Skill: 40},
	}
	teamIDs := []int{1, 1, 2, 3}
	// Team 1: 30, Team 2: 30, Team 3: 40 → diff = 10
	diff := computeDiffFromTeams(players, teamIDs)
	if diff != 10 {
		t.Errorf("expected diff 10, got %f", diff)
	}
}

func TestShouldApplyNewPartition(t *testing.T) {
	// Improvement of 10 with total skill 100 and threshold 5% → 10 > 5 → true
	if !shouldApplyNewPartition(50, 40, 100, 0.05) {
		t.Error("expected true: improvement exceeds threshold")
	}

	// Improvement of 2 with total skill 100 and threshold 5% → 2 > 5 → false
	if shouldApplyNewPartition(50, 48, 100, 0.05) {
		t.Error("expected false: improvement below threshold")
	}

	// Threshold 0 → always apply
	if !shouldApplyNewPartition(50, 50, 100, 0) {
		t.Error("expected true: threshold 0 means always apply")
	}

	// Zero total skill → always apply
	if !shouldApplyNewPartition(0, 0, 0, 0.05) {
		t.Error("expected true: zero total skill")
	}
}

func TestGeneratePermutations(t *testing.T) {
	perms := generatePermutations(3)
	if len(perms) != 6 {
		t.Errorf("expected 6 permutations for n=3, got %d", len(perms))
	}

	// Check that all permutations are unique.
	seen := make(map[string]bool)
	for _, p := range perms {
		key := ""
		for _, v := range p {
			key += string(rune('0' + v))
		}
		if seen[key] {
			t.Errorf("duplicate permutation: %v", p)
		}
		seen[key] = true
	}
}

func TestMinimizeSwapsThreeTeams(t *testing.T) {
	// 3 teams, partition has group labels swapped vs previous.
	assignments := []int{2, 0, 1}
	teamIDs := []int{10, 20, 30}
	prevTeamIDs := []int{10, 20, 30}

	result := minimizeSwaps(assignments, teamIDs, prevTeamIDs)

	// Best mapping should match all 3 previous teams.
	swaps := 0
	for i := range result {
		if result[i] != prevTeamIDs[i] {
			swaps++
		}
	}
	if swaps != 0 {
		t.Errorf("expected 0 swaps with 3-team relabeling, got %d (result: %v)", swaps, result)
	}
}
