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
