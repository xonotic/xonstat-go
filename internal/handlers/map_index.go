package handlers

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"gitlab.com/xonotic/xonstat/pkg/mmap"
	"gitlab.com/xonotic/xonstat/pkg/models"
	"strconv"
)

// mapIndexJSONResponse is the JSON response type for the map index.
type mapIndexJSONResponse struct {
	Start    int           `json:"start"`
	Next     int           `json:"next"`
	ShowMore bool          `json:"show_more"`
	HTML     template.HTML `json:"HTML"`
}

// mapIndexResponse is the regular response type for the map index.
type mapIndexResponse struct {
	Start        int
	Next         int
	NameFragment string
	ShowMore     bool
	Maps         []*mmap.InfoBase
}

// MapIndexHandler is the web handler for showing the server index and map search results.
func (ae *AppEnv) MapIndexHandler(w http.ResponseWriter, r *http.Request) {
	acceptHeader := r.Header.Get("Accept")

	limit := 20

	params := r.URL.Query()
	var start int

	start, err := strconv.Atoi(params.Get("start"))
	if err != nil {
		start = models.BlankStart
	}

	nameFragment := params.Get("name")

	maps, err := mmap.IndexData(ae.db, start, limit, nameFragment)
	if err != nil {
		log.Printf("Error: %s", err)
		ae.FiveHundredHandler(w, r)
		return
	}

	var next int
	if len(maps) < 1 {
		next = -1
	} else {
		next = maps[len(maps)-1].MapID
	}

	var showMore bool
	if len(maps) >= limit {
		showMore = true
	}

	response := mapIndexResponse{
		Start:        start,
		Next:         next,
		NameFragment: nameFragment,
		ShowMore:     showMore,
		Maps:         maps,
	}

	if acceptHeader == "application/json" {
		// JSON response
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		var fragment bytes.Buffer
		err := ae.templates["mapindex.fragment.page.html"].Execute(&fragment, response)
		if err != nil {
			log.Printf("Error: %s", err)
			ae.FiveHundredHandler(w, r)
			return
		}

		jsonResponse := mapIndexJSONResponse{
			Start:    response.Start,
			Next:     response.Next,
			ShowMore: response.ShowMore,
			HTML:     template.HTML(fragment.String()),
		}

		bytes, _ := json.Marshal(jsonResponse)
		w.Write(bytes)
	} else {
		// HTML response
		err := ae.templates["mapindex.page.html"].Execute(w, response)
		if err != nil {
			log.Printf("Error: %s", err)
			ae.FiveHundredHandler(w, r)
			return
		}
	}
}
