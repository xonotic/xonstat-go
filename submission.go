package submission

import (
	"bufio"
	"fmt"
	"github.com/antzucaro/qstr"
	"github.com/antzucaro/xonstat-go/models"
	"io"
	"strconv"
	"strings"
	"time"
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
	} else {
		err := h.scanner.Err()
		if err != nil {
			return "", err
		} else {
			return "", io.EOF
		}
	}
}

// Return puts a line back into the Queue for the next call to Read()
func (h *ReadReturner) Return(line string) {
	h.queue = append(h.queue, line)
}

// RawSubmission is an untyped game stats submission
type RawSubmission struct {
	// game metadata
	GameMeta map[string]string

	// raw team events: key/value pairs related to teams
	TeamEvents []map[string]string

	// raw player events: key/value pairs related to players
	PlayerEvents []map[string]string

	// ReadReturner used to parse the submission
	rr *ReadReturner
}

// NewRawSubmission creates a new RawSubmission from the given reader
func NewRawSubmission(body io.Reader) *RawSubmission {
	return &RawSubmission{
		GameMeta:     make(map[string]string),
		TeamEvents:   make([]map[string]string, 0),
		PlayerEvents: make([]map[string]string, 0),
		rr:           NewReadReturner(body),
	}
}

// getPair returns the space-separated key/value pair from a given string
func getPair(s string) (string, string, error) {
	tokens := strings.SplitN(s, " ", 2)
	if len(tokens) != 2 {
		return "", "", nil
	} else {
		return tokens[0], tokens[1], nil
	}
}

// nextPair is a helper utility to fetch the next key:value pair from the ReadReturner
func (s *RawSubmission) nextPair() (string, string, error) {
	line, err := s.rr.Read()
	if err != nil {
		return "", "", err
	}

	return getPair(line)
}

// parseTeam reads all of the team metadata in lines, until hitting a non-team line
func (s *RawSubmission) parseTeam(label, teamID string) {
	team := make(map[string]string)
	team[label] = teamID

	key, value, err := s.nextPair()
	for err == nil {
		switch key {
		case "e":
			subkey, subvalue, _ := getPair(value)
			team[subkey] = subvalue
		case "#", "":
			// no-op: comment or blank line
		default:
			// hit a non-team key, put that line back in the queue
			s.rr.Return(fmt.Sprintf("%s %s", key, value))

			// we're done with the team we were working on
			s.TeamEvents = append(s.TeamEvents, team)

			return
		}

		// grab the next set
		key, value, err = s.nextPair()
	}

	return
}

// parsePlayer reads all of the player metadata in lines, until hitting a non-player line
func (s *RawSubmission) parsePlayer(label, playerID string) {
	player := make(map[string]string)
	player[label] = playerID

	key, value, err := s.nextPair()
	for err == nil {
		switch key {
		case "i", "n", "t", "r":
			player[key] = value
		case "e":
			subkey, subvalue, _ := getPair(value)
			player[subkey] = subvalue
		case "#", "":
			// no-op: comment or blank line
		default:
			// hit a non-player key, so return that line
			s.rr.Return(fmt.Sprintf("%s %s", key, value))

			// we must be done w/ the player we were working on...
			s.PlayerEvents = append(s.PlayerEvents, player)

			return
		}

		// grab the next key/value pair
		key, value, err = s.nextPair()
	}

	if err == io.EOF && len(player) > 1 {
		// special case: the last player in the file
		s.PlayerEvents = append(s.PlayerEvents, player)
	}

	return
}

// Parse parses the submission's body
func (s *RawSubmission) Parse() error {
	key, value, err := s.nextPair()
	for err == nil {
		switch key {
		case "V", "R", "G", "O", "M", "I", "S", "C", "U", "D", "L":
			// metadata about the game
			s.GameMeta[key] = value
		case "Q":
			s.parseTeam(key, value)
		case "P":
			s.parsePlayer(key, value)
		case "#", "":
			// no-op: a comment or blank line
		default:
			return fmt.Errorf("Invalid top-level key '%s'", key)
		}

		key, value, err = s.nextPair()
	}

	return nil
}

// When a submission's "header" information is missing or invalid
var InvalidGameMeta = fmt.Errorf("invalid game metadata")

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
		return InvalidGameMeta
	}

	if durationSecsStr, ok := rs.GameMeta["D"]; ok {
		if durationSecs, err := strconv.ParseFloat(durationSecsStr, 32); err == nil {
			d := time.Duration(durationSecs) * time.Second
			s.Game.Duration = &d
		}
	} else {
		return InvalidGameMeta
	}

	if matchId, ok := rs.GameMeta["I"]; ok {
		s.Game.MatchId = &matchId
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
		return InvalidGameMeta
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
		return InvalidGameMeta
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

		playerId, err := strconv.Atoi(pe["i"])
		if err != nil {
			playerId = -1
		}

		player := models.Player{
			PlayerId:     playerId,
			Nick:         &nick,
			StrippedNick: &strippedNick,
			CreateDt:     s.CreateDt,
		}

		playerHashkey := models.PlayerHashkey{
			PlayerId: playerId,
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
