package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"gitlab.com/xonotic/xonstat/pkg/player"
	"gitlab.com/xonotic/xonstat/pkg/util"
)

// playerInfoResponse is the view-specific information about a player.
type playerInfoResponse struct {
	Player            *player.InfoBase
	OverallStats      map[string]*player.OverallStatsBase
	GameTypeSummaries []*player.GameTypeSummaryBase
}

// playerJSON is the JSON-specific representation of a player.
type playerJSON struct {
	PlayerID     int    `json:"player_id"`
	Nick         string `json:"nick"`
	StrippedNick string `json:"stripped_nick"`
	ActiveInd    bool   `json:"active_ind"`
	Joined       string `json:"joined"`
	JoinedEpoch  int    `json:"joined_epoch"`
	JoinedFuzzy  string `json:"joined_fuzzy"`
}

// gamesPlayedJSON is the JSON-specific representation of the games played by a player.
type gamesPlayedJSON struct {
	GameTypeCd string  `json:"game_type_cd"`
	Games      int     `json:"games"`
	Wins       int     `json:"wins"`
	Losses     int     `json:"losses"`
	WinPct     float32 `json:"win_pct"`
}

// overallStatsJSON is the JSON-specific representation of the overall stats of a player.
type overallStatsJSON struct {
	GameTypeCd       string  `json:"game_type_cd"`
	Kills            int     `json:"total_kills"`
	Deaths           int     `json:"total_deaths"`
	KDRatio          float32 `json:"k_d_ratio"`
	Captures         int     `json:"total_captures"`
	Pickups          int     `json:"total_pickups"`
	CapRatio         float32 `json:"cap_ratio"`
	CarrierFrags     int     `json:"total_carrier_frags"`
	LastPlayed       string  `json:"last_played"`
	LastPlayedEpoch  int     `json:"last_played_epoch"`
	LastPlayedFuzzy  string  `json:"last_played_fuzzy"`
	TotalPlayingTime int     `json:"total_playing_time"`
}

// playerInfoJSONResponse is the JSON-specific response type
type playerInfoJSONResponse struct {
	Player       playerJSON                  `json:"player"`
	GamesPlayed  map[string]gamesPlayedJSON  `json:"games_played"`
	OverallStats map[string]overallStatsJSON `json:"overall_stats"`
}

// newPlayerInfoJSONResponse converts a regular response into a JSON-specific one
func newPlayerInfoJSONResponse(response *playerInfoResponse) playerInfoJSONResponse {
	var jsonResponse playerInfoJSONResponse

	jsonResponse.Player = playerJSON{
		PlayerID:     response.Player.PlayerID,
		Nick:         response.Player.Nick.Nick,
		StrippedNick: response.Player.Nick.NickStripped,
		ActiveInd:    response.Player.ActiveInd,
		Joined:       response.Player.CreateDt.Dt.Format(time.RFC3339),
		JoinedEpoch:  int(response.Player.CreateDt.Epoch),
		JoinedFuzzy:  response.Player.CreateDt.Fuzzy,
	}

	jsonResponse.GamesPlayed = make(map[string]gamesPlayedJSON, len(response.GameTypeSummaries))
	for _, summary := range response.GameTypeSummaries {
		jsonResponse.GamesPlayed[summary.GameTypeCd] = gamesPlayedJSON{
			GameTypeCd: summary.GameTypeCd,
			Games:      summary.Games,
			Wins:       summary.Wins,
			Losses:     summary.Losses,
			WinPct:     summary.WinRatio,
		}
	}

	jsonResponse.OverallStats = make(map[string]overallStatsJSON, len(response.OverallStats))
	for _, overall := range response.OverallStats {
		jsonResponse.OverallStats[overall.GameTypeCd] = overallStatsJSON{
			GameTypeCd:       overall.GameTypeCd,
			Kills:            overall.Kills,
			Deaths:           overall.Deaths,
			KDRatio:          overall.KDRatio,
			Captures:         overall.Captures,
			Pickups:          overall.Pickups,
			CapRatio:         overall.CapRatio,
			CarrierFrags:     overall.CarrierFrags,
			LastPlayed:       overall.LastPlayed.Dt.Format(time.RFC3339),
			LastPlayedEpoch:  int(overall.LastPlayed.Epoch),
			LastPlayedFuzzy:  overall.LastPlayed.Fuzzy,
			TotalPlayingTime: int(overall.TimePlayed.Duration.Seconds()),
		}
	}

	return jsonResponse
}

// PlayerInfoHandler godoc
// @Summary Player information
// @Accept  json
// @Produce  json
// @Param id path int true "player_id"
// @Success 200 {object} playerInfoJSONResponse
// @Router /player/{id} [get]
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

	summaries, err := player.GameTypeSummaryData(ae.db, playerID)
	if err != nil {
		log.Printf("Error: %s", err)
		ae.FiveHundredHandler(w, r)
		return
	}

	rawOverallStats, err := player.OverallStatsData(ae.db, playerID)
	if err != nil {
		log.Printf("Error: %s", err)
		ae.FiveHundredHandler(w, r)
		return
	}

	// The raw results here include K:D for CTS. While we need this for showing an
	// accurate scoreboard, we do not want it included for other things like the
	// player's overall KD ratio. We thus need to remove it or subtract it for just
	// that game type here.
	ctsSeen := false

	overallStats := make(map[string]*player.OverallStatsBase, len(rawOverallStats))
	for _, e := range rawOverallStats {
		if e.GameTypeCd == "cts" {
			ctsSeen = true
		}

		overallStats[e.GameTypeCd] = e
	}

	// This player has CTS information, so here we remove the K:D 
	// pieces from the overall numbers. This prevents displaying a
	// super low overall K:D.
	if ctsSeen {
		overall := overallStats["overall"]
		cts := overallStats["cts"]
		overall.Kills -= cts.Kills
		overall.Deaths -= cts.Deaths
		overall.KDRatio = util.Ratio(overall.Kills, overall.Deaths)
	}

	response := &playerInfoResponse{
		Player:            info,
		GameTypeSummaries: summaries,
		OverallStats:      overallStats,
	}

	if acceptHeader == "application/json" {
		jsonResponse := newPlayerInfoJSONResponse(response)
		bytes, _ := json.Marshal(jsonResponse)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	} else {
		err = ae.templates["playerinfo.page.html"].Execute(w, response)
		if err != nil {
			log.Printf("Error: %s", err)
			ae.FiveHundredHandler(w, r)
			return
		}
	}
}
