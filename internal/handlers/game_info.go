package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"gitlab.com/xonotic/xonstat/pkg/game"
	"gitlab.com/xonotic/xonstat/pkg/mmap"
	"gitlab.com/xonotic/xonstat/pkg/models"
	"gitlab.com/xonotic/xonstat/pkg/server"
)

type teamGameStatJSON struct {
	Score  int    `json:"score"`
	Rounds int    `json:"rounds"`
	Caps   int    `json:"caps"`
	Color  string `json:"color"`
}

type playerGameStatJSON struct {
	PlayerGameStatID int     `json:"player_game_stat_id"`
	PlayerID         int     `json:"player_id"`
	Nick             string  `json:"nick"`
	Team             int     `json:"team"`
	Color            string  `json:"color"`
	AliveTime        string  `json:"alivetime"`
	AliveTimeMS      int     `json:"alivetime_ms"`
	Kills            int     `json:"kills"`
	Deaths           int     `json:"deaths"`
	Suicides         int     `json:"suicides"`
	Score            int     `json:"score"`
	Time             string  `json:"time"`
	TimeMS           int     `json:"time_ms"`
	Captures         int     `json:"captures"`
	Pickups          int     `json:"pickups"`
	Drops            int     `json:"drops"`
	Returns          int     `json:"returns"`
	Collects         int     `json:"collects"`
	Destroys         int     `json:"destroys"`
	Pushes           int     `json:"pushes"`
	CarrierFrags     int     `json:"carrier_frags"`
	Fastest          float64 `json:"fastest"`
	FastestMS        int     `json:"fastest_ms"`
	AvgLatency       int     `json:"avg_latency"`
	ScoreboardPos    int     `json:"scoreboard_pos"`
	Laps             int     `json:"laps"`
	Revivals         int     `json:"revivals"`
	Lives            int     `json:"lives"`
}

type nonParticipantJSON struct {
	PlayerGameNonParticipantID int    `json:"player_game_nonparticipant_id"`
	PlayerID                   int    `json:"player_id"`
	GameID                     int    `json:"game_id"`
	Nick                       string `json:"nick"`
	AliveTime                  string `json:"alivetime"`
	AliveTimeMS                int    `json:"alivetime_ms"`
	Score                      int    `json:"score"`
}

type gameInfoJSONResponse struct {
	GameID              int                      `json:"game_id"`
	GameTypeCd          string                   `json:"game_type_cd"`
	GameTypeDescr       string                   `json:"game_type_descr"`
	Duration            string                   `json:"duration"`
	DurationSecs        int                      `json:"duration_secs"`
	Winner              int                      `json:"winner"`
	MatchID             string                   `json:"match_id"`
	Mod                 string                   `json:"mod"`
	CreateDt            string                   `json:"create_dt"`
	CreateDtFuzzy       string                   `json:"create_dt_fuzzy"`
	ServerID            int                      `json:"server_id"`
	MapID               int                      `json:"map_id"`
	TeamGameStatsByTeam map[int]teamGameStatJSON `json:"team_game_stats"`
	PlayerGameStats     []playerGameStatJSON     `json:"player_game_stats"`
	Forfeits            []nonParticipantJSON     `json:"forfeits"`
	Spectators          []nonParticipantJSON     `json:"spectators"`
}

func gameInfoBaseToJSON(ib game.InfoBase) ([]byte, error) {

	teamGameStats := make(map[int]teamGameStatJSON, len(ib.TeamGameStatsByTeam))
	for i, v := range ib.TeamGameStatsByTeam {
		teamGameStats[i] = teamGameStatJSON{
			Score:  v.Score,
			Rounds: v.Rounds,
			Caps:   v.Caps,
			Color:  v.Color,
		}
	}

	playerGameStats := make([]playerGameStatJSON, len(ib.PlayerGameStats))
	for i, v := range ib.PlayerGameStats {
		playerGameStats[i] = playerGameStatJSON{
			PlayerGameStatID: v.PlayerGameStatID,
			PlayerID:         v.PlayerID,
			Nick:             v.Nick.NickStripped,
			Team:             v.Team,
			Color:            v.Color,
			AliveTime:        v.AliveTime.Short,
			AliveTimeMS:      int(v.AliveTime.Milliseconds),
			Kills:            v.Kills,
			Deaths:           v.Deaths,
			Suicides:         v.Suicides,
			Score:            v.Score,
			Time:             v.Time.Short,
			TimeMS:           int(v.Time.Milliseconds),
			Captures:         v.Captures,
			Pickups:          v.Pickups,
			Drops:            v.Drops,
			Returns:          v.Returns,
			Collects:         v.Collects,
			Destroys:         v.Destroys,
			Pushes:           v.Pushes,
			CarrierFrags:     v.CarrierFrags,
			Fastest:          v.Fastest.Seconds,
			FastestMS:        int(v.Fastest.Milliseconds),
			AvgLatency:       v.AvgLatency,
			ScoreboardPos:    v.ScoreboardPos,
			Laps:             v.Laps,
			Revivals:         v.Revivals,
			Lives:            v.Lives,
		}
	}

	toNonParticipantJSON := func(l []*game.NonParticipantBase) []nonParticipantJSON {
		npJSON := make([]nonParticipantJSON, len(l))
		for i, v := range l {
			npJSON[i] = nonParticipantJSON{
				PlayerGameNonParticipantID: v.PlayerGameNonParticipantID,
				PlayerID:                   v.PlayerID,
				GameID:                     v.GameID,
				Nick:                       v.Nick.NickStripped,
				AliveTime:                  v.AliveTime.Short,
				AliveTimeMS:                int(v.AliveTime.Milliseconds),
				Score:                      int(v.Score),
			}
		}
		return npJSON
	}

	forfeits := toNonParticipantJSON(ib.Forfeits)
	spectators := toNonParticipantJSON(ib.Spectators)

	response := gameInfoJSONResponse{
		GameID:              ib.GameID,
		GameTypeCd:          ib.GameTypeCd,
		GameTypeDescr:       ib.GameTypeDescr,
		Duration:            ib.Duration.Short,
		DurationSecs:        int(ib.Duration.Seconds),
		Winner:              ib.Winner,
		MatchID:             ib.MatchID,
		Mod:                 ib.Mod,
		CreateDt:            ib.CreateDt.Dt.Format(time.RFC3339),
		CreateDtFuzzy:       ib.CreateDt.Fuzzy,
		ServerID:            ib.ServerID,
		MapID:               ib.MapID,
		TeamGameStatsByTeam: teamGameStats,
		PlayerGameStats:     playerGameStats,
		Forfeits:            forfeits,
		Spectators:          spectators,
	}

	return json.Marshal(response)
}

