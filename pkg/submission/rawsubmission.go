package submission

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

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
	Humans []map[string]string

	// bots who played in the match
	Bots []map[string]string

	// humans who either spectated the entire match or forfeited before its end
	NonParticipants []map[string]string

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
		Humans:                make([]map[string]string, 0),
		Bots:                  make([]map[string]string, 0),
		NonParticipants:       make([]map[string]string, 0),
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

// joinedGame determines if the set of player events is for a player who joined the match
func joinedGame(events map[string]string) bool {
	_, joins := events["joins"]
	return joins
}

// addPlayerEvents adds a set of player events to the running list and performs some bookkeeping for the same
func (s *RawSubmission) addPlayerEvents(events map[string]string, hashkey string, index int, firedWeapon, nonZeroScore, hasFastestLap bool) {
	if len(events) <= 0 {
		return
	}

	human := isHuman(events)
	joined := joinedGame(events)
	playedTillEnd := playedInGame(events)
	if human {
		if (joined && !playedTillEnd) || (!joined && !playedTillEnd) {
			// If they joined the match but didn't actually finish it, they forfeited.
			// If they did not join, they spectated. The actual designation is done later,
			// after we have parsed AliveTime properly.
			s.NonParticipants = append(s.NonParticipants, events)
		} else if joined && playedTillEnd {
			if firedWeapon {
				s.HumanFiredWeapon = true
			}

			if nonZeroScore {
				s.HumanNonZeroScore = true
			}

			if hasFastestLap {
				s.HumanFastestLap = true
			}

			s.Humans = append(s.Humans, events)
		}
	} else if !human && playedTillEnd {
		s.Bots = append(s.Bots, events)
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

// IsSupportedGameType determines if XonStat supports the given game type
func IsSupportedGameType(gameTypeCd string) bool {
	switch gameTypeCd {
	case "as", "ca", "ctf", "cts", "dm", "dom", "duel", "ft", "freezetag", "ka":
		return true
	case "keepaway", "kh", "nb", "nexball", "rune", "tdm":
		return true
	default:
		return false
	}
}

func (s *RawSubmission) isSupportedGameType() error {
	if !IsSupportedGameType(s.GameMeta["G"]) {
		return ErrUnsupportedGameType
	}

	return nil
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
