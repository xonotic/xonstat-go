package submission

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/antzucaro/qstr"
	"gitlab.com/antibody/xonstat/pkg/models"
)

// ReadReturner represents a streaming line reader that you can put lines back into. On subsequent
// calls to Read(), lines put back are returned in the order in which they were received.
type ReadReturner struct {
	// the underlying scanner that we're reading from
	scanner *bufio.Scanner

	// the queue of strings returned to the reader for processing elsewhere
	queue []string
}

// NewReadReturner creates a new ReadReturner.
func NewReadReturner(r io.Reader) *ReadReturner {
	queue := make([]string, 0)
	scanner := bufio.NewScanner(r)
	return &ReadReturner{scanner, queue}
}

// Read returns the next line in the queue (if non-empty) or the next line in the request's body
// if the queue is empty.
func (h *ReadReturner) Read() (string, error) {
	// if there are items on the queue, return those first (in order)
	if len(h.queue) > 0 {
		val := h.queue[0]
		h.queue = append(h.queue[:0], h.queue[1:]...)
		return val, nil
	}

	// otherwise we will pull a new line from the scanner if one is available
	scanned := h.scanner.Scan()
	if scanned {
		return h.scanner.Text(), nil
	}

	err := h.scanner.Err()
	if err != nil {
		return "", err
	}

	return "", io.EOF
}

// Return puts a line back into the Queue for the next call to Read()
func (h *ReadReturner) Return(line string) {
	h.queue = append(h.queue, line)
}

// ErrInvalidGameMeta is when a submission's "header" information is missing or invalid
var ErrInvalidGameMeta = fmt.Errorf("invalid game metadata")

// ErrUnsupportedGameType is when a submission is for an unsupported game type
var ErrUnsupportedGameType = fmt.Errorf("unsupported game type")

// ErrBlankGame is when a submission is blank
var ErrBlankGame = fmt.Errorf("blank game")

// RawSubmission is an untyped game stats submission
type RawSubmission struct {
	// game metadata
	GameMeta map[string]string

	// raw team events: key/value pairs related to teams
	TeamEvents []map[string]string

	// raw player events: key/value pairs related to players
	PlayerEvents []map[string]string

	// humans who played in the match
	Humans []*map[string]string

	// bots who played in the match
	Bots []*map[string]string

	// weapons used during the match
	WeaponsUsed map[string]struct{}

	// did a human who played in the game fire a weapon?
	HumanFiredWeapon bool

	// did a human who played in the game have a non-zero score?
	HumanNonZeroScore bool

	// did a human who played in the game record a fastest lap?
	HumanFastestLap bool

	// references to player events by player hashkey
	PlayerEventsByHashkey map[string]*map[string]string

	// references to player events by player index
	PlayerEventsByIndex map[int]*map[string]string

	// ReadReturner used to parse the submission
	rr *ReadReturner
}

// NewRawSubmission creates a new RawSubmission from the given reader
func NewRawSubmission(body io.Reader) (*RawSubmission, error) {
	rs := &RawSubmission{
		GameMeta:              make(map[string]string),
		TeamEvents:            make([]map[string]string, 0),
		PlayerEvents:          make([]map[string]string, 0),
		Humans:                make([]*map[string]string, 0),
		Bots:                  make([]*map[string]string, 0),
		WeaponsUsed:           make(map[string]struct{}, 0),
		PlayerEventsByHashkey: make(map[string]*map[string]string, 0),
		PlayerEventsByIndex:   make(map[int]*map[string]string, 0),
		rr:                    NewReadReturner(body),
	}

	err := rs.parse()
	if err != nil {
		return nil, err
	}

	err = rs.analyze()
	if err != nil {
		return nil, err
	}

	err = rs.validate()
	if err != nil {
		return nil, err
	}

	return rs, nil
}

// getPair returns the space-separated key/value pair from a given string
func getPair(s string) (string, string, error) {
	tokens := strings.SplitN(s, " ", 2)
	if len(tokens) != 2 {
		return "", "", nil
	}

	return tokens[0], tokens[1], nil
}

// nextPair is a helper utility to fetch the next key:value pair from the ReadReturner
func (s *RawSubmission) nextPair() (string, string, error) {
	line, err := s.rr.Read()
	if err != nil {
		return "", "", err
	}

	return getPair(line)
}

// parseTeamEvents reads all of the team metadata in lines, until hitting a non-team line
func (s *RawSubmission) parseTeamEvents(label, teamID string) {
	events := make(map[string]string)
	events[label] = teamID

	key, value, err := s.nextPair()
	for err == nil {
		// consume all team-related events under the team label Q: e
		switch key {
		case "e":
			subkey, subvalue, _ := getPair(value)
			events[subkey] = subvalue
		case "#", "":
			// no-op: comment or blank line
		default:
			// hit a non-team key, put that line back in the queue
			s.rr.Return(fmt.Sprintf("%s %s", key, value))

			// we're done with the team we were working on
			s.TeamEvents = append(s.TeamEvents, events)

			return
		}

		// grab the next set
		key, value, err = s.nextPair()
	}

	return
}

