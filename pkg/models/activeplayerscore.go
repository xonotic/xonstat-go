package models

import (
	"context"
	"database/sql"
	"time"
)

// ActivePlayerScore is the sum score of an active player over a length of time.
type ActivePlayerScore struct {
	SortOrder int
	PlayerID  int
	Nick      string
	Score     int
}

func scanActivePlayerScores(rows *sql.Rows) ([]*ActivePlayerScore, error) {
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

// RServerActivePlayerScores retrieves the top scoring players for a server over a given period of time.
func (ds *PGDatastore) RServerActivePlayerScores(serverID int, cutoff *time.Time, limit int) ([]*ActivePlayerScore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sql := `SELECT 
	row_number() OVER (ORDER BY sum(rgs.score) DESC) AS rank, 
	p.player_id, p.nick, 
	sum(rgs.score) AS total_score

	FROM recent_game_stats_mv rgs
	INNER JOIN players p USING (player_id)
	
	WHERE rgs.server_id = $1

	GROUP BY p.nick, p.player_id 
	ORDER BY total_score DESC
	LIMIT $2`

	rows, err := ds.db.QueryContext(ctx, sql, serverID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanActivePlayerScores(rows)
}

// RMapActivePlayerScores retrieves the top scoring players for a map over a given period of time.
func (ds *PGDatastore) RMapActivePlayerScores(mapID int, cutoff *time.Time, limit int) ([]*ActivePlayerScore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sql := `SELECT
	row_number() OVER (ORDER BY sum(rgs.score) DESC) AS rank,
	p.player_id, p.nick,
	sum(rgs.score) AS total_score

	FROM recent_game_stats_mv rgs
	INNER JOIN players p USING (player_id)

	WHERE rgs.map_id = $1

	GROUP BY p.nick, p.player_id
	ORDER BY total_score DESC
	LIMIT $2`

	rows, err := ds.db.QueryContext(ctx, sql, mapID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanActivePlayerScores(rows)
}
