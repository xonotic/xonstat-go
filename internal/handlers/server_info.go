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

type serverInfoJSONResponse struct {
	ServerID  int    `json:"server_id"`
	Name      string `json:"name"`
	IPAddr    string `json:"ip_addr"`
	Port      int    `json:"port"`
	Revision  string `json:"revision"`
	ActiveInd bool   `json:"active_ind"`
	CreateDt  string `json:"create_dt"`
}

func serverInfoBaseToJSON(ib *server.InfoBase) ([]byte, error) {
	r := serverInfoJSONResponse{
		ServerID:  ib.ServerID,
		Name:      ib.Name,
		IPAddr:    ib.IPAddr,
		Port:      ib.Port,
		Revision:  ib.Revision,
		ActiveInd: ib.ActiveInd,
		CreateDt:  ib.CreateDt.UTCStr,
	}

	return json.Marshal(r)
}

type serverInfoResponse struct {
	Server            server.InfoBase
	TopScoringPlayers []*server.TopScorerBase
	TopActivePlayers  []*leaderboard.ActivePlayerBase
	TopMapsPlayed     []*models.ActiveMap
	RecentGames       []game.RecentGameBase
}

// ServerInfoHandler godoc
// @Summary Server information
// @Accept  json
// @Produce  json
// @Param id path int true "server_id"
// @Success 200 {object} serverInfoJSONResponse
// @Router /server/{id} [get]
func (ae *AppEnv) ServerInfoHandler(w http.ResponseWriter, r *http.Request) {
	acceptHeader := r.Header.Get("Accept")

	serverID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("Invalid or missing server ID value: %s", err)
		ae.NotFoundHandler(w, r)
		return
	}

	s, err := server.InfoData(ae.db, serverID)
	if err != nil {
		log.Printf("Error: %s", err)
		ae.NotFoundHandler(w, r)
		return
	}

	if acceptHeader == "application/json" {
		// JSON response
		s, err := serverInfoBaseToJSON(s)
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
		topScoringPlayers, _ := server.TopScorerData(ae.db, serverID)
		topActivePlayers, _ := server.TopActivePlayersData(ae.db, serverID)
		topMapsPlayed, _ := server.TopMapsData(ae.db, serverID)

		recentGamesCutoff := time.Now().UTC().AddDate(0, 0, -1*viper.GetInt("RecentGamesDays"))
		recentGames, _ := game.RecentGamesData(ae.db, serverID, game.EmptyMapID, game.EmptyPlayerID,
			game.EmptyGameTypeCd, &recentGamesCutoff, game.EmptyStartGameID, game.EmptyEndGameID, 20)

		response := serverInfoResponse{
			Server:            *s,
			TopScoringPlayers: topScoringPlayers,
			TopActivePlayers:  topActivePlayers,
			TopMapsPlayed:     topMapsPlayed,
			RecentGames:       recentGames,
		}

		err = ae.templates["serverinfo.page.html"].Execute(w, response)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}
	}
}
