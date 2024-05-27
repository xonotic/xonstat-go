package submission

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/antzucaro/qstr"
	"gitlab.com/xonotic/xonstat/pkg/models"
)

// Submission is a fully-formatted statistics POST request
type Submission struct {
	Game                 *models.Game
	Server               *models.Server
	Map                  *models.Map
	Players              []*models.Player
	NonParticipants      []*models.PlayerGameNonParticipant
	PlayerGameStats      []*models.PlayerGameStat
	PlayerWeaponStats    []*models.PlayerWeaponStat
	TeamGameStats        []*models.TeamGameStat
	PlayerGameAnticheats []models.PlayerGameAnticheat
	CreateDt             time.Time

	// References by player index (initially the player events 'i' or index value) for easier processing
	PlayersByIndex    map[int]*models.Player
	HashkeysByIndex   map[int]string
	FragMatrixByIndex map[int]map[int]int

	// References by hashkey for easier processing
	PlayersByHashkey           map[string]*models.Player
	PlayerGameStatsByHashkey   map[string]*models.PlayerGameStat
	NonParticipantsByHashkey   map[string]*models.PlayerGameNonParticipant
	PlayerWeaponStatsByHashkey map[string][]*models.PlayerWeaponStat
}

// gameCategory determines the game's "category" field
func gameCategory(rs *RawSubmission) string {
	// allowed weapons in each of the various categories
	vanillaAllowedWeapons := map[string]struct{}{
		"shotgun":    struct{}{},
		"devastator": struct{}{},
		"blaster":    struct{}{},
		"mortar":     struct{}{},
		"vortex":     struct{}{},
		"electro":    struct{}{},
		"arc":        struct{}{},
		"hagar":      struct{}{},
		"crylink":    struct{}{},
		"machinegun": struct{}{},
	}

	instaAllowedWeapons := map[string]struct{}{
		"vaporizer": struct{}{},
		"blaster":   struct{}{},
	}

	overkillAllowedWeapons := map[string]struct{}{
		"okhmg":        struct{}{},
		"oknex":        struct{}{},
		"okshotgun":    struct{}{},
		"okmachinegun": struct{}{},
		"okrpc":        struct{}{},
		"blaster":      struct{}{},
	}

	// for each category, have we seen all allowed weapons?
	vanillaOK := true
	instaOK := true
	overkillOK := true

	// loop through the weapons fired to see if any disallowed weapons were fired for each category
	for weapon := range rs.WeaponsUsed {
		if _, ok := vanillaAllowedWeapons[weapon]; !ok {
			vanillaOK = false
		}

		if _, ok := instaAllowedWeapons[weapon]; !ok {
			instaOK = false
		}

		if _, ok := overkillAllowedWeapons[weapon]; !ok {
			overkillOK = false
		}
	}

	var mod string
	modVal, ok := rs.GameMeta["O"]
	if ok {
		mod = modVal
	} else {
		mod = "Xonotic"
	}

	if mod == "Xonotic" {
		if vanillaOK {
			return "vanilla"
		}
	} else if mod == "InstaGib" {
		if instaOK {
			return "insta"
		}
	} else if mod == "Overkill" {
		if overkillOK {
			return "overkill"
		}
	} else {
		return "general"
	}

	return "general"
}

// fillGame fills in the Game attribute from the raw submission
func (s *Submission) fillGame(rs *RawSubmission) error {
	if gameTypeCd, ok := rs.GameMeta["G"]; ok {
		s.Game.GameTypeCd = gameTypeCd
	} else {
		return ErrInvalidGameMeta
	}

	if durationSecsStr, ok := rs.GameMeta["D"]; ok {
		s.Game.Duration = durationFromString(durationSecsStr, 1.0)
	} else {
		return ErrInvalidGameMeta
	}

	if matchID, ok := rs.GameMeta["I"]; ok {
		s.Game.MatchID = sql.NullString{Valid: true, String: matchID}
	}

	if mod, ok := rs.GameMeta["O"]; ok {
		s.Game.Mod = sql.NullString{Valid: true, String: mod}
	}

	// Category is not supported yet.
	// category := gameCategory(rs)
	// s.Game.Category = &category

	s.Game.StartDt = s.CreateDt
	s.Game.CreateDt = s.CreateDt

	return nil
}

