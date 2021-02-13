package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/alehano/reverse"
	"github.com/spf13/viper"
	"gitlab.com/xonotic/xonstat/pkg/game"
	"gitlab.com/xonotic/xonstat/pkg/leaderboard"
	"golang.org/x/text/message"
)

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
