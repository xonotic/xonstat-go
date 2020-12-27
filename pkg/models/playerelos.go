package models

import (
	"database/sql"
)

// scanElos scans PlayerElo values from a SQL query
func scanElos(rows *sql.Rows) ([]*PlayerElo, error) {
	var elos []*PlayerElo
	for rows.Next() {
		var e PlayerElo
		err := rows.Scan(&e.PlayerID, &e.GameTypeCd, &e.Games, &e.Elo, &e.ActiveInd, 
			&e.CreateDt, &e.UpdateDt)

		if err != nil {
			return nil, err
		}

		elos = append(elos, &e)
	}

	return elos, nil
}

// RPlayerElosByHashkey retrives player Elo records by the player's hashkey value.
func (ds *PGDatastore) RPlayerElosByHashkey(hashkey string) ([]*PlayerElo, error) {
	sql := `select e.player_id, e.game_type_cd, e.games, e.elo, e.active_ind, e.create_dt, e.update_dt
	from player_elos e join hashkeys h on e.player_id = h.player_id
	join players p on p.player_id = h.player_id
	where h.hashkey = $1
	and p.active_ind=true`

	rows, err := ds.db.Query(sql, hashkey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanElos(rows)
}