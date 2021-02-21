package handlers

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
)

// FiveHundredHandler handles app-level errors (HTTP 500s).
func (ae *AppEnv) FiveHundredHandler(w http.ResponseWriter, r *http.Request) {
	sayingIndex := rand.Intn(8)

	acceptHeader := r.Header.Get("Accept")
	if acceptHeader == "application/json" {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		// HTML response
		type Data struct {
			SayingIndex int
		}

		data := Data{
			SayingIndex: sayingIndex,
		}

		err := ae.templates["500.page.html"].Execute(w, data)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}
	}
}
