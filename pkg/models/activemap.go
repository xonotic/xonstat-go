package models

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
