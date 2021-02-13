package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"gitlab.com/xonotic/xonstat/pkg/leaderboard"
)

type topServerJSONResponse struct {
	Rank       int    `json:"rank"`
	ServerID   int    `json:"server_id"`
	ServerName string `json:"server_name"`
	PlayTime   string `json:"playtime"`
}

func activeServerBaseToJSON(rawData []leaderboard.ActiveServerBase) ([]byte, error) {
	var activeServers []topServerJSONResponse
	for _, as := range rawData {
		resp := topServerJSONResponse{
			Rank:       as.SortOrder,
			ServerID:   as.ServerID,
			ServerName: as.ServerName,
			PlayTime:   as.PlayTime,
		}
		activeServers = append(activeServers, resp)
	}

	return json.Marshal(activeServers)
}

type topServersResponse struct {
	ActiveServers []leaderboard.ActiveServerBase
	Start         int
	Next          int
	ShowMoreLink  bool
}

// TopServersHandler retrieves information about the top active servers by player aggregate playtime
func (ae *AppEnv) TopServersHandler(w http.ResponseWriter, r *http.Request) {
	acceptHeader := r.Header.Get("Accept")

	startStr := r.URL.Query().Get("start")
	start, err := strconv.Atoi(startStr)
	if err != nil {
		start = 1
	}

	activeServers, err := leaderboard.ActiveServersData(20, start, ae.db)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}

	if acceptHeader == "application/json" {
		// JSON response
		bytes, err := activeServerBaseToJSON(activeServers)
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
		if len(activeServers) > 0 {
			next = activeServers[len(activeServers)-1].SortOrder + 1
		}

		response := topServersResponse{
			ActiveServers: activeServers,
			Start:         start,
			Next:          next,
			ShowMoreLink:  len(activeServers) == 20,
		}

		err = ae.templates["activeservers.page.html"].Execute(w, response)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}
	}
}
