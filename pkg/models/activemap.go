package models

import (
	"context"
	"database/sql"
	"time"
)

func scanActiveMaps(rows *sql.Rows) ([]*ActiveMap, error) {
	var activeMaps []*ActiveMap
	for rows.Next() {
		var am ActiveMap
		err := rows.Scan(&am.SortOrder, &am.MapID, &am.MapName, &am.Games, &am.CreateDt)
		if err != nil {
			return nil, err
		}

		activeMaps = append(activeMaps, &am)
	}

	return activeMaps, nil
}

// RActiveMaps retrieves the active players from the "materialized view".
func (ds *PGDatastore) RActiveMaps(limit, start int) ([]*ActiveMap, error) {
	sql := `SELECT sort_order, map_id, map_name, games, create_dt
	FROM active_maps_mv
	WHERE sort_order >= $2
	LIMIT $1`

	rows, err := ds.db.Query(sql, limit, start)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanActiveMaps(rows)
}

// RActiveMapsByServer finds the most active maps played on a server over a given time period.
func (ds *PGDatastore) RActiveMapsByServer(serverID int, cutoff *time.Time, limit int) ([]*ActiveMap, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sql := `SELECT 
	row_number() OVER (ORDER BY count(*) DESC) AS rank, 
	games.map_id AS games_map_id, maps.name AS maps_name, count(*) AS times_played,
	now() at time zone 'UTC' AS create_dt

	FROM games
	INNER JOIN maps USING (map_id)

	WHERE games.server_id = $1 
	AND games.create_dt BETWEEN $2 AND (now() at time zone 'UTC' + interval '1 day')

	GROUP BY games.map_id, maps.name 
	ORDER BY times_played DESC
	LIMIT $3`

	rows, err := ds.db.QueryContext(ctx, sql, serverID, cutoff, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanActiveMaps(rows)
}
