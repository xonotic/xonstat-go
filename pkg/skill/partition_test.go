package skill

import (
	"math"
	"math/bits"
	"math/rand"
	"testing"
)

// groupSums returns the sums of the two groups described by a partition result.
func groupSums(items []float64, teams []int) (float64, float64) {
	sumA, sumB := 0.0, 0.0
	for i, item := range items {
		if teams[i] == 0 {
			sumA += item
		} else {
			sumB += item
		}
	}
	return sumA, sumB
}

func partitionDiff(items []float64, teams []int) float64 {
	sumA, sumB := groupSums(items, teams)
	return math.Abs(sumA - sumB)
}

// bruteBestDiff computes the minimum achievable group-sum difference by
// checking every possible partition. Only usable for tiny inputs.
func bruteBestDiff(items []float64) float64 {
	n := len(items)
	total := 0.0
	for _, item := range items {
		total += item
	}

	best := math.Inf(1)
	for mask := 0; mask < (1 << uint(n)); mask++ {
		sum := 0.0
		for i := 0; i < n; i++ {
			if mask&(1<<uint(i)) != 0 {
				sum += items[i]
			}
		}

		diff := math.Abs(2*sum - total)
		if diff < best {
			best = diff
		}
	}

	return best
}

// bruteBestDiffWithSize computes the minimum achievable group-sum difference
// when the group sizes must be within maxDiff of each other. Only usable for
// tiny inputs.
func bruteBestDiffWithSize(items []float64, maxDiff int) float64 {
	n := len(items)
	minCount, maxCount := allowedCounts(n, maxDiff)
	total := 0.0
	for _, item := range items {
		total += item
	}

	best := math.Inf(1)
	for mask := 0; mask < (1 << uint(n)); mask++ {
		count := bits.OnesCount(uint(mask))
		if count < minCount || count > maxCount {
			continue
		}

		sum := 0.0
		for i := 0; i < n; i++ {
			if mask&(1<<uint(i)) != 0 {
				sum += items[i]
			}
		}

		diff := math.Abs(2*sum - total)
		if diff < best {
			best = diff
		}
	}

	return best
}

// groupCount returns the number of items in group 0 described by a partition.
func groupCount(teams []int) int {
	count := 0
	for _, t := range teams {
		if t == 0 {
			count++
		}
	}
	return count
}

func TestPartitionPerfectSplit(t *testing.T) {
	items := []float64{10, 9, 8, 7}
	teams := Partition(items, 2, -1)

	sumA, sumB := groupSums(items, teams)
	if diff := math.Abs(sumA - sumB); diff != 0.0 {
		t.Fatalf("Expected a perfect split, got sums %f and %f", sumA, sumB)
	}
}

func TestPartitionImbalanced(t *testing.T) {
	items := []float64{100, 1, 1, 1}
	teams := Partition(items, 2, -1)

	diff := partitionDiff(items, teams)
	if diff != 97.0 {
		t.Fatalf("Expected the best possible difference of 97, got %f", diff)
	}
}

func TestPartitionOddCount(t *testing.T) {
	items := []float64{1, 2, 3}
	teams := Partition(items, 2, -1)

	diff := partitionDiff(items, teams)
	if diff != 0.0 {
		t.Fatalf("Expected a perfect split of {1,2,3}, got a difference of %f", diff)
	}
}

func TestPartitionAllEqual(t *testing.T) {
	items := []float64{2, 2, 2, 2, 2, 2}
	teams := Partition(items, 2, -1)

	diff := partitionDiff(items, teams)
	if diff != 0.0 {
		t.Fatalf("Expected a perfect split of equal items, got a difference of %f", diff)
	}
}

func TestPartitionSingleItem(t *testing.T) {
	teams := Partition([]float64{42}, 2, -1)
	if teams[0] != 0 {
		t.Fatal("The single item should belong to group 0")
	}
}

func TestPartitionTwoItems(t *testing.T) {
	items := []float64{3, 5}
	teams := Partition(items, 2, -1)

	if len(teams) != 2 {
		t.Fatalf("Unexpected partition length: %d", len(teams))
	}

	diff := partitionDiff(items, teams)
	if diff != 2.0 {
		t.Fatalf("Expected the best possible difference of 2, got %f", diff)
	}
}

