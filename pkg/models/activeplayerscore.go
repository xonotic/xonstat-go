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

	return scanActivePlayerScores(rows)
}

// RMapActivePlayerScores retrieves the top scoring players for a map over a given period of time.
func (ds *PGDatastore) RMapActivePlayerScores(mapID int, cutoff *time.Time, limit int) ([]*ActivePlayerScore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sql := `SELECT
	row_number() OVER (ORDER BY sum(player_game_stats.score) DESC) AS rank,
	players.player_id AS players_player_id, players.nick AS players_nick,
	sum(player_game_stats.score) AS total_score

	FROM player_game_stats
	INNER JOIN players USING (player_id)
	INNER JOIN games USING (game_id)

	WHERE games.map_id = $1
	AND players.player_id > 2
	AND player_game_stats.create_dt BETWEEN $2 AND (now() at time zone 'UTC' + interval '1 day')
	AND games.create_dt BETWEEN $2 AND (now() at time zone 'UTC' + interval '1 day')

	GROUP BY players_nick, players_player_id
	ORDER BY total_score DESC

	LIMIT $3`

	rows, err := ds.db.QueryContext(ctx, sql, mapID, cutoff, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanActivePlayerScores(rows)
}
