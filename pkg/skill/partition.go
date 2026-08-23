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

// maxHorowitzSahni is the largest field size we solve exactly. Above this we
// fall back to the Complete Karmarkar-Karp differencing algorithm, since the
// exact meet-in-the-middle search is O(2^(n/2)) in both time and space.
const maxHorowitzSahni = 28

// maxN is the largest number of players the algorithms will process. Fields larger
// than this are returned unbalanced to avoid excessive computation.
const maxN = 128

// allowedCounts returns the inclusive range of sizes that a single group may
// have for the difference between the two group sizes to be at most
// maxDifference. A negative maxDifference means any size is allowed.
func allowedCounts(n, maxDifference int) (int, int) {
	if maxDifference < 0 || maxDifference >= n {
		return 0, n
	}

	minCount := (n - maxDifference + 1) / 2
	maxCount := (n + maxDifference) / 2
	if maxCount < minCount {
		maxCount = minCount
	}

	return minCount, maxCount
}

// ckkElement holds a value and the sign assignment of the original items that
// contribute to it during the Complete Karmarkar-Karp differencing process. For
// each original item i in this element, bit i of pos indicates sign +1 and a
// set bit in all but not pos indicates sign -1. This compact representation
// replaces per-element map allocations with two uint64 bitmasks, reducing
// allocations from ~1M to ~0 per call.
type ckkElement struct {
	value        float64
	posLo, posHi uint64 // bits for items with sign +1 (low/high 64)
	allLo, allHi uint64 // bits for all items in this element
}

// mergeCKK combines two elements for the operation a - b: a's signs are kept,
// b's signs are negated. The negation of b is simply all ^ pos.
func mergeCKK(a, b ckkElement) ckkElement {
	return ckkElement{
		value: a.value - b.value,
		posLo: a.posLo | (b.allLo ^ b.posLo),
		posHi: a.posHi | (b.allHi ^ b.posHi),
		allLo: a.allLo | b.allLo,
		allHi: a.allHi | b.allHi,
	}
}

// ckkDepthLimit bounds the backtracking depth in the CKK algorithm. At each
// level the largest element is paired with up to ckkMaxBranch other elements,
// so the total nodes explored is at most ckkMaxBranch^ckkDepthLimit.
const ckkDepthLimit = 7

// ckkMaxBranch is the maximum number of alternative pairings explored at each
// differencing step. The largest element is paired with each of the next
// ckkMaxBranch elements by absolute value. Combined with ckkDepthLimit=7,
// this gives at most 3^7 ≈ 2,187 nodes, running in well under a millisecond.
const ckkMaxBranch = 3

// ckkTwoWay solves the 2-way number partitioning problem using the Complete
// Karmarkar-Karp differencing algorithm with backtracking. It returns a slice
// of 0-indexed team assignments (0 or 1) that minimises the difference
// between the sums of the two groups. No constraint is placed on group sizes.
func ckkTwoWay(items []float64) []int {
	n := len(items)
	if n == 0 {
		return nil
	}
	if n == 1 {
		return []int{0}
	}

	// Pre-allocate a single working buffer for the entire recursion. Each
	// branch saves/restores the elements it modifies rather than allocating
	// a new slice, eliminating ~1M slice allocations per call.
	buf := make([]ckkElement, n)
	for i, item := range items {
		if i < 64 {
			buf[i] = ckkElement{value: item, posLo: 1 << uint(i), allLo: 1 << uint(i)}
		} else {
			buf[i] = ckkElement{value: item, posHi: 1 << uint(i-64), allHi: 1 << uint(i-64)}
		}
	}

	bestResidue := math.Inf(1)
	bestPosLo, bestPosHi := uint64(0), uint64(0)

	// saveStack is indexed by recursion depth. Each level saves its state
	// to a separate region so child calls don't overwrite the parent's data.
	saveStack := make([]ckkElement, n*n)

	var recurse func(depth int, count int)
	recurse = func(depth int, count int) {
		if count <= 1 {
			if count == 1 {
				residue := math.Abs(buf[0].value)
				if residue < bestResidue {
					bestResidue = residue
					bestPosLo = buf[0].posLo
					bestPosHi = buf[0].posHi
				}
			}
			return
		}

		// Sort by absolute value descending.
		sort.Slice(buf[:count], func(i, j int) bool {
			ai, bi := buf[i].value, buf[j].value
			if ai < 0 {
				ai = -ai
			}
			if bi < 0 {
				bi = -bi
			}
			return ai > bi
		})

		// Branch: pair the largest element a with each of the next
		// min(count-1, ckkMaxBranch) elements.
		limit := count
		if depth < ckkDepthLimit && limit > ckkMaxBranch+1 {
			limit = ckkMaxBranch + 1
		} else if depth >= ckkDepthLimit {
			limit = 2 // only the standard KK move when depth exhausted
		}

		// Snapshot current state to this depth's region of saveStack.
		saveBase := depth * n
		copy(saveStack[saveBase:saveBase+count], buf[:count])

		for j := 1; j < limit; j++ {
			a, b := saveStack[saveBase], saveStack[saveBase+j]
			buf[0] = mergeCKK(a, b)

			// Compact: copy elements [1..count) excluding index j.
			pos := 1
			for k := 1; k < count; k++ {
				if k != j {
					buf[pos] = saveStack[saveBase+k]
					pos++
				}
			}

			recurse(depth+1, count-1)
		}
	}

	recurse(0, n)

	// Reconstruct the partition from the best bitmask found.
	result := make([]int, n)
	for i := 0; i < n; i++ {
		if i < 64 {
			if bestPosLo&(1<<uint(i)) != 0 {
				result[i] = 0
			} else {
				result[i] = 1
			}
		} else {
			if bestPosHi&(1<<uint(i-64)) != 0 {
				result[i] = 0
			} else {
				result[i] = 1
			}
		}
	}

	return result
}

