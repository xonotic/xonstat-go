package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/spf13/viper"
	"gitlab.com/xonotic/xonstat/pkg/game"
)

// RecentGamesHandler retrieves information about games played by varios filter criteria
func (ae *AppEnv) RecentGamesHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: lots of GET query params and input validation here. For now just go
	// for the simple case for the leaderboard.

	recentGamesDays := viper.GetInt("RecentGamesDays")

	now := time.Now()
	cutoff := now.AddDate(0, 0, -1 * recentGamesDays)

	recentGames, err := game.RecentGamesJSON(ae.db, -1, -1, -1, "", cutoff, -1, -1, 20)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(recentGames)
}