type gameInfoResponse struct {
	GameID                int
	GameTypeCd            string
	GameTypeDescr         string
	Duration              *models.MultiDuration
	Winner                int
	MatchID               string
	Mod                   string
	CreateDt              *models.MultiDt
	Server                *server.InfoBase
	Map                   *mmap.InfoBase
	TeamGameStatsByTeam   map[int]*game.TeamGameStatBase
	TeamOrdering          []int
	PlayerGameStats       []*game.PlayerGameStatBase
	PlayerGameStatsByTeam map[int][]*game.PlayerGameStatBase
	FragMatrix            map[int][]int
	Forfeits              []*game.NonParticipantBase
	Spectators            []*game.NonParticipantBase
	ShowWeaponCharts      bool
	ShowFragMatrix        bool
}

// It doesn't make sense to show weapon charts for certain game modes.
func shouldShowWeaponCharts(gameTypeCd string) bool {
	switch gameTypeCd {
	case "cts", "nb", "nexball":
		return false
	default:
		return true
	}
}

// GameInfoHandler godoc
// @Summary Game information
// @Accept  json
// @Produce  json
// @Param id path int true "game_id"
// @Success 200 {object} gameInfoJSONResponse
// @Router /game/{id} [get]
func (ae *AppEnv) GameInfoHandler(w http.ResponseWriter, r *http.Request) {
	acceptHeader := strings.ToLower(r.Header.Get("Accept"))

	gameID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("Invalid or missing game ID value: %s", err)
		ae.NotFoundHandler(w, r)
		return
	}

	gameInfo, err := game.InfoData(ae.db, gameID)
	if err != nil {
		log.Printf("Error: %s", err)
		ae.FiveHundredHandler(w, r)
		return
	}

	if acceptHeader == "application/json" {
		// JSON response
		bytes, err := gameInfoBaseToJSON(*gameInfo)
		if err != nil {
			log.Printf("Error: %s", err)
			ae.FiveHundredHandler(w, r)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	} else {
		// HTML response
		serverInfo, err := server.InfoData(ae.db, gameInfo.ServerID)
		if err != nil {
			log.Printf("Error: %s", err)
			ae.NotFoundHandler(w, r)
			return
		}

		mapInfo, err := mmap.InfoData(ae.db, gameInfo.MapID)
		if err != nil {
			log.Printf("Error: %s", err)
			ae.NotFoundHandler(w, r)
			return
		}

		response := gameInfoResponse{
			GameID:                gameInfo.GameID,
			GameTypeCd:            gameInfo.GameTypeCd,
			GameTypeDescr:         gameInfo.GameTypeDescr,
			Duration:              gameInfo.Duration,
			Winner:                gameInfo.Winner,
			MatchID:               gameInfo.MatchID,
			Mod:                   gameInfo.Mod,
			CreateDt:              gameInfo.CreateDt,
			Server:                serverInfo,
			Map:                   mapInfo,
			TeamGameStatsByTeam:   gameInfo.TeamGameStatsByTeam,
			TeamOrdering:          gameInfo.TeamOrdering,
			PlayerGameStats:       gameInfo.PlayerGameStats,
			PlayerGameStatsByTeam: gameInfo.PlayerGameStatsByTeam,
			FragMatrix:            gameInfo.FragMatrix,
			Forfeits:              gameInfo.Forfeits,
			Spectators:            gameInfo.Spectators,
			ShowWeaponCharts:      shouldShowWeaponCharts(gameInfo.GameTypeCd),
			ShowFragMatrix:        len(gameInfo.FragMatrix) > 1,
		}

		err = ae.templates["gameinfo.page.html"].Execute(w, response)
		if err != nil {
			log.Printf("Error: %s", err)
			ae.FiveHundredHandler(w, r)
			return
		}
	}
}
