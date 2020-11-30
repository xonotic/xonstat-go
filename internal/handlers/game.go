package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"gitlab.com/xonotic/xonstat/pkg/game"
	"gitlab.com/xonotic/xonstat/pkg/submission"
)

// RecentGamesHandler retrieves information about games played by varios filter criteria
func (ae *AppEnv) RecentGamesHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: lots of GET query params and input validation here. For now just go
	// for the simple case for the leaderboard.
	acceptHeader := r.Header.Get("Accept")

	params := r.URL.Query()

	gameTypeCd := params.Get("game_type_cd")
	if !submission.IsSupportedGameType(gameTypeCd) {
		// It is not a supported game type. Use the default for the DB query and don't include it
		// in the resulting query string.
		gameTypeCd = ""
		params.Del("game_type_cd")
	}

	startGameID, err := strconv.Atoi(params.Get("start_game_id"))
	params.Del("start_game_id")
	if err != nil {
		startGameID = -1
	}

	if acceptHeader == "application/json" {
		// JSON response
		recentGames, err := game.RecentGamesJSON(ae.db, -1, -1, -1, gameTypeCd, nil, startGameID, -1, 20)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(recentGames)
	} else {
		// HTML response
		gameTypeCds := []string{"overall", "duel", "ctf", "dm", "tdm", "ca", "kh", "ft", "as", "dom", "nb", "cts"}

		recentGames, err := game.RecentGamesData(ae.db, -1, -1, -1, gameTypeCd, nil, startGameID, -1, 20)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}

		// Set the query string for pagination
		if len(recentGames) > 0 {
			params.Set("start_game_id", fmt.Sprintf("%d", recentGames[len(recentGames)-1].GameID-1))
		}

		nextQueryStr := template.URL(params.Encode())

		type Data struct {
			GameTypeCds      []string
			ActiveGameTypeCd string
			RecentGames      []game.RecentGameBase
			ShowMoreLink     bool
			NextQueryStr     template.URL
		}

		data := Data{
			GameTypeCds:      gameTypeCds,
			ActiveGameTypeCd: gameTypeCd,
			RecentGames:      recentGames,
			ShowMoreLink:     len(recentGames) == 20,
			NextQueryStr:     nextQueryStr,
		}

		err = ae.templates["gameindex.page.html"].Execute(w, data)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}
	}
}

// GameInfoHandler retrieves information about a game by its ID.
func (ae *AppEnv) GameInfoHandler(w http.ResponseWriter, r *http.Request) {
	acceptHeader := r.Header.Get("Accept")

	gameID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("Invalid or missing game ID value: %s", err)
		ae.NotFoundHandler(w, r)
		return
	}

	if acceptHeader == "application/json" {
		// JSON response
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// w.Write()
	} else {
		// HTML response
		gameInfo, err := game.GameInfoData(ae.db, gameID)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}

		err = ae.templates["gameinfo.page.html"].Execute(w, gameInfo)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
			return
		}
	}
}

// Determine the primary background color for weapon charts by weapon code.
func weaponBackgroundColor(weaponCd string) string {
	weaponColors := map[string]string{
		"arc":             "#7C9CEB",
		"laser":           "#F7717B",
		"blaster":         "#F7717B",
		"shotgun":         "#849BA8",
		"uzi":             "#81F13D",
		"machinegun":      "#81F13D",
		"grenadelauncher": "#FD7865",
		"mortar":          "#FD7865",
		"minelayer":       "#FD7865",
		"electro":         "#6899F2",
		"crylink":         "#EA6FF9",
		"nex":             "#75C3D5",
		"vortex":          "#75C3D5",
		"hagar":           "#E39160",
		"rocketlauncher":  "#E9BE57",
		"devastator":      "#E9BE57",
		"porto":           "#6899F2",
		"minstanex":       "#978ED2",
		"vaporizer":       "#978ED2",
		"hook":            "#81F13D",
		"hlac":            "#E5965B",
		"seeker":          "#F7717B",
		"rifle":           "#E39160",
		"tuba":            "#E9BE57",
		"fireball":        "#F0855F",
	}

	color, ok := weaponColors[weaponCd]
	if ok {
		return color
	}

	return ""
}

