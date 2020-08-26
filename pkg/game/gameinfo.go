package game

import (
	"time"

	"gitlab.com/xonotic/xonstat/pkg/models"
	"gitlab.com/xonotic/xonstat/pkg/server"
)

// InfoBase is the view-agnostic representation of a Game.
type InfoBase struct {
	GameID        int
	GameTypeCd    string
	GameTypeDescr string
	Duration      time.Duration
	Winner        int
	MatchID       string
	Mod           string
	CreateDt      *models.MultiDt
	Server        *server.InfoBase
}

// InfoData returns the view-agnostic data for a given game by its ID.
func InfoData(db models.Datastore, gameID int) (*InfoBase, error) {
	game, err := db.RGameByID(gameID)
	if err != nil {
		return nil, err
	}

	server, err := server.InfoData(db, game.ServerID)
	if err != nil {
		return nil, err
	}

	dt, err := models.NewMultiDt(game.CreateDt)
	if err != nil {
		return nil, err
	}

	return &InfoBase{
		GameID:        gameID,
		GameTypeCd:    game.GameTypeCd,
		GameTypeDescr: game.GameTypeDescr,
		Duration:      *game.Duration,
		Winner:        int(game.Winner.Int64),
		MatchID:       game.MatchID.String,
		Mod:           game.Mod.String,
		CreateDt:      dt,
		Server:        server,
	}, nil
}
