package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"gitlab.com/xonotic/xonstat/pkg/server"
)

// serverTopScorerEntry is one entry in the list of a server's top scoring players.
type serverTopScorerEntry struct {
	PlayerID int    `json:"player_id"`
	Nick     string `json:"nick"`
	Score    int    `json:"score"`
	Rank     int    `json:"rank"`
}

// newServerTopScorerEntry creates a new ServerTopScorerEntry from its base type.
func newServerTopScorerEntry(base *server.TopScorerBase) *serverTopScorerEntry {
	return &serverTopScorerEntry{
		PlayerID: base.PlayerID,
		Nick:     base.Nick,
		Score:    base.Score,
		Rank:     base.SortOrder,
	}
}

// serverTopScorersResponse is the response type for the ServerTopScorersHandler (JSON only).
type serverTopScorersResponse struct {
	ServerID   int                     `json:"server_id"`
	TopScorers []*serverTopScorerEntry `json:"top_scorers"`
}

// ServerTopScorersHandler godoc
// @Summary Top scoring players on a particular server over a fixed time period
// @Accept  json
// @Produce  json
// @Param id path int false "server_id"
// @Success 200 {object} serverTopScorersResponse
// @Router /server/{id}/topscorers [get]
func (ae *AppEnv) ServerTopScorersHandler(w http.ResponseWriter, r *http.Request) {
	acceptHeader := r.Header.Get("Accept")

	// This handler only accepts JSON.
	if acceptHeader != "application/json" {
		log.Printf("Invalid or missing Accept header: %s", acceptHeader)
		ae.NotFoundHandler(w, r)
		return
	}

	serverID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("Invalid or missing server ID value: %s", err)
		ae.NotFoundHandler(w, r)
		return
	}

	topScoringPlayers, _ := server.TopScorerData(ae.db, serverID)
	topScorers := make([]*serverTopScorerEntry, len(topScoringPlayers))
	for i, base := range topScoringPlayers {
		topScorers[i] = newServerTopScorerEntry(base)
	}

	response := serverTopScorersResponse{
		ServerID:   serverID,
		TopScorers: topScorers,
	}

	bytes, _ := json.Marshal(response)

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}
