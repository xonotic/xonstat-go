package models

import (
	"database/sql"
)

// CPlayerWeaponStat inserts a PlayerWeaponStat record into the database.
func (ds *PGDatastore) CPlayerWeaponStat(tx *sql.Tx, pws PlayerWeaponStat) (int64, error) {
	// The player_weapon_stats table is partitioned, so the "returning" clause will not work like
	// the other tables. We must grab the sequence value explicitly.
	var pwsID int64

	seqVal, err := ds.nextSeqVal("player_weapon_stats_player_weapon_stats_id_seq")
	if err != nil {
		return pwsID, err
	}
	pwsID = seqVal

	sql := `insert into player_weapon_stats (player_weapon_stats_id, player_id, game_id, 
		player_game_stat_id, weapon_cd, actual, max, hit, fired, frags) 
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) returning player_weapon_stats_id`

	_, err = tx.Exec(sql, pwsID, pws.PlayerID, pws.GameID, pws.PlayerGameStatID, pws.WeaponCd, 
		pws.Actual, pws.Max, pws.Hit, pws.Fired, pws.Frags)

	if err != nil {
		return pwsID, err
	}

	return pwsID, nil
}

// scanPlayerWeaponStats is a helper function to parse full player weapon stat records out of a resultset.
func scanPlayerWeaponStats(rows *sql.Rows) ([]*PlayerWeaponStat, error) {
	var pws []*PlayerWeaponStat
	for rows.Next() {
		var s PlayerWeaponStat

		err := rows.Scan(&s.PlayerWeaponStatID, &s.PlayerID, &s.GameID, &s.PlayerGameStatID, 
			&s.WeaponCd, &s.Actual, &s.Max, &s.Hit, &s.Fired, &s.Frags)

		if err != nil {
			return nil, err
		}

		pws = append(pws, &s)
	}

	return pws, nil
}

// RPlayerWeaponStatsByGameID retrieves player weapon stat records by their game ID
func (ds *PGDatastore) RPlayerWeaponStatsByGameID(gameID int) ([]*PlayerWeaponStat, error) {
	sql := `select ws.player_weapon_stats_id, ws.player_id, ws.game_id, 
	ws.player_game_stat_id, ws.weapon_cd, ws.actual, ws.max, ws.hit, ws.fired, ws.frags
	from player_weapon_stats ws join player_game_stats gs on ws.player_game_stat_id = gs.player_game_stat_id
	where ws.game_id = $1
    order by gs.scoreboardpos;`

	rows, err := ds.db.Query(sql, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pgs, err := scanPlayerWeaponStats(rows)
	if err != nil {
		return nil, err
	}

	return pgs, nil
}