// enforceCardinality adjusts a 2-way partition to meet the size constraint.
// It flips the minimum-cost items between groups so that the group A count
// falls within [minCount, maxCount].
func enforceCardinality(items []float64, result []int, maxDifference int) []int {
	n := len(items)
	minCount, maxCount := allowedCounts(n, maxDifference)

	countA := 0
	for _, v := range result {
		if v == 0 {
			countA++
		}
	}

	if countA >= minCount && countA <= maxCount {
		return result
	}

	// Compute current sums so we can pick the best items to flip.
	sumA, sumB := 0.0, 0.0
	for i, v := range result {
		if v == 0 {
			sumA += items[i]
		} else {
			sumB += items[i]
		}
	}

	if countA > maxCount {
		// Move excess items from A to B, picking those that bring the
		// sums closest together.
		toMove := countA - maxCount
		diff := sumA - sumB
		candidates := make([]int, 0, toMove)
		for i, v := range result {
			if v == 0 {
				candidates = append(candidates, i)
			}
		}
		sort.Slice(candidates, func(i, j int) bool {
			di := math.Abs(diff - 2*items[candidates[i]])
			dj := math.Abs(diff - 2*items[candidates[j]])
			return di < dj
		})
		for i := 0; i < toMove && i < len(candidates); i++ {
			result[candidates[i]] = 1
		}
	} else {
		// Move items from B to A.
		toMove := minCount - countA
		diff := sumA - sumB
		candidates := make([]int, 0, toMove)
		for i, v := range result {
			if v == 1 {
				candidates = append(candidates, i)
			}
		}
		sort.Slice(candidates, func(i, j int) bool {
			di := math.Abs(diff + 2*items[candidates[i]])
			dj := math.Abs(diff + 2*items[candidates[j]])
			return di < dj
		})
		for i := 0; i < toMove && i < len(candidates); i++ {
			result[candidates[i]] = 0
		}
	}

	return result
}

// hsTwoWay solves the 2-way number partitioning problem using the
// Horowitz-Sahni meet-in-the-middle algorithm. It returns a slice of 0-indexed
// team assignments (0 or 1) that minimises the difference between group sums
// while respecting the cardinality constraint. Only called for n ≤ maxHorowitzSahni.
func hsTwoWay(items []float64, maxDifference int) []int {
	n := len(items)
	result := make([]int, n)
	if n <= 1 {
		return result
	}

	total := 0.0
	for _, item := range items {
		total += item
	}
	target := total / 2.0

	minCount, maxCount := allowedCounts(n, maxDifference)

	mid := n / 2
	sumsA := enumerateSubsetSums(items[:mid])
	sumsB := enumerateSubsetSums(items[mid:])

	// Bucket the right-half subsets by their cardinality so the search can
	// restrict itself to subsets that respect the size constraint.
	sumsBByCount := make(map[int][]subsetSum, mid+1)
	for _, b := range sumsB {
		count := bits.OnesCount(uint(b.mask))
		sumsBByCount[count] = append(sumsBByCount[count], subsetSum{sum: b.sum, mask: b.mask})
	}
	for _, list := range sumsBByCount {
		sort.Slice(list, func(i, j int) bool { return list[i].sum < list[j].sum })
	}

	bestDiff := math.Inf(1)
	bestMaskA, bestMaskB := 0, 0

	// A group consisting of a subset from each half has a sum of a.sum + b.sum.
	// We want that as close to target as possible, since that minimizes the
	// difference between the sums of the two groups, while keeping the total
	// number of items in the group within [minCount, maxCount].
	for _, a := range sumsA {
		countA := bits.OnesCount(uint(a.mask))

		lo := minCount - countA
		if lo < 0 {
			lo = 0
		}
		hi := maxCount - countA
		if hi > n-mid {
			hi = n - mid
		}

		for countB := lo; countB <= hi; countB++ {
			list, ok := sumsBByCount[countB]
			if !ok {
				continue
			}

			index := sort.Search(len(list), func(i int) bool { return list[i].sum >= target-a.sum })
			for _, i := range []int{index - 1, index, index + 1} {
				if i < 0 || i >= len(list) {
					continue
				}

				diff := math.Abs(a.sum + list[i].sum - target)
				if diff < bestDiff {
					bestDiff = diff
					bestMaskA = a.mask
					bestMaskB = list[i].mask
				}
			}
		}
	}

	for i := 0; i < mid; i++ {
		if bestMaskA&(1<<uint(i)) != 0 {
			result[i] = 0
		} else {
			result[i] = 1
		}
	}

	for i := 0; i < n-mid; i++ {
		if bestMaskB&(1<<uint(i)) != 0 {
			result[mid+i] = 0
		} else {
			result[mid+i] = 1
		}
	}

	return result
}

