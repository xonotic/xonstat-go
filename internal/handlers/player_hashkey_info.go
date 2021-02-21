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

// PlayerHashkeyInfoHandler is the web handler for retrieving player information
// geared towards in-game display (on the profile page).
func (ae *AppEnv) PlayerHashkeyInfoHandler(w http.ResponseWriter, r *http.Request) {
	// Pull the D0 verification information out, if it is present.
	hashkey := r.Header.Get("X-D0-Blind-Id-IDFP")
	if hashkey == "" {
		log.Printf("Invalid or missing D0 signature")
		ae.NotFoundHandler(w, r)
		return
	}

	// hashkey -> player (and we should only find one player by this hashkey value)
	players, err := ae.db.RPlayersByHashkeyMulti([]string{hashkey})
	if err != nil || len(players) != 1 {
		log.Printf("Invalid or missing player ID value: %s", err)
		ae.NotFoundHandler(w, r)
		return
	}

	playerID := players[hashkey].PlayerID

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

	// Index by game type
	gameTypesByGamesPlayed := make([]string, len(summaries))
	summariesByGTC := make(map[string]*player.GameTypeSummaryBase, len(summaries))
	for i, elem := range summaries {
		summariesByGTC[elem.GameTypeCd] = elem
		gameTypesByGamesPlayed[i] = elem.GameTypeCd
	}

	overallStats, err := player.OverallStatsData(ae.db, playerID)
	if err != nil {
		log.Printf("Error: %s", err)
		ae.FiveHundredHandler(w, r)
		return
	}

	// Index by game type
	overallStatsByGTC := make(map[string]*player.OverallStatsBase, len(overallStats))
	for _, elem := range overallStats {
		overallStatsByGTC[elem.GameTypeCd] = elem
	}

	t, _ := models.NewMultiDt(time.Now().UTC())

	// This response type is a flat textual response. Rather than putting this in a template
	// and worrying about collapsing whitespace with the logic involved, we'll build it up in a buffer
	// directly (similar to PlayerEloInfoHandler).
	var buf bytes.Buffer
	buf.WriteString("V 1\n")
	buf.WriteString("R XonStat/1.0\n")
	buf.WriteString(fmt.Sprintf("T %d\n", t.Epoch))
	buf.WriteString(fmt.Sprintf("S /player/%d\n", playerID))
	buf.WriteString(fmt.Sprintf("P %s\n", hashkey))
	buf.WriteString(fmt.Sprintf("n %s\n", info.Nick.Nick))
	buf.WriteString(fmt.Sprintf("i %d\n", playerID))
	buf.WriteString(fmt.Sprintf("e joined_dt %s\n", info.CreateDt.Dt.Format("2006-01-02")))

	if overall, ok := overallStatsByGTC["overall"]; ok {
		buf.WriteString(fmt.Sprintf("e last_seen_dt %s\n", overall.LastPlayed.Dt.Format("2006-01-02")))
	}

	if info.ActiveInd {
		buf.WriteString(fmt.Sprintf("e active-ind 1\n"))
	} else {
		buf.WriteString(fmt.Sprintf("e active-ind 0\n"))
	}

	buf.WriteString(fmt.Sprintf("e location\n"))

	if len(summaries) > 0 {
		buf.WriteString(fmt.Sprintf("e matches %d\n", summaries[0].Games))
	}

	if overall, ok := overallStatsByGTC["overall"]; ok {
		buf.WriteString(fmt.Sprintf("e total-deaths %d\n", overall.Deaths))
		buf.WriteString(fmt.Sprintf("e total-fckills %d\n", overall.CarrierFrags))
		buf.WriteString(fmt.Sprintf("e alivetime %d\n", int(overall.TimePlayed.Duration.Seconds())))
		buf.WriteString(fmt.Sprintf("e total-kills %d\n", overall.Kills))
	}

	if summary, ok := summariesByGTC["overall"]; ok {
		buf.WriteString(fmt.Sprintf("e wins %d\n", summary.Wins))
	}

	for _, gameTypeCd := range gameTypesByGamesPlayed {
		if gameTypeCd == "overall" {
			continue
		}

		buf.WriteString(fmt.Sprintf("G %s\n", gameTypeCd))
		// Omitting the "e elo" entry for the game type here.
		// Omitting the "e percentile" for the game type here.

		if summary, ok := summariesByGTC[gameTypeCd]; ok {
			buf.WriteString(fmt.Sprintf("e matches %d\n", summary.Games))
			buf.WriteString(fmt.Sprintf("e wins %d\n", summary.Wins))
		}

		// Omitting the "e rank" for the game type here.

		if overall, ok := overallStatsByGTC[gameTypeCd]; ok {
			buf.WriteString(fmt.Sprintf("e total-deaths %d\n", overall.Deaths))
			buf.WriteString(fmt.Sprintf("e alivetime %.2f\n", overall.TimePlayed.Duration.Seconds()))
			buf.WriteString(fmt.Sprintf("e total-kills %d\n", overall.Kills))
		}

		// Omitting the "e favorite-map" for the game type here.
	}

	w.Write(buf.Bytes())
}
