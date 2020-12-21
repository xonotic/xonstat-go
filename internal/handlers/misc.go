package handlers

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