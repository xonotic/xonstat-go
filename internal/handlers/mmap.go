package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"gitlab.com/xonotic/xonstat/pkg/leaderboard"
	"gitlab.com/xonotic/xonstat/pkg/mmap"
	"gitlab.com/xonotic/xonstat/pkg/server"
	"strconv"
)

// MapInfoResponse is the view-specific information about a map related information.
type MapInfoResponse struct {
	Map               *mmap.InfoBase
	TopScoringPlayers []*server.TopScorerBase
	TopActivePlayers  []*leaderboard.ActivePlayerBase
	TopActiveServers  []*leaderboard.ActiveServerBase
}

// MapInfoHandler is the web handler for retrieving map information
func (ae *AppEnv) MapInfoHandler(w http.ResponseWriter, r *http.Request) {
	acceptHeader := r.Header.Get("Accept")

	mapID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("Invalid or missing map ID value: %s", err)
		ae.NotFoundHandler(w, r)
	}

	info, err := mmap.InfoData(ae.db, mapID)
	if err != nil {
		log.Printf("Invalid or missing map ID value: %s", err)
		ae.NotFoundHandler(w, r)
	}

	topScorers, _ := mmap.TopScorerData(ae.db, mapID)
	topActive, _ := mmap.TopActivePlayersData(ae.db, mapID)
	topServers, _ := mmap.TopActiveServersData(ae.db, mapID)

	response := &MapInfoResponse{
		Map:               info,
		TopScoringPlayers: topScorers,
		TopActivePlayers:  topActive,
		TopActiveServers:  topServers,
	}

	if acceptHeader == "application/json" {
		bytes, _ := json.Marshal(response)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	} else {
		err = ae.templates["mapinfo.page.html"].Execute(w, response)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}
	}
}
