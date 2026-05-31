package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"gitlab.com/xonotic/xonstat/pkg/game"
	"gitlab.com/xonotic/xonstat/pkg/leaderboard"
	"gitlab.com/xonotic/xonstat/pkg/server"
)

type heatmapResponse struct {
	Server   *server.InfoBase
}

// HeatmapHandler godoc
// @summary Data around the games played in hourly intervals along the week as a matrix.
// @Accept  json
// @Produce  json
// @Param server_id query int false "server_id filter"
// @Param hashkey query string false "server hashkey filter"
// @Success 200 {object} [][]int
// @Router /heatmap [get]
func (ae *AppEnv) HeatmapHandler(w http.ResponseWriter, r *http.Request) {
	acceptHeader := r.Header.Get("Accept")

	params := r.URL.Query()

	var err error
	serverID := game.EmptyServerID

	if idStr := params.Get("server_id"); idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err == nil {
			serverID = id
		}
	} else if hashkey := params.Get("hashkey"); hashkey != "" {
		servers, err := ae.db.RServersByHashkey(hashkey)
		if err != nil {
			log.Printf("Error looking up server by hashkey: %s", err)
			ae.FiveHundredHandler(w, r)
			return
		}

		if len(servers) > 0 {
			serverID = servers[0].ServerID
		}
	}

	if acceptHeader == "application/json" {

		heatmap, err := leaderboard.HeatmapData(ae.db, serverID)
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
	} else {
		var response heatmapResponse

		if serverID != game.EmptyServerID {
			s, err := server.InfoData(ae.db, serverID)
			if err != nil {
				log.Printf("Error: %s", err)
				ae.NotFoundHandler(w, r)
				return
			}
			response.Server = s
		}

		err = ae.templates["heatmap.page.html"].Execute(w, response)
		if err != nil {
			log.Printf("Error: %s", err)
			ae.FiveHundredHandler(w, r)
			return
		}
	}
}