// isHuman determines of the set of player events represents a human
func isHuman(events map[string]string) bool {
	return !strings.HasPrefix(events["P"], "bot")
}

// playedInGame determines of the set of player events represents a player who played the match (is on the scoreboard)
func playedInGame(events map[string]string) bool {
	_, matches := events["matches"]
	_, scoreboardvalid := events["scoreboardvalid"]
	return matches && scoreboardvalid
}

// addPlayerEvents adds a set of player events to the running list and performs some bookkeeping for the same
func (s *RawSubmission) addPlayerEvents(events map[string]string, hashkey string, index int, firedWeapon, nonZeroScore, hasFastestLap bool) {
	if len(events) <= 0 {
		return
	}

	human := isHuman(events)
	played := playedInGame(events)
	if human && played {
		if firedWeapon {
			s.HumanFiredWeapon = true
		}

		if nonZeroScore {
			s.HumanNonZeroScore = true
		}

		if hasFastestLap {
			s.HumanFastestLap = true
		}

		s.Humans = append(s.Humans, &events)
	} else if !human && played {
		s.Bots = append(s.Bots, &events)
	}

	// reference by hashkey or player index
	s.PlayerEventsByHashkey[hashkey] = &events
	s.PlayerEventsByIndex[index] = &events

	// we are done w/ the events we were working on...
	s.PlayerEvents = append(s.PlayerEvents, events)
}

// parsePlayerEvents reads all of the player metadata in lines, until hitting a non-player line
func (s *RawSubmission) parsePlayerEvents(label, hashkey string) {
	events := make(map[string]string)
	events[label] = hashkey
	index := -1
	firedWeapon := false
	hasFastestLap := false
	nonZeroScore := false

	key, value, err := s.nextPair()
	for err == nil {
		// consume all player-related keys below the player label P: i, n, t, r, e
		switch key {
		case "i":
			index, _ = strconv.Atoi(value)
			events[key] = value
		case "n", "t", "r":
			events[key] = value
		case "e":
			subkey, subvalue, _ := getPair(value)
			events[subkey] = subvalue

			// did this player fire a weapon?
			if strings.HasPrefix(subkey, "acc-") && strings.HasSuffix(subkey, "cnt-fired") {
				firedWeapon = true
			}

			if subkey == "scoreboard-score" {
				score, err := strconv.ParseFloat(subvalue, 32)
				if err == nil && (score > 0.0 || score < 0.0) {
					nonZeroScore = true
				}
			}

			// did this player have a fastest lap
			if subkey == "scoreboard-fastest" {
				hasFastestLap = true
			}
		case "#", "":
			// no-op: comment or blank line
		default:
			// hit a non-player key, so return that line
			s.rr.Return(fmt.Sprintf("%s %s", key, value))
			s.addPlayerEvents(events, hashkey, index, firedWeapon, nonZeroScore, hasFastestLap)

			return
		}

		// grab the next key/value pair
		key, value, err = s.nextPair()
	}

	if err == io.EOF {
		// special case: the last player in the file
		s.addPlayerEvents(events, hashkey, index, firedWeapon, nonZeroScore, hasFastestLap)
	}

	return
}

// parse parses the submission's body
func (s *RawSubmission) parse() error {
	key, value, err := s.nextPair()
	for err == nil {
		switch key {
		case "V", "R", "G", "O", "M", "I", "S", "C", "U", "D", "L":
			// metadata about the game
			s.GameMeta[key] = value
		case "Q":
			s.parseTeamEvents(key, value)
		case "P":
			s.parsePlayerEvents(key, value)
		case "#", "":
			// no-op: a comment or blank line
		default:
			return fmt.Errorf("Invalid top-level key '%s'", key)
		}

		key, value, err = s.nextPair()
	}

	return nil
}

// hasRequiredMetadata checks that the required top-level metadata is present
func (s *RawSubmission) hasRequiredMetadata() error {
	for _, requiredKey := range []string{"G", "V", "I", "S", "M"} {
		if _, ok := s.GameMeta[requiredKey]; !ok {
			return ErrInvalidGameMeta
		}
	}
	return nil
}

