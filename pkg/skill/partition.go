package skill

import (
	"math"
	"math/bits"
	"sort"
)

// subsetSum holds the sum of the items in a subset of a list, along with the
// bitmask describing which items belong to the subset.
type subsetSum struct {
	sum  float64
	mask int
}

// enumerateSubsetSums returns the sum and bitmask of every subset of items.
// The bitmask uses bit i to indicate that items[i] is part of the subset.
func enumerateSubsetSums(items []float64) []subsetSum {
	size := 1 << uint(len(items))
	sums := make([]subsetSum, size)

	for mask := 1; mask < size; mask++ {
		bit := bits.TrailingZeros(uint(mask))
		prev := mask &^ (1 << uint(bit))
		sums[mask] = subsetSum{
			sum:  sums[prev].sum + items[bit],
			mask: mask,
		}
	}

	return sums
}

// Partition splits items into two groups such that the difference between the
// sums of the two groups is as small as possible. It returns a boolean slice
// where true means the item at that index belongs to group A and false means
// it belongs to group B.
//
// This is the 2-way number partitioning problem, solved exactly using the
// meet-in-the-middle (Horowitz-Sahni) approach. It runs in O(2^(n/2)) time and
// space, which is effectively instant for the small player counts this is
// intended for (a maximum of 24 players).
func Partition(items []float64) []bool {
	n := len(items)
	result := make([]bool, n)
	if n <= 1 {
		if n == 1 {
			result[0] = true
		}
		return result
	}

	total := 0.0
	for _, item := range items {
		total += item
	}
	target := total / 2.0

	mid := n / 2
	sumsA := enumerateSubsetSums(items[:mid])
	sumsB := enumerateSubsetSums(items[mid:])

	// Only one subset needs to be sorted for correctness: sumsA is iterated
	// in place, while sumsB is binary-searched for the sum closest to the
	// target.
	sort.Slice(sumsB, func(i, j int) bool { return sumsB[i].sum < sumsB[j].sum })

	bestDiff := math.Inf(1)
	bestMaskA, bestMaskB := 0, 0

	// A group consisting of a subset from each half has a sum of a.sum + b.sum.
	// We want that as close to target as possible, since that minimizes the
	// difference between the sums of the two groups.
	for _, a := range sumsA {
		index := sort.Search(len(sumsB), func(i int) bool { return sumsB[i].sum >= target-a.sum })
		for _, i := range []int{index - 1, index, index + 1} {
			if i < 0 || i >= len(sumsB) {
				continue
			}

			diff := math.Abs(a.sum + sumsB[i].sum - target)
			if diff < bestDiff {
				bestDiff = diff
				bestMaskA = a.mask
				bestMaskB = sumsB[i].mask
			}
		}
	}

	for i := 0; i < mid; i++ {
		result[i] = bestMaskA&(1<<uint(i)) != 0
	}

	for i := 0; i < n-mid; i++ {
		result[mid+i] = bestMaskB&(1<<uint(i)) != 0
	}

	return result
}