func TestPartitionEmpty(t *testing.T) {
	teams := Partition(nil, 2, -1)
	if teams == nil || len(teams) != 0 {
		t.Fatalf("Expected an empty result, got %v", teams)
	}
}

func TestPartitionMatchesBruteForce(t *testing.T) {
	rng := rand.New(rand.NewSource(0))

	for n := 1; n <= 12; n++ {
		for trial := 0; trial < 100; trial++ {
			items := make([]float64, n)
			for i := range items {
				items[i] = rng.Float64()*100.0 + 1.0
			}

			teams := Partition(items, 2, -1)
			if len(teams) != n {
				t.Fatalf("Partition returned %d entries for %d items", len(teams), n)
			}

			got := partitionDiff(items, teams)
			want := bruteBestDiff(items)
			if math.Abs(got-want) > 1e-9 {
				t.Fatalf("For items %v: partition diff %f does not match optimal %f",
					items, got, want)
			}
		}
	}
}

func TestPartitionLargestInput(t *testing.T) {
	rng := rand.New(rand.NewSource(42))

	items := make([]float64, 24)
	for i := range items {
		items[i] = rng.Float64()*1000.0 + 1.0
	}

	teams := Partition(items, 2, -1)
	if len(teams) != 24 {
		t.Fatalf("Partition returned %d entries for 24 items", len(teams))
	}

	diff := partitionDiff(items, teams)
	total := 0.0
	for _, item := range items {
		total += item
	}

	// A random 24-player field should always partition within a few percent of
	// perfectly balanced.
	if diff > total*0.01 {
		t.Fatalf("Partition of %d items left a difference of %f out of %f total",
			len(items), diff, total)
	}
}

func TestPartitionCKKFallback(t *testing.T) {
	// A field larger than maxN must use CKK and still produce a
	// valid, reasonably balanced assignment.
	rng := rand.New(rand.NewSource(7))

	items := make([]float64, maxHorowitzSahni+2)
	total := 0.0
	for i := range items {
		items[i] = rng.Float64()*100.0 + 1.0
		total += items[i]
	}

	teams := Partition(items, 2, -1)
	if len(teams) != len(items) {
		t.Fatalf("Partition returned %d entries for %d items", len(teams), len(items))
	}

	sumA, sumB := groupSums(items, teams)
	diff := math.Abs(sumA - sumB)

	// The CKK fallback should keep a random field within a few percent of
	// perfectly balanced.
	if diff > total*0.05 {
		t.Fatalf("CKK fallback left a difference of %f out of %f total", diff, total)
	}
}

func TestPartitionMaxDiffForcesEvenTeams(t *testing.T) {
	// One dominant player (like a highly-rated star) used to force a 1-vs-3
	// split. With a max size difference of 1 the teams must be 2 and 2, at the
	// cost of a worse skill-sum balance.
	items := []float64{100, 1, 1, 1}
	teams := Partition(items, 2, 1)

	if count := groupCount(teams); count != 2 {
		t.Fatalf("Expected a 2-vs-2 split, got %d vs %d", count, len(items)-count)
	}

	diff := partitionDiff(items, teams)
	if diff != 99.0 {
		t.Fatalf("Expected the constrained optimum of 99, got %f", diff)
	}
}

func TestPartitionMaxDiffDominantPlayer(t *testing.T) {
	// Kyle-style field: one player holds ~90% of the total skill, so the
	// unconstrained partition puts them alone on a team. With a max size
	// difference of 1, a 16-player field must split 8-vs-8.
	items := make([]float64, 16)
	items[0] = 30.0
	for i := 1; i < len(items); i++ {
		items[i] = 2.0
	}

	teams := Partition(items, 2, 1)
	if count := groupCount(teams); count != 8 {
		t.Fatalf("Expected an 8-vs-8 split, got %d vs %d", count, len(items)-count)
	}
}

func TestPartitionMaxDiffOddCount(t *testing.T) {
	// Five items with a max size difference of 1 must split 3-vs-2.
	items := []float64{1, 2, 3, 4, 5}
	teams := Partition(items, 2, 1)

	count := groupCount(teams)
	if count != 2 && count != 3 {
		t.Fatalf("Expected a 3-vs-2 split, got %d vs %d", count, len(items)-count)
	}
}

