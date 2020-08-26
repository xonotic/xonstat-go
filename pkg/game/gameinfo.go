package game

import (
	"time"

	"github.com/nleeper/goment"
	"gitlab.com/xonotic/xonstat/pkg/models"
	"gitlab.com/xonotic/xonstat/pkg/server"
)

// InfoBase is the view-agnostic representation of a Game.
type InfoBase struct {
	GameID     int
	GameTypeCd string
	Duration   time.Duration
	Winner     int
	MatchID    string
	Mod        string
	CreateDt   time.Time
	CreateDtEpoch  int64
	CreateDtUTCStr string
	CreateDtFuzzy  string
	Server     *server.InfoBase
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

	dtUTC := game.CreateDt.UTC()
	fuzzyDt, _ := goment.New(dtUTC)

	return &InfoBase{
		GameID: gameID,
		GameTypeCd: game.GameTypeCd,
		Duration: *game.Duration,
		Winner: int(game.Winner.Int64),
		MatchID: game.MatchID.String,
		Mod: game.Mod.String,
		CreateDt: game.CreateDt,
		CreateDtEpoch:  game.CreateDt.Unix(),
		CreateDtUTCStr: dtUTC.Format("Mon, 2 Jan 2006 15:04:05 MST"),
		CreateDtFuzzy:  fuzzyDt.FromNow(),
		Server: server,
	}, nil
}