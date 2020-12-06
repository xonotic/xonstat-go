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
	"gitlab.com/xonotic/xonstat/pkg/player"
)

// PlayerInfoResponse is the view-specific information about a map related information.
type PlayerInfoResponse struct {
	Player            *player.InfoBase
	GameTypeSummaries []*player.GameTypeSummaryBase
	RecentGames       []game.RecentGameBase
}

// PlayerInfoHandler is the web handler for retrieving player information
func (ae *AppEnv) PlayerInfoHandler(w http.ResponseWriter, r *http.Request) {
	acceptHeader := r.Header.Get("Accept")

	playerID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("Invalid or missing player ID value: %s", err)
		ae.NotFoundHandler(w, r)
		return
	}

	info, err := player.InfoData(ae.db, playerID)
	if err != nil {
		log.Printf("Invalid or missing player ID value: %s", err)
		ae.NotFoundHandler(w, r)
		return
	}

	summaries, _ := player.GameTypeSummaryData(ae.db, playerID)

	recentGamesCutoff := time.Now().UTC().AddDate(0, 0, -1*viper.GetInt("RecentGamesDays"))
	recentGames, _ := game.RecentGamesData(ae.db, game.EmptyServerID, game.EmptyMapID, playerID,
		game.EmptyGameTypeCd, &recentGamesCutoff, game.EmptyStartGameID, game.EmptyEndGameID, 20)

	response := &PlayerInfoResponse{
		Player:            info,
		GameTypeSummaries: summaries,
		RecentGames:       recentGames,
	}

	if acceptHeader == "application/json" {
		bytes, _ := json.Marshal(response)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	} else {
		err = ae.templates["playerinfo.page.html"].Execute(w, response)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}
	}
}