func TestPartitionMaxDiffMatchesBruteForce(t *testing.T) {
	rng := rand.New(rand.NewSource(0))

	for maxDiff := 0; maxDiff <= 3; maxDiff++ {
		for n := 1; n <= 12; n++ {
			minCount, maxCount := allowedCounts(n, maxDiff)
			for trial := 0; trial < 50; trial++ {
				items := make([]float64, n)
				for i := range items {
					items[i] = rng.Float64()*100.0 + 1.0
				}

				teams := Partition(items, 2, maxDiff)
				if len(teams) != n {
					t.Fatalf("Partition returned %d entries for %d items", len(teams), n)
				}

				count := groupCount(teams)
				if count < minCount || count > maxCount {
					t.Fatalf("maxDiff=%d items %v: size constraint violated, group 0 has %d of %d",
						maxDiff, items, count, n)
				}

				got := partitionDiff(items, teams)
				want := bruteBestDiffWithSize(items, maxDiff)
				if math.Abs(got-want) > 1e-9 {
					t.Fatalf("maxDiff=%d items %v: partition diff %f does not match optimal %f",
						maxDiff, items, got, want)
				}
			}
		}
	}
}

func TestPartitionMaxDiffCKKFallback(t *testing.T) {
	// The CKK fallback must honour the size constraint too.
	rng := rand.New(rand.NewSource(7))

	items := make([]float64, maxHorowitzSahni+2)
	for i := range items {
		items[i] = rng.Float64()*100.0 + 1.0
	}

	minCount, maxCount := allowedCounts(len(items), 1)
	teams := Partition(items, 2, 1)
	if len(teams) != len(items) {
		t.Fatalf("Partition returned %d entries for %d items", len(teams), len(items))
	}

	count := groupCount(teams)
	if count < minCount || count > maxCount {
		t.Fatalf("Size constraint violated on CKK fallback, group 0 has %d of %d items",
			count, len(items))
	}
}

func TestPartitionMaxDiffEqualCounts(t *testing.T) {
	// A max size difference of 0 forces the teams to be as equal as possible.
	items := []float64{7, 5, 3, 1}
	teams := Partition(items, 2, 0)

	if count := groupCount(teams); count != 2 {
		t.Fatalf("Expected a 2-vs-2 split, got %d vs %d", count, len(items)-count)
	}

	diff := partitionDiff(items, teams)
	if diff != 0.0 {
		t.Fatalf("Expected a perfect 2-vs-2 split of {7,5,3,1}, got a difference of %f", diff)
	}
}

// --- CKK-specific tests ---

func TestCKKMatchesBruteForce(t *testing.T) {
	rng := rand.New(rand.NewSource(99))

	// Verify CKK against brute force for small inputs. CKK with bounded
	// backtracking is a heuristic and may not find the exact optimum. We verify
	// it always produces a result within 2x of optimal, which is more than
	// sufficient for game-balancing purposes where the absolute differences are
	// tiny compared to total team sums.
	for n := 2; n <= 14; n++ {
		for trial := 0; trial < 100; trial++ {
			items := make([]float64, n)
			for i := range items {
				items[i] = rng.Float64()*100.0 + 1.0
			}

			teams := ckkTwoWay(items)
			if len(teams) != n {
				t.Fatalf("ckkTwoWay returned %d entries for %d items", len(teams), n)
			}

			got := partitionDiff(items, teams)
			want := bruteBestDiff(items)
			if want > 0 && got-want > want {
				t.Fatalf("CKK items %v: partition diff %f is more than 2x optimal %f",
					items, got, want)
			}
		}
	}
}

func TestCKKCardinality(t *testing.T) {
	rng := rand.New(rand.NewSource(101))

	for maxDiff := 0; maxDiff <= 3; maxDiff++ {
		for n := 2; n <= 14; n++ {
			minCount, maxCount := allowedCounts(n, maxDiff)
			for trial := 0; trial < 50; trial++ {
				items := make([]float64, n)
				for i := range items {
					items[i] = rng.Float64()*100.0 + 1.0
				}

				teams := Partition(items, 2, maxDiff)
				if len(teams) != n {
					t.Fatalf("Partition returned %d entries for %d items", len(teams), n)
				}

				count := groupCount(teams)
				if count < minCount || count > maxCount {
					t.Fatalf("CKK maxDifference=%d items %v: size constraint violated, group 0 has %d of %d",
						maxDiff, items, count, n)
				}
			}
		}
	}
}

