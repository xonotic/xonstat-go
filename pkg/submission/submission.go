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
	PlayerGameStats      []*models.PlayerGameStat
	PlayerWeaponStats    []*models.PlayerWeaponStat
	TeamGameStats        []*models.TeamGameStat
	PlayerGameAnticheats []models.PlayerGameAnticheat
	CreateDt             time.Time

	// References by player index (initially the player events 'i' or index value) for easier processing
	PlayersByIndex map[int]*models.Player

	// References by hashkey for easier processing
	PlayersByHashkey           map[string]*models.Player
	PlayerGameStatsByHashkey   map[string]*models.PlayerGameStat
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
			// TODO: if the game type is ctf, do fastest cap processing
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
	}

	// there is no "winning team" field, so we derive it
	if wins && pgs.Team.Valid {
		s.Game.Winner = sql.NullInt64{Valid: true, Int64: int64(pgs.Team.Int32)}
	}

	s.PlayerGameStats = append(s.PlayerGameStats, pgs)
	s.PlayerGameStatsByHashkey[hashkey] = pgs

	return nil
}

// fillPlayers fills in the Players and PlayerHashKeys slices from the raw submission
func (s *Submission) fillPlayers(rs *RawSubmission) error {
	// Only consider events from humans or bots who actually played in the game.
	playersInGame := append(rs.Humans, rs.Bots...)

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
		s.PlayersByIndex[playerIndex] = &player
		s.PlayersByHashkey[hashkey] = &player

		s.fillPlayerGameStat(events, &player)
	}

	return nil
}

