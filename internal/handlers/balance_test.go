package handlers

import (
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
D 100.000000
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
e alivetime 100
e matches 1
e scoreboardvalid 1
e joins 1
P B
i 2
n Bob
t 5
e scoreboard-score 20
e acc-blaster-cnt-fired 1
e alivetime 100
e matches 1
e scoreboardvalid 1
e joins 1
P C
i 3
n Carol
t 14
e scoreboard-score 40
e acc-blaster-cnt-fired 1
e alivetime 100
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
