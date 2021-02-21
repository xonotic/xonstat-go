package handlers

import (
	"log"
	"math/rand"
	"net/http"
)

// NotFoundHandler handles route patterns that aren't registered.
func (ae *AppEnv) NotFoundHandler(w http.ResponseWriter, r *http.Request) {

	sayingIndex := rand.Intn(8)

	acceptHeader := r.Header.Get("Accept")
	if acceptHeader == "application/json" {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
	} else {
		// HTML response
		type Data struct {
			SayingIndex int
		}

		data := Data{
			SayingIndex: sayingIndex,
		}

		err := ae.templates["404.page.html"].Execute(w, data)
		if err != nil {
			log.Printf("Error: %s", err)
			ae.FiveHundredHandler(w, r)
			return
		}
	}
}
