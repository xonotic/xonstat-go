package handlers

import (
	"fmt"
	"log"
	"net/http"

	"gitlab.com/xonotic/xonstat/pkg/leaderboard"
)

// SummaryStatsHandler retrieves information about the summary stats
func (ae *AppEnv) SummaryStatsHandler(w http.ResponseWriter, r *http.Request) {
	scope := r.URL.Query().Get("scope")
	if scope != "all" && scope != "day" {
		scope = "all"
	}

	summaryStats, err := leaderboard.SummaryStatsJSON(scope, ae.db)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(summaryStats)
}
