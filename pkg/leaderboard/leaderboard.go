package leaderboard

import (
	"encoding/json"
	"html/template"
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

// ActivePlayerBase is the base type used to represent active players for all 
// marshalled types (HTML/JSON/etc). 
type ActivePlayerBase struct {
	SortOrder    int
	PlayerID     int
	Nick         qstr.QStr
	HTMLNick     template.HTML
	StrippedNick string
	AliveTime    string
}

// ActivePlayersData retrieves the active players
func ActivePlayersData(limit, start int, db models.Datastore) ([]ActivePlayerBase, error) {
	rawActivePlayers, err := db.RActivePlayers(limit, start)
	if err != nil {
		return nil, err
	}

	var activePlayersHTML []ActivePlayerBase
	for _, v := range rawActivePlayers {
		nick := qstr.QStr(v.Nick)
		ap := ActivePlayerBase{
			SortOrder: v.SortOrder, 
			PlayerID: v.PlayerID, 
			Nick: nick,
			HTMLNick: nick.HTML(), 
			StrippedNick: nick.Stripped(), 
			AliveTime: models.DurationString(v.AliveTime, true),
		}
		activePlayersHTML = append(activePlayersHTML, ap)
	}

	return activePlayersHTML, nil
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
		resp := Response {
			Rank: ap.SortOrder,
			PlayerID: ap.PlayerID,
			Nick: ap.StrippedNick,
			AliveTime: ap.AliveTime,
		}

		activePlayers = append(activePlayers, resp)
	}

	return json.Marshal(activePlayers)
}

// ActiveServerBase is the base type for marshalling active servers to other formats (HTML, JSON, etc).
type ActiveServerBase struct {
	SortOrder int
	ServerID int
	ServerName string
	PlayTime string
}

// ActiveServersData retrieves the active servers
func ActiveServersData(limit, start int, db models.Datastore) ([]ActiveServerBase, error) {
	rawActiveServers, err := db.RActiveServers(limit, start)
	if err != nil {
		return nil, err
	}

	var activeServers []ActiveServerBase
	for _, v := range rawActiveServers {
		as := ActiveServerBase {
			SortOrder: v.SortOrder,
			ServerID: v.ServerID,
			ServerName: v.ServerName,
			PlayTime: models.DurationString(v.PlayTime, true),
		}
		activeServers = append(activeServers, as)
	}

	return activeServers, nil
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
		resp := Response{
			Rank: as.SortOrder, 
			ServerID: as.ServerID, 
			ServerName: as.ServerName, 
			PlayTime: as.PlayTime,
		}
		activeServers = append(activeServers, resp)
	}

	return json.Marshal(activeServers)
}

// ActiveMapsData retrieves the active maps
func ActiveMapsData(limit, start int, db models.Datastore) ([]*models.ActiveMap, error) {
	return db.RActiveMaps(limit, start)
}

// ActiveMapsJSON returns active map stats in JSON form.
func ActiveMapsJSON(limit, start int, db models.Datastore) ([]byte, error) {
	rawData, err := ActiveMapsData(limit, start, db)
	if err != nil {
		return nil, err
	}

	// the JSON response (a single entry in the list)
	type Response struct {
		Rank int `json:"rank"`
		MapID int `json:"map_id"`
		MapName string `json:"map_name"`
		Games int `json:"games"`
	}

	var activeMaps []Response
	for _, am := range rawData {
		activeMaps = append(activeMaps, Response{am.SortOrder, am.MapID, am.MapName, am.Games})
	}

	return json.Marshal(activeMaps)
}