package handlers

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"gitlab.com/xonotic/xonstat/pkg/submission"
)

type balancePlayer struct {
	Hashkey        string
	PlayerID       int
	Nick           string
	Skill          float64
	ScorePerSecond float64
}

// BySkillandSPS implements sort.Interface for []balancePlayer based on
// the Skill and then ScorePerSecond fields, greatest the least.
type BySkillandSPS []*balancePlayer

func (a BySkillandSPS) Len() int      { return len(a) }
func (a BySkillandSPS) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a BySkillandSPS) Less(i, j int) bool {
	if a[i].Skill != a[j].Skill {
		return a[i].Skill > a[j].Skill
	}
	return a[i].ScorePerSecond > a[j].ScorePerSecond
}

type balanceResponse struct {
	Version int
	Release string
	Time    int64
	Players []*balancePlayer
}

// randomSwap picks two random positions in the array and swaps them. 
func randomSwap(players []*balancePlayer) {
	rand.Seed(time.Now().UnixNano())

	i := rand.Intn(len(players))
	j := rand.Intn(len(players))

	if i != j {
		players[i], players[j] = players[j], players[i]
	}
}

// BalanceHandler takes player info from servers and returns back a best-guess
// ordering of those players according to their skill.
func (ae *AppEnv) BalanceHandler(w http.ResponseWriter, r *http.Request) {
	bodyReader := bufio.NewReader(r.Body)

	params := r.URL.Query()
	jitter, err := strconv.Atoi(params.Get("jitter"))
	if err != nil || jitter < 1 || jitter > 5 {
		jitter = 0
	}

	rawSubmission, err := submission.NewRawSubmission(bodyReader)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("422 %s", http.StatusText(422)), 422)
		return
	}

	// It doesn't make sense to balance games with only one or two players.
	if rawSubmission.NumHumansPlayed < 3 {
		log.Printf("Error: not enough players to balance")
		http.Error(w, fmt.Sprintf("422 %s", http.StatusText(422)), 422)
		return
	}

	sub, err := submission.NewSubmission(rawSubmission)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("422 %s", http.StatusText(422)), 422)
		return
	}

	// The destination for the balanced (sorted) players.
	players := make([]*balancePlayer, 0)

	// Use hashkeys to make updates to entries in the above slice.
	hashkeysToPlayers := make(map[string]*balancePlayer, 0)

	// Hashkeys for tracked players
	tracked := make([]string, 0)

	for hashkey, player := range sub.PlayersByHashkey {
		gamestat, hasGameStat := sub.PlayerGameStatsByHashkey[hashkey]
		if strings.HasPrefix(hashkey, "bot#") || !hasGameStat {
			// Bots and players that don't have a stat record aren't considered.
			continue
		}

		score := gamestat.Score.Int32
		alivetimesecs := sub.PlayerGameStatsByHashkey[hashkey].AliveTime.Seconds()
		sps := float64(score) / alivetimesecs

		player := balancePlayer{
			Hashkey:        hashkey,
			PlayerID:       player.PlayerID,
			Nick:           player.Nick.String,
			ScorePerSecond: sps,
		}

		players = append(players, &player)

		hashkeysToPlayers[hashkey] = &player

		if !strings.HasPrefix(hashkey, "player#") {
			// This is a tracked player, and might have a skill record.
			// 1. Save these types of hashkeys
			tracked = append(tracked, hashkey)
		}
	}

	if len(tracked) > 0 {
		// 2. Query for skills for those hashkeys
		skills, err := ae.db.RPlayerSkillsBatch(tracked, sub.Game.GameTypeCd)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, fmt.Sprintf("422 %s", http.StatusText(422)), 422)
			return
		}

		// 3. Update the corresponding balancePlayer record w/ skill info (if found)
		for _, skill := range skills {
			// Optimistic (right end of the confidence interval): we are ~97% confident that
			// this player's skill is LESS than this number.
			player := hashkeysToPlayers[skill.Hashkey]
			player.Skill = skill.Mu + (3 * skill.Sigma)
		}
	}

	// 4. Sort the untracked players by skill (if available), falling back to score per second.
	sort.Sort(BySkillandSPS(players))

	// 5. Introduce jitter to randomize results, if asked.
	for i:=0; i < jitter; i++ {
		randomSwap(players)
	}

	response := balanceResponse{
		Version: 1,
		Release: "XonStat/1.0",
		Time:    sub.CreateDt.Unix(),
		Players: players,
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "text/plain")

	err = ae.templates["balance.page.html"].Execute(w, response)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, fmt.Sprintf("500 %s", http.StatusText(500)), 500)
		return
	}
}
