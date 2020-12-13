package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/spf13/viper"
	"gitlab.com/xonotic/xonstat/pkg/game"
	"gitlab.com/xonotic/xonstat/pkg/player"
)

// PlayerInfoResponse is the view-specific information about a map related information.
type PlayerInfoResponse struct {
	Player            *player.InfoBase
	GameTypeSummaries []*player.GameTypeSummaryBase
	RecentGames       []game.RecentGameBase
}

// PlayerInfoHandler is the web handler for retrieving player information
func (ae *AppEnv) PlayerInfoHandler(w http.ResponseWriter, r *http.Request) {
	acceptHeader := r.Header.Get("Accept")

	playerID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("Invalid or missing player ID value: %s", err)
		ae.NotFoundHandler(w, r)
		return
	}

	info, err := player.InfoData(ae.db, playerID)
	if err != nil {
		log.Printf("Invalid or missing player ID value: %s", err)
		ae.NotFoundHandler(w, r)
		return
	}

	summaries, _ := player.GameTypeSummaryData(ae.db, playerID)

	recentGamesCutoff := time.Now().UTC().AddDate(0, 0, -1*viper.GetInt("RecentGamesDays"))
	recentGames, _ := game.RecentGamesData(ae.db, game.EmptyServerID, game.EmptyMapID, playerID,
		game.EmptyGameTypeCd, &recentGamesCutoff, game.EmptyStartGameID, game.EmptyEndGameID, 20)

	response := &PlayerInfoResponse{
		Player:            info,
		GameTypeSummaries: summaries,
		RecentGames:       recentGames,
	}

	if acceptHeader == "application/json" {
		bytes, _ := json.Marshal(response)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	} else {
		err = ae.templates["playerinfo.page.html"].Execute(w, response)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}
	}
}

// PlayerAccuracyRichData is the more detailed set of information for a given slice of the line chart for accuracy.
type PlayerAccuracyRichData struct {
	GameID           int      `json:"game_id"`
	WeaponCd         string   `json:"weapon_cd"`
	WeaponCdInitCaps string   `json:"weapon_cd_init_caps"`
	Hit              int      `json:"hit"`
	Fired            int      `json:"fired"`
	PctAccuracy      *float32 `json:"pct_accuracy"`
}

// NewPlayerAccuracyRichData converts an AccuracyBase into a PlayerAccuracyRichData object or a blank entry.
func NewPlayerAccuracyRichData(weaponCd string, gameID int, ab *player.AccuracyBase) *PlayerAccuracyRichData {
	if ab == nil {
		// Blank entry (nil for accuracy to let Chart.js know it is missing data)
		return &PlayerAccuracyRichData{
			WeaponCd:    weaponCd,
			PctAccuracy: nil,
		}
	}

	return &PlayerAccuracyRichData{
		GameID:           gameID,
		WeaponCd:         ab.WeaponCd,
		WeaponCdInitCaps: strings.Title(ab.WeaponCd),
		Hit:              ab.Hit,
		Fired:            ab.Fired,
		PctAccuracy:      &ab.PctAccuracy,
	}
}

// PlayerAccuracyDataset is player accuracy data in the "shape" that Chart.js wants.
type PlayerAccuracyDataset struct {
	Label           string                    `json:"label"`
	BackgroundColor string                    `json:"backgroundColor"`
	BorderColor     string                    `json:"borderColor"`
	RichData        []*PlayerAccuracyRichData `json:"richData"`
	Data            []float32                 `json:"data"`
}

// NewPlayerAccuracyDataset creates a new AccuracyDataSet from a weapon code.
func NewPlayerAccuracyDataset(weaponCd string) *AccuracyDataset {
	return &AccuracyDataset{
		Label:           weaponCd,
		BackgroundColor: weaponBackgroundColor(weaponCd),
		BorderColor:     weaponBorderColor(weaponCd),
		RichData:        make([]*AccuracyRichData, 0),
		Data:            make([]float32, 0),
	}
}

// PlayerWeaponInfoResponse is the response type for the PlayerWeaponInfoHandler.
type PlayerWeaponInfoResponse struct {
	PlayerID int      `json:"player_id"`
	Weapons  []string `json:"weapons"`
	GameIDs  []int    `json:"game_ids"`
}

// PlayerWeaponInfoHandler is the web handler for retrieving player weapon information
func (ae *AppEnv) PlayerWeaponInfoHandler(w http.ResponseWriter, r *http.Request) {
	playerID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("Invalid or missing player ID value: %s", err)
		ae.NotFoundHandler(w, r)
		return
	}

	info, err := player.PlayerWeaponInfoData(ae.db, playerID, 20, "")
	if err != nil {
		log.Printf("Unable to retrieve weapon info data: %s", err)
		ae.NotFoundHandler(w, r)
		return
	}

	response := &PlayerWeaponInfoResponse{
		PlayerID: playerID,
		Weapons:  info.Weapons,
		GameIDs:  info.GameIDs,
	}

	bytes, _ := json.Marshal(response)

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}
