package models

// RHeatmap retrieves the metadata about games played in hour intervals within the week.
// Its scope is controlled by the amount of data in the underlying materialized view.
func (ds *PGDatastore) RHeatmap() ([]*HeatmapEntry, error) {
	sql := `select 
	extract(isodow from create_dt) as day, 
	extract(hour from create_dt) as hour, 
	count(*)
	from recent_game_stats_mv rgs 
	group by 1, 2;`

	rows, err := ds.db.Query(sql)
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