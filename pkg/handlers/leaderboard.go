package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

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

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(summaryStats)
}

// TopActiveHandler retrieves information about the top active players by playing time
func (ae *AppEnv) TopActiveHandler(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start")
	start, err := strconv.Atoi(startStr)
	if err != nil {
		start = 1
	}

	bytes, err := leaderboard.ActivePlayersJSON(10, start, ae.db)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}
