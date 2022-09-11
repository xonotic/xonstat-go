package handlers

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"gitlab.com/xonotic/xonstat/pkg/skill"
	wenglin "gitlab.com/xonotic/xonstat/pkg/skill"
	"gitlab.com/xonotic/xonstat/pkg/submission"
)

type balanceResponse struct {
	Version int
	Release string
	Time    int64
	Players []*skill.BalancePlayer
}

// preprocess handles assembling the POST body into a type that we can work with
func preprocess(w http.ResponseWriter, r *http.Request) (*submission.Submission, error) {
	bodyReader := bufio.NewReader(r.Body)

	rawSubmission, err := submission.NewRawSubmission(bodyReader)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("422 %s", http.StatusText(422)), 422)
		return nil, err
	}

	// It doesn't make sense to balance games with only one or two players.
	if rawSubmission.NumHumansPlayed < 3 {
		log.Printf("Error: not enough players to balance")
		http.Error(w, fmt.Sprintf("422 %s", http.StatusText(422)), 422)
		return nil, fmt.Errorf("not enough players to balance")
	}

	sub, err := submission.NewSubmission(rawSubmission)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("422 %s", http.StatusText(422)), 422)
		return nil, err
	}

	return sub, nil
}

// BalanceHandler godoc
// @Summary Best guess ordering of players according to skill and score data.
// @Accept  text/plain
// @Produce  text/plain
// @Success 200 {object} balanceResponse
// @Router /balance [post]
func (ae *AppEnv) BalanceHandler(w http.ResponseWriter, r *http.Request) {
	sub, err := preprocess(w, r)
	if err != nil {
		return
	}

	params := r.URL.Query()

	scoreFactorInt, err := strconv.Atoi(params.Get("scorefactor"))
	if err != nil || scoreFactorInt < 0 || scoreFactorInt > 100{
		scoreFactorInt = 25
	}

	scoreFactor := float64(scoreFactorInt)/100.0

	bp := skill.BalanceParams{
		DefaultMu: wenglin.DefaultParams.DefaultMu,
		DefaultSigma: wenglin.DefaultParams.DefaultSigma,
		DefaultBeta: wenglin.DefaultParams.DefaultBeta,
		ScoreFactor: scoreFactor,
	}

	players, err := skill.Balance(bp, ae.db, sub)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("422 %s", http.StatusText(422)), 422)
		return
	}

	response := balanceResponse{
		Version: 1,
		Release: "XonStat/1.0",
		Time:    sub.CreateDt.Unix(),
		Players: players,
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
