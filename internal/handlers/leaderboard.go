package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"gitlab.com/xonotic/xonstat/pkg/leaderboard"
	"gitlab.com/xonotic/xonstat/pkg/models"
	"golang.org/x/text/message"
)

// SummaryStatsHandler retrieves information about the summary stats
func (ae *AppEnv) SummaryStatsHandler(w http.ResponseWriter, r *http.Request) {
	scope := r.URL.Query().Get("scope")
	if scope != "all" && scope != "day" {
		scope = "all"
	}

	summaryStats, err := leaderboard.SummaryStatsJSON(scope, ae.db)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(summaryStats)
}

// TopActiveHandler retrieves information about the top active players by playing time
func (ae *AppEnv) TopActiveHandler(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start")
	start, err := strconv.Atoi(startStr)
	if err != nil {
		start = 1
	}

	bytes, err := leaderboard.ActivePlayersJSON(10, start, ae.db)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}

// TopServersHandler retrieves information about the top active servers by player aggregate playtime
func (ae *AppEnv) TopServersHandler(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start")
	start, err := strconv.Atoi(startStr)
	if err != nil {
		start = 1
	}

	bytes, err := leaderboard.ActiveServersJSON(10, start, ae.db)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}

// TopMapsHandler retrieves information about the top active maps
func (ae *AppEnv) TopMapsHandler(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start")
	start, err := strconv.Atoi(startStr)
	if err != nil {
		start = 1
	}

	bytes, err := leaderboard.ActiveMapsJSON(10, start, ae.db)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}

// Assemble the stats line at the top of the leaderboard. Can accept either the "all" or "day"
// scoped version of SummaryStat array.
func makeStatLine(prefix string, summaryStats []*models.SummaryStat, suffix string) template.HTML {
	// TODO: Add links to the game types when that handler/template is ready.
	// Derive the URL if possible instead of hard coding it.

	// This is used to get the commas in the output's numbers.
	p := message.NewPrinter(message.MatchLanguage("en"))

	if len(summaryStats) == 0 {
		return ""
	}

	// The total number of games.
	var totalGameCount int
	for _, v := range summaryStats {
		totalGameCount += v.GameCount
	}

	// We can't show the counts for *all* game types, so we'll group all the ones past the top five
	// into an "other" category.
	var otherGameCount int
	if len(summaryStats) > 5 {
		for _, v := range summaryStats[5:] {
			otherGameCount += v.GameCount
		}
	}

	var buf bytes.Buffer
	buf.WriteString(p.Sprintf("%d players and %d games (", summaryStats[1].PlayerCount, totalGameCount))

	// If for some reason we don't have 5 "top" game types...
	topN := 5
	if len(summaryStats) < topN {
		topN = len(summaryStats)
	}

	for i, v := range summaryStats[:topN] {
		buf.WriteString(p.Sprintf("%d %s", v.GameCount, v.GameTypeCd))

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

// LeaderboardHandler is the main page of the site
func (ae *AppEnv) LeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	// The summary stat line for all activity tracked thus far.
	allSummaryStats, err := leaderboard.SummaryStatsData("all", ae.db)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}

	// The summary stat line typically for the past day's worth of activity.
	daySummaryStats, err := leaderboard.SummaryStatsData("day", ae.db)
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

	// The structure passed to the template.
	type Data struct {
		StatLine      template.HTML
		DayStatLine   template.HTML
		ActivePlayers []leaderboard.ActivePlayerBase
		ActiveServers []leaderboard.ActiveServerBase
		ActiveMaps    []leaderboard.ActiveMapBase
	}

	data := Data{
		StatLine:      makeStatLine("", allSummaryStats, ""),
		DayStatLine:   makeStatLine("", daySummaryStats, ""),
		ActivePlayers: activePlayers,
		ActiveServers: activeServers,
		ActiveMaps:    activeMaps,
	}

	err = ae.templates.ExecuteTemplate(w, "leaderboard.page.html", data)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}
}
