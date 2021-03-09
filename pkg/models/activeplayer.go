package models

import (
	"context"
	"database/sql"
	"time"
)

func scanActivePlayers(rows *sql.Rows) ([]*ActivePlayer, error) {
	var activePlayers []*ActivePlayer
	alivetime := 0
	for rows.Next() {
		var ap ActivePlayer
		err := rows.Scan(&ap.SortOrder, &ap.PlayerID, &ap.Nick, &alivetime, &ap.CreateDt)
		if err != nil {
			return nil, err
		}

		d := time.Duration(alivetime) * time.Second
		ap.AliveTime = d

		activePlayers = append(activePlayers, &ap)
	}
	return activePlayers, nil
}

// RActivePlayers retrieves the active players from the "materialized view".
func (ds *PGDatastore) RActivePlayers(limit, start int) ([]*ActivePlayer, error) {
	sql := `SELECT sort_order, player_id, nick, cast(extract(epoch from alivetime) as INTEGER) alivetime, create_dt
	FROM active_players_mv
	WHERE sort_order >= $2
	LIMIT $1`

	rows, err := ds.db.Query(sql, limit, start)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanActivePlayers(rows)
}

// RActivePlayersByServer retrieves the active players for a given server.
func (ds *PGDatastore) RActivePlayersByServer(serverID int, cutoff *time.Time, limit int) ([]*ActivePlayer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sql := `SELECT 
	row_number() OVER (ORDER BY sum(player_game_stats.alivetime) DESC) AS rank, 
	players.player_id AS players_player_id, players.nick AS players_nick, 
	cast(extract(epoch from sum(player_game_stats.alivetime)) AS INTEGER) AS total_alivetime,
	now() at time zone 'UTC' AS create_dt

	FROM player_game_stats
	INNER JOIN players USING (player_id)
	INNER JOIN games USING (game_id)

	WHERE games.server_id = $1
	AND players.player_id > 2
	AND player_game_stats.create_dt BETWEEN $2 AND (now() at time zone 'UTC' + interval '1 day')
	AND games.create_dt BETWEEN $2 AND (now() at time zone 'UTC' + interval '1 day')

	GROUP BY players.nick, players.player_id 
	ORDER BY total_alivetime DESC
	LIMIT $3`

	rows, err := ds.db.QueryContext(ctx, sql, serverID, cutoff, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanActivePlayers(rows)
}

// RActivePlayersByMap retrieves the active players for a given map.
func (ds *PGDatastore) RActivePlayersByMap(mapID int, cutoff *time.Time, limit int) ([]*ActivePlayer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sql := `SELECT 
	row_number() OVER (ORDER BY sum(player_game_stats.alivetime) DESC) AS rank, 
	players.player_id AS players_player_id, players.nick AS players_nick, 
	cast(extract(epoch from sum(player_game_stats.alivetime)) AS INTEGER) AS total_alivetime,
	now() at time zone 'UTC' AS create_dt

	FROM player_game_stats
	INNER JOIN players USING (player_id)
	INNER JOIN games USING (game_id)

	WHERE games.map_id = $1
	AND players.player_id > 2
	AND player_game_stats.create_dt BETWEEN $2 AND (now() at time zone 'UTC' + interval '1 day')
	AND games.create_dt BETWEEN $2 AND (now() at time zone 'UTC' + interval '1 day')

	GROUP BY players.nick, players.player_id 
	ORDER BY total_alivetime DESC
	LIMIT $3`

	rows, err := ds.db.QueryContext(ctx, sql, mapID, cutoff, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanActivePlayers(rows)
}