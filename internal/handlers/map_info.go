package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/spf13/viper"
	"gitlab.com/xonotic/xonstat/pkg/game"
	"gitlab.com/xonotic/xonstat/pkg/leaderboard"
	"gitlab.com/xonotic/xonstat/pkg/mmap"
	"gitlab.com/xonotic/xonstat/pkg/server"
	"strconv"
)

type mapInfoJSONResponse struct {
	MapID    int    `json:"map_id"`
	Name     string `json:"name"`
	CreateDt string `json:"create_dt"`
}

func mapInfoBaseToJSON(m *mmap.InfoBase) ([]byte, error) {
	r := mapInfoJSONResponse{
		MapID:    m.MapID,
		Name:     m.Name,
		CreateDt: m.CreateDt.Dt.Format(time.RFC3339),
	}
	return json.Marshal(r)
}

type mapInfoResponse struct {
	Map               *mmap.InfoBase
	TopScoringPlayers []*server.TopScorerBase
	TopActivePlayers  []*leaderboard.ActivePlayerBase
	TopActiveServers  []*leaderboard.ActiveServerBase
	RecentGames       []game.RecentGameBase
}

// MapInfoHandler godoc
// @Summary Map information
// @Accept  json
// @Produce  json
// @Param id path int true "map_id"
// @Success 200 {object} mapInfoJSONResponse
// @Router /map/{id} [get]
func (ae *AppEnv) MapInfoHandler(w http.ResponseWriter, r *http.Request) {
	acceptHeader := r.Header.Get("Accept")

	mapID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("Invalid or missing map ID value: %s", err)
		ae.NotFoundHandler(w, r)
		return
	}

	info, err := mmap.InfoData(ae.db, mapID)
	if err != nil {
		log.Printf("Invalid or missing map ID value: %s", err)
		ae.NotFoundHandler(w, r)
		return
	}

	if acceptHeader == "application/json" {
		// JSON response
		bytes, err := mapInfoBaseToJSON(info)
		if err != nil {
			log.Printf("Error: %s", err)
			ae.FiveHundredHandler(w, r)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	} else {
		// HTML response
		topScorers, _ := mmap.TopScorerData(ae.db, mapID)
		topActive, _ := mmap.TopActivePlayersData(ae.db, mapID)
		topServers, _ := mmap.TopActiveServersData(ae.db, mapID)

		recentGamesCutoff := time.Now().UTC().AddDate(0, 0, -1*viper.GetInt("RecentGamesDays"))
		recentGames, _ := game.RecentGamesData(ae.db, game.EmptyServerID, mapID, game.EmptyPlayerID,
			game.EmptyGameTypeCd, &recentGamesCutoff, game.EmptyStartGameID, game.EmptyEndGameID, 20, 
			game.EmptyMatchID)

		response := &mapInfoResponse{
			Map:               info,
			TopScoringPlayers: topScorers,
			TopActivePlayers:  topActive,
			TopActiveServers:  topServers,
			RecentGames:       recentGames,
		}
		err = ae.templates["mapinfo.page.html"].Execute(w, response)
		if err != nil {
			log.Printf("Error: %s", err)
			ae.FiveHundredHandler(w, r)
			return
		}
	}
}
