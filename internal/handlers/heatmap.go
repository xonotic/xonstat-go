package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"gitlab.com/xonotic/xonstat/pkg/leaderboard"
)

// HeatmapHandler godoc
// @summary Data around the games played in hourly intervals along the week as a matrix.
// @Accept  json
// @Produce  json
// @Success 200 {object} [][]int
// @Router /heatmap [get]
func (ae *AppEnv) HeatmapHandler(w http.ResponseWriter, r *http.Request) {
	heatmap, err := leaderboard.HeatmapData(ae.db)
	if err != nil {
		log.Printf("Error: %s", err)
		ae.FiveHundredHandler(w, r)
		return
	}

	bytes, err := json.Marshal(heatmap)
	if err != nil {
		log.Printf("Error: %s", err)
		ae.FiveHundredHandler(w, r)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}
