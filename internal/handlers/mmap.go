package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/spf13/viper"
	"github.com/go-chi/chi"
	"gitlab.com/xonotic/xonstat/pkg/leaderboard"
	"gitlab.com/xonotic/xonstat/pkg/mmap"
	"gitlab.com/xonotic/xonstat/pkg/models"
	"gitlab.com/xonotic/xonstat/pkg/game"
	"gitlab.com/xonotic/xonstat/pkg/server"
	"strconv"
)

// MapInfoResponse is the view-specific information about a map related information.
type MapInfoResponse struct {
	Map               *mmap.InfoBase
	TopScoringPlayers []*server.TopScorerBase
	TopActivePlayers  []*leaderboard.ActivePlayerBase
	TopActiveServers  []*leaderboard.ActiveServerBase
	RecentGames       []game.RecentGameBase
}

// MapInfoHandler is the web handler for retrieving map information
func (ae *AppEnv) MapInfoHandler(w http.ResponseWriter, r *http.Request) {
	acceptHeader := r.Header.Get("Accept")

	mapID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("Invalid or missing map ID value: %s", err)
		ae.NotFoundHandler(w, r)
		return
	}

	info, err := mmap.InfoData(ae.db, mapID)
	if err != nil {
		log.Printf("Invalid or missing map ID value: %s", err)
		ae.NotFoundHandler(w, r)
		return
	}

	topScorers, _ := mmap.TopScorerData(ae.db, mapID)
	topActive, _ := mmap.TopActivePlayersData(ae.db, mapID)
	topServers, _ := mmap.TopActiveServersData(ae.db, mapID)

	recentGamesCutoff := time.Now().UTC().AddDate(0, 0, -1*viper.GetInt("RecentGamesDays"))
	recentGames, _ := game.RecentGamesData(ae.db, game.EmptyServerID, mapID, game.EmptyPlayerID,
		game.EmptyGameTypeCd, &recentGamesCutoff, game.EmptyStartGameID, game.EmptyEndGameID, 20)

	response := &MapInfoResponse{
		Map:               info,
		TopScoringPlayers: topScorers,
		TopActivePlayers:  topActive,
		TopActiveServers:  topServers,
		RecentGames: recentGames,
	}

	if acceptHeader == "application/json" {
		bytes, _ := json.Marshal(response)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	} else {
		err = ae.templates["mapinfo.page.html"].Execute(w, response)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}
	}
}

// MapIndexJSONResponse is the JSON response type for the map index.
type MapIndexJSONResponse struct {
	Start    int           `json:"start"`
	Next     int           `json:"next"`
	ShowMore bool          `json:"show_more"`
	HTML     template.HTML `json:"HTML"`
}

// MapIndexResponse is the regular response type for the map index.
type MapIndexResponse struct {
	Start        int
	Next         int
	NameFragment string
	ShowMore     bool
	Maps      []*mmap.InfoBase
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
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
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

	response := MapIndexResponse{
		Start:        start,
		Next:         next,
		NameFragment: nameFragment,
		ShowMore:     showMore,
		Maps:      maps,
	}

	if acceptHeader == "application/json" {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		var fragment bytes.Buffer
		err := ae.templates["mapindex.fragment.page.html"].Execute(&fragment, response)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}

		jsonResponse := MapIndexJSONResponse{
			Start:    response.Start,
			Next:     response.Next,
			ShowMore: response.ShowMore,
			HTML:     template.HTML(fragment.String()),
		}

		bytes, _ := json.Marshal(jsonResponse)
		w.Write(bytes)
	} else {
		err := ae.templates["mapindex.page.html"].Execute(w, response)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}
	}
}
