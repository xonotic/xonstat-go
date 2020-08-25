package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"gitlab.com/xonotic/xonstat/pkg/game"
	"gitlab.com/xonotic/xonstat/pkg/submission"
)

// RecentGamesHandler retrieves information about games played by varios filter criteria
func (ae *AppEnv) RecentGamesHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: lots of GET query params and input validation here. For now just go
	// for the simple case for the leaderboard.
	acceptHeader := r.Header.Get("Accept")

	params := r.URL.Query()

	gameTypeCd := params.Get("game_type_cd")
	if !submission.IsSupportedGameType(gameTypeCd) {
		// It is not a supported game type. Use the default for the DB query and don't include it
		// in the resulting query string.
		gameTypeCd = ""
		params.Del("game_type_cd")
	}

	startGameID, err := strconv.Atoi(params.Get("start_game_id"))
	params.Del("start_game_id")
	if err != nil {
		startGameID = -1
	}

	if acceptHeader == "application/json" {
		// JSON response
		recentGames, err := game.RecentGamesJSON(ae.db, -1, -1, -1, gameTypeCd, nil, startGameID, -1, 20)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(recentGames)
	} else {
		// HTML response
		gameTypeCds := []string{"overall", "duel", "ctf", "dm", "tdm", "ca", "kh", "ft", "as", "dom", "nb", "cts"}

		recentGames, err := game.RecentGamesData(ae.db, -1, -1, -1, gameTypeCd, nil, startGameID, -1, 20)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}

		// Set the query string for pagination
		if len(recentGames) > 0 {
			params.Set("start_game_id", fmt.Sprintf("%d", recentGames[len(recentGames)-1].GameID-1))
		}

		nextQueryStr := template.URL(params.Encode())

		type Data struct {
			GameTypeCds      []string
			ActiveGameTypeCd string
			RecentGames      []game.RecentGameBase
			ShowMoreLink     bool
			NextQueryStr     template.URL
		}

		data := Data{
			GameTypeCds:      gameTypeCds,
			ActiveGameTypeCd: gameTypeCd,
			RecentGames:      recentGames,
			ShowMoreLink:     len(recentGames) == 20,
			NextQueryStr:     nextQueryStr,
		}

		err = ae.templates["gameindex.page.html"].Execute(w, data)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}
	}
}

// GameInfoHandler retrieves information about a game by its ID.
func (ae *AppEnv) GameInfoHandler(w http.ResponseWriter, r *http.Request) {
	acceptHeader := r.Header.Get("Accept")

	gameID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("Invalid or missing game ID value: %s", err)
		http.Error(w, fmt.Sprintf("404 %s", http.StatusText(404)), 404)
		return
	}

	if acceptHeader == "application/json" {
		// JSON response
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// w.Write()
	} else {
		// HTML response
		gameInfo, err := game.InfoData(ae.db, gameID)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}

		err = ae.templates["gameinfo.page.html"].Execute(w, gameInfo)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}
	}
}
