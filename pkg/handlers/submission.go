package handlers

import (
	"bufio"
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"gitlab.com/antibody/xonstat/pkg/d0"
	"gitlab.com/antibody/xonstat/pkg/submission"
	"github.com/spf13/viper"
)

// SubmissionHandler is the main stats submission handler. It takes stats submissions
// from the servers as a signed POST request, parses them, and submits them to the
// database.
func (ae *AppEnv) SubmissionHandler(w http.ResponseWriter, r *http.Request) {
	// Grab the body for later logging, if warranted.
	body, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	bodyReader := bufio.NewReader(r.Body)
	rawSubmission, err := submission.NewRawSubmission(bodyReader)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("422 %s", http.StatusText(422)), 422)
		return
	}

	minimumRequiredPlayers := viper.GetInt("MinimumRequiredPlayers")
	if len(rawSubmission.Humans) < minimumRequiredPlayers {
		log.Printf("Error: not enough players (want %d, found %d)", minimumRequiredPlayers, 
			len(rawSubmission.Humans))
		http.Error(w, fmt.Sprintf("422 %s", http.StatusText(422)), 422)
		return
	}

	sub, err := submission.NewSubmission(rawSubmission)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("422 %s", http.StatusText(422)), 422)
		return
	}

	// Pull the D0 verification information out, if it is present.
	value := r.Context().Value(D0VerifyResultKey)
    d0Result, ok := value.(d0.VerifyResult)
	if ok {
		if d0Result.IDFP != "" {
			sub.Server.HashKey = sql.NullString{Valid: true, String: d0Result.IDFP}
		}
	}

	// If we've gotten here, it's likely that we have a valid submission, so we'll log it.
	bodyLogMsg := fmt.Sprintf("\n----- BEGIN REQUEST BODY -----\n%s%s----- END REQUEST BODY -----\n\n",
		fmt.Sprintf("IDFP %s\n", d0Result.IDFP), string(body))
	log.Printf(bodyLogMsg)

	err = submission.Submit(sub, ae.db)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("200 OK\n"))
}