// Determine the border color for a given weapon in charts
func weaponBorderColor(weaponCd string) string {
	borderColors := map[string]string{
		"arc":             "#7c9ceb",
		"laser":           "#f7717b",
		"blaster":         "#f7717b",
		"shotgun":         "#849ba8",
		"uzi":             "#81f13d",
		"machinegun":      "#81f13d",
		"grenadelauncher": "#fd7865",
		"mortar":          "#fd7865",
		"minelayer":       "#fd7865",
		"electro":         "#6899f2",
		"crylink":         "#ea6ff9",
		"nex":             "#75c3d5",
		"vortex":          "#75c3d5",
		"hagar":           "#e39160",
		"rocketlauncher":  "#e9be57",
		"devastator":      "#e9be57",
		"porto":           "#6899f2",
		"minstanex":       "#978ed2",
		"vaporizer":       "#978ed2",
		"hook":            "#81f13d",
		"hlac":            "#e5965b",
		"seeker":          "#f7717b",
		"rifle":           "#e39160",
		"tuba":            "#e9be57",
		"fireball":        "#f0855f",
	}

	color, ok := borderColors[weaponCd]
	if ok {
		return color
	}

	return ""
}

// DamageRichData is the more detailed set of information for a given slice of the stacked bar chart for damage.
type DamageRichData struct {
	PlayerID         int     `json:"player_id"`
	Nick             string  `json:"nick"`
	GameID           int     `json:"game_id"`
	PlayerGameStatID int     `json:"player_game_stat_id"`
	WeaponCd         string  `json:"weapon_cd"`
	WeaponCdInitCaps string  `json:"weapon_cd_init_caps"`
	Actual           int     `json:"actual"`
	Max              int     `json:"max"`
	PctTotalDamage   float32 `json:"pct_total_damage"`
	Frags            int     `json:"frags"`
}

// NewDamageRichData converts a WeaponInfoBase into a DamageRichData object or a blank entry.
func NewDamageRichData(weaponCd string, wi *game.WeaponInfoBase) *DamageRichData {
	if wi == nil {
		// Blank entry
		return &DamageRichData{
			WeaponCd: weaponCd,
		}
	}

	return &DamageRichData{
		PlayerID:         wi.PlayerID,
		Nick:             wi.Nick.NickStripped,
		GameID:           wi.GameID,
		PlayerGameStatID: wi.PlayerGameStatID,
		WeaponCd:         wi.WeaponCd,
		WeaponCdInitCaps: wi.WeaponCdInitCaps,
		Actual:           wi.Actual,
		Max:              wi.Max,
		PctTotalDamage:   wi.PctTotalDamage,
		Frags:            wi.Frags,
	}
}

// DamageDataset is damage data in the "shape" that chart.js wants.
type DamageDataset struct {
	Label           string            `json:"label"`
	BackgroundColor string            `json:"backgroundColor"`
	BorderColor     string            `json:"borderColor"`
	MaxBarThickness int               `json:"maxBarThickness"`
	RichData        []*DamageRichData `json:"richData"`
	Data            []int             `json:"data"`
}

// NewDamageDataset creates a new DamageDataSet from a weapon code.
func NewDamageDataset(weaponCd string) *DamageDataset {
	return &DamageDataset{
		Label:           weaponCd,
		BackgroundColor: weaponBackgroundColor(weaponCd),
		BorderColor:     weaponBorderColor(weaponCd),
		MaxBarThickness: 25,
		RichData:        make([]*DamageRichData, 0),
		Data:            make([]int, 0),
	}
}