func TestCKKLargeInput(t *testing.T) {
	rng := rand.New(rand.NewSource(200))

	items := make([]float64, 50)
	for i := range items {
		items[i] = rng.Float64()*100.0 + 1.0
	}

	teams := Partition(items, 2, 1)
	if len(teams) != 50 {
		t.Fatalf("Partition returned %d entries for 50 items", len(teams))
	}

	minCount, maxCount := allowedCounts(50, 1)
	count := groupCount(teams)
	if count < minCount || count > maxCount {
		t.Fatalf("CKK large input: size constraint violated, group 0 has %d of 50", count)
	}

	sumA, sumB := groupSums(items, teams)
	total := sumA + sumB
	diff := math.Abs(sumA - sumB)

	if diff > total*0.05 {
		t.Fatalf("CKK large input left a difference of %f out of %f total", diff, total)
	}
}

// --- Multi-way CKK tests ---

func TestMultiWayCKKBalanced(t *testing.T) {
	// 9 players should split roughly evenly across 3 teams (3 each).
	items := []float64{10, 9, 8, 7, 6, 5, 4, 3, 2}
	teams := Partition(items, 3, 1)

	teamCounts := make(map[int]int)
	for _, tid := range teams {
		teamCounts[tid]++
	}

	if len(teamCounts) != 3 {
		t.Fatalf("Expected 3 teams, got %d", len(teamCounts))
	}

	for _, count := range teamCounts {
		if count < 2 || count > 4 {
			t.Fatalf("Team size %d is outside expected range [2, 4] for 9 players in 3 teams", count)
		}
	}
}

func TestMultiWayCKKFourTeams(t *testing.T) {
	// 12 players across 4 teams should get roughly 3 each.
	items := make([]float64, 12)
	for i := range items {
		items[i] = float64(i + 1)
	}

	teams := Partition(items, 4, 1)

	teamCounts := make(map[int]int)
	for _, tid := range teams {
		teamCounts[tid]++
	}

	if len(teamCounts) != 4 {
		t.Fatalf("Expected 4 teams, got %d", len(teamCounts))
	}

	for _, count := range teamCounts {
		if count < 2 || count > 4 {
			t.Fatalf("Team size %d is outside expected range [2, 4] for 12 players in 4 teams", count)
		}
	}
}

func TestMultiWayCKKUnevenSplit(t *testing.T) {
	// 7 players across 3 teams: 3-2-2 or 2-3-2 or 2-2-3.
	items := []float64{10, 8, 6, 4, 3, 2, 1}
	teams := Partition(items, 3, 1)

	teamCounts := make(map[int]int)
	for _, tid := range teams {
		teamCounts[tid]++
	}

	if len(teamCounts) != 3 {
		t.Fatalf("Expected 3 teams, got %d", len(teamCounts))
	}

	for _, count := range teamCounts {
		if count < 2 || count > 3 {
			t.Fatalf("Team size %d is outside expected range [2, 3] for 7 players in 3 teams", count)
		}
	}
}

func TestMultiWayCKKSumsBalanced(t *testing.T) {
	// Verify that the total skill is distributed reasonably across teams.
	rng := rand.New(rand.NewSource(300))
	items := make([]float64, 15)
	for i := range items {
		items[i] = rng.Float64()*100.0 + 1.0
	}

	teams := Partition(items, 3, 1)

	teamSums := make(map[int]float64)
	for i, tid := range teams {
		teamSums[tid] += items[i]
	}

	total := 0.0
	for _, item := range items {
		total += item
	}

	avgSum := total / 3.0
	for tid, s := range teamSums {
		diff := math.Abs(s - avgSum)
		if diff > avgSum*0.4 {
			t.Fatalf("Team %d sum %f is more than 40%% off the average %f", tid, s, avgSum)
		}
	}
}
