package models

import (
	"context"
	"database/sql"
	"log"
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
func (ds *PGDatastore) RActivePlayersByServer(serverID int, limit int) ([]*ActivePlayer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sql := `SELECT 
	row_number() OVER (ORDER BY sum(rgs.alivetime) DESC) AS rank, 
	p.player_id, p.nick, 
	cast(extract(epoch from sum(rgs.alivetime)) AS INTEGER) AS total_alivetime,
	now() at time zone 'UTC' AS create_dt

	FROM recent_game_stats_mv rgs
	INNER JOIN players p USING (player_id)

	WHERE rgs.server_id = $1

	GROUP BY p.nick, p.player_id 
	ORDER BY total_alivetime DESC
	LIMIT $2`

	rows, err := ds.db.QueryContext(ctx, sql, serverID, limit)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	return scanActivePlayers(rows)
}

// RActivePlayersByMap retrieves the active players for a given map.
func (ds *PGDatastore) RActivePlayersByMap(mapID int, limit int) ([]*ActivePlayer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sql := `SELECT 
	row_number() OVER (ORDER BY sum(rgs.alivetime) DESC) AS rank, 
	p.player_id, p.nick, 
	cast(extract(epoch from sum(rgs.alivetime)) AS INTEGER) AS total_alivetime,
	now() at time zone 'UTC' AS create_dt

	FROM recent_game_stats_mv rgs
	INNER JOIN players p USING (player_id)

	WHERE rgs.map_id = $1

	GROUP BY p.nick, p.player_id 
	ORDER BY total_alivetime DESC
	LIMIT $2`

	rows, err := ds.db.QueryContext(ctx, sql, mapID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanActivePlayers(rows)
}
