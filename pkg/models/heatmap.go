package models

import (
	"bytes"
)

// RHeatmap retrieves the metadata about games played in hour intervals within the week.
// Its scope is controlled by the amount of data in the underlying materialized view.
func (ds *PGDatastore) RHeatmap(serverID int) ([]*HeatmapEntry, error) {

	// Build up the SQL that will eventually be executed.
	var sqlBuf bytes.Buffer

	sqlBuf.WriteString(`select 
	extract(isodow from create_dt) as day, 
	extract(hour from create_dt) as hour, 
	count(*)
	from recent_game_stats_mv rgs `)

	params := make([]interface{}, 0)

	if serverID != -1 {
		sqlBuf.WriteString("where server_id = $1 ")
		params = append(params, serverID)
	}

	sqlBuf.WriteString(`group by 1, 2;`)

	rows, err := ds.db.Query(sqlBuf.String(), params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*HeatmapEntry
	for rows.Next() {
		var hme HeatmapEntry
		err := rows.Scan(&hme.DayOfWeek, &hme.HourOfDay, &hme.Count)
		if err != nil {
			return nil, err
		}

		entries = append(entries, &hme)
	}

	return entries, nil
}