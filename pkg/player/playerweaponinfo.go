package player

import (
	"fmt"
	"gitlab.com/xonotic/xonstat/pkg/models"
	"gitlab.com/xonotic/xonstat/pkg/util"
)

// AccuracyBase is the accuracy of a weapon in a single game.
type AccuracyBase struct {
	WeaponCd    string
	Fired       int
	Hit         int
	PctAccuracy float32
}

// DamageBase is the damage of a weapon in a single game.
type DamageBase struct {
	WeaponCd string
	Actual   int
	Max      int
}

// PlayerWeaponInfoBase houses the summary weapon info for a player.
type PlayerWeaponInfoBase struct {
	// Player ID who these weapon stats belong to.
	PlayerID int

	// List of game IDs used to calculate these figures.
	GameIDs []int

	// List of distinct weapons used in the figures.
	Weapons []string

	// Accuracies accessible by the key <gameID>-<weaponCd>.
	Accuracy map[string]*AccuracyBase

	// Damages accessible by the key <gameID><weaponCd>.
	Damage map[string]*DamageBase
}

// PlayerWeaponInfoData returns the view-agnostic weapon data for a given player by ID.
func PlayerWeaponInfoData(db models.Datastore, playerID, limit int, gameTypeCd string) (*PlayerWeaponInfoBase, error) {
	if !models.IsWeaponInfoGameType(gameTypeCd) {
		return nil, fmt.Errorf("unsupported gameTypeCd=%s for PlayerWeaponInfoData", gameTypeCd)
	}

	gameIDs, err := db.RGameIDsByPlayerID(playerID, limit, gameTypeCd)
	if err != nil {
		return nil, err
	}

	weaponStats, err := db.RPlayerWeaponStatsByGameList(playerID, gameIDs)
	if err != nil {
		return nil, err
	}

	weaponSet := make(map[string]struct{})
	weapons := make([]string, 0)
	accuracy := make(map[string]*AccuracyBase)
	damage := make(map[string]*DamageBase)
	for _, ws := range weaponStats {
		// Keep track of the distinct weapons we've seen.
		_, seen := weaponSet[ws.WeaponCd]
		if !seen {
			weaponSet[ws.WeaponCd] = struct{}{}
			weapons = append(weapons, ws.WeaponCd)
		}

		// Add data entries to the map using the key.
		key := fmt.Sprintf("%d-%s", ws.GameID, ws.WeaponCd)
		accuracy[key] = &AccuracyBase{
			WeaponCd:    ws.WeaponCd,
			Fired:       ws.Fired,
			Hit:         ws.Hit,
			PctAccuracy: util.Percentage(ws.Hit, ws.Fired),
		}

		damage[key] = &DamageBase{
			WeaponCd: ws.WeaponCd,
			Actual:   ws.Actual,
			Max:      ws.Max,
		}
	}

	return &PlayerWeaponInfoBase{
		PlayerID: playerID,
		GameIDs:  gameIDs,
		Weapons:  weapons,
		Accuracy: accuracy,
		Damage:   damage,
	}, nil
}
