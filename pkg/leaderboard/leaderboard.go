package leaderboard

import (
	"encoding/json"
	"time"

	"github.com/antzucaro/qstr"
	"gitlab.com/xonotic/xonstat/pkg/models"
)

// SummaryStatsData retrieves summary stat information for the leaderboard.
func SummaryStatsData(scope string, db models.Datastore) ([]*models.SummaryStat, error) {
	return db.RSummaryStats(scope)
}

// SummaryStatsJSON returns summary stats in JSON form.
func SummaryStatsJSON(scope string, db models.Datastore) ([]byte, error) {
	rawData, err := SummaryStatsData(scope, db)
	if err != nil {
		return nil, err
	}

	// the JSON response
	type GameCounts struct {
		GameTypeCd string `json:"game_type_cd"`
		GameCount int `json:"num_games"`
	}

	type Response struct {
		Players int `json:"players"`
		Games []GameCounts `json:"games"`
		Scope string `json:"scope"`
		LastRefreshed string `json:"last_refreshed"`
	}

	var players int
	lastRefreshed := "unknown"
	var games []GameCounts

	for i, ss := range rawData {
		if i == 0 {
			// We'll use the first record's refresh date and num_players for all of them.
			lastRefreshed = ss.RefreshDt.Format(time.RFC3339)
			players = ss.PlayerCount
		}
		games = append(games, GameCounts{ss.GameTypeCd, ss.GameCount})
	}

	return json.Marshal(Response{Players: players, Games: games, Scope: scope, LastRefreshed: lastRefreshed})
}

// ActivePlayersData retrieves the active players
func ActivePlayersData(limit, start int, db models.Datastore) ([]*models.ActivePlayer, error) {
	return db.RActivePlayers(limit, start)
}

// ActivePlayersJSON returns active player stats in JSON form.
func ActivePlayersJSON(limit, start int, db models.Datastore) ([]byte, error) {
	rawData, err := ActivePlayersData(limit, start, db)
	if err != nil {
		return nil, err
	}

	// the JSON response (a single entry in the list)
	type Response struct {
		Rank int `json:"rank"`
		PlayerID int `json:"player_id"`
		Nick string `json:"nick"`
		AliveTime string `json:"alivetime"`
	}

	var activePlayers []Response
	for _, ap := range rawData {
		var resp Response

		resp.Rank = ap.SortOrder
		resp.PlayerID = ap.PlayerID

		nick := qstr.QStr(ap.Nick)
		resp.Nick = nick.Stripped()
		resp.AliveTime = models.DurationString(ap.AliveTime, true)

		activePlayers = append(activePlayers, resp)
	}

	return json.Marshal(activePlayers)
}

// ActiveServersData retrieves the active servers
func ActiveServersData(limit, start int, db models.Datastore) ([]*models.ActiveServer, error) {
	return db.RActiveServers(limit, start)
}

// ActiveServersJSON returns active server stats in JSON form.
func ActiveServersJSON(limit, start int, db models.Datastore) ([]byte, error) {
	rawData, err := ActiveServersData(limit, start, db)
	if err != nil {
		return nil, err
	}

	// the JSON response (a single entry in the list)
	type Response struct {
		Rank int `json:"rank"`
		ServerID int `json:"server_id"`
		ServerName string `json:"server_name"`
		PlayTime string `json:"playtime"`
	}

	var activeServers []Response
	for _, as := range rawData {
		resp := Response{as.SortOrder, as.ServerID, as.ServerName, models.DurationString(as.PlayTime, true)}
		activeServers = append(activeServers, resp)
	}

	return json.Marshal(activeServers)
}