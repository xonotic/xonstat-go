package game

import (
	"math"
	"strings"
	"time"

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

// TeamColorFromTeam takes a team number and converts it to its corresponding color
func TeamColorFromTeam(team int) string {
	color := "red"
	if team == 5 {
		color = "red"
	} else if team == 14 {
		color = "blue"
	} else if team == 13 {
		color = "yellow"
	} else if team == 10 {
		color = "pink"
	}

	return color
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

	color := TeamColorFromTeam(tgs.Team)
	colorInitCap := strings.Title(color)

	return &TeamGameStatBase{
		TeamGameStatID: tgs.TeamGameStatID,
		GameID:         tgs.GameID,
		Team:           tgs.Team,
		Score:          score,
		Rounds:         rounds,
		Caps:           caps,
		Color:          color,
		ColorInitCap:   colorInitCap,
	}
}

// PlayerGameStatBase is the view-agnostic type for a given player's stats in game.
type PlayerGameStatBase struct {
	PlayerGameStatID int
	PlayerID         int
	GameID           int
	Nick             *models.MultiNick
	Team             int
	Rank             int
	AliveTime        *models.MultiDuration
	Kills            int
	Deaths           int
	Suicides         int
	Score            int
	Time             *models.MultiDuration
	Captures         int
	Pickups          int
	Drops            int
	Returns          int
	Collects         int
	Destroys         int
	Pushes           int
	CarrierFrags     int
	EloDelta         float64
	Fastest          *models.MultiDuration
	AvgLatency       int
	TeamRank         int
	ScoreboardPos    int
	Laps             int
	Revivals         int
	Lives            int
	CreateDt         time.Time
}

// NewPlayerGameStatBase converts a raw player game stat record into a PlayerGameStatBase
func NewPlayerGameStatBase(pgs *models.PlayerGameStat) *PlayerGameStatBase {
	noDuration := time.Duration(0) * time.Second

	var alivetime, ttime, fastest *models.MultiDuration
	if pgs.AliveTime != nil {
		alivetime = models.NewMultiDuration(*pgs.AliveTime)
	} else {
		alivetime = models.NewMultiDuration(noDuration)
	}

	if pgs.Time != nil {
		ttime = models.NewMultiDuration(*pgs.Time)
	} else {
		ttime = models.NewMultiDuration(noDuration)
	}

	if pgs.Fastest != nil {
		fastest = models.NewMultiDuration(*pgs.Fastest)
	} else {
		fastest = models.NewMultiDuration(noDuration)
	}

	return &PlayerGameStatBase{
		PlayerGameStatID: pgs.PlayerGameStatID,
		PlayerID:         pgs.PlayerID,
		GameID:           pgs.GameID,
		Nick:             models.NewMultiNick(pgs.Nick.String),
		Team:             int(pgs.Team.Int32),
		Rank:             int(pgs.Rank.Int32),
		AliveTime:        alivetime,
		Kills:            int(pgs.Kills.Int32),
		Deaths:           int(pgs.Deaths.Int32),
		Suicides:         int(pgs.Suicides.Int32),
		Score:            int(pgs.Score.Int32),
		Time:             ttime,
		Captures:         int(pgs.Captures.Int32),
		Pickups:          int(pgs.Pickups.Int32),
		Drops:            int(pgs.Drops.Int32),
		Returns:          int(pgs.Returns.Int32),
		Collects:         int(pgs.Collects.Int32),
		Destroys:         int(pgs.Destroys.Int32),
		Pushes:           int(pgs.Pushes.Int32),
		CarrierFrags:     int(pgs.CarrierFrags.Int32),
		EloDelta:         pgs.EloDelta.Float64,
		Fastest:          fastest,
		AvgLatency:       int(math.Round(pgs.AvgLatency.Float64)),
		TeamRank:         int(pgs.Team.Int32),
		ScoreboardPos:    int(pgs.ScoreboardPos.Int32),
		Laps:             int(pgs.Laps.Int32),
		Revivals:         int(pgs.Revivals.Int32),
		Lives:            int(pgs.Lives.Int32),
		CreateDt:         pgs.CreateDt,
	}
}

// GameInfoBase is the view-agnostic representation of a Game.
type GameInfoBase struct {
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
	PlayerGameStats   []*PlayerGameStatBase
	PlayerWeaponStats []*models.PlayerWeaponStat
}

// GameInfoData returns the view-agnostic data for a given game by its ID.
func GameInfoData(db models.Datastore, gameID int) (*GameInfoBase, error) {
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

	rawPlayerGameStats, err := db.RPlayerGameStatsByGameID(gameID)
	if err != nil {
		return nil, err
	}

	var playerGameStats []*PlayerGameStatBase
	for _, v := range rawPlayerGameStats {
		playerGameStats = append(playerGameStats, NewPlayerGameStatBase(v))
	}

	playerWeaponStats, err := db.RPlayerWeaponStatsByGameID(gameID)
	if err != nil {
		return nil, err
	}

	return &GameInfoBase{
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
