package models

import "time"

// RActiveServers retrieves the active players from the "materialized view".
func (ds *PGDatastore) RActiveServers(limit, start int) ([]*ActiveServer, error) {
	sql := `SELECT sort_order, server_id, server_name, extract(epoch from play_time) player_time, create_dt
	FROM active_servers_mv
	WHERE sort_order >= $2
	LIMIT $1`

	rows, err := ds.db.Query(sql, limit, start)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activeServers []*ActiveServer
	playtime := 0
	for rows.Next() {
		var as ActiveServer
		err := rows.Scan(&as.SortOrder, &as.ServerID, &as.ServerName, &playtime, &as.CreateDt)
		if err != nil {
			return nil, err
		}

		d := time.Duration(playtime) * time.Second
		as.PlayTime = d

		activeServers = append(activeServers, &as)
	}

	return activeServers, nil
}
