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
	{"../../test/submissions/cts_normal.txt", 0, 1},
	{"../../test/submissions/ca_normal.txt", 2, 10},
	{"../../test/submissions/ctf_normal.txt", 2, 14},
	{"../../test/submissions/dm_normal.txt", 0, 8},
	{"../../test/submissions/duel_normal.txt", 0, 4},
	{"../../test/submissions/ft_normal.txt", 2, 11},
	{"../../test/submissions/ka_normal.txt", 0, 7},
	{"../../test/submissions/kh_normal.txt", 4, 10},
	{"../../test/submissions/tdm_normal.txt", 2, 8},
	{"../../test/submissions/overkill_normal.txt", 0, 5},
}

// test the correct counts of team and player events
func TestCorrectCounts(t *testing.T) {
	for _, testCase := range eventCountTests {
		f, err := os.Open(testCase.filename)
		if err != nil {
			t.Fatalf("Unable to open file %s for testing", testCase.filename)
		}
		defer f.Close()

		body := bufio.NewReader(f)
		rawSubmission, err := NewRawSubmission(body)
		if err != nil {
			t.Fatalf("Unable to parse %s: %s", testCase.filename, err)
		}

		if len(rawSubmission.TeamEvents) != testCase.expectedTeamsCount {
			t.Errorf("Incorrect number of teams found in %s: found %d, expected %d",
				testCase.filename, len(rawSubmission.TeamEvents), testCase.expectedTeamsCount)
		}

		if len(rawSubmission.PlayerEvents) != testCase.expectedPlayersCount {
			t.Errorf("Incorrect number of players found in %s: found %d, expected %d",
				testCase.filename, len(rawSubmission.PlayerEvents), testCase.expectedPlayersCount)
		}

		_, err = NewSubmission(rawSubmission)
		if err != nil {
			t.Errorf("Could not parse %s into an actual submission: %s", testCase.filename, err)
		}
	}
}

// test that errors are thrown if a submission is missing required metadata
func TestInvalidMetadata(t *testing.T) {
	filename := "../../test/submissions/missing_metadata.txt"

	f, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Unable to open file %s for testing", filename)
	}
	defer f.Close()

	body := bufio.NewReader(f)
	_, err = NewRawSubmission(body)
	if err != ErrInvalidGameMeta {
		t.Errorf("Did not receive ErrInvalidGameMeta")
	}
}

// test that errors are thrown if a submission is for an unsupported game type
func TestUnsupportedGameType(t *testing.T) {
	filename := "../../test/submissions/unsupported_game_type.txt"

	f, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Unable to open file %s for testing", filename)
	}
	defer f.Close()

	body := bufio.NewReader(f)
	_, err = NewRawSubmission(body)
	if err != ErrUnsupportedGameType {
		t.Errorf("Did not receive ErrUnsupportedGameType")
	}
}

// test that errors are thrown if a submission is for a blank game
func TestBlankGame(t *testing.T) {
	files := []string{
		"../../test/submissions/cts_blank_game.txt",
	}

	for _, filename := range files {
		f, err := os.Open(filename)
		if err != nil {
			t.Fatalf("Unable to open file %s for testing", filename)
		}
		defer f.Close()

		body := bufio.NewReader(f)
		_, err = NewRawSubmission(body)
		if err != ErrBlankGame {
			t.Errorf("Did not receive ErrBlankGame")
		}
	}
}

// test that the submissions have the correct category
func TestCategory(t *testing.T) {
	var categoryTests = []struct {
		filename string
		category string
	}{
		{"../../test/submissions/ctf_normal.txt", "vanilla"},
		{"../../test/submissions/overkill_normal.txt", "overkill"},
		{"../../test/submissions/ctf_insta.txt", "insta"},
		{"../../test/submissions/ctf_jeff.txt", "general"},
	}

	for _, categoryTest := range categoryTests {
		f, err := os.Open(categoryTest.filename)
		if err != nil {
			t.Fatalf("Unable to open file %s for testing", categoryTest.filename)
		}
		defer f.Close()

		body := bufio.NewReader(f)
		rs, err := NewRawSubmission(body)
		if err != nil {
			t.Fatalf("Couldn't parse %s as a valid RawSubmission: %s", categoryTest.filename, err)
		}

		category := gameCategory(rs)
		if category != categoryTest.category {
			t.Fatalf("Incorrect category for %s: expected %s, got %s", categoryTest.filename,
				categoryTest.category, category)
		}
	}
}
