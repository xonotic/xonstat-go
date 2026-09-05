package skill

import "math"

// defaultStabilityThreshold is the fraction of total skill above which a
// partition improvement is considered worth applying (i.e. worth swapping
// players for). A value of 0.05 means 5% of total skill.
const defaultStabilityThreshold = 0.05

// minimizeSwaps finds the bijection from partition indices to game team IDs
// that minimises the number of players who change teams. It returns the
// actual team ID assigned to each player. For T teams it tries all T!
// permutations of the index→teamID mapping.
func minimizeSwaps(assignments []int, teamIDs []int, prevTeamIDs []int) []int {
	t := len(teamIDs)
	n := len(assignments)
	if t <= 1 || n == 0 {
		result := make([]int, n)
		for i, idx := range assignments {
			if idx < t {
				result[i] = teamIDs[idx]
			} else {
				result[i] = teamIDs[0]
			}
		}
		return result
	}

	bestSwaps := n + 1
	bestResult := make([]int, n)

	for _, perm := range generatePermutations(t) {
		swaps := 0
		for i, idx := range assignments {
			if idx < t && teamIDs[perm[idx]] != prevTeamIDs[i] {
				swaps++
			}
		}

		if swaps < bestSwaps {
			bestSwaps = swaps
			for i, idx := range assignments {
				if idx < t {
					bestResult[i] = teamIDs[perm[idx]]
				} else {
					bestResult[i] = teamIDs[0]
				}
			}

			if bestSwaps == 0 {
				break
			}
		}
	}
	return bestResult
}

// generatePermutations returns all permutations of {0, 1, ..., n-1}.
func generatePermutations(n int) [][]int {
	if n <= 0 {
		return [][]int{{}}
	}
	result := make([][]int, 0, factorial(n))
	genPermHelper(make([]int, n), make([]bool, n), 0, &result)
	return result
}

func genPermHelper(current []int, used []bool, depth int, result *[][]int) {
	if depth == len(current) {
		perm := make([]int, len(current))
		copy(perm, current)
		*result = append(*result, perm)
		return
	}
	for i := 0; i < len(current); i++ {
		if used[i] {
			continue
		}
		used[i] = true
		current[depth] = i
		genPermHelper(current, used, depth+1, result)
		used[i] = false
	}
}

func factorial(n int) int {
	result := 1
	for i := 2; i <= n; i++ {
		result *= i
	}
	return result
}

// computeDiffFromTeams returns max(teamSum) - min(teamSum) given each
// player's team assignment (as actual team IDs).
func computeDiffFromTeams(eligible []*BalancePlayer, teamIDs []int) float64 {
	if len(eligible) == 0 {
		return 0
	}

	sums := make(map[int]float64)
	for i, bp := range eligible {
		sums[teamIDs[i]] += bp.Skill
	}

	minSum := math.Inf(1)
	maxSum := math.Inf(-1)
	for _, s := range sums {
		if s < minSum {
			minSum = s
		}
		if s > maxSum {
			maxSum = s
		}
	}
	return maxSum - minSum
}

// shouldApplyNewPartition returns true if the new partition improves balance
// by more than the threshold fraction of total skill.
func shouldApplyNewPartition(oldDiff, newDiff, totalSkill, threshold float64) bool {
	if threshold <= 0 || totalSkill == 0 {
		return true
	}
	return (oldDiff - newDiff) > threshold*totalSkill
}
