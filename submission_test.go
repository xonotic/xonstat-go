package submission

import (
	"bufio"
	"os"
	"testing"
)

var eventCountTests = []struct {
	filename             string
	expectedTeamsCount   int
	expectedPlayersCount int
}{
	{"resources/submissions/ca_normal.txt", 2, 3},
	{"resources/submissions/ctf_normal.txt", 2, 11},
	{"resources/submissions/cts_normal.txt", 0, 3},
	{"resources/submissions/dm_normal.txt", 0, 10},
	{"resources/submissions/duel_normal.txt", 0, 3},
	{"resources/submissions/ft_normal.txt", 2, 6},
	{"resources/submissions/ka_normal.txt", 0, 7},
	{"resources/submissions/kh_normal.txt", 3, 9},
	{"resources/submissions/tdm_normal.txt", 2, 6},
}

// test the correct counts of team and player events
func TestCorrectCounts(t *testing.T) {
	for _, testCase := range eventCountTests {
		f, err := os.Open(testCase.filename)
		if err != nil {
			t.Errorf("Unable to open file %s for testing", testCase.filename)
		}
		defer f.Close()

		body := bufio.NewReader(f)
		submission := NewRawSubmission(body)

		err = submission.Parse()
		if err != nil {
			t.Errorf("Unable to parse.")
		}

		if len(submission.TeamEvents) != testCase.expectedTeamsCount {
			t.Errorf("Incorrect number of teams found: found %d, expected %d",
				len(submission.TeamEvents), testCase.expectedTeamsCount)
		}

		if len(submission.PlayerEvents) != testCase.expectedPlayersCount {
			t.Errorf("Incorrect number of players found: found %d, expected %d",
				len(submission.PlayerEvents), testCase.expectedPlayersCount)
		}
	}
}
