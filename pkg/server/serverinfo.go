package server

import (
	"encoding/json"
	"html/template"
	"time"

	"github.com/antzucaro/qstr"
	"github.com/nleeper/goment"
	"github.com/spf13/viper"
	"gitlab.com/xonotic/xonstat/pkg/leaderboard"
	"gitlab.com/xonotic/xonstat/pkg/models"
)

// ServerInfoBase is the base type used to represent servers for all
// marshalled types (HTML/JSON/etc).
type ServerInfoBase struct {
	ServerID           int
	Name               string
	NameHTML           template.HTML
	IPAddr             string
	Port               int
	Revision           string
	ActiveInd          bool
	CreateDt           time.Time
	CreateDtEpoch      int64
	CreateDtUTCStr     string
	CreateDtFuzzy      string
	ActivePlayerScores []*models.ActivePlayerScore
	ActivePlayers      []leaderboard.ActivePlayerBase
	ActiveMaps         []*models.ActiveMap
}

// ServerInfoData retrieves information about a given server.
func ServerInfoData(db models.Datastore, ID int) (*ServerInfoBase, error) {
	rawServer, err := db.RServerByID(ID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	cutoffDays := viper.GetInt("TopMapsByGamesDays")
	cutoff := now.AddDate(0, 0, -1*cutoffDays)

	activePlayerScores, err := db.RActivePlayerScores(ID, &cutoff, 10)
	rawActivePlayers, _ := db.RActivePlayersByServer(ID, &cutoff, 10)
	activePlayers := leaderboard.ActivePlayerToActivePlayerBase(rawActivePlayers)

	activeMaps, _ := db.RActiveMapsByServer(ID, &cutoff, 10)

	// Conversions.
	name := qstr.QStr(rawServer.Name.String)
	dtUTC := rawServer.CreateDt.UTC()
	fuzzyDt, _ := goment.New(dtUTC)

	return &ServerInfoBase{
		ServerID:           rawServer.ServerID,
		Name:               rawServer.Name.String,
		NameHTML:           name.HTML(),
		IPAddr:             rawServer.IPAddr.String,
		Port:               int(rawServer.Port.Int64),
		Revision:           rawServer.Revision.String,
		ActiveInd:          rawServer.ActiveInd,
		CreateDt:           rawServer.CreateDt,
		CreateDtEpoch:      rawServer.CreateDt.Unix(),
		CreateDtUTCStr:     dtUTC.Format("Mon, 2 Jan 2006 15:04:05 MST"),
		CreateDtFuzzy:      fuzzyDt.FromNow(),
		ActivePlayerScores: activePlayerScores,
		ActivePlayers:      activePlayers,
		ActiveMaps:         activeMaps,
	}, nil
}

// ServerInfoJSON returns server records as JSON.
func ServerInfoJSON(db models.Datastore, ID int) ([]byte, error) {
	rawData, err := ServerInfoData(db, ID)
	if err != nil {
		return nil, err
	}

	// the JSON response (a single entry in the list)
	type ActivePlayerScore struct {
		Rank     int    `json:"rank"`
		PlayerID int    `json:"player_id"`
		Nick     string `json:"nick"`
		Score    int    `json:"score"`
	}

	type ActivePlayer struct {
		Rank      int    `json:"rank"`
		PlayerID  int    `json:"player_id"`
		Nick      string `json:"nick"`
		AliveTime string `json:"alivetime"`
	}

	type ActiveMap struct {
		Rank    int    `json:"rank"`
		MapName string `json:"map_name"`
		MapID   int    `json:"map_id"`
		Games   int    `json:"games"`
	}

	type Response struct {
		ServerID           int                 `json:"server_id"`
		Name               string              `json:"name"`
		IPAddr             string              `json:"ip_addr"`
		Port               int                 `json:"port"`
		Revision           string              `json:"revision"`
		ActiveInd          bool                `json:"active_ind"`
		CreateDt           string              `json:"create_dt"`
		ActivePlayerScores []ActivePlayerScore `json:"active_player_scores"`
		ActivePlayers      []ActivePlayer      `json:"active_players"`
		ActiveMaps         []ActiveMap         `json:"active_maps"`
	}

	var aps []ActivePlayerScore
	for _, v := range rawData.ActivePlayerScores {
		aps = append(aps, ActivePlayerScore{v.SortOrder, v.PlayerID, v.NickStripped, v.Score})
	}

	var am []ActiveMap
	for _, v := range rawData.ActiveMaps {
		am = append(am, ActiveMap{v.SortOrder, v.MapName, v.MapID, v.Games})
	}

	var ap []ActivePlayer
	for _, v := range rawData.ActivePlayers {
		nick := qstr.QStr(v.Nick)
		ap = append(ap, ActivePlayer{v.SortOrder, v.PlayerID, nick.Stripped(), v.AliveTime})
	}

	r := Response{
		ServerID:           rawData.ServerID,
		Name:               rawData.Name,
		IPAddr:             rawData.IPAddr,
		Port:               rawData.Port,
		Revision:           rawData.Revision,
		ActiveInd:          rawData.ActiveInd,
		CreateDt:           rawData.CreateDtUTCStr,
		ActivePlayerScores: aps,
		ActivePlayers:      ap,
		ActiveMaps:         am,
	}

	return json.Marshal(r)
}
