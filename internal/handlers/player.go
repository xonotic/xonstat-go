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
	OverallStats      map[string]*player.OverallStatsBase
	GameTypeSummaries []*player.GameTypeSummaryBase
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

	summaries, err := player.GameTypeSummaryData(ae.db, playerID)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}

	rawOverallStats, err := player.OverallStatsData(ae.db, playerID)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}

	overallStats := make(map[string]*player.OverallStatsBase, len(rawOverallStats))
	for _, e := range rawOverallStats {
		overallStats[e.GameTypeCd] = e
	}

	response := &PlayerInfoResponse{
		Player:            info,
		GameTypeSummaries: summaries,
		OverallStats:      overallStats,
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

// PlayerAccuracySummary is the more detailed set of information for a given slice of the line chart for accuracy.
type PlayerAccuracySummary struct {
	WeaponCd         string   `json:"weapon_cd"`
	WeaponCdInitCaps string   `json:"weapon_cd_init_caps"`
	Hit              int      `json:"hit"`
	Fired            int      `json:"fired"`
	PctAccuracy      *float32 `json:"pct_accuracy"`
}

// PlayerAccuracyDataset is player accuracy data in the "shape" that Chart.js wants.
type PlayerAccuracyDataset struct {
	Label           string                `json:"label"`
	BackgroundColor string                `json:"backgroundColor"`
	BorderColor     string                `json:"borderColor"`
	Summary         PlayerAccuracySummary `json:"summary"`
	Data            []*float32            `json:"data"`
	Fill            bool                  `json:"fill"`
	LineTension     float32               `json:"lineTension"`
	PointHitRadius  int                   `json:"pointHitRadius"`
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
	TotalDamage      int    `json:"t_damage"`
}

// PlayerDamageDataset is player damage data in the "shape" that Chart.js wants.
type PlayerDamageDataset struct {
	Label           string                  `json:"label"`
	BackgroundColor string                  `json:"backgroundColor"`
	BorderColor     string                  `json:"borderColor"`
	RichData        []*PlayerDamageRichData `json:"richData"`
	Data            []int                   `json:"data"`
}

// NewPlayerDamageDataset creates a new PlayerDamageDataSet from a weapon code.
func NewPlayerDamageDataset(weaponCd string) *PlayerDamageDataset {
	return &PlayerDamageDataset{
		Label:           weaponCd,
		BackgroundColor: weaponBackgroundColor(weaponCd),
		BorderColor:     weaponBorderColor(weaponCd),
		RichData:        make([]*PlayerDamageRichData, 0),
		Data:            make([]int, 0),
	}
}

// PlayerWeaponInfoResponse is the response type for the PlayerWeaponInfoHandler.
type PlayerWeaponInfoResponse struct {
	PlayerID           int                      `json:"player_id"`
	Weapons            []string                 `json:"weapons"`
	GameIDs            []int                    `json:"game_ids"`
	AccuracyData       []*PlayerAccuracyDataset `json:"accuracy"`
	DamageData         []*PlayerDamageDataset   `json:"damage"`
	TotalDamagePerGame []int                    `json:"total_damage_per_game"`
}

// Assemble the accuracy data in a Chart.js friendly format.
func assembleAccuracy(weaponsUsed []string, gameIDs []int, rawAccuracy map[string]*player.AccuracyBase) []*PlayerAccuracyDataset {
	accuracy := make([]*PlayerAccuracyDataset, 0)
	for _, weaponCd := range weaponsUsed {
		if !isAccuracyWeapon(weaponCd) {
			continue
		}

		dataset := NewPlayerAccuracyDataset(weaponCd)
		summary := PlayerAccuracySummary{
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
				summary.Hit += base.Hit
				summary.Fired += base.Fired
			}
		}
		// Calculate the overall accuracy from total hits and fired-s.
		overallAccuracy := util.Percentage(summary.Hit, summary.Fired)
		summary.PctAccuracy = &overallAccuracy
		dataset.Summary = summary

		accuracy = append(accuracy, dataset)
	}

	return accuracy
}

