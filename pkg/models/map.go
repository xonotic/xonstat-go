package models

import (
	"database/sql"
)

// CMap inserts a Map record into the database.
func (ds *PGDatastore) CMap(tx *sql.Tx, m Map) (int64, error) {
	sql := `insert into maps (name, version, pk3_name, curl_url)
		values ($1, $2, $3, $4) returning map_id`

	row := tx.QueryRow(sql, m.Name, m.Version, m.Pk3Name, m.CurlUrl)

	var mapID int64
	err := row.Scan(&mapID)
	if err != nil {
		return mapID, err
	}

	return mapID, nil
}

// scanMaps is a helper function to parse full Map records out of a resultset.
func scanMaps(rows *sql.Rows) ([]*Map, error) {
	var maps []*Map
	for rows.Next() {
		var m Map
		err := rows.Scan(&m.MapId, &m.Name, &m.Version, &m.Pk3Name, &m.CurlUrl, &m.CreateDt)

		if err != nil {
			return nil, err
		}

		maps = append(maps, &m)
	}

	return maps, nil
}

// RMapsByName retrives map records by their name value.
func (ds *PGDatastore) RMapsByName(name string) ([]*Map, error) {
	sql := `select map_id, name, version, pk3_name, curl_url, create_dt 
	from maps
	where name = $1
	order by create_dt`

	rows, err := ds.db.Query(sql, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMaps(rows)
}