// fillServer fills in the Server attribute from the raw submission
func (s *Submission) fillServer(rs *RawSubmission) error {
	s.Server.ActiveInd = true

	if serverName, ok := rs.GameMeta["S"]; ok {
		if len(serverName) > 64 {
			serverName = serverName[:64]
		}

		s.Server.Name = sql.NullString{String: serverName, Valid: true}
	} else {
		return ErrInvalidGameMeta
	}

	if portStr, ok := rs.GameMeta["U"]; ok {
		if port, err := strconv.Atoi(portStr); err == nil {
			s.Server.Port = sql.NullInt64{Int64: int64(port), Valid: true}
		}
	}

	if impureCvarsStr, ok := rs.GameMeta["C"]; ok {
		if impureCvars, err := strconv.Atoi(impureCvarsStr); err == nil {
			s.Server.ImpureCvars = sql.NullInt64{Int64: int64(impureCvars), Valid: true}
		}

		if s.Server.ImpureCvars.Valid && s.Server.ImpureCvars.Int64 == 0 {
			s.Server.PureInd = true
		}
	}

	if revision, ok := rs.GameMeta["R"]; ok {
		s.Server.Revision = sql.NullString{String: revision, Valid: true}
	}

	s.Server.CreateDt = s.CreateDt

	return nil
}

// fillMap fills in the Map attribute from the raw submission
func (s *Submission) fillMap(rs *RawSubmission) error {
	if mapName, ok := rs.GameMeta["M"]; ok {
		s.Map.Name = mapName
	} else {
		return ErrInvalidGameMeta
	}

	s.Map.CreateDt = s.CreateDt

	return nil
}

// fillTeamStats fills in the stats attributable to teams
func (s *Submission) fillTeamStats(rs *RawSubmission) error {
	// helper function to return a NULLable SQL int from a rounded float
	var intFromFloat = func(value string) sql.NullInt32 {
		floatVal, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return sql.NullInt32{Valid: false}
		}
		rounded := int(math.Round(floatVal))

		return sql.NullInt32{Valid: true, Int32: int32(rounded)}
	}

	for _, events := range rs.TeamEvents {
		tgs := models.NewTeamGameStat(s.Game.GameTypeCd)
		tgs.GameID = s.Game.GameID
		tgs.CreateDt = s.Game.CreateDt

		team, err := strconv.Atoi(strings.Split(events["Q"], "#")[1])
		if err != nil {
			return err
		}
		tgs.Team = team

		for key, value := range events {
			switch key {
			case "scoreboard-score":
				tgs.Score = intFromFloat(value)
			case "scoreboard-caps", "scoreboard-goals":
				tgs.Caps = intFromString(value)
			case "scoreboard-rounds":
				tgs.Rounds = intFromString(value)
			}
		}

		s.TeamGameStats = append(s.TeamGameStats, tgs)
	}

	return nil
}

// intFromStringDefault converts a string to an int if possible, and if not returns a default value
func intFromStringDefault(value string, defaultVal int) int {
	intVal, err := strconv.Atoi(value)
	if err != nil {
		return defaultVal
	}
	return intVal
}

// intFromString converts a string to a NULLable SQL int
func intFromString(value string) sql.NullInt32 {
	intVal, err := strconv.Atoi(value)
	if err != nil {
		return sql.NullInt32{Valid: false}
	}
	return sql.NullInt32{Valid: true, Int32: int32(intVal)}
}

// floatFromString converts a string to a NULLable SQL float
func floatFromString(value string) sql.NullFloat64 {
	floatVal, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return sql.NullFloat64{Valid: false}
	}
	return sql.NullFloat64{Valid: true, Float64: floatVal}
}

// durationFromString converts a string representing some multiple of seconds to a duration.
// Adjust the divisor argument to account for the scale of the raw value (sometimes raw values
// are reported in hundredths of seconds, etc). A divisor of 1.0 is if the input string represents
// a value in seconds.
func durationFromString(value string, divisor float64) *time.Duration {
	floatVal, err := strconv.ParseFloat(value, 32)
	if err != nil {
		return nil
	}

	seconds := floatVal / divisor
	duration, err := time.ParseDuration(fmt.Sprintf("%fs", seconds))
	if err != nil {
		return nil
	}

	return &duration
}

