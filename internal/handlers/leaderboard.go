package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/alehano/reverse"
	"github.com/spf13/viper"
	"gitlab.com/xonotic/xonstat/pkg/game"
	"gitlab.com/xonotic/xonstat/pkg/leaderboard"
	"golang.org/x/text/message"
)

// TopServersHandler retrieves information about the top active servers by player aggregate playtime
func (ae *AppEnv) TopServersHandler(w http.ResponseWriter, r *http.Request) {
	acceptHeader := r.Header.Get("Accept")

	startStr := r.URL.Query().Get("start")
	start, err := strconv.Atoi(startStr)
	if err != nil {
		start = 1
	}

	if acceptHeader == "application/json" {
		// JSON response
		bytes, err := leaderboard.ActiveServersJSON(10, start, ae.db)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	} else {
		// HTML response
		activeServers, err := leaderboard.ActiveServersData(20, start, ae.db)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}

		next := 1
		if len(activeServers) > 0 {
			next = activeServers[len(activeServers)-1].SortOrder + 1
		}

		// The structure passed to the template.
		type Data struct {
			ActiveServers []leaderboard.ActiveServerBase
			Start         int
			Next          int
			ShowMoreLink  bool
		}

		data := Data{
			ActiveServers: activeServers,
			Start:         start,
			Next:          next,
			ShowMoreLink:  len(activeServers) == 20,
		}

		err = ae.templates["activeservers.page.html"].Execute(w, data)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}
	}
}

// TopMapsHandler retrieves information about the top active maps
func (ae *AppEnv) TopMapsHandler(w http.ResponseWriter, r *http.Request) {
	acceptHeader := r.Header.Get("Accept")

	startStr := r.URL.Query().Get("start")
	start, err := strconv.Atoi(startStr)
	if err != nil {
		start = 1
	}

	if acceptHeader == "application/json" {
		// JSON response
		bytes, err := leaderboard.ActiveMapsJSON(10, start, ae.db)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	} else {
		// HTML response
		activeMaps, err := leaderboard.ActiveMapsData(20, start, ae.db)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}

		// The structure passed to the template.
		type Data struct {
			ActiveMaps   []leaderboard.ActiveMapBase
			Start        int
			Next         int
			ShowMoreLink bool
		}

		next := 1
		if len(activeMaps) > 0 {
			next = activeMaps[len(activeMaps)-1].SortOrder + 1
		}

		data := Data{
			ActiveMaps:   activeMaps,
			Start:        start,
			Next:         next,
			ShowMoreLink: len(activeMaps) == 20,
		}

		err = ae.templates["activemaps.page.html"].Execute(w, data)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}
	}
}

// Assemble the stats line at the top of the leaderboard. Can accept either the "all" or "day"
// scoped version of SummaryStat array.
func makeStatLine(prefix string, summaryStats *leaderboard.SummaryBase, suffix string) template.HTML {
	// TODO: Add links to the game types when that handler/template is ready.
	// Derive the URL if possible instead of hard coding it.

	// This is used to get the commas in the output's numbers.
	p := message.NewPrinter(message.MatchLanguage("en"))

	if len(summaryStats.Games) == 0 {
		return ""
	}

	// The total number of games.
	var totalGameCount int
	for _, v := range summaryStats.Games {
		totalGameCount += v.GameCount
	}

	// We can't show the counts for *all* game types, so we'll group all the ones past the top five
	// into an "other" category.
	var otherGameCount int
	if len(summaryStats.Games) > 5 {
		for _, v := range summaryStats.Games[5:] {
			otherGameCount += v.GameCount
		}
	}

	var buf bytes.Buffer
	buf.WriteString(p.Sprintf("%d players and %d games (", summaryStats.Players, totalGameCount))

	// If for some reason we don't have 5 "top" game types...
	topN := 5
	if len(summaryStats.Games) < topN {
		topN = len(summaryStats.Games)
	}

	for i, v := range summaryStats.Games[:topN] {
		buf.WriteString(p.Sprintf("%d <a href=\"%s?game_type_cd=%s\">%s</a>", v.GameCount, reverse.Rev("games"), v.GameTypeCd, v.GameTypeCd))

		if i < topN-1 {
			buf.WriteString("; ")
		}
	}

	if otherGameCount > 0 {
		buf.WriteString(p.Sprintf("; %d other", otherGameCount))
	}
	buf.WriteString(")")

	return template.HTML(buf.String())
}

type leaderboardResponse struct {
	StatLine      template.HTML
	DayStatLine   template.HTML
	ActivePlayers []*leaderboard.ActivePlayerBase
	ActiveServers []leaderboard.ActiveServerBase
	ActiveMaps    []leaderboard.ActiveMapBase
	RecentGames   []game.RecentGameBase
}

// LeaderboardHandler is the main page of the site
func (ae *AppEnv) LeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	// The summary stat line for all activity tracked thus far.
	allSummaryStats, err := leaderboard.SummaryData("all", ae.db)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}

	// The summary stat line typically for the past day's worth of activity.
	daySummaryStats, err := leaderboard.SummaryData("day", ae.db)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}

	// Active players by playing (alive) time.
	activePlayers, _ := leaderboard.ActivePlayersData(10, 1, ae.db)

	// Active servers by total accumulated player time on the server
	activeServers, _ := leaderboard.ActiveServersData(10, 1, ae.db)

	// Active maps by number of times played
	activeMaps, _ := leaderboard.ActiveMapsData(10, 1, ae.db)

	// Recent games
	recentGamesDays := viper.GetInt("RecentGamesDays")
	now := time.Now().UTC()
	cutoff := now.AddDate(0, 0, -1*recentGamesDays)

	recentGames, _ := game.RecentGamesData(ae.db, game.EmptyServerID, game.EmptyMapID,
		game.EmptyPlayerID, game.EmptyGameTypeCd, &cutoff, game.EmptyStartGameID,
		game.EmptyEndGameID, 20)

	response := leaderboardResponse{
		StatLine:      makeStatLine("", allSummaryStats, ""),
		DayStatLine:   makeStatLine("", daySummaryStats, ""),
		ActivePlayers: activePlayers,
		ActiveServers: activeServers,
		ActiveMaps:    activeMaps,
		RecentGames:   recentGames,
	}

	err = ae.templates["leaderboard.page.html"].Execute(w, response)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}
}
