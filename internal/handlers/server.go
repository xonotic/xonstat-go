package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/spf13/viper"
	"gitlab.com/xonotic/xonstat/pkg/game"
	"gitlab.com/xonotic/xonstat/pkg/leaderboard"
	"gitlab.com/xonotic/xonstat/pkg/models"
	"gitlab.com/xonotic/xonstat/pkg/server"
)

// ServerInfoHandler retrieves information about individual servers.
func (ae *AppEnv) ServerInfoHandler(w http.ResponseWriter, r *http.Request) {
	acceptHeader := r.Header.Get("Accept")

	serverID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("Invalid or missing server ID value: %s", err)
		ae.NotFoundHandler(w, r)
		return
	}

	if acceptHeader == "application/json" {
		// JSON response
		s, err := server.InfoJSON(ae.db, serverID)
		if err != nil {
			log.Printf("Error: %s", err)
			ae.NotFoundHandler(w, r)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(s)
	} else {
		// HTML response
		s, err := server.InfoData(ae.db, serverID)
		if err != nil {
			log.Printf("Error: %s", err)
			ae.NotFoundHandler(w, r)
			return
		}

		topScoringPlayers, _ := server.TopScorerData(ae.db, serverID)
		topActivePlayers, _ := server.TopActivePlayersData(ae.db, serverID)
		topMapsPlayed, _ := server.TopMapsData(ae.db, serverID)

		recentGamesCutoff := time.Now().UTC().AddDate(0, 0, -1*viper.GetInt("RecentGamesDays"))
		recentGames, _ := game.RecentGamesData(ae.db, serverID, game.EmptyMapID, game.EmptyPlayerID,
			game.EmptyGameTypeCd, &recentGamesCutoff, game.EmptyStartGameID, game.EmptyEndGameID, 20)

		// The structure passed to the template.
		type Data struct {
			Server            server.InfoBase
			TopScoringPlayers []*server.TopScorerBase
			TopActivePlayers  []*leaderboard.ActivePlayerBase
			TopMapsPlayed     []*models.ActiveMap
			RecentGames       []game.RecentGameBase
		}

		data := Data{
			Server:            *s,
			TopScoringPlayers: topScoringPlayers,
			TopActivePlayers:  topActivePlayers,
			TopMapsPlayed:     topMapsPlayed,
			RecentGames:       recentGames,
		}

		err = ae.templates["serverinfo.page.html"].Execute(w, data)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}
	}
}

// ServerTopScorerEntry is one entry in the list of a server's top scoring players.
type ServerTopScorerEntry struct {
	PlayerID int    `json:"player_id"`
	Nick     string `json:"nick"`
	Score    int    `json:"score"`
	Rank     int    `json:"rank"`
}

// NewServerTopScorerEntry creates a new ServerTopScorerEntry from its base type.
func NewServerTopScorerEntry(base *server.TopScorerBase) *ServerTopScorerEntry {
	return &ServerTopScorerEntry{
		PlayerID: base.PlayerID,
		Nick:     base.Nick,
		Score:    base.Score,
		Rank:     base.SortOrder,
	}
}

// ServerTopScorersResponse is the response type for the ServerTopScorersHandler (JSON only).
type ServerTopScorersResponse struct {
	ServerID   int                     `json:"server_id"`
	TopScorers []*ServerTopScorerEntry `json:"top_scorers"`
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
	topScorers := make([]*ServerTopScorerEntry, len(topScoringPlayers))
	for i, base := range topScoringPlayers {
		topScorers[i] = NewServerTopScorerEntry(base)
	}

	response := ServerTopScorersResponse{
		ServerID:   serverID,
		TopScorers: topScorers,
	}

	bytes, _ := json.Marshal(response)

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}


