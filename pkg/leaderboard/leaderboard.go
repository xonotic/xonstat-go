package leaderboard

import (
	"encoding/json"
	"html/template"

	"github.com/antzucaro/qstr"
	"gitlab.com/xonotic/xonstat/pkg/models"
)

// GameCounts represent how many games have been played for a particular game type.
type GameCounts struct {
	GameTypeCd string
	GameCount  int
}

// SummaryBase provides high-level summary information about what has been recorded by stats
// over a given time period (scope).
type SummaryBase struct {
	Players       int
	Games         []GameCounts
	Scope         string
	LastRefreshed *models.MultiDt
}

// SummaryData retrieves summary stat information for the leaderboard.
func SummaryData(scope string, db models.Datastore) (*SummaryBase, error) {
	rawData, err := db.RSummaryStats(scope)
	if err != nil {
		return nil, err
	}

	var players int
	var lastRefreshed *models.MultiDt
	var games []GameCounts

	for i, ss := range rawData {
		if i == 0 {
			// We'll use the first record's refresh date and num_players for all of them.
			lastRefreshed, _ = models.NewMultiDt(ss.RefreshDt)
			players = ss.PlayerCount
		}
		games = append(games, GameCounts{ss.GameTypeCd, ss.GameCount})
	}

	return &SummaryBase{Players: players, Games: games, Scope: scope, LastRefreshed: lastRefreshed}, nil

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

// ActivePlayerToActivePlayerBase converts a raw ActivePlayer model into a better format for presentation
func ActivePlayerToActivePlayerBase(in []*models.ActivePlayer) []*ActivePlayerBase {
	var out []*ActivePlayerBase
	for _, v := range in {
		nick := qstr.QStr(v.Nick)
		ap := ActivePlayerBase{
			SortOrder:    v.SortOrder,
			PlayerID:     v.PlayerID,
			Nick:         nick,
			HTMLNick:     nick.HTML(),
			StrippedNick: nick.Stripped(),
			AliveTime:    models.DurationString(v.AliveTime, true),
		}
		out = append(out, &ap)
	}

	return out
}

// ActivePlayersData retrieves the active players
func ActivePlayersData(limit, start int, db models.Datastore) ([]*ActivePlayerBase, error) {
	rawActivePlayers, err := db.RActivePlayers(limit, start)
	if err != nil {
		return nil, err
	}

	return ActivePlayerToActivePlayerBase(rawActivePlayers), nil
}

// ActivePlayersJSON returns active player stats in JSON form.
func ActivePlayersJSON(limit, start int, db models.Datastore) ([]byte, error) {
	rawData, err := ActivePlayersData(limit, start, db)
	if err != nil {
		return nil, err
	}

	// the JSON response (a single entry in the list)
	type Response struct {
		Rank      int    `json:"rank"`
		PlayerID  int    `json:"player_id"`
		Nick      string `json:"nick"`
		AliveTime string `json:"alivetime"`
	}

	var activePlayers []Response
	for _, ap := range rawData {
		resp := Response{
			Rank:      ap.SortOrder,
			PlayerID:  ap.PlayerID,
			Nick:      ap.StrippedNick,
			AliveTime: ap.AliveTime,
		}

		activePlayers = append(activePlayers, resp)
	}

	return json.Marshal(activePlayers)
}

// ActiveServerBase is the base type for marshalling active servers to other formats (HTML, JSON, etc).
type ActiveServerBase struct {
	SortOrder  int
	ServerID   int
	ServerName string
	PlayTime   string
}

// ActiveServersData retrieves the active servers
func ActiveServersData(limit, start int, db models.Datastore) ([]ActiveServerBase, error) {
	rawActiveServers, err := db.RActiveServers(limit, start)
	if err != nil {
		return nil, err
	}

	var activeServers []ActiveServerBase
	for _, v := range rawActiveServers {
		as := ActiveServerBase{
			SortOrder:  v.SortOrder,
			ServerID:   v.ServerID,
			ServerName: v.ServerName,
			PlayTime:   models.DurationString(v.PlayTime, true),
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
		Rank       int    `json:"rank"`
		ServerID   int    `json:"server_id"`
		ServerName string `json:"server_name"`
		PlayTime   string `json:"playtime"`
	}

	var activeServers []Response
	for _, as := range rawData {
		resp := Response{
			Rank:       as.SortOrder,
			ServerID:   as.ServerID,
			ServerName: as.ServerName,
			PlayTime:   as.PlayTime,
		}
		activeServers = append(activeServers, resp)
	}

	return json.Marshal(activeServers)
}

// ActiveMapBase is the base type returned for all formats (HTML, JSON, etc.).
type ActiveMapBase struct {
	SortOrder int
	MapID     int
	MapName   string
	Games     int
}

// ActiveMapsData retrieves the active maps
func ActiveMapsData(limit, start int, db models.Datastore) ([]ActiveMapBase, error) {
	rawActiveMaps, err := db.RActiveMaps(limit, start)
	if err != nil {
		return nil, err
	}

	var activeMaps []ActiveMapBase
	for _, v := range rawActiveMaps {
		am := ActiveMapBase{
			SortOrder: v.SortOrder,
			MapID:     v.MapID,
			MapName:   v.MapName,
			Games:     v.Games,
		}

		activeMaps = append(activeMaps, am)
	}

	return activeMaps, nil
}

// ActiveMapsJSON returns active map stats in JSON form.
func ActiveMapsJSON(limit, start int, db models.Datastore) ([]byte, error) {
	rawData, err := ActiveMapsData(limit, start, db)
	if err != nil {
		return nil, err
	}

	// the JSON response (a single entry in the list)
	type Response struct {
		Rank    int    `json:"rank"`
		MapID   int    `json:"map_id"`
		MapName string `json:"map_name"`
		Games   int    `json:"games"`
	}

	var activeMaps []Response
	for _, am := range rawData {
		activeMaps = append(activeMaps, Response{am.SortOrder, am.MapID, am.MapName, am.Games})
	}

	return json.Marshal(activeMaps)
}
