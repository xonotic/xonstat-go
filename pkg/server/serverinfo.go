package server

import (
	"encoding/json"
	"html/template"
	"time"

	"github.com/antzucaro/qstr"
	"github.com/nleeper/goment"
	"gitlab.com/xonotic/xonstat/pkg/models"
)

// ServerInfoBase is the base type used to represent servers for all
// marshalled types (HTML/JSON/etc).
type ServerInfoBase struct {
	ServerID       int
	Name           string
	NameHTML       template.HTML
	IPAddr         string
	Port           int
	Revision       string
	ActiveInd      bool
	CreateDt       time.Time
	CreateDtEpoch  int64
	CreateDtUTCStr string
	CreateDtFuzzy  string
}

// ServerInfoData retrieves information about a given server.
func ServerInfoData(db models.Datastore, ID int) (*ServerInfoBase, error) {
	rawServer, err := db.RServerByID(ID)
	if err != nil {
		return nil, err
	}

	// Conversions.
	name := qstr.QStr(rawServer.Name.String)
	dtUTC := rawServer.CreateDt.UTC()
	fuzzyDt, _ := goment.New(dtUTC)

	return &ServerInfoBase{
		ServerID:       rawServer.ServerID,
		Name:           rawServer.Name.String,
		NameHTML:       name.HTML(),
		IPAddr:         rawServer.IPAddr.String,
		Port:           int(rawServer.Port.Int64),
		Revision:       rawServer.Revision.String,
		ActiveInd:      rawServer.ActiveInd,
		CreateDt:       rawServer.CreateDt,
		CreateDtEpoch:  rawServer.CreateDt.Unix(),
		CreateDtUTCStr: dtUTC.Format("Mon, 2 Jan 2006 15:04:05 MST"),
		CreateDtFuzzy:  fuzzyDt.FromNow(),
	}, nil
}

// ServerInfoJSON returns server records as JSON.
func ServerInfoJSON(db models.Datastore, ID int) ([]byte, error) {
	rawData, err := ServerInfoData(db, ID)
	if err != nil {
		return nil, err
	}

	// the JSON response (a single entry in the list)
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
		CreateDt:  rawData.CreateDtUTCStr,
	}

	return json.Marshal(r)
}
