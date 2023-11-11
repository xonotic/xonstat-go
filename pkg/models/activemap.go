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
func (ds *PGDatastore) RActiveMapsByServer(serverID int, limit int) ([]*ActiveMap, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sql := `
	WITH rgs AS (
		SELECT 
		    map_id, count(*) AS times_played
		FROM 
		    recent_game_stats_mv
		WHERE 
		    server_id = $1
		GROUP BY map_id
		ORDER BY 2 desc
		LIMIT $2
	)
	SELECT 
	    row_number() OVER (ORDER BY times_played DESC) AS rank, 
		rgs.map_id, m.name, rgs.times_played, now() at time zone 'UTC' AS create_dt
	FROM 
	    rgs INNER JOIN maps m USING (map_id) `

	rows, err := ds.db.QueryContext(ctx, sql, serverID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanActiveMaps(rows)
}
