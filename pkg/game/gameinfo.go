package game

import (
	"strings"

	"gitlab.com/xonotic/xonstat/pkg/models"
	"gitlab.com/xonotic/xonstat/pkg/server"
)

// TeamGameStatBase is the view agnostic representation of a team's stats
type TeamGameStatBase struct {
	TeamGameStatID int
	GameID         int
	Team           int
	Score          int
	Rounds         int
	Caps           int
	Color          string
	ColorInitCap   string
}

// NewTeamGameStatBase creates an instance of this class from the model type 
// returned from the DB.
func NewTeamGameStatBase(tgs *models.TeamGameStat) *TeamGameStatBase {
	score := 0
	if tgs.Score.Valid {
		score = int(tgs.Score.Int32)
	}

	rounds := 0
	if tgs.Rounds.Valid {
		rounds = int(tgs.Rounds.Int32)
	}

	caps := 0
	if tgs.Caps.Valid {
		caps = int(tgs.Caps.Int32)
	}

	color := "red"
	if tgs.Team == 5 {
		color = "red"
	} else if tgs.Team == 14 {
		color = "blue"
	} else if tgs.Team == 13 {
		color = "yellow"
	} else if tgs.Team == 10 {
		color = "pink"
	}

	colorInitCap := strings.Title(color)

	return &TeamGameStatBase{
		TeamGameStatID: tgs.TeamGameStatID,
		GameID: tgs.GameID,
		Team: tgs.Team,
		Score: score,
		Rounds: rounds,
		Caps: caps,
		Color: color,
		ColorInitCap: colorInitCap,
	}
}

// InfoBase is the view-agnostic representation of a Game.
type InfoBase struct {
	GameID            int
	GameTypeCd        string
	GameTypeDescr     string
	Duration          *models.MultiDuration
	Winner            int
	MatchID           string
	Mod               string
	CreateDt          *models.MultiDt
	Server            *server.InfoBase
	TeamGameStats     []*TeamGameStatBase
	PlayerGameStats   []*models.PlayerGameStat
	PlayerWeaponStats []*models.PlayerWeaponStat
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

	rawTeamGameStats, err := db.RTeamGameStatsByGameID(gameID)
	if err != nil {
		return nil, err
	}

	var teamGameStats []*TeamGameStatBase
	for _, v := range rawTeamGameStats {
		teamGameStats = append(teamGameStats, NewTeamGameStatBase(v))
	}

	playerGameStats, err := db.RPlayerGameStatsByGameID(gameID)
	if err != nil {
		return nil, err
	}

	playerWeaponStats, err := db.RPlayerWeaponStatsByGameID(gameID)
	if err != nil {
		return nil, err
	}

	return &InfoBase{
		GameID:            gameID,
		GameTypeCd:        game.GameTypeCd,
		GameTypeDescr:     game.GameTypeDescr,
		Duration:          models.NewMultiDuration(*game.Duration),
		Winner:            int(game.Winner.Int64),
		MatchID:           game.MatchID.String,
		Mod:               game.Mod.String,
		CreateDt:          dt,
		Server:            server,
		TeamGameStats:     teamGameStats,
		PlayerGameStats:   playerGameStats,
		PlayerWeaponStats: playerWeaponStats,
	}, nil
}
