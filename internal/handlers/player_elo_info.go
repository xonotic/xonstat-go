package handlers

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	"gitlab.com/xonotic/xonstat/pkg/models"
	"gitlab.com/xonotic/xonstat/pkg/player"
)

// PlayerEloInfoHandler is the web handler for retrieving player Elo information
func (ae *AppEnv) PlayerEloInfoHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	hashkey := params.Get("hashkey")

	r.Header.Set("Accept", "text/plain")

	eloInfo, err := player.EloInfoData(ae.db, hashkey)
	if err != nil {
		log.Printf("Error retrieving Elo information for hashkey %s: %s", hashkey, err)
		ae.NotFoundHandler(w, r)
		return
	}

	if len(eloInfo) == 0 {
		log.Printf("No Elos found for hashkey %s", hashkey)
		ae.NotFoundHandler(w, r)
		return
	}

	playerID := eloInfo[0].PlayerID

	t, _ := models.NewMultiDt(time.Now().UTC())
	player, err := player.InfoData(ae.db, playerID)
	if err != nil {
		log.Printf("No player with ID %d found", playerID)
		ae.NotFoundHandler(w, r)
		return
	}

	// The Elo info response type is a flat textual response. Rather than putting this in a template
	// and worrying about collapsing whitespace with the logic involved, we'll build it up in a buffer
	// directly.
	var buf bytes.Buffer
	buf.WriteString("V 1\n")
	buf.WriteString("R XonStat/1.0\n")
	buf.WriteString(fmt.Sprintf("T %d\n", t.Epoch))
	buf.WriteString(fmt.Sprintf("S /player/%d\n", eloInfo[0].PlayerID))
	buf.WriteString(fmt.Sprintf("P %s\n", hashkey))
	buf.WriteString(fmt.Sprintf("n %s\n", player.Nick.Nick))
	buf.WriteString(fmt.Sprintf("i %d\n", player.PlayerID))

	if player.ActiveInd {
		buf.WriteString(fmt.Sprintf("e active-ind 1\n"))
	} else {
		buf.WriteString(fmt.Sprintf("e active-ind 0\n"))
	}

	buf.WriteString("e location\n")

	for _, elo := range eloInfo {
		buf.WriteString(fmt.Sprintf("G %s\n", elo.GameTypeCd))
		buf.WriteString(fmt.Sprintf("e %.2f\n", elo.Elo))
	}

	w.Write(buf.Bytes())
}
