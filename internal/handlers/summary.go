package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"gitlab.com/xonotic/xonstat/pkg/leaderboard"
)

type gameCountsJSON struct {
	GameTypeCd string `json:"game_type_cd"`
	GameCount  int    `json:"num_games"`
}

type summaryJSONResponse struct {
	Players       int              `json:"players"`
	Games         []gameCountsJSON `json:"games"`
	Scope         string           `json:"scope"`
	LastRefreshed string           `json:"last_refreshed"`
}

func summaryBaseToJSONResponse(base *leaderboard.SummaryBase) summaryJSONResponse {
	games := make([]gameCountsJSON, len(base.Games))
	for i, v := range base.Games {
		games[i] = gameCountsJSON{v.GameTypeCd, v.GameCount}
	}

	return summaryJSONResponse{base.Players, games, base.Scope, base.LastRefreshed.Dt.Format(time.RFC3339)}
}

// SummaryHandler godoc
// @summary Summary information about the games processed by the server in a given period of time.
// @Accept  json
// @Produce  json
// @Param scope query string false "scope (either 'day' or 'all')"
// @Success 200 {object} summaryJSONResponse
// @Router /summary [get]
func (ae *AppEnv) SummaryHandler(w http.ResponseWriter, r *http.Request) {
	scope := r.URL.Query().Get("scope")
	if scope != "all" && scope != "day" {
		scope = "all"
	}

	summaryStats, err := leaderboard.SummaryData(scope, ae.db)
	if err != nil {
		log.Printf("Error: %s", err)
		ae.FiveHundredHandler(w, r)
		return
	}

	response := summaryBaseToJSONResponse(summaryStats)
	bytes, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error: %s", err)
		ae.FiveHundredHandler(w, r)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}
