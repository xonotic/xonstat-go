package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"html/template"

	"gitlab.com/xonotic/xonstat/pkg/models"
	"gitlab.com/xonotic/xonstat/pkg/server"
)

// serverIndexJSONResponse is the response type for the server index.
type serverIndexJSONResponse struct {
	Start    int           `json:"start"`
	Next     int           `json:"next"`
	ShowMore bool          `json:"show_more"`
	HTML     template.HTML `json:"HTML"`
}

// serverIndexResponse is the response type for the server index.
type serverIndexResponse struct {
	Start        int
	Next         int
	NameFragment string
	ShowMore     bool
	Servers      []*server.InfoBase
}

// ServerIndexHandler is the web handler for showing the server index and player search results.
func (ae *AppEnv) ServerIndexHandler(w http.ResponseWriter, r *http.Request) {
	acceptHeader := r.Header.Get("Accept")

	limit := 20

	params := r.URL.Query()
	var start int

	start, err := strconv.Atoi(params.Get("start"))
	if err != nil {
		start = models.BlankStart
	}

	nameFragment := params.Get("name")

	servers, err := server.IndexData(ae.db, start, limit, nameFragment)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}

	var next int
	if len(servers) < 1 {
		next = -1
	} else {
		next = servers[len(servers)-1].ServerID
	}

	var showMore bool
	if len(servers) >= limit {
		showMore = true
	}

	response := serverIndexResponse{
		Start:        start,
		Next:         next,
		NameFragment: nameFragment,
		ShowMore:     showMore,
		Servers:      servers,
	}

	if acceptHeader == "application/json" {
		// JSON response
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		var fragment bytes.Buffer
		err := ae.templates["serverindex.fragment.page.html"].Execute(&fragment, response)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}

		jsonResponse := serverIndexJSONResponse{
			Start:    response.Start,
			Next:     response.Next,
			ShowMore: response.ShowMore,
			HTML:     template.HTML(fragment.String()),
		}

		bytes, _ := json.Marshal(jsonResponse)
		w.Write(bytes)
	} else {
		// HTML response
		err := ae.templates["serverindex.page.html"].Execute(w, response)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}
	}
}