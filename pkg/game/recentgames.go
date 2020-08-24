package game

import (
	"encoding/json"
	"html/template"
	"time"

	"github.com/antzucaro/qstr"
	"github.com/nleeper/goment"
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
	CreateDt            time.Time
	CreateDtEpoch       int64
	CreateDtUTCStr      string
	CreateDtFuzzy       string
}

// RecentGamesData retrieves recent games data based on filter criteria
func RecentGamesData(db models.Datastore, serverID int, mapID int, playerID int, gameTypeCd string,
	cutoff *time.Time, startGameID int, endGameID int, limit int) ([]RecentGameBase, error) {

	if gameTypeCd != "" && !submission.IsSupportedGameType(gameTypeCd) {
		return nil, submission.ErrUnsupportedGameType
	}

	rawRecentGames, err := db.RRecentGames(serverID, mapID, playerID, gameTypeCd, cutoff,
		startGameID, endGameID, limit)

	if err != nil {
		return nil, err
	}

	var recentGames []RecentGameBase
	for _, v := range rawRecentGames {
		nick := qstr.QStr(v.WinningNick)
		winningTeam := -1
		if v.WinningTeam.Valid {
			winningTeam = int(v.WinningTeam.Int32)
		}

		fuzzyDt, err := goment.New(v.CreateDt.UTC())
		if err != nil {
			return recentGames, err
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
			WinningNick:     nick.Stripped(),
			WinningNickHTML: nick.HTML(),
			CreateDt:        v.CreateDt,
			CreateDtEpoch:   v.CreateDt.Unix(),
			CreateDtUTCStr:  v.CreateDt.UTC().Format("Mon, 2 Jan 2006 15:04:05 MST"),
			CreateDtFuzzy:   fuzzyDt.FromNow(),
		}

		recentGames = append(recentGames, rg)
	}

	return recentGames, err
}

// RecentGamesJSON returns recent games in JSON format
func RecentGamesJSON(db models.Datastore, serverID int, mapID int, playerID int, gameTypeCd string,
	cutoff *time.Time, startGameID int, endGameID int, limit int) ([]byte, error) {

	rawData, err := RecentGamesData(db, serverID, mapID, playerID, gameTypeCd, cutoff,
		startGameID, endGameID, limit)
	if err != nil {
		return nil, err
	}

	// the JSON response
	type Response struct {
		GameID          int    `json:"game_id"`
		GameTypeCd      string `json:"game_type_cd"`
		GameTypeDescr   string `json:"game_type_descr"`
		ServerID        int    `json:"server_id"`
		ServerName      string `json:"server_name"`
		MapID           int    `json:"map_id"`
		MapName         string `json:"map_name"`
		WinningTeam     int    `json:"winning_team"`
		WinningPlayerID int    `json:"winning_player_id"`
		WinningNick     string `json:"nick"`
		CreateDt        string `json:"create_dt"`
	}

	var rgs []Response
	for _, rg := range rawData {
		r := Response{
			GameID:          rg.GameID,
			GameTypeCd:      rg.GameTypeCd,
			GameTypeDescr:   rg.GameTypeDescr,
			ServerID:        rg.ServerID,
			ServerName:      rg.ServerName,
			MapID:           rg.MapID,
			MapName:         rg.MapName,
			WinningTeam:     rg.WinningTeam,
			WinningPlayerID: rg.WinningPlayerID,
			WinningNick:     rg.WinningNick,
			CreateDt:        rg.CreateDt.Format(time.RFC3339),
		}

		rgs = append(rgs, r)
	}

	return json.Marshal(rgs)
}
