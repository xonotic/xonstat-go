package mmap

import (
	"github.com/antzucaro/qstr"
	"github.com/spf13/viper"
	"gitlab.com/xonotic/xonstat/pkg/leaderboard"
	"gitlab.com/xonotic/xonstat/pkg/models"
	"gitlab.com/xonotic/xonstat/pkg/server"
	"log"
	"time"
)

// InfoBase is the view-agnostic information about a map.
type InfoBase struct {
	MapID    int
	Name     string
	CreateDt *models.MultiDt
}

// InfoData retrieves all of the relevant information about a map and returns it.
func InfoData(db models.Datastore, mapID int) (*InfoBase, error) {
	m, err := db.RMapByID(mapID)
	if err != nil {
		return nil, err
	}

	dt, err := models.NewMultiDt(m.CreateDt)
	if err != nil {
		return nil, err
	}

	return &InfoBase{
		MapID:    m.MapID,
		Name:     m.Name,
		CreateDt: dt,
	}, nil
}

// TopScorerData returns view-agnostic data for the top scorers on a given map.
// The topscore struct is borrowed from the server package since they have the same fields (only the
// query is different).
func TopScorerData(db models.Datastore, mapID int) ([]*server.TopScorerBase, error) {
	activePlayersScoresDays := viper.GetInt("TopPlayersByScoreDays")
	if activePlayersScoresDays > 60 {
		log.Printf("Warning: limiting TopPlayersByScoreDays to 60 for performance reasons. Original value: %d.", activePlayersScoresDays)
		activePlayersScoresDays = 60
	}

	activePlayerScoresCutoff := time.Now().UTC().AddDate(0, 0, -1*activePlayersScoresDays)
	rawActivePlayerScores, err := db.RMapActivePlayerScores(mapID, &activePlayerScoresCutoff, 10)
	if err != nil {
		return nil, err
	}

	var topScorers []*server.TopScorerBase
	for _, v := range rawActivePlayerScores {
		nick := qstr.QStr(v.Nick)
		ts := server.TopScorerBase{
			SortOrder:    v.SortOrder,
			PlayerID:     v.PlayerID,
			Nick:         v.Nick,
			NickStripped: nick.Stripped(),
			NickHTML:     nick.HTML(),
		}

		topScorers = append(topScorers, &ts)
	}
	return topScorers, nil
}

// TopActivePlayersData returns view-agnostic data for the most active players on a given map.
// NOTE: the base type returned here is shared with the leaderboard package.
func TopActivePlayersData(db models.Datastore, mapID int) ([]*leaderboard.ActivePlayerBase, error) {
	// Top players by alive time over the time period.
	activePlayersDays := viper.GetInt("TopPlayersByTimeDays")
	if activePlayersDays > 60 {
		log.Printf("Warning: limiting TopPlayersByTimeDays to 60 for performance reasons. Original value: %d.", activePlayersDays)
		activePlayersDays = 60
	}

	activePlayersCutoff := time.Now().UTC().AddDate(0, 0, -1*activePlayersDays)
	rawActivePlayers, err := db.RActivePlayersByMap(mapID, &activePlayersCutoff, 10)
	if err != nil {
		return nil, err
	}

	activePlayers := leaderboard.ActivePlayerToActivePlayerBase(rawActivePlayers)
	return activePlayers, nil
}

// TopActiveServersData returns view-agnostic data about the servers who have played a given map
// the most.
// NOTE: the base type returned here is shared with the leaderboard package.
func TopActiveServersData(db models.Datastore, mapID int) ([]*leaderboard.ActiveServerBase, error) {
	playtimeDays := viper.GetInt("TopServersByPlaytimeDays")
	if playtimeDays > 60 {
		log.Printf("Warning: limiting TopServersByPlaytimeDays to 60 for performance reasons. Original value: %d.", playtimeDays)
		playtimeDays = 60
	}

	cutoff := time.Now().UTC().AddDate(0, 0, -1*playtimeDays)
	rawActiveServers, err := db.RActiveServersByMap(mapID, &cutoff, 10)
	if err != nil {
		return nil, err
	}

	var activeServers []*leaderboard.ActiveServerBase
	for _, v := range rawActiveServers {
		as := leaderboard.ActiveServerBase{
			SortOrder:  v.SortOrder,
			ServerID:   v.ServerID,
			ServerName: v.ServerName,
			PlayTime:   models.DurationString(v.PlayTime, true),
		}
		activeServers = append(activeServers, &as)
	}
	return activeServers, nil
}
