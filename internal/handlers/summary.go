package handlers

import (
	"encoding/json"
	"fmt"
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

// SummaryHandler retrieves information about the summary stats
func (ae *AppEnv) SummaryHandler(w http.ResponseWriter, r *http.Request) {
	scope := r.URL.Query().Get("scope")
	if scope != "all" && scope != "day" {
		scope = "all"
	}

	summaryStats, err := leaderboard.SummaryData(scope, ae.db)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}

	response := summaryBaseToJSONResponse(summaryStats)
	bytes, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}
