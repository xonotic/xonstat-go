package handlers

import (
	"bufio"
	"fmt"
	"net/http"

	"gitlab.com/antibody/xonstat/pkg/submission"
)

// SubmissionHandler is the main stats submission handler. It takes stats submissions
// from the servers as a signed POST request, parses them, and submits them to the
// database.
func SubmissionHandler(w http.ResponseWriter, r *http.Request) {
	body := bufio.NewReader(r.Body)
	rawSubmission, err := submission.NewRawSubmission(body)
	if err != nil {
		http.Error(w, fmt.Sprintf("422 %s", http.StatusText(422)), 422)
		return
	}

	_, err = submission.NewSubmission(rawSubmission)
	if err != nil {
		http.Error(w, fmt.Sprintf("422 %s", http.StatusText(422)), 422)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("200 OK\n"))
}
