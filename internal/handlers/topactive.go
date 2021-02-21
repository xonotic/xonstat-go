package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"gitlab.com/xonotic/xonstat/pkg/leaderboard"
)

type topActiveJSONResponse struct {
	Rank      int    `json:"rank"`
	PlayerID  int    `json:"player_id"`
	Nick      string `json:"nick"`
	AliveTime string `json:"alivetime"`
}

func activePlayerBaseToJSON(aps []*leaderboard.ActivePlayerBase) ([]byte, error) {
	var activePlayers []topActiveJSONResponse
	for _, ap := range aps {
		resp := topActiveJSONResponse{
			Rank:      ap.SortOrder,
			PlayerID:  ap.PlayerID,
			Nick:      ap.StrippedNick,
			AliveTime: ap.AliveTime,
		}

		activePlayers = append(activePlayers, resp)
	}

	return json.Marshal(activePlayers)
}

type topActiveResponse struct {
	ActivePlayers []*leaderboard.ActivePlayerBase
	Start         int
	Next          int
	ShowMoreLink  bool
}

// TopActiveHandler godoc
// @Summary Top active players by playing time
// @Accept  json
// @Produce  json
// @Param start query int false "rank pagination offset"
// @Success 200 {object} topActiveJSONResponse
// @Router /topactive [get]
func (ae *AppEnv) TopActiveHandler(w http.ResponseWriter, r *http.Request) {
	acceptHeader := r.Header.Get("Accept")

	startStr := r.URL.Query().Get("start")
	start, err := strconv.Atoi(startStr)
	if err != nil {
		start = 1
	}

	activePlayers, err := leaderboard.ActivePlayersData(20, start, ae.db)
	if err != nil {
		log.Printf("Error: %s", err)
		ae.FiveHundredHandler(w, r)
		return
	}

	if acceptHeader == "application/json" {
		// JSON response
		bytes, err := activePlayerBaseToJSON(activePlayers)
		if err != nil {
			log.Printf("Error: %s", err)
			ae.FiveHundredHandler(w, r)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	} else {
		// HTML response
		next := 1
		if len(activePlayers) > 0 {
			next = activePlayers[len(activePlayers)-1].SortOrder + 1
		}

		response := topActiveResponse{
			ActivePlayers: activePlayers,
			Start:         start,
			Next:          next,
			ShowMoreLink:  len(activePlayers) == 20,
		}

		err = ae.templates["activeplayers.page.html"].Execute(w, response)
		if err != nil {
			log.Printf("Error: %s", err)
			ae.FiveHundredHandler(w, r)
			return
		}
	}
}
