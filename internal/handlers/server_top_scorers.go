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

// NewServerTopScorerEntry creates a new ServerTopScorerEntry from its base type.
func NewServerTopScorerEntry(base *server.TopScorerBase) *serverTopScorerEntry {
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

// ServerTopScorersHandler retrieves information about the top scoring players on a given server.
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
		topScorers[i] = NewServerTopScorerEntry(base)
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
