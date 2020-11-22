package handlers

import (
	"encoding/json"
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
		gameInfo, err := game.GameInfoData(ae.db, gameID)
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

// GameWeaponInfoHandler retrieves information about the weapons in a game by its ID.
func (ae *AppEnv) GameWeaponInfoHandler(w http.ResponseWriter, r *http.Request) {
	gameID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("Invalid or missing game ID value: %s", err)
		http.Error(w, fmt.Sprintf("404 %s", http.StatusText(404)), 404)
		return
	}

	gameWeaponInfo, err := game.GameWeaponInfoData(ae.db, gameID)
	if err != nil {
		log.Printf("Could not process weapon info for game ID %d: %s", gameID, err)
		http.Error(w, fmt.Sprintf("404 %s", http.StatusText(404)), 404)
		return
	}

	type Series struct {
		Key string `json:"key"`
		Color string `json:"color"`
		Values []*game.PlayerWeaponStatBase `json:"values"`
	}

	weaponToSeries := make(map[string]*Series)

	// Assemble the weapon data in the "shape" that NVD3 wants. The series is the 
	// distinct set of weapons, with each entry in the set being the list of weapon
	// stats by the players. 
	for _, ws := range gameWeaponInfo.WeaponStats {
		series, ok := weaponToSeries[ws.WeaponCd]
		if ok {
			// Series already created, append
			series.Values = append(series.Values, ws)
		} else {
			// Series not created yet. Create and add.
			series := Series{Key: ws.WeaponCd, Values: make([]*game.PlayerWeaponStatBase, 0)}
			series.Values = append(series.Values, ws)
			weaponToSeries[ws.WeaponCd] = &series
		}
	}

	// Flatten the map into a plain list of series
	series := make([]*Series, 0)
	for _, s := range weaponToSeries {
		series = append(series, s)
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	bytes, err := json.Marshal(series)
	if err != nil {
		log.Printf("Could not marshal weapon info to JSON for game ID %d: %s", gameID, err)
		http.Error(w, fmt.Sprintf("404 %s", http.StatusText(404)), 404)
		return
	}

	w.Write(bytes)
}
