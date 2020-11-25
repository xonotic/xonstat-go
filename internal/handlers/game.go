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
		http.Error(w, fmt.Sprintf("404 %s", http.StatusText(404)), 404)
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
	} else {
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

// Determine the background color for a given weapon in charts
func weaponBackgroundColor(weaponCd string) string {
	weaponColors := map[string]string{
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

	color, ok := weaponColors[weaponCd]
	if ok {
		return color
	}

	return ""
}

// NewDamageDataset creates a new DamageDataSet from a weapon code.
func NewDamageDataset(weaponCd string) *DamageDataset {
	return &DamageDataset{
		Label:           weaponCd,
		BackgroundColor: weaponBackgroundColor(weaponCd),
		BorderColor:     "", // TODO: calculate border color based on weapon
		MaxBarThickness: 25,
		RichData:        make([]*DamageRichData, 0),
		Data:            make([]int, 0),
	}
}

// GameWeaponInfoResponse is the JSON response type for a game's weapon information.
type GameWeaponInfoResponse struct {
	GameID          int              `json:"game_id"`
	GameTypeCd      string           `json:"game_type_cd"`
	GameTypeDescr   string           `json:"game_type_descr"`
	Duration        string           `json:"duration"`
	Winner          int              `json:"winning_team"`
	MatchID         string           `json:"match_id"`
	Mod             string           `json:"mod"`
	DistinctWeapons []string         `json:"distinct_weapons"`
	DistinctPlayers []string         `json:"distinct_players"`
	DamageData      []*DamageDataset `json:"damage_data"`
}

// GameWeaponInfoHandler retrieves information about the weapons in a game by its ID.
func (ae *AppEnv) GameWeaponInfoHandler(w http.ResponseWriter, r *http.Request) {
	gameID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("Invalid or missing game ID value: %s", err)
		http.Error(w, fmt.Sprintf("404 %s", http.StatusText(404)), 404)
		return
	}

	gameWeaponInfo, err := game.GameWeaponInfoData(ae.db, gameID)
	if err != nil {
		log.Printf("Could not process weapon info for game ID %d: %s", gameID, err)
		http.Error(w, fmt.Sprintf("404 %s", http.StatusText(404)), 404)
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

	// A damage dataset is added for each player, for each distinct weapon used in the match.
	damageData := make([]*DamageDataset, 0)
	for _, weaponCd := range gameWeaponInfo.DistinctWeapons {
		dmgDataset := NewDamageDataset(weaponCd)
		for _, pgsID := range pgsIDOrder {
			wi, ok := pgsIDToWeapons[pgsID][weaponCd]
			if ok {
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
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	bytes, err := json.Marshal(gameWeaponInfoResponse)
	if err != nil {
		log.Printf("Could not marshal weapon info to JSON for game ID %d: %s", gameID, err)
		http.Error(w, fmt.Sprintf("404 %s", http.StatusText(404)), 404)
		return
	}

	w.Write(bytes)
}
