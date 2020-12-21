package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/spf13/viper"
	"gitlab.com/xonotic/xonstat/pkg/game"
	"gitlab.com/xonotic/xonstat/pkg/player"
	"gitlab.com/xonotic/xonstat/pkg/util"
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
	WeaponCd         string   `json:"weapon_cd"`
	WeaponCdInitCaps string   `json:"weapon_cd_init_caps"`
	Hit              int      `json:"hit"`
	Fired            int      `json:"fired"`
	PctAccuracy      *float32 `json:"pct_accuracy"`
}

// PlayerAccuracyDataset is player accuracy data in the "shape" that Chart.js wants.
type PlayerAccuracyDataset struct {
	Label           string                 `json:"label"`
	BackgroundColor string                 `json:"backgroundColor"`
	BorderColor     string                 `json:"borderColor"`
	Summary         PlayerAccuracyRichData `json:"summary"`
	Data            []*float32             `json:"data"`
	Fill            bool                   `json:"fill"`
	LineTension     float32                `json:"lineTension"`
	PointHitRadius  int                    `json:"pointHitRadius"`
}

// NewPlayerAccuracyDataset creates a new PlayerAccuracyDataSet from a weapon code.
func NewPlayerAccuracyDataset(weaponCd string) *PlayerAccuracyDataset {
	return &PlayerAccuracyDataset{
		Label:           weaponCd,
		BackgroundColor: weaponBackgroundColor(weaponCd),
		BorderColor:     weaponBorderColor(weaponCd),
		Data:            make([]*float32, 0),
		LineTension:     0.1,
		PointHitRadius:  5,
	}
}

// PlayerDamageRichData is the more detailed set of information for a given slice of the bar chart for damage.
type PlayerDamageRichData struct {
	WeaponCd         string `json:"weapon_cd"`
	WeaponCdInitCaps string `json:"weapon_cd_init_caps"`
	Actual           int    `json:"actual"`
	Max              int    `json:"max"`
}

// PlayerDamageDataset is player damage data in the "shape" that Chart.js wants.
type PlayerDamageDataset struct {
	Label           string               `json:"label"`
	BackgroundColor string               `json:"backgroundColor"`
	BorderColor     string               `json:"borderColor"`
	RichData        PlayerDamageRichData `json:"richData"`
	Data            []int                `json:"data"`
}

// NewPlayerDamageDataset creates a new PlayerDamageDataSet from a weapon code.
func NewPlayerDamageDataset(weaponCd string) *PlayerDamageDataset {
	return &PlayerDamageDataset{
		Label:           weaponCd,
		BackgroundColor: weaponBackgroundColor(weaponCd),
		BorderColor:     weaponBorderColor(weaponCd),
		Data:            make([]int, 0),
	}
}

// PlayerWeaponInfoResponse is the response type for the PlayerWeaponInfoHandler.
type PlayerWeaponInfoResponse struct {
	PlayerID     int                      `json:"player_id"`
	Weapons      []string                 `json:"weapons"`
	GameIDs      []int                    `json:"game_ids"`
	AccuracyData []*PlayerAccuracyDataset `json:"accuracy"`
	DamageData   []*PlayerDamageDataset   `json:"damage"`
}

// Assemble the accuracy data in a Chart.js friendly format.
func assembleAccuracy(weaponsUsed []string, gameIDs []int, rawAccuracy map[string]*player.AccuracyBase) []*PlayerAccuracyDataset {
	accuracy := make([]*PlayerAccuracyDataset, 0)
	for _, weaponCd := range weaponsUsed {
		if !isAccuracyWeapon(weaponCd) {
			continue
		}

		dataset := NewPlayerAccuracyDataset(weaponCd)
		richData := PlayerAccuracyRichData{
			WeaponCd:         weaponCd,
			WeaponCdInitCaps: strings.Title(weaponCd),
		}

		// For each game ID, we'll pull that weapon's data
		for _, gameID := range gameIDs {
			base, _ := rawAccuracy[fmt.Sprintf("%d-%s", gameID, weaponCd)]
			if base == nil {
				// Player did not use this weapon in this game, so put a blank marker.
				dataset.Data = append(dataset.Data, nil)
			} else {
				dataset.Data = append(dataset.Data, &base.PctAccuracy)

				// Keep track of the overall numbers to provide hover information.
				richData.Hit += base.Hit
				richData.Fired += base.Fired
			}
		}
		// Calculate the overall accuracy from total hits and fired-s.
		overallAccuracy := util.Percentage(richData.Hit, richData.Fired)
		richData.PctAccuracy = &overallAccuracy
		dataset.Summary = richData

		accuracy = append(accuracy, dataset)
	}

	return accuracy
}

// Assemble the damage data in a Chart.js friendly format.
func assembleDamage(weaponsUsed []string, gameIDs []int, rawDamage map[string]*player.DamageBase) []*PlayerDamageDataset {
	damage := make([]*PlayerDamageDataset, 0)
	for _, weaponCd := range weaponsUsed {
		if !isDamageWeapon(weaponCd) {
			continue
		}

		dataset := NewPlayerDamageDataset(weaponCd)
		richData := PlayerDamageRichData{
			WeaponCd:         weaponCd,
			WeaponCdInitCaps: strings.Title(weaponCd),
		}

		// For each game ID, we'll pull that weapon's data
		for _, gameID := range gameIDs {
			base, _ := rawDamage[fmt.Sprintf("%d-%s", gameID, weaponCd)]
			if base == nil {
				// Player did not use this weapon in this game, so put a blank marker.
				dataset.Data = append(dataset.Data, 0)
			} else {
				dataset.Data = append(dataset.Data, base.Actual)

				// Keep track of the overall numbers to provide hover information.
				richData.Actual += base.Actual
				richData.Max += base.Max
			}
		}
		dataset.RichData = richData

		damage = append(damage, dataset)
	}

	return damage
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

	// Sort the game IDs in ascending order for processing.
	sort.Ints(info.GameIDs)

	// Note that we do two passes through the raw weapon data here. This is largely because of the different
	// aggregation between accuracy and damage. For accuracy, we want to show the average accuracy over
	// all games (e.g. 40% vortex accuracy) whereas for damage we want to know percentage per game (e.g.
	// devastator was 35% of the total damage for this game).
	accuracy := assembleAccuracy(info.Weapons, info.GameIDs, info.Accuracy)
	damage := assembleDamage(info.Weapons, info.GameIDs, info.Damage)

	response := &PlayerWeaponInfoResponse{
		PlayerID:     playerID,
		Weapons:      info.Weapons,
		GameIDs:      info.GameIDs,
		AccuracyData: accuracy,
		DamageData:   damage,
	}

	bytes, _ := json.Marshal(response)

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}
