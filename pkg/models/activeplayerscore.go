package models

import (
	"time"
)

// ActivePlayerScore is the sum score of an active player over a length of time.
type ActivePlayerScore struct {
	SortOrder int
	PlayerID  int
	Nick      string
	Score     int
}

// RActivePlayerScores retrieves the top scoring players for a server over a given period of time.
func (ds *PGDatastore) RActivePlayerScores(serverID int, cutoff *time.Time, limit int) ([]*ActivePlayerScore, error) {
	sql := `SELECT row_number() OVER (ORDER BY sum(player_game_stats.score) DESC) AS rank, 
	players.player_id AS players_player_id, players.nick AS players_nick, 
	sum(player_game_stats.score) AS total_score
	FROM player_game_stats, players, games
	WHERE players.player_id = player_game_stats.player_id 
	AND games.game_id = player_game_stats.game_id 
	AND games.server_id = $1
	AND players.player_id > 2 
	AND player_game_stats.create_dt > $2
	GROUP BY players.nick, players.player_id 
	ORDER BY sum(player_game_stats.score) DESC
	LIMIT $3`

	rows, err := ds.db.Query(sql, serverID, cutoff, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activePlayerScores []*ActivePlayerScore
	for rows.Next() {
		var aps ActivePlayerScore

		err := rows.Scan(&aps.SortOrder, &aps.PlayerID, &aps.Nick, &aps.Score)
		if err != nil {
			return nil, err
		}

		activePlayerScores = append(activePlayerScores, &aps)
	}
	return activePlayerScores, nil
}
