package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"gitlab.com/xonotic/xonstat/pkg/game"
	"gitlab.com/xonotic/xonstat/pkg/submission"
)

type recentGamesJSONResponse struct {
	GameID          int    `json:"game_id"`
	GameTypeCd      string `json:"game_type_cd"`
	GameTypeDescr   string `json:"game_type_descr"`
	ServerID        int    `json:"server_id"`
	ServerName      string `json:"server_name"`
	MapID           int    `json:"map_id"`
	MapName         string `json:"map_name"`
	WinningTeam     int    `json:"winning_team"`
	WinningPlayerID int    `json:"winning_player_id"`
	WinningNick     string `json:"nick"`
	CreateDt        string `json:"create_dt"`
}

func recentGameBaseToJSON(rawData []game.RecentGameBase) ([]byte, error) {
	var rgs []recentGamesJSONResponse
	for _, rg := range rawData {
		r := recentGamesJSONResponse{
			GameID:          rg.GameID,
			GameTypeCd:      rg.GameTypeCd,
			GameTypeDescr:   rg.GameTypeDescr,
			ServerID:        rg.ServerID,
			ServerName:      rg.ServerName,
			MapID:           rg.MapID,
			MapName:         rg.MapName,
			WinningTeam:     rg.WinningTeam,
			WinningPlayerID: rg.WinningPlayerID,
			WinningNick:     rg.WinningNick,
			CreateDt:        rg.CreateDt.Dt.Format(time.RFC3339),
		}

		rgs = append(rgs, r)
	}

	return json.Marshal(rgs)
}

type recentGamesResponse struct {
	GameTypeCds      []string
	ActiveGameTypeCd string
	RecentGames      []game.RecentGameBase
	ShowMoreLink     bool
	NextQueryStr     template.URL
}

// RecentGamesHandler godoc
// @Summary Games played by various filter criteria
// @Accept  json
// @Produce  json
// @Param server_id query int false "server_id filter"
// @Param map_id query int false "map_id filter"
// @Param player_id query int false "player_id filter"
// @Param game_type_cd query string false "game_type_cd filter"
// @Param start_game_id query int false "game_id range lower bound"
// @Param end_game_id query int false "game_id range upper bound"
// @Success 200 {object} []recentGamesJSONResponse
// @Router /games [get]
func (ae *AppEnv) RecentGamesHandler(w http.ResponseWriter, r *http.Request) {
	acceptHeader := r.Header.Get("Accept")

	params := r.URL.Query()

	serverID, err := strconv.Atoi(params.Get("server_id"))
	if err != nil {
		serverID = game.EmptyServerID
	}

	mapID, err := strconv.Atoi(params.Get("map_id"))
	if err != nil {
		mapID = game.EmptyMapID
	}

	playerID, err := strconv.Atoi(params.Get("player_id"))
	if err != nil {
		playerID = game.EmptyPlayerID
	}

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
		startGameID = game.EmptyStartGameID
	}

	endGameID, err := strconv.Atoi(params.Get("end_game_id"))
	params.Del("end_game_id")
	if err != nil {
		endGameID = game.EmptyEndGameID
	}

	recentGames, err := game.RecentGamesData(ae.db, serverID, mapID, playerID, gameTypeCd, nil, startGameID, endGameID, 20)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}

	if acceptHeader == "application/json" {
		// JSON response
		recentGames, err := recentGameBaseToJSON(recentGames)
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

		// Set the query string for pagination
		if len(recentGames) > 0 {
			params.Set("start_game_id", fmt.Sprintf("%d", recentGames[len(recentGames)-1].GameID-1))
		}

		nextQueryStr := template.URL(params.Encode())

		response := recentGamesResponse{
			GameTypeCds:      gameTypeCds,
			ActiveGameTypeCd: gameTypeCd,
			RecentGames:      recentGames,
			ShowMoreLink:     len(recentGames) == 20,
			NextQueryStr:     nextQueryStr,
		}

		err = ae.templates["gameindex.page.html"].Execute(w, response)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}
	}
}
