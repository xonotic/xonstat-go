package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
)

// PlayerSkillHashkeyHandler godoc
// @Summary Player skill information by hashkey
// @Accept  json
// @Produce  json
// @Param hashkey query string true "hashkey"
// @Param game_type_cd query string true "hashkey"
// @Success 200 {object} []playerSkillJSONResponse
// @Router /skill [get]
func (ae *AppEnv) PlayerSkillHashkeyHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	hashkey := params.Get("hashkey")

	nextParams := url.Values{}
	nextParams.Add("game_type_cd", params.Get("game_type_cd"))

	// hashkey -> player (and we should only find one player by this hashkey value)
	players, err := ae.db.RPlayersByHashkeyMulti([]string{hashkey})
	if err != nil || len(players) != 1 {
		log.Printf("Invalid player hashkey value: %s", err)
		ae.NotFoundHandler(w, r)
		return
	}

	playerID := players[hashkey].PlayerID
	url := fmt.Sprintf("/player/%d/skill?%s", playerID, nextParams.Encode())
	http.Redirect(w, r, url, http.StatusFound)
}
