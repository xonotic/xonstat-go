package mmap

import (
	"github.com/antzucaro/qstr"
	"github.com/spf13/viper"
	"gitlab.com/xonotic/xonstat/pkg/models"
	"gitlab.com/xonotic/xonstat/pkg/server"
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
	activePlayerScoresCutoff := time.Now().UTC().AddDate(0, 0, -1*viper.GetInt("TopPlayersByScoreDays"))
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
