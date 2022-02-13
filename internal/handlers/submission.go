package handlers

import (
	"bufio"
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/spf13/viper"
	"gitlab.com/xonotic/xonstat/pkg/submission"
)

// SubmissionHandler is the main stats submission handler. It takes stats submissions
// from the servers as a signed POST request, parses them, and submits them to the
// database.
func (ae *AppEnv) SubmissionHandler(w http.ResponseWriter, r *http.Request) {
	logAllRequests := viper.GetBool("RequestsLogAll")

	// Grab the body for logging, either now or after some validation checks.
	body, _ := ioutil.ReadAll(r.Body)

	// Logging all requests is really useful for debugging. Since it is for that purpose,
	// we log as early as possible. Note that this doesn't include stuff like the d0 header.
	if logAllRequests {
		ae.requestLogger.Write(body)
	}

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
	gtc := rawSubmission.GameMeta["G"]
	if rawSubmission.NumHumansPlayed < minimumRequiredPlayers {
		// CTS is allowed to record games with only one player.
		if gtc == "cts" && rawSubmission.NumHumansPlayed == 1 {
			return
		}

		log.Printf("Error: not enough players (want %d, found %d)", minimumRequiredPlayers,
			len(rawSubmission.Humans))
		http.Error(w, fmt.Sprintf("422 %s", http.StatusText(422)), 422)
		return
	}

	sub, err := submission.NewSubmission(rawSubmission)

	// Server IP addresses come from "X-Real-IP" or "X-Forwarded-For" headers.
	// These are determined/evaluated by the middleware.RealIP in chi.
	serverIPAddr := r.RemoteAddr
	if serverIPAddr != "" {
		sub.Server.IPAddr.String = serverIPAddr
		sub.Server.IPAddr.Valid = true
	}

	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("422 %s", http.StatusText(422)), 422)
		return
	}

	// Pull the D0 verification information out, if it is present.
	idfp := r.Header.Get("X-D0-Blind-Id-IDFP")
	if idfp != "" {
		sub.Server.HashKey = sql.NullString{Valid: true, String: idfp}
	}

	if !logAllRequests {
		// If we've gotten here, it's likely that we have a valid submission, so we'll log it.
		bodyLogMsg := fmt.Sprintf("----- BEGIN REQUEST BODY -----\n%s%s----- END REQUEST BODY -----\n\n",
			fmt.Sprintf("IDFP %s\n", idfp), string(body))
		ae.requestLogger.Write([]byte(bodyLogMsg))
	}

	err = submission.Submit(sub, ae.db)
	if err == submission.ErrDuplicateGame {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("422 %s", http.StatusText(422)), 422)
		return
	} else if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}

	w.WriteHeader(http.StatusOK)

	err = ae.templates["submission.page.html"].Execute(w, sub)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}

}
