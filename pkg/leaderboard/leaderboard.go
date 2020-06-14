package leaderboard

import (
	"encoding/json"
	"time"

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