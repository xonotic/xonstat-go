package game

import (
	"strings"
	"time"

	"gitlab.com/xonotic/xonstat/pkg/models"
)

// WeaponInfoBase is the weapon info by a single player in a match, with some added fields for convenience.
type WeaponInfoBase struct {
	PlayerWeaponStatID int
	PlayerID           int
	Nick               *models.MultiNick
	GameID             int
	PlayerGameStatID   int
	WeaponCd           string
	WeaponCdInitCaps   string
	Actual             int
	Max                int
	PctTotalDamage     float32
	Hit                int
	Fired              int
	PctAccuracy        float32
	Frags              int
	CreateDt           time.Time
}

// NewWeaponInfoBaseList takes the list of weapon stats and returns back their view-ready versions.
// NOTE: We take in a slice of weapon infos because we need to calculate percentage damage out of all
// damage dealt by the player in that game. This requires all the damage info from the other weapons.
func NewWeaponInfoBaseList(wiList []*models.WeaponInfo) []*WeaponInfoBase {
	var wib []*WeaponInfoBase
	if len(wiList) == 0 {
		return wib
	}

	// Totals to support a total damage percentage and a special "meta" entry
	var totalDamage, totalFrags, totalHits, totalFired int
	for _, v := range wiList {
		totalDamage += v.Actual
		totalFrags += v.Frags
		totalHits += v.Hit
		totalFired += v.Fired
	}

	if totalDamage <= 0 {
		totalDamage = 1
	}

	if totalFired <= 0 {
		totalFired = 1
	}

	for _, v := range wiList {
		if v.Fired <= 0 {
			v.Fired = 1
		}

		pctTotalDamage := 100 * (float32(v.Actual) / float32(totalDamage))
		pctAccuracy := 100 * (float32(v.Hit) / float32(v.Fired))

		wib = append(wib, &WeaponInfoBase{
			PlayerWeaponStatID: v.PlayerWeaponStatID,
			PlayerID:           v.PlayerID,
			Nick:               models.NewMultiNick(v.Nick),
			GameID:             v.GameID,
			PlayerGameStatID:   v.PlayerGameStatID,
			WeaponCd:           v.WeaponCd,
			WeaponCdInitCaps:   strings.Title(v.WeaponCd),
			Actual:             v.Actual,
			Max:                v.Max,
			PctTotalDamage:     pctTotalDamage,
			Hit:                v.Hit,
			Fired:              v.Fired,
			PctAccuracy:        pctAccuracy,
			Frags:              v.Frags,
			CreateDt:           v.CreateDt,
		})
	}

	return wib
}

// GameWeaponInfoBase is the view-agnostic representation of a game's weapon information.
type GameWeaponInfoBase struct {
	GameID          int
	GameTypeCd      string
	GameTypeDescr   string
	Duration        *models.MultiDuration
	Winner          int
	MatchID         string
	Mod             string
	DistinctWeapons []string
	WeaponInfo      []*WeaponInfoBase
	CreateDt        *models.MultiDt
}

// GameWeaponInfoData returns the view-agnostic weapon data for a given game by its ID.
func GameWeaponInfoData(db models.Datastore, gameID int) (*GameWeaponInfoBase, error) {
	game, err := db.RGameByID(gameID)
	if err != nil {
		return nil, err
	}

	dt, err := models.NewMultiDt(game.CreateDt)
	if err != nil {
		return nil, err
	}

	// Weapon stats processing
	rawWeaponInfos, err := db.RWeaponInfoByGameID(game.GameID)
	if err != nil {
		return nil, err
	}

	// Find the unique set of weapons, separate them into per-pgstatID groups
	weaponSet := make(map[string]struct{})
	distinctWeapons := make([]string, 0)
	weaponInfosByPGSID := make(map[int][]*models.WeaponInfo)
	pgsidOrder := make([]int, 0)

	for _, w := range rawWeaponInfos {
		if _, ok := weaponSet[w.WeaponCd]; !ok {
			weaponSet[w.WeaponCd] = struct{}{}
			distinctWeapons = append(distinctWeapons, w.WeaponCd)
		}

		if _, ok := weaponInfosByPGSID[w.PlayerGameStatID]; !ok {
			weaponInfosByPGSID[w.PlayerGameStatID] = make([]*models.WeaponInfo, 0)
			pgsidOrder = append(pgsidOrder, w.PlayerGameStatID)
		}
		weaponInfosByPGSID[w.PlayerGameStatID] = append(weaponInfosByPGSID[w.PlayerGameStatID], w)
	}

	// Iterate over the raw weapon infos, taking care to maintain order
	weaponInfoBases := make([]*WeaponInfoBase, 0)
	for _, pgsid := range pgsidOrder {
		weaponInfoBases = append(weaponInfoBases, NewWeaponInfoBaseList(weaponInfosByPGSID[pgsid])...)
	}

	return &GameWeaponInfoBase{
		GameID:          gameID,
		GameTypeCd:      game.GameTypeCd,
		GameTypeDescr:   game.GameTypeDescr,
		Duration:        models.NewMultiDuration(*game.Duration),
		Winner:          int(game.Winner.Int64),
		MatchID:         game.MatchID.String,
		Mod:             game.Mod.String,
		DistinctWeapons: distinctWeapons,
		WeaponInfo:      weaponInfoBases,
		CreateDt:        dt,
	}, nil
}