// NewSubmission converts a RawSubmission into a fully-formed one
func NewSubmission(rs *RawSubmission) (*Submission, error) {
	var game models.Game
	var server models.Server
	var _map models.Map
	players := make([]*models.Player, 0, len(rs.PlayerEvents))
	playerGameStats := make([]*models.PlayerGameStat, 0, len(rs.PlayerEvents))
	var playerWeaponStats []*models.PlayerWeaponStat
	var teamGameStats []*models.TeamGameStat
	var playerGameAnticheats []models.PlayerGameAnticheat
	playersByID := make(map[int]*models.Player, 0)
	playerWeaponStatsByHashkey := make(map[string][]*models.PlayerWeaponStat, 0)
	playersByHashkey := make(map[string]*models.Player, 0)
	playerGameStatsByHashkey := make(map[string]*models.PlayerGameStat, 0)

	s := &Submission{
		Game:                       &game,
		Server:                     &server,
		Map:                        &_map,
		Players:                    players,
		PlayerGameStats:            playerGameStats,
		PlayerWeaponStats:          playerWeaponStats,
		TeamGameStats:              teamGameStats,
		PlayerGameAnticheats:       playerGameAnticheats,
		CreateDt:                   time.Now().UTC(),
		PlayersByIndex:             playersByID,
		PlayersByHashkey:           playersByHashkey,
		PlayerGameStatsByHashkey:   playerGameStatsByHashkey,
		PlayerWeaponStatsByHashkey: playerWeaponStatsByHashkey,
	}

	// one at a time, we fill the members
	err := s.fillGame(rs)
	if err != nil {
		return nil, err
	}

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

// ShouldUpdateServer determines if the database server record needs to be updated with
// new information coming from the submission.
func ShouldUpdateServer(incoming, existing *models.Server) bool {
	if incoming.Name.Valid && incoming.Name.String != existing.Name.String {
		return true
	}

	if incoming.HashKey.Valid && incoming.HashKey.String != existing.HashKey.String {
		return true
	}

	if incoming.IPAddr.Valid && incoming.IPAddr.String != existing.IPAddr.String {
		return true
	}

	if incoming.Port.Valid && incoming.Port.Int64 != existing.Port.Int64 {
		return true
	}

	if incoming.Revision.Valid && incoming.Revision.String != existing.Revision.String {
		return true
	}

	if incoming.ImpureCvars.Valid && incoming.ImpureCvars.Int64 != existing.ImpureCvars.Int64 {
		return true
	}

	return false
}

// GetOrCreateServer finds an existing server matching the one provided or constructs a new one.
func GetOrCreateServer(tx *sql.Tx, db models.Datastore, rawServer *models.Server) (*models.Server, error) {
	var servers []*models.Server
	var err error

	if rawServer.HashKey.Valid {
		log.Printf("Looking for server by hashkey '%s'", rawServer.HashKey.String)
		servers, err = db.RServersByHashkey(rawServer.HashKey.String)
		if err != nil {
			return nil, err
		}
	}

	// Fall back to searching by name if hashkey is not provided.
	if len(servers) == 0 && rawServer.Name.Valid {
		log.Printf("Looking for server by name '%s'", rawServer.Name.String)
		servers, err = db.RServersByName(rawServer.Name.String)
		if err != nil {
			return nil, err
		}
	}

	if len(servers) == 0 {
		// We haven't found a matching server. Create one.
		serverID, err := db.CServer(tx, *rawServer)
		if err != nil {
			return nil, err
		}
		rawServer.ServerID = int(serverID)
		log.Printf("Created new server %d.", serverID)
		return rawServer, nil
	}

	if len(servers) == 1 {
		log.Printf("Found matching server %d.", servers[0].ServerID)
	} else {
		log.Printf("Multiple matching servers found. Using the first one (%d).", servers[0].ServerID)
	}
	rawServer.ServerID = servers[0].ServerID

	if ShouldUpdateServer(rawServer, servers[0]) {
		log.Printf("Updating server %d.", rawServer.ServerID)
		err := db.UServer(tx, *rawServer)
		if err != nil {
			return nil, err
		}
	}

	// Do not proceed if a server is inactive (banned, broken, etc).
	if !servers[0].ActiveInd {
		return nil, fmt.Errorf("server %d is inactive", servers[0].ServerID)
	}

	return servers[0], nil
}

// GetOrCreateMap finds an existing map matching the one provided or constructs a new one.
func GetOrCreateMap(tx *sql.Tx, db models.Datastore, rawMap *models.Map) (*models.Map, error) {
	var maps []*models.Map
	var err error

	log.Printf("Looking for map by name '%s'", rawMap.Name)
	maps, err = db.RMapsByName(rawMap.Name)
	if err != nil {
		return nil, err
	}

	if len(maps) == 0 {
		// We haven't found a matching map. Create one.
		mapID, err := db.CMap(tx, *rawMap)
		if err != nil {
			return nil, err
		}
		rawMap.MapID = int(mapID)
		log.Printf("Created new map %d.", mapID)
		return rawMap, nil
	}

	if len(maps) == 1 {
		log.Printf("Found matching map %d.", maps[0].MapID)
	} else {
		log.Printf("Multiple matching maps found. Using the first one (%d).", maps[0].MapID)
	}
	rawMap.MapID = maps[0].MapID

	return maps[0], nil
}

// CreateGame creates a game record in the database, first checking if it exists using the MatchID.
// We expect a game to be inserted upon each submission, so this method only returns an error.
func CreateGame(tx *sql.Tx, db models.Datastore, s *Submission) error {
	if s.Game.MatchID.Valid {
		games, err := db.RGamesByMatchID(s.Game.MatchID.String)
		if err != nil {
			return err
		}

		if len(games) > 0 {
			log.Printf("A game with match_id %s already exists in the database.", s.Game.MatchID.String)
			return fmt.Errorf("duplicate game found via match_id")
		}
	}

	// For easier queries later, we store the PIDs right on the game entry as an array.
	var humansInGame []int
	for _, player := range s.Players {
		if player.PlayerID > 2 {
			humansInGame = append(humansInGame, player.PlayerID)
		}
	}
	s.Game.Players = humansInGame

	gameID, err := db.CGame(tx, *s.Game)
	if err != nil {
		return err
	}
	s.Game.GameID = int(gameID)
	log.Printf("Created game %d.", gameID)

	// Update the next record along the way.
	for _, pgs := range s.PlayerGameStats {
		pgs.GameID = int(gameID)
	}

	return nil
}

// ShouldUpdatePlayer determines if the incoming data has a new piece of information
// that should be persisted to the database with an update.
func ShouldUpdatePlayer(incoming, existing *models.Player) bool {
	if incoming.Nick.Valid && incoming.Nick.String != existing.Nick.String {
		return true
	}
	// TODO: register a nick change, if that is something we still want...
	return false
}

// GetOrCreatePlayers fetches existing players or creates new ones based upon the data
// in the submission. This one is done in batch to reduce SQL calls.
func GetOrCreatePlayers(tx *sql.Tx, db models.Datastore, s *Submission) (map[string]*models.Player, error) {
	// These records are fixed for bots and anons (untracked players) respectively
	bot := models.Player{PlayerID: 1, Nick: sql.NullString{Valid: true, String: "bot"}}
	anon := models.Player{PlayerID: 2, Nick: sql.NullString{Valid: true, String: "Anonymous Player"}}

	// This is the final return value. All player records found or created in the database.
	playersByHashkey := make(map[string]*models.Player)

	var hashkeys []string                      // used for searching
	hashkeySet := make(map[string]struct{}, 0) // used for keeping track of who we haven't processed

	// Bots and untracked players need no fetches from the database.
	for hashkey := range s.PlayersByHashkey {
		if strings.HasPrefix(hashkey, "bot#") {
			// bot
			playersByHashkey[hashkey] = &bot
		} else if strings.HasPrefix(hashkey, "player#") {
			// untracked player
			playersByHashkey[hashkey] = &anon
		} else {
			// human that we need to look for/create
			hashkeys = append(hashkeys, hashkey)
			hashkeySet[hashkey] = struct{}{}
		}
	}

	// Players that already exist are more complicated. We first fetch who we can,
	// then update them if need be.
	playersByHashkeyDB, err := db.RPlayersByHashkeyMulti(hashkeys)
	if err != nil {
		return nil, err
	}

	for hashkey, dbPlayer := range playersByHashkeyDB {
		rawPlayer := s.PlayersByHashkey[hashkey]
		if ShouldUpdatePlayer(rawPlayer, dbPlayer) {
			// Apply the update to the DB value...
			dbPlayer.Nick = rawPlayer.Nick
			dbPlayer.StrippedNick = rawPlayer.StrippedNick

			// ...and save it to the database.
			db.UPlayer(tx, *dbPlayer)
		}

		playersByHashkey[hashkey] = dbPlayer

		delete(hashkeySet, hashkey) // done processing this one
	}

	// The remaining players left in hashkeySet need to be created.
	for hashkey := range hashkeySet {
		newPlayer := s.PlayersByHashkey[hashkey]

		playerID, err := db.CPlayer(tx, *newPlayer)
		if err != nil {
			return nil, err
		}

		err = db.CHashkey(tx, models.PlayerHashkey{Hashkey: hashkey, PlayerID: int(playerID)})
		if err != nil {
			return nil, err
		}

		newPlayer.PlayerID = int(playerID)
		playersByHashkey[hashkey] = newPlayer

		log.Printf("Created player %d '%s'", playerID, newPlayer.StrippedNick.String)
	}

	// Reflect the new PIDs in the submission accordingly (at least for the next table being modified).
	for hashkey, player := range playersByHashkey {
		*s.PlayersByHashkey[hashkey] = *player
		s.PlayerGameStatsByHashkey[hashkey].PlayerID = player.PlayerID
	}

	return playersByHashkey, nil
}

// CreatePlayerGameStats inserts all of the game stat records to the database.
func CreatePlayerGameStats(tx *sql.Tx, db models.Datastore, s *Submission) error {
	for _, pgs := range s.PlayerGameStats {
		pgsID, err := db.CPlayerGameStat(tx, *pgs)
		if err != nil {
			return err
		}
		pgs.PlayerGameStatID = int(pgsID)
	}

	return nil
}

// CreatePlayerWeaponStats inserts all of the weapon stat records to the database.
func CreatePlayerWeaponStats(tx *sql.Tx, db models.Datastore, s *Submission) error {
	for hashkey, pwsList := range s.PlayerWeaponStatsByHashkey {
		for _, pws := range pwsList {
			pws.PlayerID = s.PlayersByHashkey[hashkey].PlayerID

			// We don't store weapon information for bots.
			if pws.PlayerID == 1 {
				break
			}

			pws.GameID = s.Game.GameID
			pws.PlayerGameStatID = s.PlayerGameStatsByHashkey[hashkey].PlayerGameStatID

			pwsID, err := db.CPlayerWeaponStat(tx, *pws)
			if err != nil {
				return err
			}

			pws.PlayerWeaponStatID = int(pwsID)
		}
	}

	return nil
}

// CreateTeamGameStats inserts all of the team game stat records to the database.
func CreateTeamGameStats(tx *sql.Tx, db models.Datastore, s *Submission) error {
	for _, tgs := range s.TeamGameStats {
		tgs.GameID = s.Game.GameID
		tgsID, err := db.CTeamGameStat(tx, *tgs)
		if err != nil {
			return err
		}

		tgs.TeamGameStatID = int(tgsID)
	}

	return nil
}

// Submit takes a fully-formed submission and stores it in the database, filling out all the
// missing information (like primary key values) along the way.
func Submit(s *Submission, db models.Datastore) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	server, err := GetOrCreateServer(tx, db, s.Server)
	if err != nil {
		return err
	}
	s.Game.ServerID = server.ServerID

	m, err := GetOrCreateMap(tx, db, s.Map)
	if err != nil {
		return err
	}
	s.Game.MapID = m.MapID

	_, err = GetOrCreatePlayers(tx, db, s)
	if err != nil {
		return err
	}

	err = CreateGame(tx, db, s)
	if err != nil {
		return err
	}

	err = CreatePlayerGameStats(tx, db, s)
	if err != nil {
		return err
	}

	err = CreatePlayerWeaponStats(tx, db, s)
	if err != nil {
		return err
	}

	err = CreateTeamGameStats(tx, db, s)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
