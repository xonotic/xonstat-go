package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/spf13/viper"
	"gitlab.com/xonotic/xonstat/pkg/game"
)

// PlayerRecentGamesFragmentHandler retrieves just the recent games table.
func (ae *AppEnv) PlayerRecentGamesFragmentHandler(w http.ResponseWriter, r *http.Request) {
	playerID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("invalid or missing player ID value: %s", err)
		ae.NotFoundHandler(w, r)
		return
	}

	params := r.URL.Query()
	gameTypeCd := params.Get("game_type_cd")

	// The "overall" game type means no filter.
	if gameTypeCd == "overall" {
		gameTypeCd = ""
	}

	recentGamesCutoff := time.Now().UTC().AddDate(0, 0, -1*viper.GetInt("RecentGamesDays"))
	recentGames, _ := game.RecentGamesData(ae.db, game.EmptyServerID, game.EmptyMapID, playerID,
		gameTypeCd, &recentGamesCutoff, game.EmptyStartGameID, game.EmptyEndGameID, 20)

	err = ae.templates["recentgames.fragment.page.html"].Execute(w, recentGames)
	if err != nil {
		log.Printf("Error: %s", err)
		ae.FiveHundredHandler(w, r)
		return
	}
}