// AccuracyRichData is the more detailed set of information for a given slice of the bar chart for accuracy.
type AccuracyRichData struct {
	PlayerID         int     `json:"player_id"`
	Nick             string  `json:"nick"`
	GameID           int     `json:"game_id"`
	PlayerGameStatID int     `json:"player_game_stat_id"`
	WeaponCd         string  `json:"weapon_cd"`
	WeaponCdInitCaps string  `json:"weapon_cd_init_caps"`
	Hit              int     `json:"hit"`
	Fired            int     `json:"fired"`
	PctAccuracy      float32 `json:"pct_accuracy"`
	Frags            int     `json:"frags"`
}

// NewAccuracyRichData converts a WeaponInfoBase into an AccuracyRichData object or a blank entry.
func NewAccuracyRichData(weaponCd string, wi *game.WeaponInfoBase) *AccuracyRichData {
	if wi == nil {
		// Blank entry
		return &AccuracyRichData{
			WeaponCd: weaponCd,
		}
	}

	return &AccuracyRichData{
		PlayerID:         wi.PlayerID,
		Nick:             wi.Nick.NickStripped,
		GameID:           wi.GameID,
		PlayerGameStatID: wi.PlayerGameStatID,
		WeaponCd:         wi.WeaponCd,
		WeaponCdInitCaps: wi.WeaponCdInitCaps,
		Hit:              wi.Hit,
		Fired:            wi.Fired,
		PctAccuracy:      wi.PctAccuracy,
		Frags:            wi.Frags,
	}
}

// AccuracyDataset is damage data in the "shape" that chart.js wants.
type AccuracyDataset struct {
	Label           string              `json:"label"`
	BackgroundColor string              `json:"backgroundColor"`
	BorderColor     string              `json:"borderColor"`
	MaxBarThickness int                 `json:"maxBarThickness"`
	RichData        []*AccuracyRichData `json:"richData"`
	Data            []float32           `json:"data"`
}

// NewAccuracyDataset creates a new AccuracyDataSet from a weapon code.
func NewAccuracyDataset(weaponCd string) *AccuracyDataset {
	return &AccuracyDataset{
		Label:           weaponCd,
		BackgroundColor: weaponBackgroundColor(weaponCd),
		BorderColor:     weaponBorderColor(weaponCd),
		MaxBarThickness: 25,
		RichData:        make([]*AccuracyRichData, 0),
		Data:            make([]float32, 0),
	}
}

// GameWeaponInfoResponse is the JSON response type for a game's weapon information.
type GameWeaponInfoResponse struct {
	GameID          int                `json:"game_id"`
	GameTypeCd      string             `json:"game_type_cd"`
	GameTypeDescr   string             `json:"game_type_descr"`
	Duration        string             `json:"duration"`
	Winner          int                `json:"winning_team"`
	MatchID         string             `json:"match_id"`
	Mod             string             `json:"mod"`
	DistinctWeapons []string           `json:"distinct_weapons"`
	DistinctPlayers []string           `json:"distinct_players"`
	DamageData      []*DamageDataset   `json:"damage_data"`
	AccuracyData    []*AccuracyDataset `json:"accuracy_data"`
}

// Should the weapon be included in the accuracy dataset?
func isAccuracyWeapon(weaponCd string) bool {
	switch weaponCd {
	case "arc", "assaultrifle", "huntingrifle", "machinegun", "okhmg", "okmachinegun":
		return true
	case "oknex", "rifle", "shockwave", "shotgun", "vaporizer", "vortex":
		return true
	}

	return false
}