// fillPlayerWeaponStat populates a PlayerWeaponStat object from the events in the events map
func (s *Submission) fillPlayerWeaponStat(weapon string, events map[string]string, player *models.Player) error {
	var ws models.PlayerWeaponStat
	hashkey := events["P"]
	ws.WeaponCd = weapon
	ws.CreateDt = s.Game.CreateDt

	// helper function to pull weapon stat values, rounded from the floats they might be
	var intFromFloat = func(key string) int {
		if s, ok := events[key]; ok {
			val, err := strconv.ParseFloat(s, 32)
			if err != nil {
				return 0
			}
			return int(math.Round(val))
		}
		return 0
	}

	ws.Fired = intFromFloat(fmt.Sprintf("acc-%s-cnt-fired", weapon))
	ws.Hit = intFromFloat(fmt.Sprintf("acc-%s-cnt-hit", weapon))
	ws.Max = intFromFloat(fmt.Sprintf("acc-%s-fired", weapon))
	ws.Actual = intFromFloat(fmt.Sprintf("acc-%s-hit", weapon))
	ws.Frags = intFromFloat(fmt.Sprintf("acc-%s-frags", weapon))

	s.PlayerWeaponStats = append(s.PlayerWeaponStats, &ws)
	s.PlayerWeaponStatsByHashkey[hashkey] = append(s.PlayerWeaponStatsByHashkey[hashkey], &ws)

	return nil
}

// parseFragMatrix parses the opponent's player index and the number of frags this player
// has made against them. For example: "kills-3 5" means that the current player (index unknown
// to this function) has fragged the player identified by index value 3 a count of 5 times.
func parseFragMatrix(key, value string) (int, int) {
	opponentIndexStr := strings.Replace(key, "kills-", "", 1)
	opponentIndex, err := strconv.Atoi(opponentIndexStr)
	if err != nil {
		return -1, -1
	}

	fragsAgainstOpponent, err := strconv.Atoi(value)
	if err != nil {
		return -1, -1
	}

	return opponentIndex, fragsAgainstOpponent
}

