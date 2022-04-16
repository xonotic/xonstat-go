package handlers

import (
	"bufio"
	"fmt"
	"log"
	"net/http"

	"gitlab.com/xonotic/xonstat/pkg/submission"
)

type balancePlayer struct {
	Hashkey string
	PlayerID int
	Nick string
}

type balanceResponse struct {
	Version int
	Release string
	Time int64
	Players []balancePlayer
}

// BalanceHandler takes player info from servers and returns back a best-guess 
// ordering of those players according to their skill.
func (ae *AppEnv) BalanceHandler(w http.ResponseWriter, r *http.Request) {
	bodyReader := bufio.NewReader(r.Body)
	rawSubmission, err := submission.NewRawSubmission(bodyReader)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("422 %s", http.StatusText(422)), 422)
		return
	}

	// It doesn't make sense to balance games with only one or two players.
	if rawSubmission.NumHumansPlayed < 3 {
		log.Printf("Error: not enough players to balance")
		http.Error(w, fmt.Sprintf("422 %s", http.StatusText(422)), 422)
		return
	}

	sub, err := submission.NewSubmission(rawSubmission)

	response := balanceResponse {
		Version: 1,
		Release: "XonStat/1.0",
		Time: sub.CreateDt.Unix(),
		Players: make([]balancePlayer, 0),
	}

	hashkeys := make([]string, len(sub.PlayersByHashkey))
	i := 0
	for hashkey, _ := range sub.PlayersByHashkey {
		hashkeys[i] = hashkey
		i++
	}

	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("422 %s", http.StatusText(422)), 422)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "text/plain")

	err = ae.templates["balance.page.html"].Execute(w, response)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}
}