// GameWeaponInfoHandler retrieves information about the weapons in a game by its ID.
func (ae *AppEnv) GameWeaponInfoHandler(w http.ResponseWriter, r *http.Request) {
	gameID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("Invalid or missing game ID value: %s", err)
		ae.NotFoundHandler(w, r)
		return
	}

	gameWeaponInfo, err := game.GameWeaponInfoData(ae.db, gameID)
	if err != nil {
		log.Printf("Could not process weapon info for game ID %d: %s", gameID, err)
		ae.NotFoundHandler(w, r)
		return
	}

	pgsIDOrder := make([]int, 0)                                    // the order we've seen the pgsIDs
	pgsIDToWeapons := make(map[int]map[string]*game.WeaponInfoBase) // game stat IDs to weapons
	nicks := make([]string, 0)                                      //  keep track of the stripped nicks of each player
	for _, wi := range gameWeaponInfo.WeaponInfo {
		pgsID := wi.PlayerGameStatID
		weaponCd := wi.WeaponCd
		if _, ok := pgsIDToWeapons[wi.PlayerGameStatID]; !ok {
			pgsIDToWeapons[pgsID] = make(map[string]*game.WeaponInfoBase)
			nicks = append(nicks, wi.Nick.NickStripped)
			pgsIDOrder = append(pgsIDOrder, pgsID)
		}

		pgsIDToWeapons[pgsID][weaponCd] = wi
	}

	// Datasets are added for each player, for each distinct weapon used in the match.
	damageData := make([]*DamageDataset, 0)
	for _, weaponCd := range gameWeaponInfo.DistinctWeapons {
		dmgDataset := NewDamageDataset(weaponCd)
		for _, pgsID := range pgsIDOrder {
			// Check if the player actually used that weapon.
			wi, usedWeapon := pgsIDToWeapons[pgsID][weaponCd]

			// Damage
			if usedWeapon {
				// convert wi and add to dataset
				dmgDataset.RichData = append(dmgDataset.RichData, NewDamageRichData(weaponCd, wi))
				dmgDataset.Data = append(dmgDataset.Data, wi.Actual)
			} else {
				// blank entry
				dmgDataset.RichData = append(dmgDataset.RichData, NewDamageRichData(weaponCd, nil))
				dmgDataset.Data = append(dmgDataset.Data, 0)
			}
		}
		damageData = append(damageData, dmgDataset)
	}

	// Accuracy datasets are also added for each player, for each distinct weapon used in the match.
	accData := make([]*AccuracyDataset, 0)
	for _, weaponCd := range gameWeaponInfo.DistinctWeapons {
		// Not all weapons are put into the accuracy charts.
		if !isAccuracyWeapon(weaponCd) {
			continue
		}

		accDataset := NewAccuracyDataset(weaponCd)
		for _, pgsID := range pgsIDOrder {
			// Check if the player actually used that weapon.
			wi, usedWeapon := pgsIDToWeapons[pgsID][weaponCd]

			// Accuracy
			if usedWeapon {
				// convert wi and add to dataset
				accDataset.RichData = append(accDataset.RichData, NewAccuracyRichData(wi.WeaponCd, wi))
				accDataset.Data = append(accDataset.Data, wi.PctAccuracy)
			} else {
				// blank entry
				accDataset.RichData = append(accDataset.RichData, NewAccuracyRichData(weaponCd, nil))
				accDataset.Data = append(accDataset.Data, 0.0)
			}
		}
		accData = append(accData, accDataset)
	}

	gameWeaponInfoResponse := GameWeaponInfoResponse{
		GameID:          gameWeaponInfo.GameID,
		GameTypeCd:      gameWeaponInfo.GameTypeCd,
		GameTypeDescr:   gameWeaponInfo.GameTypeDescr,
		Duration:        gameWeaponInfo.Duration.Long,
		Winner:          gameWeaponInfo.Winner,
		MatchID:         gameWeaponInfo.MatchID,
		Mod:             gameWeaponInfo.Mod,
		DistinctWeapons: gameWeaponInfo.DistinctWeapons,
		DistinctPlayers: nicks,
		DamageData:      damageData,
		AccuracyData:    accData,
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	bytes, err := json.Marshal(gameWeaponInfoResponse)
	if err != nil {
		log.Printf("Could not marshal weapon info to JSON for game ID %d: %s", gameID, err)
		ae.NotFoundHandler(w, r)
		return
	}

	w.Write(bytes)
}