// fillPlayerGameStat fills in a single PlayerGameStat struct from the raw submission events
func (s *Submission) fillPlayerGameStat(events map[string]string, player *models.Player) error {
	// an initialized pgstat based on the game type being played
	pgs := models.NewPlayerGameStat(s.Game.GameTypeCd)

	hashkey := events["P"]

	// fields passed on from other objects
	pgs.PlayerID = player.PlayerID
	pgs.PlayerGameStatID = player.PlayerID
	pgs.GameID = s.Game.GameID
	pgs.CreateDt = s.CreateDt
	pgs.Nick = player.Nick
	pgs.StrippedNick = player.StrippedNick

	// required fields
	score := 0
	if scoreStr, ok := events["scoreboard-score"]; ok {
		scoreFloat, err := strconv.ParseFloat(scoreStr, 32)
		if err == nil {
			score = int(math.Round(scoreFloat))
		}
	}
	pgs.Score = sql.NullInt32{Valid: true, Int32: int32(score)}

	if alivetimeStr, ok := events["alivetime"]; ok {
		pgs.AliveTime = durationFromString(alivetimeStr, 1.0)
	}

	if rankStr, ok := events["rank"]; ok {
		pgs.Rank = sql.NullInt32{Valid: true, Int32: int32(intFromStringDefault(rankStr, 0))}
	}

	if scoreboardPosStr, ok := events["scoreboardpos"]; ok {
		pgs.ScoreboardPos = sql.NullInt32{
			Valid: true,
			Int32: int32(intFromStringDefault(scoreboardPosStr, 0)),
		}
	}

	wins := false

	playerIndex, _ := strconv.Atoi(events["i"])

	for key, value := range events {
		switch key {
		case "wins":
			wins = true
		case "t":
			pgs.Team = intFromString(value)
		case "scoreboard-drops", "scoreboard-released", "scoreboard-ticks", "scoreboard-losses":
			pgs.Drops = intFromString(value)
		case "scoreboard-returns":
			pgs.Returns = intFromString(value)
		case "scoreboard-fckills", "scoreboard-bckills", "scoreboard-kckills":
			pgs.CarrierFrags = intFromString(value)
		case "scoreboard-pickups", "scoreboard-takes":
			pgs.Pickups = intFromString(value)
		case "scoreboard-caps", "scoreboard-captured", "scoreboard-goals":
			pgs.Captures = intFromString(value)
		case "scoreboard-deaths":
			pgs.Deaths = intFromString(value)
		case "scoreboard-kills":
			pgs.Kills = intFromString(value)
		case "scoreboard-suicides":
			pgs.Suicides = intFromString(value)
		case "scoreboard-objectives":
			pgs.Collects = intFromString(value)
		case "scoreboard-fastest", "scoreboard-captime":
			pgs.Fastest = durationFromString(value, 100.0)
		case "total-fastest":
			// We allow CTS to log fastest laps even if they've specced. In these
			// cases they would only have a "total" entry and not a "scoreboard"
			// entry since they wouldn't show up on the in-game scoreboard.
			if pgs.Fastest == nil {
				pgs.Fastest = durationFromString(value, 100.0)
			}
		case "scoreboard-revivals":
			pgs.Revivals = intFromString(value)
		case "scoreboard-bctime":
			pgs.Time = durationFromString(value, 1.0)
		case "scoreboard-pushes":
			pgs.Pushes = intFromString(value)
		case "scoreboard-destroyed":
			pgs.Destroys = intFromString(value)
		case "scoreboard-lives":
			pgs.Lives = intFromString(value)
		case "scoreboard-faults":
			pgs.Drops = intFromString(value)
		case "scoreboard-laps":
			pgs.Laps = intFromString(value)
		case "avglatency":
			pgs.AvgLatency = floatFromString(value)
		case "scoreboard-dmg":
			// TODO: database field and parsing
		case "scoreboard-dmgtaken":
			// TODO: database field and parsing
		case "scoreboard-fps":
			// TODO: database field and parsing? Not sure if we want this saved.
		}

		if strings.HasSuffix(key, "cnt-fired") {
			weapon := weaponFromKey(key)
			s.fillPlayerWeaponStat(weapon, events, player)
		}

		if strings.HasPrefix(key, "anticheat") {
			floatVal, _ := strconv.ParseFloat(value, 64)
			ac := models.PlayerGameAnticheat{
				PlayerID: pgs.PlayerID,
				GameID:   pgs.GameID,
				Key:      key,
				Value:    floatVal,
				CreateDt: pgs.CreateDt,
			}

			s.PlayerGameAnticheats = append(s.PlayerGameAnticheats, ac)
		}

		if strings.HasPrefix(key, "kills-") {
			opponentIndex, fragsAgainstOpponent := parseFragMatrix(key, value)
			s.FragMatrixByIndex[playerIndex][opponentIndex] = fragsAgainstOpponent
		}
	}

	// there is no "winning team" field, so we derive it
	if wins && pgs.Team.Valid {
		s.Game.Winner = sql.NullInt64{Valid: true, Int64: int64(pgs.Team.Int32)}
	}

	s.PlayerGameStats = append(s.PlayerGameStats, pgs)
	s.PlayerGameStatsByHashkey[hashkey] = pgs
	s.HashkeysByIndex[playerIndex] = hashkey

	return nil
}

// fillNonParticipants fills in a single struct from the raw submission events
func (s *Submission) fillNonParticipants(events map[string]string, player *models.Player) error {
	// an initialized pgstat based on the game type being played
	var pgnp models.PlayerGameNonParticipant

	hashkey := events["P"]
	playerIndex, _ := strconv.Atoi(events["i"])

	// fields passed on from other objects
	pgnp.PlayerID = player.PlayerID
	pgnp.PlayerGameNonParticipantID = player.PlayerID
	pgnp.GameID = s.Game.GameID
	pgnp.CreateDt = s.CreateDt
	pgnp.Nick = player.Nick
	pgnp.StrippedNick = player.StrippedNick

	// required fields
	score := 0
	if scoreStr, ok := events["total-score"]; ok {
		scoreFloat, err := strconv.ParseFloat(scoreStr, 32)
		if err == nil {
			score = int(math.Round(scoreFloat))
		}
	}
	pgnp.Score = sql.NullInt32{Valid: true, Int32: int32(score)}

	zero := time.Duration(0 * time.Second)
	if alivetimeStr, ok := events["alivetime"]; ok {
		pgnp.AliveTime = durationFromString(alivetimeStr, 1.0)
	} else {
		pgnp.AliveTime = &zero
	}

	// Determine spectator or not purely by alivetime
	if *pgnp.AliveTime > zero {
		pgnp.Status = "forfeit"
	} else {
		pgnp.Status = "spectator"
	}

	s.NonParticipants = append(s.NonParticipants, &pgnp)
	s.NonParticipantsByHashkey[hashkey] = &pgnp
	s.HashkeysByIndex[playerIndex] = hashkey

	return nil
}