// Assemble the damage data in a Chart.js friendly format.
func assembleDamage(weaponsUsed []string, gameIDs []int, rawDamage map[string]*player.DamageBase) ([]*PlayerDamageDataset, []int) {
	// Map of weapon codes to datasets
	datasets := make(map[string]*PlayerDamageDataset)

	// We need the total damage by weapon to determine placement of the bar segments in the chart.
	// The weapons with the most total damage would form the "base" of the bars. Since we'd like
	// to control the ordering of the datasets based on this order, we keep track of those in a
	// map as well.
	totalDamageByWeapon := make(map[string]int)
	for _, weaponCd := range weaponsUsed {
		datasets[weaponCd] = NewPlayerDamageDataset(weaponCd)
		totalDamageByWeapon[weaponCd] = 0
	}

	// We need the total damage dealt per game to show the contribution any given weapon
	// had towards it as a percentage (e.g. your damage in game 123 was 45% from vortex).
	totalDamagePerGame := make([]int, 0, len(gameIDs))

	for _, gameID := range gameIDs {
		totalDamage := 0
		for _, weaponCd := range weaponsUsed {
			base, _ := rawDamage[fmt.Sprintf("%d-%s", gameID, weaponCd)]
			if base == nil {
				// Player did not use this weapon in this game, so put a blank marker.
				datasets[weaponCd].Data = append(datasets[weaponCd].Data, 0)
			} else {
				datasets[weaponCd].Data = append(datasets[weaponCd].Data, base.Actual)
				totalDamage += base.Actual
				totalDamageByWeapon[weaponCd] += base.Actual
			}
		}
		totalDamagePerGame = append(totalDamagePerGame, totalDamage)
	}

	// Here is where we add the datasets in a particular order so they are shown properly.
	weaponOrder := sortDescByIntValue(totalDamageByWeapon)
	datasetList := make([]*PlayerDamageDataset, 0, len(datasets))
	for _, weaponCd := range weaponOrder {
		datasetList = append(datasetList, datasets[weaponCd])
	}

	return datasetList, totalDamagePerGame
}

// PlayerWeaponInfoHandler is the web handler for retrieving player weapon information
func (ae *AppEnv) PlayerWeaponInfoHandler(w http.ResponseWriter, r *http.Request) {
	playerID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("Invalid or missing player ID value: %s", err)
		ae.NotFoundHandler(w, r)
		return
	}

	params := r.URL.Query()
	gameTypeCd := params.Get("game_type_cd")

	info, err := player.PlayerWeaponInfoData(ae.db, playerID, 20, gameTypeCd)
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
	damage, totalDamagePerGame := assembleDamage(info.Weapons, info.GameIDs, info.Damage)

	response := &PlayerWeaponInfoResponse{
		PlayerID:           playerID,
		Weapons:            info.Weapons,
		GameIDs:            info.GameIDs,
		AccuracyData:       accuracy,
		DamageData:         damage,
		TotalDamagePerGame: totalDamagePerGame,
	}

	bytes, _ := json.Marshal(response)

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}

// PlayerRecentGamesFragmentHandler retrieves just the recent games table.
func (ae *AppEnv) PlayerRecentGamesFragmentHandler(w http.ResponseWriter, r *http.Request) {
	playerID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("invalid or missing player ID value: %s", err)
		ae.NotFoundHandler(w, r)
		return
	}

	params := r.URL.Query()
	gameTypeCd := params.Get("game_type_cd")

	recentGamesCutoff := time.Now().UTC().AddDate(0, 0, -1*viper.GetInt("RecentGamesDays"))
	recentGames, _ := game.RecentGamesData(ae.db, game.EmptyServerID, game.EmptyMapID, playerID,
		gameTypeCd, &recentGamesCutoff, game.EmptyStartGameID, game.EmptyEndGameID, 20)

	err = ae.templates["recentgames.fragment.page.html"].Execute(w, recentGames)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}
}
