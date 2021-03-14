package models

import (
	"context"
	"time"
)

// RActiveServers retrieves the active players from the "materialized view".
func (ds *PGDatastore) RActiveServers(limit, start int) ([]*ActiveServer, error) {
	sql := `SELECT sort_order, server_id, server_name, extract(epoch from play_time) play_time, create_dt
	FROM active_servers_mv
	WHERE sort_order >= $2
	LIMIT $1`

	rows, err := ds.db.Query(sql, limit, start)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activeServers []*ActiveServer
	var playtime float64
	for rows.Next() {
		var as ActiveServer
		err := rows.Scan(&as.SortOrder, &as.ServerID, &as.ServerName, &playtime, &as.CreateDt)
		if err != nil {
			return nil, err
		}

		d := time.Duration(int64(playtime)) * time.Second
		as.PlayTime = d

		activeServers = append(activeServers, &as)
	}

	return activeServers, nil
}

// RActiveServersByMap finds the servers who have played a given map (by its ID) most frequently
// over a given time period.
func (ds *PGDatastore) RActiveServersByMap(mapID int, cutoff *time.Time, limit int) ([]*ActiveServer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sql := `SELECT 
	row_number() over(order by sum(extract(epoch from pgs.alivetime)) desc) as rank,
	s.server_id, s.name, sum(extract(epoch from pgs.alivetime)) time_played, 
	now() at time zone 'utc' 

	FROM player_game_stats pgs
	INNER JOIN games g USING (game_id)
	INNER JOIN servers s USING (server_id)

	WHERE g.map_id = $1
	AND pgs.player_id > 1
	AND pgs.create_dt BETWEEN $2 AND (now() at time zone 'UTC' + interval '1 day')
	AND g.create_dt BETWEEN $2 AND (now() at time zone 'UTC' + interval '1 day')

	GROUP BY s.server_id, s.name
	ORDER BY time_played desc
	LIMIT $3`

	rows, err := ds.db.QueryContext(ctx, sql, mapID, cutoff, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activeServers []*ActiveServer
	var playtime float64
	for rows.Next() {
		var as ActiveServer
		err := rows.Scan(&as.SortOrder, &as.ServerID, &as.ServerName, &playtime, &as.CreateDt)
		if err != nil {
			return nil, err
		}

		d := time.Duration(int64(playtime)) * time.Second
		as.PlayTime = d

		activeServers = append(activeServers, &as)
	}

	return activeServers, nil
}