package skill

import (
	"math"
	"math/rand"
	"testing"
)

// groupSums returns the sums of the two groups described by a partition result.
func groupSums(items []float64, inGroupA []bool) (float64, float64) {
	sumA, sumB := 0.0, 0.0
	for i, item := range items {
		if inGroupA[i] {
			sumA += item
		} else {
			sumB += item
		}
	}
	return sumA, sumB
}

func partitionDiff(items []float64, inGroupA []bool) float64 {
	sumA, sumB := groupSums(items, inGroupA)
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

func TestPartitionPerfectSplit(t *testing.T) {
	items := []float64{10, 9, 8, 7}
	inGroupA := Partition(items)

	sumA, sumB := groupSums(items, inGroupA)
	if diff := math.Abs(sumA - sumB); diff != 0.0 {
		t.Fatalf("Expected a perfect split, got sums %f and %f", sumA, sumB)
	}
}

func TestPartitionImbalanced(t *testing.T) {
	items := []float64{100, 1, 1, 1}
	inGroupA := Partition(items)

	diff := partitionDiff(items, inGroupA)
	if diff != 97.0 {
		t.Fatalf("Expected the best possible difference of 97, got %f", diff)
	}
}

func TestPartitionOddCount(t *testing.T) {
	items := []float64{1, 2, 3}
	inGroupA := Partition(items)

	diff := partitionDiff(items, inGroupA)
	if diff != 0.0 {
		t.Fatalf("Expected a perfect split of {1,2,3}, got a difference of %f", diff)
	}
}

func TestPartitionAllEqual(t *testing.T) {
	items := []float64{2, 2, 2, 2, 2, 2}
	inGroupA := Partition(items)

	diff := partitionDiff(items, inGroupA)
	if diff != 0.0 {
		t.Fatalf("Expected a perfect split of equal items, got a difference of %f", diff)
	}
}

func TestPartitionSingleItem(t *testing.T) {
	inGroupA := Partition([]float64{42})
	if !inGroupA[0] {
		t.Fatal("The single item should belong to group A")
	}
}

func TestPartitionTwoItems(t *testing.T) {
	items := []float64{3, 5}
	inGroupA := Partition(items)

	if len(inGroupA) != 2 {
		t.Fatalf("Unexpected partition length: %d", len(inGroupA))
	}

	diff := partitionDiff(items, inGroupA)
	if diff != 2.0 {
		t.Fatalf("Expected the best possible difference of 2, got %f", diff)
	}
}

func TestPartitionEmpty(t *testing.T) {
	inGroupA := Partition(nil)
	if inGroupA == nil || len(inGroupA) != 0 {
		t.Fatalf("Expected an empty result, got %v", inGroupA)
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

			inGroupA := Partition(items)
			if len(inGroupA) != n {
				t.Fatalf("Partition returned %d entries for %d items", len(inGroupA), n)
			}

			got := partitionDiff(items, inGroupA)
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

	inGroupA := Partition(items)
	if len(inGroupA) != 24 {
		t.Fatalf("Partition returned %d entries for 24 items", len(inGroupA))
	}

	diff := partitionDiff(items, inGroupA)
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
