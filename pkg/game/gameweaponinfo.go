package game

import (
	"gitlab.com/xonotic/xonstat/pkg/models"
)

// GameWeaponInfoBase is the view-agnostic representation of a game's weapon information.
type GameWeaponInfoBase struct {
	GameID        int
	GameTypeCd    string
	GameTypeDescr string
	Duration      *models.MultiDuration
	Winner        int
	MatchID       string
	Mod           string
	WeaponStats   []*PlayerWeaponStatBase
	CreateDt      *models.MultiDt
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
	rawWeaponStats, err := db.RPlayerWeaponStatsByGameID(gameID)
	if err != nil {
		return nil, err
	}

	weaponStats := NewPlayerWeaponStatBaseList(rawWeaponStats)

	return &GameWeaponInfoBase{
		GameID:        gameID,
		GameTypeCd:    game.GameTypeCd,
		GameTypeDescr: game.GameTypeDescr,
		Duration:      models.NewMultiDuration(*game.Duration),
		Winner:        int(game.Winner.Int64),
		MatchID:       game.MatchID.String,
		Mod:           game.Mod.String,
		WeaponStats:   weaponStats,
		CreateDt:      dt,
	}, nil
}
