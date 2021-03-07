package leaderboard

import (
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
			AliveTime:    models.DurationString(v.AliveTime, "short"),
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
			PlayTime:   models.DurationString(v.PlayTime, "short"),
		}
		activeServers = append(activeServers, as)
	}

	return activeServers, nil
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