func (s *RawSubmission) isSupportedGameType() error {
	switch s.GameMeta["G"] {
	case "as":
		return nil
	case "ca":
		return nil
	case "ctf":
		return nil
	case "cts":
		return nil
	case "dm":
		return nil
	case "dom":
		return nil
	case "duel":
		return nil
	case "ft":
		return nil
	case "freezetag":
		return nil
	case "ka":
		return nil
	case "keepaway":
		return nil
	case "kh":
		return nil
	case "nb":
		return nil
	case "nexball":
		return nil
	case "rune":
		return nil
	case "tdm":
		return nil
	default:
		return ErrUnsupportedGameType
	}
}

// weaponFromKey extracts the weapon code from an accuracy event key (e.g. acc-blaster-cnt-fired -> blaster)
func weaponFromKey(key string) string {
	pieces := strings.SplitN(key, "-", 3)
	if len(pieces) == 3 && pieces[0] == "acc" {
		return pieces[1]
	}
	return ""
}

// analyze looks over the various events and captures information about them for later validation
func (s *RawSubmission) analyze() error {
	for _, playerEvents := range s.PlayerEvents {
		for key := range playerEvents {
			// keep track of which weapons were used (i.e. fired) in the match via accuracy events
			if strings.HasPrefix(key, "acc-") && strings.HasSuffix(key, "cnt-fired") {
				s.WeaponsUsed[weaponFromKey(key)] = struct{}{}
			}
		}
	}
	return nil
}

// isBlankGame determines if the game has data worth processing
func (s *RawSubmission) isBlankGame() error {
	gameType := s.GameMeta["G"]
	if gameType == "cts" {
		if !s.HumanFastestLap {
			// CTS requires a human to capture a fastest lap
			return ErrBlankGame
		}
	} else if (gameType == "nb" || gameType == "nexball") && !s.HumanNonZeroScore {
		// in Nexball, we need a human to have a non-zero score
		return ErrBlankGame
	} else if !(s.HumanNonZeroScore && s.HumanFiredWeapon) {
		return ErrBlankGame
	}

	return nil
}

// validate runs the preconditions checks possible for raw submissions
func (s *RawSubmission) validate() error {
	err := s.hasRequiredMetadata()
	if err != nil {
		return err
	}

	err = s.isSupportedGameType()
	if err != nil {
		return err
	}

	err = s.isBlankGame()
	if err != nil {
		return err
	}
	return nil
}

// Submission is a fully-formatted statistics POST request
type Submission struct {
	Game              models.Game
	Server            models.Server
	Map               models.Map
	Players           []models.Player
	PlayerHashkeys    []models.PlayerHashkey
	PlayerGameStats   []models.PlayerGameStat
	PlayerWeaponStats []models.PlayerWeaponStat
	TeamGameStats     []models.TeamGameStat
	CreateDt          time.Time

	// References by player ID (initially the player events 'i' or index value) for easier processing
	PlayersByID           map[int]*models.Player
	PlayerHashkeysByID    map[int]*models.PlayerHashkey
	PlayerGameStatsByID   map[int]*models.PlayerGameStat
	PlayerWeaponStatsByID map[int][]*models.PlayerWeaponStat
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
		"hmg":        struct{}{},
		"vortex":     struct{}{},
		"shotgun":    struct{}{},
		"blaster":    struct{}{},
		"machinegun": struct{}{},
		"rpc":        struct{}{},
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
		if durationSecs, err := strconv.ParseFloat(durationSecsStr, 32); err == nil {
			d := time.Duration(durationSecs) * time.Second
			s.Game.Duration = &d
		}
	} else {
		return ErrInvalidGameMeta
	}

	if matchID, ok := rs.GameMeta["I"]; ok {
		s.Game.MatchId = &matchID
	}

	if mod, ok := rs.GameMeta["O"]; ok {
		s.Game.Mod = &mod
	}

	category := gameCategory(rs)
	s.Game.Category = &category

	s.Game.StartDt = s.CreateDt
	s.Game.CreateDt = s.CreateDt

	return nil
}

