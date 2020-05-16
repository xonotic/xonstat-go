package handlers

import (
	"bufio"
	"fmt"
	"log"
	"net/http"

	"gitlab.com/antibody/xonstat/pkg/submission"
)

// SubmissionHandler is the main stats submission handler. It takes stats submissions
// from the servers as a signed POST request, parses them, and submits them to the
// database.
func (ae *AppEnv) SubmissionHandler(w http.ResponseWriter, r *http.Request) {
	body := bufio.NewReader(r.Body)
	rawSubmission, err := submission.NewRawSubmission(body)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("422 %s", http.StatusText(422)), 422)
		return
	}

	sub, err := submission.NewSubmission(rawSubmission)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("422 %s", http.StatusText(422)), 422)
		return
	}

	err = submission.Submit(sub, ae.db)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("200 OK\n"))
}
