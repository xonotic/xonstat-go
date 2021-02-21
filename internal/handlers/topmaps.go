package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"gitlab.com/xonotic/xonstat/pkg/leaderboard"
)

type topMapsJSONResponse struct {
	Rank    int    `json:"rank"`
	MapID   int    `json:"map_id"`
	MapName string `json:"map_name"`
	Games   int    `json:"games"`
}

func activeMapsBaseToJSON(rawData []leaderboard.ActiveMapBase) ([]byte, error) {
	var activeMaps []topMapsJSONResponse
	for _, am := range rawData {
		activeMaps = append(activeMaps, topMapsJSONResponse{am.SortOrder, am.MapID, am.MapName, am.Games})
	}

	return json.Marshal(activeMaps)
}

type topMapsResponse struct {
	ActiveMaps   []leaderboard.ActiveMapBase
	Start        int
	Next         int
	ShowMoreLink bool
}

// TopMapsHandler godoc
// @Summary Top maps by number of games played
// @Accept  json
// @Produce  json
// @Param start query int false "rank pagination offset"
// @Success 200 {object} topMapsJSONResponse
// @Router /topmaps [get]
func (ae *AppEnv) TopMapsHandler(w http.ResponseWriter, r *http.Request) {
	acceptHeader := r.Header.Get("Accept")

	startStr := r.URL.Query().Get("start")
	start, err := strconv.Atoi(startStr)
	if err != nil {
		start = 1
	}

	activeMaps, err := leaderboard.ActiveMapsData(20, start, ae.db)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}

	if acceptHeader == "application/json" {
		// JSON response
		bytes, err := activeMapsBaseToJSON(activeMaps)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	} else {
		// HTML response

		next := 1
		if len(activeMaps) > 0 {
			next = activeMaps[len(activeMaps)-1].SortOrder + 1
		}

		data := topMapsResponse{
			ActiveMaps:   activeMaps,
			Start:        start,
			Next:         next,
			ShowMoreLink: len(activeMaps) == 20,
		}

		err = ae.templates["activemaps.page.html"].Execute(w, data)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}
	}
}
