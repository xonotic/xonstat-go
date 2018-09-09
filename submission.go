package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
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
	// header data
	Headers map[string]string

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
func NewRawSubmission(headers map[string]string, body io.Reader) *RawSubmission {
	return &RawSubmission{
		Headers:      headers,
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

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println("Couldn't open the file")
		os.Exit(1)
	}

	body := bufio.NewReader(f)
	headers := make(map[string]string)

	submission := NewRawSubmission(headers, body)
	err = submission.Parse()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		fmt.Printf("Meta: %+v\n", submission.GameMeta)
		fmt.Printf("# Teams: %+v\n", len(submission.TeamEvents))
		fmt.Printf("# Players: %+v\n", len(submission.PlayerEvents))
	}

}
