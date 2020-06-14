package models

// RSummaryStats retrives summary stats by scope. Scope may be either "all" or "day".
func (ds *PGDatastore) RSummaryStats(scope string) ([]*SummaryStat, error) {
	sql := `SELECT num_players, game_type_cd, num_games, create_dt refresh_dt 
	FROM summary_stats_mv 
	WHERE scope = $1
	ORDER BY sort_order `

	rows, err := ds.db.Query(sql, scope)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summarystats []*SummaryStat
	for rows.Next() {
		var ss SummaryStat
		err := rows.Scan(&ss.PlayerCount, &ss.GameTypeCd, &ss.GameCount, &ss.RefreshDt)
		if err != nil {
			return nil, err
		}

		summarystats = append(summarystats, &ss)
	}

	return summarystats, nil
}
