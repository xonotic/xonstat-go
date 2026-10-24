package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func bodyFromFile(t *testing.T, filename string) string {
	t.Helper()

	body, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Unable to read test submission %s: %s", filename, err)
	}

	return string(body)
}

// tdmBody is a minimal team game submission with three human players.
const tdmBody = `V 9
R test
G tdm
O Xonotic
M testmap
I 0.123456789
S Test Server
C 0
U 26000
D 240.000000
Q team#5
e scoreboard-score 0
Q team#14
e scoreboard-score 0
P A
i 1
n Alice
t 5
e scoreboard-score 30
e acc-blaster-cnt-fired 1
e alivetime 240
e matches 1
e scoreboardvalid 1
e joins 1
P B
i 2
n Bob
t 5
e scoreboard-score 20
e acc-blaster-cnt-fired 1
e alivetime 240
e matches 1
e scoreboardvalid 1
e joins 1
P C
i 3
n Carol
t 14
e scoreboard-score 40
e acc-blaster-cnt-fired 1
e alivetime 240
e matches 1
e scoreboardvalid 1
e joins 1
`

// tooManyTeamsBody is a team game submission with 5 teams, exceeding the
// maxTeams limit of 4.
const tooManyTeamsBody = `V 9
R test
G tdm
O Xonotic
M testmap
I 0.123456789
S Test Server
C 0
U 26000
D 240.000000
Q team#1
e scoreboard-score 0
Q team#2
e scoreboard-score 0
Q team#3
e scoreboard-score 0
Q team#4
e scoreboard-score 0
Q team#5
e scoreboard-score 0
P A
i 1
n Alice
t 1
e scoreboard-score 10
e acc-blaster-cnt-fired 1
e alivetime 240
e matches 1
e scoreboardvalid 1
e joins 1
P B
i 2
n Bob
t 2
e scoreboard-score 20
e acc-blaster-cnt-fired 1
e alivetime 240
e matches 1
e scoreboardvalid 1
e joins 1
P C
i 3
n Carol
t 3
e scoreboard-score 30
e acc-blaster-cnt-fired 1
e alivetime 240
e matches 1
e scoreboardvalid 1
e joins 1
P D
i 4
n Dave
t 4
e scoreboard-score 40
e acc-blaster-cnt-fired 1
e alivetime 240
e matches 1
e scoreboardvalid 1
e joins 1
P E
i 5
n Eve
t 5
e scoreboard-score 50
e acc-blaster-cnt-fired 1
e alivetime 240
e matches 1
e scoreboardvalid 1
e joins 1
`

// TestBalanceRejectsNonTeamGameType verifies that /balance rejects games
// without teams, since there is nothing to balance.
func TestBalanceRejectsNonTeamGameType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/balance",
		strings.NewReader(bodyFromFile(t, "../../test/submissions/dm_normal.txt")))
	w := httptest.NewRecorder()

	sub, err := preprocess(w, req)
	if err == nil {
		t.Fatal("Expected an error for a non-team game, got a submission")
	}
	if sub != nil {
		t.Fatal("Expected no submission for a non-team game")
	}
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("Expected status %d, got %d", http.StatusUnprocessableEntity, w.Code)
	}
}

// TestBalanceAcceptsTeamGameType verifies that team games still pass
// preprocessing.
func TestBalanceAcceptsTeamGameType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/balance", strings.NewReader(tdmBody))
	w := httptest.NewRecorder()

	sub, err := preprocess(w, req)
	if err != nil {
		t.Fatalf("Expected a team game to pass preprocessing, got: %s", err)
	}
	if sub == nil {
		t.Fatal("Expected a submission for a team game")
	}
}

// TestBalanceRejectsTooManyTeams verifies that /balance rejects games with more
// than maxTeams teams, to bound processing time.
func TestBalanceRejectsTooManyTeams(t *testing.T) {
	ae := &AppEnv{
		textTemplates: loadTextTemplates("../../web/template"),
	}

	req := httptest.NewRequest(http.MethodPost, "/balance", strings.NewReader(tooManyTeamsBody))
	w := httptest.NewRecorder()

	ae.BalanceHandler(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("Expected status %d for too many teams, got %d", http.StatusUnprocessableEntity, w.Code)
	}
}

// discardWriteCloser satisfies AppEnv's requestLogger without retaining data.
type discardWriteCloser struct{ io.Writer }

func (discardWriteCloser) Close() error { return nil }

// blankTdmBody is a team game with three human players, zero scores, and no
// weapon-fire events: blank by the submission endpoint's rules (no non-zero
// score, no weapon fired), but still a match the balancer can process.
const blankTdmBody = `V 9
R test
G tdm
O Xonotic
M testmap
I 0.123456789
S Test Server
C 0
U 26000
D 240.000000
Q team#5
e scoreboard-score 0
Q team#14
e scoreboard-score 0
P player#aaaa
i 1
n Alice
t 5
e scoreboard-score 0
e alivetime 240
e matches 1
e scoreboardvalid 1
e joins 1
P player#bbbb
i 2
n Bob
t 5
e scoreboard-score 0
e alivetime 240
e matches 1
e scoreboardvalid 1
e joins 1
P player#cccc
i 3
n Carol
t 14
e scoreboard-score 0
e alivetime 240
e matches 1
e scoreboardvalid 1
e joins 1
`

// TestBalanceAcceptsBlankGame verifies that the balance handler processes a
// match the submission endpoint considers blank (no non-zero scores, no
// weapon fired) instead of rejecting it with a 422.
func TestBalanceAcceptsBlankGame(t *testing.T) {
	ae := &AppEnv{
		textTemplates: loadTextTemplates("../../web/template"),
	}

	req := httptest.NewRequest(http.MethodPost, "/balance", strings.NewReader(blankTdmBody))
	w := httptest.NewRecorder()

	ae.BalanceHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d for a blank game, got %d (body: %s)",
			http.StatusOK, w.Code, w.Body.String())
	}
}

// TestSubmissionStillRejectsBlankGame verifies that the blank game checks
// remain in place for the submission endpoint: the same body that the
// balance handler accepts is rejected with a 422 here.
func TestSubmissionStillRejectsBlankGame(t *testing.T) {
	ae := &AppEnv{
		requestLogger: discardWriteCloser{io.Discard},
	}

	req := httptest.NewRequest(http.MethodPost, "/submission", strings.NewReader(blankTdmBody))
	w := httptest.NewRecorder()

	ae.SubmissionHandler(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("Expected status %d for a blank game, got %d (body: %s)",
			http.StatusUnprocessableEntity, w.Code, w.Body.String())
	}
}