// fillServer fills in the Server attribute from the raw submission
func (s *Submission) fillServer(rs *RawSubmission) error {
	if serverName, ok := rs.GameMeta["S"]; ok {
		s.Server.Name = &serverName
	} else {
		return ErrInvalidGameMeta
	}

	if portStr, ok := rs.GameMeta["U"]; ok {
		if port, err := strconv.Atoi(portStr); err == nil {
			s.Server.Port = &port
		}
	}

	if impureCvarsStr, ok := rs.GameMeta["C"]; ok {
		if impureCvars, err := strconv.Atoi(impureCvarsStr); err == nil {
			s.Server.ImpureCvars = &impureCvars
		}
	}

	if revision, ok := rs.GameMeta["R"]; ok {
		s.Server.Revision = &revision
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

// intFromStringDefault converts a string to an int if possible, and if not returns a default value
func intFromStringDefault(value string, defaultVal int) *int {
	intVal, err := strconv.Atoi(value)
	if err != nil {
		return &defaultVal
	}
	return &intVal
}

// intFromStringDefault converts a string to an int if possible, and if not returns nil
func intFromString(value string) *int {
	intVal, err := strconv.Atoi(value)
	if err != nil {
		return nil
	}
	return &intVal
}

// fillPlayerGameStat fills in a single PlayerGameStat struct from the raw submission events
func (s *Submission) fillPlayerGameStat(events map[string]string, player *models.Player) error {
	// an initialized pgstat based on the game type being played
	pgs := models.NewPlayerGameStat(s.Game.GameTypeCd)

	// fields passed on from other objects
	pgs.PlayerId = player.PlayerId
	pgs.GameId = s.Game.GameId
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
	pgs.Score = &score

	alivetimeSecs := 0
	if alivetimeStr, ok := events["alivetime"]; ok {
		alivetimeFloat, err := strconv.ParseFloat(alivetimeStr, 64)
		if err == nil {
			alivetimeSecs = int(math.Round(alivetimeFloat))
		}
	}
	alivetime := time.Duration(alivetimeSecs) * time.Second
	pgs.AliveTime = &alivetime

	if rankStr, ok := events["rank"]; ok {
		pgs.Rank = intFromStringDefault(rankStr, 0)
	}

	if scoreboardPosStr, ok := events["scoreboardpos"]; ok {
		pgs.ScoreboardPos = intFromStringDefault(scoreboardPosStr, 0)
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
			// TODO: pgstat.fastest = datetime.timedelta(seconds=float(value)/100)
			// TODO: if the game type is ctf, do fastest cap processing
		case "scoreboard-revivals":
			pgs.Revivals = intFromString(value)
		case "scoreboard-bctime":
			// TODO: pgstat.time = datetime.timedelta(seconds=int(value))
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
			// TODO: pgstat.avg_latency = float(value)
		}

		// TODO: process anticheat records
	}

	// there is no "winning team" field, so we derive it
	if wins {
		s.Game.Winner = pgs.Team
	}

	s.PlayerGameStats = append(s.PlayerGameStats, *pgs)
	s.PlayerGameStatsByID[pgs.PlayerId] = pgs

	return nil
}

// fillPlayers fills in the Players and PlayerHashKeys slices from the raw submission
func (s *Submission) fillPlayers(rs *RawSubmission) error {
	for _, events := range rs.PlayerEvents {
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

		playerID, err := strconv.Atoi(events["i"])
		if err != nil {
			playerID = -1
		}

		player := models.Player{
			PlayerId:     playerID,
			Nick:         &nick,
			StrippedNick: &strippedNick,
			CreateDt:     s.CreateDt,
		}
		s.Players = append(s.Players, player)
		s.PlayersByID[playerID] = &player

		playerHashkey := models.PlayerHashkey{
			PlayerId: playerID,
			Hashkey:  hashkey,
			CreateDt: s.CreateDt,
		}
		s.PlayerHashkeys = append(s.PlayerHashkeys, playerHashkey)
		s.PlayerHashkeysByID[playerID] = &playerHashkey

		s.fillPlayerGameStat(events, &player)
	}
	return nil
}

// NewSubmission converts a RawSubmission into a fully-formed one
func NewSubmission(rs *RawSubmission) (*Submission, error) {
	players := make([]models.Player, 0, len(rs.PlayerEvents))
	playerHashkeys := make([]models.PlayerHashkey, 0, len(rs.PlayerEvents))
	playerGameStats := make([]models.PlayerGameStat, 0, len(rs.PlayerEvents))
	playerWeaponStats := make([]models.PlayerWeaponStat, 0)
	teamGameStats := make([]models.TeamGameStat, 0)
	playersByID := make(map[int]*models.Player, 0)
	playerHashkeysByIndex := make(map[int]*models.PlayerHashkey, 0)
	playerGameStatsByIndex := make(map[int]*models.PlayerGameStat, 0)
	playerWeaponStatsByIndex := make(map[int][]*models.PlayerWeaponStat, 0)

	s := &Submission{
		Players:               players,
		PlayerHashkeys:        playerHashkeys,
		PlayerGameStats:       playerGameStats,
		PlayerWeaponStats:     playerWeaponStats,
		TeamGameStats:         teamGameStats,
		CreateDt:              time.Now().UTC(),
		PlayersByID:           playersByID,
		PlayerHashkeysByID:    playerHashkeysByIndex,
		PlayerGameStatsByID:   playerGameStatsByIndex,
		PlayerWeaponStatsByID: playerWeaponStatsByIndex,
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

	err = s.fillPlayers(rs)
	if err != nil {
		return nil, err
	}

	return s, nil
}
