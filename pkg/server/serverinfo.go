package server

import (
	"encoding/json"
	"html/template"
	"time"

	"github.com/antzucaro/qstr"
	"github.com/spf13/viper"
	"gitlab.com/xonotic/xonstat/pkg/leaderboard"
	"gitlab.com/xonotic/xonstat/pkg/models"
)

// TopScorerBase is the base type for players on a server and their
// accumulated score
type TopScorerBase struct {
	SortOrder    int
	PlayerID     int
	Nick         string
	NickStripped string
	NickHTML     template.HTML
	Score        int
}

// TopScorerData returns view-agnostic data for the top scorers on a given server.
func TopScorerData(db models.Datastore, serverID int) ([]*TopScorerBase, error) {
	activePlayerScoresCutoff := time.Now().UTC().AddDate(0, 0, -1*viper.GetInt("TopPlayersByScoreDays"))
	rawActivePlayerScores, err := db.RServerActivePlayerScores(serverID, &activePlayerScoresCutoff, 10)
	if err != nil {
		return nil, err
	}

	var topScorers []*TopScorerBase
	for _, v := range rawActivePlayerScores {
		nick := qstr.QStr(v.Nick)
		ts := TopScorerBase{
			SortOrder:    v.SortOrder,
			PlayerID:     v.PlayerID,
			Nick:         v.Nick,
			NickStripped: nick.Stripped(),
			NickHTML:     nick.HTML(),
			Score:        v.Score,
		}

		topScorers = append(topScorers, &ts)
	}
	return topScorers, nil
}

// TopActivePlayersData returns view-agnostic data for the most active players on a given server.
// NOTE: the base type returned here is shared with the leaderboard package.
func TopActivePlayersData(db models.Datastore, serverID int) ([]*leaderboard.ActivePlayerBase, error) {
	// Top players by alive time over the time period.
	activePlayersCutoff := time.Now().UTC().AddDate(0, 0, -1*viper.GetInt("TopPlayersByTimeDays"))
	rawActivePlayers, err := db.RActivePlayersByServer(serverID, &activePlayersCutoff, 10)
	if err != nil {
		return nil, err
	}

	activePlayers := leaderboard.ActivePlayerToActivePlayerBase(rawActivePlayers)
	return activePlayers, nil
}

// TopMapsData returns view-agnostic data for the most active maps on a given server by times played.
// NOTE: there is no base type here (it is straight from model/query result) because no manipulation
// or hiding is required.
func TopMapsData(db models.Datastore, serverID int) ([]*models.ActiveMap, error) {
	// Top maps by times played over the time period.
	activeMapsCutoff := time.Now().UTC().AddDate(0, 0, -1*viper.GetInt("TopMapsByGamesDays"))
	activeMaps, err := db.RActiveMapsByServer(serverID, &activeMapsCutoff, 10)
	if err != nil {
		return nil, err
	}

	return activeMaps, nil
}

// InfoBase is the base type used to represent servers for all
// marshalled types (HTML/JSON/etc).
type InfoBase struct {
	ServerID  int
	Name      string
	NameHTML  template.HTML
	IPAddr    string
	Port      int
	Revision  string
	ActiveInd bool
	CreateDt  *models.MultiDt
}

// InfoData retrieves information about a given server.
func InfoData(db models.Datastore, serverID int) (*InfoBase, error) {
	rawServer, err := db.RServerByID(serverID)
	if err != nil {
		return nil, err
	}

	// Conversions.
	name := qstr.QStr(rawServer.Name.String)
	dt, err := models.NewMultiDt(rawServer.CreateDt)
	if err != nil {
		return nil, err
	}

	return &InfoBase{
		ServerID:  rawServer.ServerID,
		Name:      rawServer.Name.String,
		NameHTML:  name.HTML(),
		IPAddr:    rawServer.IPAddr.String,
		Port:      int(rawServer.Port.Int64),
		Revision:  rawServer.Revision.String,
		ActiveInd: rawServer.ActiveInd,
		CreateDt:  dt,
	}, nil
}

// InfoJSON returns server records as JSON.
func InfoJSON(db models.Datastore, serverID int) ([]byte, error) {
	rawData, err := InfoData(db, serverID)
	if err != nil {
		return nil, err
	}

	type Response struct {
		ServerID  int    `json:"server_id"`
		Name      string `json:"name"`
		IPAddr    string `json:"ip_addr"`
		Port      int    `json:"port"`
		Revision  string `json:"revision"`
		ActiveInd bool   `json:"active_ind"`
		CreateDt  string `json:"create_dt"`
	}

	r := Response{
		ServerID:  rawData.ServerID,
		Name:      rawData.Name,
		IPAddr:    rawData.IPAddr,
		Port:      rawData.Port,
		Revision:  rawData.Revision,
		ActiveInd: rawData.ActiveInd,
		CreateDt:  rawData.CreateDt.UTCStr,
	}

	return json.Marshal(r)
}
