package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"gitlab.com/xonotic/xonstat/pkg/server"
)

// ServerInfoHandler retrieves information about individual servers.
func (ae *AppEnv) ServerInfoHandler(w http.ResponseWriter, r *http.Request) {
	acceptHeader := r.Header.Get("Accept")

	ID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("Invalid or missing server ID value: %s", err)
		http.Error(w, fmt.Sprintf("404 %s", http.StatusText(404)), 404)
		return
	}

	if acceptHeader == "application/json" {
		// JSON response
		s, err := server.ServerInfoJSON(ae.db, ID)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("404 %s", http.StatusText(404)), 404)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(s)
	} else {
		// HTML response
		s, err := server.ServerInfoData(ae.db, ID)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("404 %s", http.StatusText(404)), 404)
			return
		}

		// The structure passed to the template.
		type Data struct {
			Server server.ServerInfoBase
		}

		data := Data{
			Server:  *s,
		}

		err = ae.templates["serverinfo.page.html"].Execute(w, data)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}
	}
}
