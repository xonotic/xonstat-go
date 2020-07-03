package game

import (
	"encoding/json"
	"time"

	"gitlab.com/xonotic/xonstat/pkg/models"
	"gitlab.com/xonotic/xonstat/pkg/submission"
)

// RecentGamesData retrieves recent games data based on filter criteria
func RecentGamesData(db models.Datastore, serverID int, mapID int, playerID int, gameTypeCd string,
	cutoff time.Time, startGameID int, endGameID int, limit int) ([]*models.RecentGame, error) {

	if gameTypeCd != "" && !submission.IsSupportedGameType(gameTypeCd) {
		return nil, submission.ErrUnsupportedGameType
	}

	return db.RRecentGames(serverID, mapID, playerID, gameTypeCd, cutoff,
		startGameID, endGameID, limit)
}

// RecentGamesJSON returns recent games in JSON format
func RecentGamesJSON(db models.Datastore, serverID int, mapID int, playerID int, gameTypeCd string,
	cutoff time.Time, startGameID int, endGameID int, limit int) ([]byte, error) {

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
		WinningTeam     *int32    `json:"winning_team"`
		WinningPlayerID int    `json:"winning_player_id"`
		WinningNick     string `json:"nick"`
		CreateDt        string `json:"create_dt"`
	}

	var rgs []Response
	for _, rg := range rawData {
		r := Response{
			GameID: rg.GameID, 
			GameTypeCd: rg.GameTypeCd, 
			GameTypeDescr: rg.GameTypeDescr, 
			ServerID: rg.ServerID, 
			ServerName: rg.ServerName,
			MapID: rg.MapID, 
			MapName: rg.MapName, 
			WinningPlayerID: rg.WinningPlayerID, 
			WinningNick: rg.WinningNick,
			CreateDt: rg.CreateDt.Format(time.RFC3339),
		}

		if rg.WinningTeam.Valid {
			r.WinningTeam = &rg.WinningTeam.Int32
		} else {
			r.WinningTeam = nil
		}

		rgs = append(rgs, r)
	}

	return json.Marshal(rgs)
}