// multiWayAssign recursively splits items[start:end] into numTeams groups,
// assigning team IDs starting from baseTeam. It reorders items[start:end] as a
// side effect so that group-A items come first.
func multiWayAssign(result []int, items []float64, start, end, numTeams, baseTeam, maxDifference int) {
	size := end - start
	if size == 0 {
		return
	}

	if numTeams <= 1 || size <= 1 {
		for i := start; i < end; i++ {
			result[i] = baseTeam
		}
		return
	}

	// Compute how many items should go to each side based on the
	// number of teams. This ensures the final team sizes are
	// balanced: each team gets roughly N/T items.
	leftTeams := (numTeams + 1) / 2
	rightTeams := numTeams - leftTeams
	targetLeft := leftTeams * size / numTeams
	if targetLeft < leftTeams {
		targetLeft = leftTeams
	}

	if targetLeft > size-rightTeams {
		targetLeft = size - rightTeams
	}

	// Use CKK to get a roughly balanced partition (no cardinality).
	subset := make([]float64, size)
	for i := 0; i < size; i++ {
		subset[i] = items[start+i]
	}

	inA := ckkTwoWay(subset)
	inA = enforceCardinality(subset, inA, maxDifference)

	countA := 0
	for _, v := range inA {
		if v == 0 {
			countA++
		}
	}

	// Adjust: swap items between groups to match the target size,
	// choosing swaps that minimise the increase in partition diff.
	for countA != targetLeft {
		sumA, sumB := 0.0, 0.0
		for i := 0; i < size; i++ {
			if inA[i] == 0 {
				sumA += subset[i]
			} else {
				sumB += subset[i]
			}
		}

		if countA < targetLeft {
			// Move an item from B to A.
			bestIdx, bestDelta := -1, math.Inf(1)
			for i := 0; i < size; i++ {
				if inA[i] == 1 {
					delta := math.Abs((sumA + subset[i]) - (sumB - subset[i]))
					if delta < bestDelta {
						bestDelta = delta
						bestIdx = i
					}
				}
			}

			if bestIdx < 0 {
				break
			}

			inA[bestIdx] = 0
			countA++
		} else {
			// Move an item from A to B.
			bestIdx, bestDelta := -1, math.Inf(1)
			for i := 0; i < size; i++ {
				if inA[i] == 0 {
					delta := math.Abs((sumA - subset[i]) - (sumB + subset[i]))
					if delta < bestDelta {
						bestDelta = delta
						bestIdx = i
					}
				}
			}

			if bestIdx < 0 {
				break
			}

			inA[bestIdx] = 1
			countA--
		}
	}

	// Partition indices into left (group A) and right (group B).
	left := make([]int, 0, countA)
	right := make([]int, 0, size-countA)
	for i := 0; i < size; i++ {
		if inA[i] == 0 {
			left = append(left, start+i)
		} else {
			right = append(right, start+i)
		}
	}

	// Rearrange so that group-A items come first in the slice.
	perm := append(left, right...)
	tmp := make([]float64, size)
	for i, idx := range perm {
		tmp[i] = items[idx]
	}
	for i := 0; i < size; i++ {
		items[start+i] = tmp[i]
	}

	multiWayAssign(result, items, start, start+countA, leftTeams, baseTeam, maxDifference)
	multiWayAssign(result, items, start+countA, end, rightTeams, baseTeam+leftTeams, maxDifference)
}

// Partition splits items into numTeams groups such that the difference between
// the sums of the groups is as small as possible. For numTeams=2 it uses
// Horowitz-Sahni (n ≤ 28) or CKK (n > 28) with a cardinality constraint.
// For numTeams>2 it uses recursive CKK bisection. It returns a slice of
// 0-indexed team assignments. Fields exceeding maxFieldSize are returned
// unbalanced (all on team 0).
func Partition(items []float64, numTeams int, maxDifference int) []int {
	n := len(items)
	if n == 0 {
		return make([]int, 0)
	}

	if n > maxN || numTeams <= 1 {
		result := make([]int, n)
		return result
	}

	if numTeams == 2 {
		// 1) Horowitz-Sahni
		if n <= maxHorowitzSahni {
			return hsTwoWay(items, maxDifference)
		}

		// 2) Complete Karmarkar-Karp
		ckk := ckkTwoWay(items)
		return enforceCardinality(items, ckk, maxDifference)
	}

	// 3) Recursive Complete Karmarkar-Karp bisection.
	result := make([]int, n)
	for i := range result {
		result[i] = -1
	}

	multiWayAssign(result, items, 0, n, numTeams, 0, maxDifference)
	return result
}
