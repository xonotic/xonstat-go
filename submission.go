package submission

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/antzucaro/qstr"
	"github.com/antzucaro/xonstat-go/models"
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

// parsePlayerEvents reads all of the player metadata in lines, until hitting a non-player line
func (s *RawSubmission) parsePlayerEvents(label, hashkey string) {
	events := make(map[string]string)
	events[label] = hashkey
	index := -1

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
		case "#", "":
			// no-op: comment or blank line
		default:
			// hit a non-player key, so return that line
			s.rr.Return(fmt.Sprintf("%s %s", key, value))

			// we must be done w/ the player we were working on...
			s.PlayerEvents = append(s.PlayerEvents, events)

			// reference by hashkey or player index
			s.PlayerEventsByHashkey[hashkey] = &events
			s.PlayerEventsByIndex[index] = &events

			return
		}

		// grab the next key/value pair
		key, value, err = s.nextPair()
	}

	if err == io.EOF && len(events) > 1 {
		// special case: the last player in the file
		s.PlayerEvents = append(s.PlayerEvents, events)
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

// weaponFromKey extracts the weapon code from an accuracy event key (e.g. acc-blaster-cnt-fired -> blaster)
func weaponFromKey(key string) string {
	pieces := strings.SplitN(key, "-", 3)
	if len(pieces) == 3 && pieces[0] == "acc" {
		return pieces[1]
	} else {
		return ""
	}
}

// analyze looks over the various events and captures information about them for later validation
func (s *RawSubmission) analyze() error {
	var human, played bool
	for _, playerEvents := range s.PlayerEvents {
		// keep track of the humans and bots that actually played in the game
		human = isHuman(playerEvents)
		played = playedInGame(playerEvents)
		if played {
			if human {
				s.Humans = append(s.Humans, &playerEvents)
			} else {
				s.Bots = append(s.Humans, &playerEvents)
			}
		}

		for key, _ := range playerEvents {
			// keep track of which weapons were used (i.e. fired) in the match via accuracy events
			if strings.HasPrefix(key, "acc-") && strings.HasSuffix(key, "cnt-fired") {
				s.WeaponsUsed[weaponFromKey(key)] = struct{}{}
			}
		}
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

// fillPlayerGameStat fills in a single PlayerGameStat struct from the raw submission events
func (s *Submission) fillPlayerGameStat(rs *RawSubmission, index int) error {
	// an initialized pgstat based on the game type being played
	pgs := models.NewPlayerGameStat(s.Game.GameTypeCd)

	// fields passed on from other objects
	pgs.GameId = s.Game.GameId
	pgs.CreateDt = s.CreateDt

	s.PlayerGameStats = append(s.PlayerGameStats, *pgs)

	return nil
}

// fillPlayers fills in the Players and PlayerHashKeys slices from the raw submission
func (s *Submission) fillPlayers(rs *RawSubmission) error {
	for i, pe := range rs.PlayerEvents {
		hashkey := pe["P"]
		nick := pe["n"]
		nickQStr := qstr.QStr(nick)
		strippedNick := nickQStr.Stripped()

		playerID, err := strconv.Atoi(pe["i"])
		if err != nil {
			playerID = -1
		}

		player := models.Player{
			PlayerId:     playerID,
			Nick:         &nick,
			StrippedNick: &strippedNick,
			CreateDt:     s.CreateDt,
		}

		playerHashkey := models.PlayerHashkey{
			PlayerId: playerID,
			Hashkey:  hashkey,
			CreateDt: s.CreateDt,
		}

		s.fillPlayerGameStat(rs, i)

		s.Players = append(s.Players, player)
		s.PlayerHashkeys = append(s.PlayerHashkeys, playerHashkey)
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

	s := &Submission{
		Players:           players,
		PlayerHashkeys:    playerHashkeys,
		PlayerGameStats:   playerGameStats,
		PlayerWeaponStats: playerWeaponStats,
		TeamGameStats:     teamGameStats,
		CreateDt:          time.Now().UTC(),
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
