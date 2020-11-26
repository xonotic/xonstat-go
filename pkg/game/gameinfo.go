package game

import (
	"math"
	"strings"
	"time"

	"gitlab.com/xonotic/xonstat/pkg/models"
	"gitlab.com/xonotic/xonstat/pkg/server"
	"gitlab.com/xonotic/xonstat/pkg/submission"
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
	color := ""
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
	Color            string
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

	color := TeamColorFromTeam(int(pgs.Team.Int32))

	return &PlayerGameStatBase{
		PlayerGameStatID: pgs.PlayerGameStatID,
		PlayerID:         pgs.PlayerID,
		GameID:           pgs.GameID,
		Nick:             models.NewMultiNick(pgs.Nick.String),
		Team:             int(pgs.Team.Int32),
		Color:            color,
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

// NonParticipantBase houses people who were around for a match but didn't complete it.
type NonParticipantBase struct {
	PlayerGameNonParticipantID int
	PlayerID                   int
	GameID                     int
	Status                     string
	Nick                       *models.MultiNick
	AliveTime                  *models.MultiDuration
	Score                      int32
}

// NewNonParticipantBase transforms a raw NonParticipant into its view-agnostic version.
func NewNonParticipantBase(np *models.PlayerGameNonParticipant) *NonParticipantBase {
	nick := models.NewMultiNick(np.Nick.String)
	noDuration := time.Duration(0) * time.Second

	var alivetime *models.MultiDuration
	if np.AliveTime != nil {
		alivetime = models.NewMultiDuration(*np.AliveTime)
	} else {
		alivetime = models.NewMultiDuration(noDuration)
	}

	return &NonParticipantBase{
		PlayerGameNonParticipantID: np.PlayerGameNonParticipantID,
		PlayerID:                   np.PlayerID,
		GameID:                     np.GameID,
		Status:                     np.Status,
		Nick:                       nick,
		AliveTime:                  alivetime,
		Score:                      np.Score.Int32,
	}
}

// GameInfoBase is the view-agnostic representation of a Game.
type GameInfoBase struct {
	GameID                int
	GameTypeCd            string
	GameTypeDescr         string
	Duration              *models.MultiDuration
	Winner                int
	MatchID               string
	Mod                   string
	CreateDt              *models.MultiDt
	Server                *server.InfoBase
	TeamGameStatsByTeam   map[int]*TeamGameStatBase
	TeamOrdering          []int
	PlayerGameStats       []*PlayerGameStatBase
	PlayerGameStatsByTeam map[int][]*PlayerGameStatBase
	ShowFragMatrix        bool
	FragMatrix            map[int][]int
	Forfeits              []*NonParticipantBase
	Spectators            []*NonParticipantBase
	ShowWeaponCharts      bool
}

func shouldShowWeaponCharts(gameTypeCd string) bool {
	switch gameTypeCd {
	case "cts", "nb", "nexball":
		return false
	default:
		return true
	}
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

	teamGameStatsByTeam := make(map[int]*TeamGameStatBase)
	for _, v := range rawTeamGameStats {
		teamGameStatsByTeam[v.Team] = NewTeamGameStatBase(v)
	}

	rawPlayerGameStats, err := db.RPlayerGameStatsByGameID(gameID)
	if err != nil {
		return nil, err
	}

	// Create map entries for each team
	var playerGameStats []*PlayerGameStatBase
	playerGameStatsByTeam := make(map[int][]*PlayerGameStatBase)
	for _, v := range teamGameStatsByTeam {
		playerGameStatsByTeam[v.Team] = make([]*PlayerGameStatBase, 0)
	}

	var forfeits, spectators []*NonParticipantBase
	rawNonParticipants, err := db.RPlayerGameNonParticipantsByGameID(gameID)
	if err != nil {
		return nil, err
	}

	for _, v := range rawNonParticipants {
		np := NewNonParticipantBase(v)
		if np.Status == "forfeit" {
			forfeits = append(forfeits, np)
		} else if np.Status == "spectator" {
			spectators = append(spectators, np)
		}
	}

	var teamOrdering []int // so we know which team "won" since maps are not ordered
	for _, v := range rawPlayerGameStats {
		pgsb := NewPlayerGameStatBase(v)
		playerGameStats = append(playerGameStats, pgsb)

		// To determine the order in which teams should be shown, we iterate over the
		// player game stats (which should be in scoreboardpos order) while keeping track
		// of the distinct team IDs we see.
		if len(teamOrdering) == 0 || teamOrdering[len(teamOrdering)-1] != pgsb.Team {
			teamOrdering = append(teamOrdering, pgsb.Team)
		}

		playerGameStatsByTeam[pgsb.Team] = append(playerGameStatsByTeam[pgsb.Team], pgsb)
	}

	// Frag matrix processing.
	showFragMatrix := false
	fragMatrix := make(map[int][]int)

	if submission.ShouldDoFragMatrix(game.GameTypeCd) {
		rawMatrix, err := db.RPlayerGameFragMatrixByGameID(game.GameID)
		if err != nil {
			return nil, err
		}

		// Map the matrix records to their player game stat records for easier display processing.
		matrixByPGSID := make(map[int]*models.PlayerGameFragMatrix)
		for _, v := range rawMatrix {
			matrixByPGSID[v.PlayerGameStatID] = v
		}

		for _, fragger := range playerGameStats {
			var cells []int
			for _, victim := range playerGameStats {
				if fragger.PlayerGameStatID == victim.PlayerGameStatID {
					// TODO: This is where the fragger is equal to the victim. Use suicides here?
					cells = append(cells, 0)
				} else {
					// Populate the cell for the intersection of the fragger and victim,
					// i.e. how many times fragger has fragged the victim in the match.
					fraggerMatrix := matrixByPGSID[fragger.PlayerGameStatID]
					victimMatrix := matrixByPGSID[victim.PlayerGameStatID]

					// Sometimes there are no entries for a given PGStatID. In these cases,
					// we treat that as a NULL intersection, assigning zero frags and moving on.
					if fraggerMatrix == nil || victimMatrix == nil {
						cells = append(cells, 0)
						continue
					}

					if frags, ok := fraggerMatrix.Matrix[victimMatrix.PlayerIndex]; ok {
						cells = append(cells, frags)
					} else {
						cells = append(cells, 0)
					}
				}
			}
			fragMatrix[fragger.PlayerGameStatID] = cells
		}

		showFragMatrix = true
	}

	return &GameInfoBase{
		GameID:                gameID,
		GameTypeCd:            game.GameTypeCd,
		GameTypeDescr:         game.GameTypeDescr,
		Duration:              models.NewMultiDuration(*game.Duration),
		Winner:                int(game.Winner.Int64),
		MatchID:               game.MatchID.String,
		Mod:                   game.Mod.String,
		CreateDt:              dt,
		Server:                server,
		TeamGameStatsByTeam:   teamGameStatsByTeam,
		TeamOrdering:          teamOrdering,
		PlayerGameStats:       playerGameStats,
		PlayerGameStatsByTeam: playerGameStatsByTeam,
		ShowFragMatrix:        showFragMatrix,
		FragMatrix:            fragMatrix,
		Forfeits:              forfeits,
		Spectators:            spectators,
		ShowWeaponCharts:      shouldShowWeaponCharts(game.GameTypeCd),
	}, nil
}
