package handlers

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"gitlab.com/xonotic/xonstat/pkg/models"
	"gitlab.com/xonotic/xonstat/pkg/player"
)

// playerIndexJSONResponse is the response type for the player index.
type playerIndexJSONResponse struct {
	Start    int           `json:"start"`
	Next     int           `json:"next"`
	ShowMore bool          `json:"show_more"`
	HTML     template.HTML `json:"HTML"`
}

// playerIndexResponse is the response type for the player index.
type playerIndexResponse struct {
	Start        int
	Next         int
	NickFragment string
	ShowMore     bool
	Players      []*player.InfoBase
}

// PlayerIndexHandler is the web handler for showing the player index and player search results.
func (ae *AppEnv) PlayerIndexHandler(w http.ResponseWriter, r *http.Request) {
	acceptHeader := r.Header.Get("Accept")

	limit := 20

	params := r.URL.Query()
	var start int

	start, err := strconv.Atoi(params.Get("start"))
	if err != nil {
		start = models.BlankStart
	}

	nickFragment := params.Get("nick")

	players, err := player.IndexData(ae.db, start, limit, nickFragment)
	if err != nil {
		log.Printf("Error: %s", err)
		ae.FiveHundredHandler(w, r)
		return
	}

	var next int
	if len(players) < 1 {
		next = -1
	} else {
		next = players[len(players)-1].PlayerID
	}

	var showMore bool
	if len(players) >= limit {
		showMore = true
	}

	response := playerIndexResponse{
		Start:        start,
		Next:         next,
		NickFragment: nickFragment,
		ShowMore:     showMore,
		Players:      players,
	}

	if acceptHeader == "application/json" {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		var fragment bytes.Buffer
		err := ae.templates["playerindex.fragment.page.html"].Execute(&fragment, response)
		if err != nil {
			log.Printf("Error: %s", err)
			ae.FiveHundredHandler(w, r)
			return
		}

		jsonResponse := playerIndexJSONResponse{
			Start:    response.Start,
			Next:     response.Next,
			ShowMore: response.ShowMore,
			HTML:     template.HTML(fragment.String()),
		}

		bytes, _ := json.Marshal(jsonResponse)
		w.Write(bytes)
	} else {
		err := ae.templates["playerindex.page.html"].Execute(w, response)
		if err != nil {
			log.Printf("Error: %s", err)
			ae.FiveHundredHandler(w, r)
			return
		}
	}
}
