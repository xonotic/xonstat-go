package handlers

import (
	"sort"
)

// Should the weapon be included in the accuracy dataset?
func isAccuracyWeapon(weaponCd string) bool {
	switch weaponCd {
	case "arc", "assaultrifle", "huntingrifle", "machinegun", "okhmg", "okmachinegun":
		return true
	case "oknex", "rifle", "shockwave", "shotgun", "vaporizer", "vortex":
		return true
	}

	return false
}

// Should the weapon be included in the damage dataset?
func isDamageWeapon(weaponCd string) bool {
	switch weaponCd {
	case "vortex", "machinegun", "shotgun", "arc", "uzi", "nex", "minstanex":
		return true
	case "rifle", "grenadelauncher", "minelayer", "rocketlauncher", "hlac", "seeker":
		return true
	case "fireball", "mortar", "electro", "crylink", "hagar", "devastator":
		return true
	}

	return false
}

// Sorting a map[string]int by value, returning the list of keys in sorted order.
type pair struct {
	Key string
	Value int
}

type pairList []pair

func (p pairList) Len() int { return len(p) }
func (p pairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p pairList) Swap(i, j int){ p[i], p[j] = p[j], p[i] }

func sortDescByIntValue(m map[string]int) []string {
	pl := make(pairList, len(m))
	i := 0
	for k, v := range m {
		pl[i] = pair{k, v}
		i++
	}

	sort.Sort(sort.Reverse(pl))

	sorted := make([]string, len(pl))
	for i, v := range pl {
		sorted[i] = v.Key
	}

	return sorted
}