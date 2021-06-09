package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"gitlab.com/xonotic/xonstat/pkg/skill"
	"gitlab.com/xonotic/xonstat/pkg/submission"
)

// playerSkillJSONResponse is the JSON-specific response type
type playerSkillJSONResponse struct {
	PlayerID   int     `json:"player_id"`
	GameTypeCd string  `json:"game_type_cd"`
	Mu         float64 `json:"mu"`
	Sigma      float64 `json:"sigma"`
	ActiveInd  bool    `json:"active_ind"`
	CreateDt   string  `json:"create_dt"`
	UpdateDt   string  `json:"update_dt"`
}

// PlayerSkillHandler godoc
// @Summary Player skill information
// @Accept  json
// @Produce  json
// @Param id path int true "player_id"
// @Param game_type_cd query string false "game_type_cd"
// @Success 200 {object} []playerSkillJSONResponse
// @Router /player/{id}/skill [get]
func (ae *AppEnv) PlayerSkillHandler(w http.ResponseWriter, r *http.Request) {
	playerID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || playerID <= 2 {
		log.Printf("Invalid or missing player ID value")
		ae.NotFoundHandler(w, r)
		return
	}

	params := r.URL.Query()

	gameTypeCd := params.Get("game_type_cd")
	if !submission.IsSupportedGameType(gameTypeCd) {
		// It is not a supported game type. Use the default for the DB query and don't include it
		// in the resulting query string.
		gameTypeCd = ""
	}

	infos, err := skill.InfoData(ae.db, playerID, gameTypeCd)
	if err != nil {
		log.Printf("Unable to retrieve the skills values: %s", err)
		ae.NotFoundHandler(w, r)
		return
	}

	response := make([]playerSkillJSONResponse, len(infos))
	for i, v := range infos {
		response[i] = playerSkillJSONResponse{
			PlayerID:   v.PlayerID,
			GameTypeCd: v.GameTypeCd,
			Mu:         v.Mu,
			Sigma:      v.Sigma,
			ActiveInd:  v.ActiveInd,
			CreateDt:   v.CreateDt.Format(time.RFC3339),
			UpdateDt:   v.UpdateDt.Format(time.RFC3339),
		}
	}

	bytes, _ := json.Marshal(response)

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}