// fillPlayers fills in the Players and PlayerHashKeys slices from the raw submission
func (s *Submission) fillPlayers(rs *RawSubmission) error {
	// Keep track of the non-participants. They need to have player and hashkey records
	// filled, but not anything else.
	nonParticipants := make(map[string]struct{})
	for _, events := range rs.NonParticipants {
		nonParticipants[events["P"]] = struct{}{}
	}

	// Humans, Bots, and Non-Participants
	playersInGame := append(rs.Humans, rs.Bots...)
	playersInGame = append(playersInGame, rs.NonParticipants...)

	for _, events := range playersInGame {
		hashkey := events["P"]

		nick := "Anonymous Player"
		if nickStr, ok := events["n"]; ok {
			if len(nickStr) > 128 {
				nick = nickStr[:128]
			} else {
				nick = nickStr
			}
		}

		nickQStr := qstr.QStr(nick)
		strippedNick := nickQStr.Stripped()

		playerIndex, err := strconv.Atoi(events["i"])
		if err != nil {
			playerIndex = -1
		}

		player := models.Player{
			PlayerID:     playerIndex,
			Nick:         sql.NullString{Valid: true, String: nick},
			StrippedNick: sql.NullString{Valid: true, String: strippedNick},
			ActiveInd:    true,
			CreateDt:     s.CreateDt,
		}

		s.Players = append(s.Players, &player)
		s.PlayersByHashkey[hashkey] = &player
		s.PlayersByIndex[playerIndex] = &player

		// Do not fill out the other records for non-participants.
		if _, ok := nonParticipants[hashkey]; ok {
			s.fillNonParticipants(events, &player)
		} else {
			s.FragMatrixByIndex[playerIndex] = make(map[int]int)
			s.fillPlayerGameStat(events, &player)
		}
	}

	return nil
}

// NewSubmission converts a RawSubmission into a fully-formed one
func NewSubmission(rs *RawSubmission) (*Submission, error) {
	var game models.Game
	var server models.Server
	var _map models.Map
	players := make([]*models.Player, 0, len(rs.PlayerEvents))
	nonParticipants := make([]*models.PlayerGameNonParticipant, 0, len(rs.PlayerEvents))
	playerGameStats := make([]*models.PlayerGameStat, 0, len(rs.PlayerEvents))
	var playerWeaponStats []*models.PlayerWeaponStat
	var teamGameStats []*models.TeamGameStat
	var playerGameAnticheats []models.PlayerGameAnticheat
	playersByIndex := make(map[int]*models.Player, 0)
	hashkeysByIndex := make(map[int]string)
	fragMatrixByIndex := make(map[int]map[int]int, 0)
	playerWeaponStatsByHashkey := make(map[string][]*models.PlayerWeaponStat, 0)
	playersByHashkey := make(map[string]*models.Player, 0)
	playerGameStatsByHashkey := make(map[string]*models.PlayerGameStat, 0)
	nonParticipantsByHashkey := make(map[string]*models.PlayerGameNonParticipant, 0)

	s := &Submission{
		Game:                       &game,
		Server:                     &server,
		Map:                        &_map,
		Players:                    players,
		NonParticipants:            nonParticipants,
		PlayerGameStats:            playerGameStats,
		PlayerWeaponStats:          playerWeaponStats,
		TeamGameStats:              teamGameStats,
		PlayerGameAnticheats:       playerGameAnticheats,
		CreateDt:                   time.Now().UTC(),
		PlayersByIndex:             playersByIndex,
		HashkeysByIndex:            hashkeysByIndex,
		FragMatrixByIndex:          fragMatrixByIndex,
		PlayersByHashkey:           playersByHashkey,
		PlayerGameStatsByHashkey:   playerGameStatsByHashkey,
		NonParticipantsByHashkey:   nonParticipantsByHashkey,
		PlayerWeaponStatsByHashkey: playerWeaponStatsByHashkey,
	}

	// one at a time, we fill the members
	err := s.fillGame(rs)
	if err != nil {
		return nil, err
	}

	// This helps with debugging requests that error out.
	log.Printf("Processing match %s", s.Game.MatchID.String)

	err = s.fillServer(rs)
	if err != nil {
		return nil, err
	}

	err = s.fillMap(rs)
	if err != nil {
		return nil, err
	}

	err = s.fillTeamStats(rs)
	if err != nil {
		return nil, err
	}

	err = s.fillPlayers(rs)
	if err != nil {
		return nil, err
	}

	return s, nil
}
