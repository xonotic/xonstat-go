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

// maxTeams is the maximum number of teams the balance handler will process.
// Games with more teams are rejected to bound processing time (the swap
// minimization algorithm is O(T!) in the number of teams).
const maxTeams = 4

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

	// There's nothing to balance without teams.
	if !submission.IsTeamGameType(rawSubmission.GameMeta["G"]) {
		log.Printf("Error: nothing to balance for non-team game type %s", rawSubmission.GameMeta["G"])
		http.Error(w, fmt.Sprintf("422 %s", http.StatusText(422)), 422)
		return nil, fmt.Errorf("nothing to balance for non-team game type %s", rawSubmission.GameMeta["G"])
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

	// Each player receives up to configurable fraction of their own skill as
	// a bonus for in-match performance. This we call the scorefactor.
	scoreFactorInt, err := strconv.Atoi(params.Get("scorefactor"))
	if err != nil || scoreFactorInt < 0 || scoreFactorInt > 100{
		scoreFactorInt = 25
	}

	scoreFactor := float64(scoreFactorInt)/100.0

	// cardinality controls the maximum allowed difference in the number of
	// players between any two teams. Default 1, clamped to [0, 4].
	cardinality, err := strconv.Atoi(params.Get("cardinality"))
	if err != nil || cardinality < 0 {
		cardinality = 1
	}

	if cardinality > 4 {
		cardinality = 4
	}

	// stability controls the threshold for swapping players between teams.
	// A value of 5 means the new partition must improve balance by more than
	// 5% of total skill to be applied. Default 5, clamped to [0, 100].
	stabilityInt, err := strconv.Atoi(params.Get("stability"))
	if err != nil || stabilityInt < 0 {
		stabilityInt = 5
	}

	if stabilityInt > 100 {
		stabilityInt = 100
	}

	stabilityFloat := float64(stabilityInt) / 100.0

	// Derive the number of teams from the submission itself.
	// Fall back to looking at the game stat entries to derive teams.
	numTeams := len(sub.TeamGameStats)
	if numTeams == 0 {
		teamSet := make(map[int]struct{})
		for _, pgs := range sub.PlayerGameStats {
			if pgs.Team.Valid {
				teamSet[int(pgs.Team.Int32)] = struct{}{}
			}
		}
		numTeams = len(teamSet)
	}

	if numTeams > maxTeams {
		log.Printf("Error: too many teams to balance (%d > %d)", numTeams, maxTeams)
		http.Error(w, fmt.Sprintf("422 %s", http.StatusText(422)), 422)
		return
	}

	bp := skill.BalanceParams{
		DefaultMu:         wenglin.DefaultParams.DefaultMu,
		DefaultSigma:      wenglin.DefaultParams.DefaultSigma,
		DefaultBeta:       wenglin.DefaultParams.DefaultBeta,
		ScoreFactor:       scoreFactor,
		StabilityThreshold: stabilityFloat,
	}

	players, err := skill.Balance(bp, ae.db, sub, cardinality, numTeams)
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

	t, ok := ae.textTemplates["balance.page.txt"]
	if !ok {
		log.Printf("Error: balance template is not loaded")
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}

	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	err = t.Execute(w, response)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}
}
