package models

import "time"

// RActivePlayers retrieves the active players from the "materialized view".
func (ds *PGDatastore) RActivePlayers(limit, start int) ([]*ActivePlayer, error) {
	sql := `SELECT sort_order, player_id, nick, extract(epoch from alivetime) alivetime, create_dt
	FROM active_players_mv
	WHERE sort_order >= $2
	LIMIT $1`

	rows, err := ds.db.Query(sql, limit, start)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
