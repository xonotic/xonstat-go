package game

import (
	"html/template"
	"time"

	"gitlab.com/xonotic/xonstat/pkg/models"
	"gitlab.com/xonotic/xonstat/pkg/submission"
)

// EmptyServerID is a blank server ID value.
const EmptyServerID = -1

// EmptyMapID is a blank map ID value.
const EmptyMapID = -1

// EmptyPlayerID is a blank player ID value.
const EmptyPlayerID = -1

// EmptyGameTypeCd is a blank game type value.
const EmptyGameTypeCd = ""

// EmptyStartGameID is a blank starting game ID value.
const EmptyStartGameID = -1

// EmptyEndGameID is a blank ending game ID value.
const EmptyEndGameID = -1

// EmptyEndGameID is a blank ending game ID value.
const EmptyMatchID = ""

// RecentGameBase is the base type for recent games of any format (HTML, JSON, etc).
type RecentGameBase struct {
	GameID              int
	GameTypeCd          string
	GameTypeDescr       string
	ServerID            int
	ServerName          string
	MapID               int
	MapName             string
	WinningTeam         int
	WinningPlayerID     int
	WinningNick         string
	WinningNickStripped string
	WinningNickHTML     template.HTML
	CreateDt            *models.MultiDt
}

// RecentGamesData retrieves recent games data based on filter criteria
func RecentGamesData(db models.Datastore, serverID int, mapID int, playerID int, gameTypeCd string,
	cutoff *time.Time, startGameID int, endGameID int, limit int, matchID string) ([]RecentGameBase, error) {

	if gameTypeCd != "" && !submission.IsSupportedGameType(gameTypeCd) {
		return nil, submission.ErrUnsupportedGameType
	}

	rawRecentGames, err := db.RRecentGames(serverID, mapID, playerID, gameTypeCd, cutoff,
		startGameID, endGameID, limit, matchID)

	if err != nil {
		return nil, err
	}

	var recentGames []RecentGameBase
	for _, v := range rawRecentGames {
		nick := models.NewMultiNick(v.WinningNick)
		winningTeam := -1
		if v.WinningTeam.Valid {
			winningTeam = int(v.WinningTeam.Int32)
		}

		dt, err := models.NewMultiDt(v.CreateDt)
		if err != nil {
			return nil, err
		}

		rg := RecentGameBase{
			GameID:          v.GameID,
			GameTypeCd:      v.GameTypeCd,
			GameTypeDescr:   v.GameTypeDescr,
			ServerID:        v.ServerID,
			ServerName:      v.ServerName,
			MapID:           v.MapID,
			MapName:         v.MapName,
			WinningTeam:     winningTeam,
			WinningPlayerID: v.WinningPlayerID,
			WinningNick:     nick.NickStripped,
			WinningNickHTML: nick.NickHTML,
			CreateDt:        dt,
		}

		recentGames = append(recentGames, rg)
	}

	return recentGames, err
}
