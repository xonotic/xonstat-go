package handlers

import (
	"fmt"
	"log"
	"net/http"

	"gitlab.com/xonotic/xonstat/pkg/game"
	"gitlab.com/xonotic/xonstat/pkg/submission"
)

// RecentGamesHandler retrieves information about games played by varios filter criteria
func (ae *AppEnv) RecentGamesHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: lots of GET query params and input validation here. For now just go
	// for the simple case for the leaderboard.
	acceptHeader := r.Header.Get("Accept")

	gameTypeCd := r.URL.Query().Get("game_type_cd")
	if !submission.IsSupportedGameType(gameTypeCd) {
		gameTypeCd = ""
	}

	if acceptHeader == "application/json" {
		// JSON response
		recentGames, err := game.RecentGamesJSON(ae.db, -1, -1, -1, gameTypeCd, nil, -1, -1, 20)
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
		gameTypeCds := []string{"overall","duel","ctf","dm","tdm","ca","kh","ft","as","dom","nb","cts"}

		recentGames, err := game.RecentGamesData(ae.db, -1, -1, -1, gameTypeCd, nil, -1, -1, 20)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}

		activeGameTypeCd := "overall"
		if gameTypeCd != "" {
			activeGameTypeCd = gameTypeCd
		}

		type Data struct {
			ActiveGameTypeCd string
			GameTypeCds []string
			RecentGames []game.RecentGameBase
		}

		data := Data{
			ActiveGameTypeCd: activeGameTypeCd,
			GameTypeCds: gameTypeCds,
			RecentGames: recentGames,
		}		
		
		err = ae.templates["gameindex.page.html"].Execute(w, data)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}
	}
}
