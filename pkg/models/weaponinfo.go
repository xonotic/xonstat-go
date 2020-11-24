package models

import (
	"database/sql"
)

// scanWeaponInfos is a helper function to parse the query results for weapons
func scanWeaponInfos(rows *sql.Rows) ([]*WeaponInfo, error) {
	var weapons []*WeaponInfo
	for rows.Next() {
		var wi WeaponInfo

		err := rows.Scan(&wi.PlayerWeaponStatID, &wi.PlayerID, &wi.Nick, &wi.GameID, &wi.PlayerGameStatID, 
			&wi.WeaponCd, &wi.Actual, &wi.Max, &wi.Hit, &wi.Fired, &wi.Frags)

		if err != nil {
			return nil, err
		}

		weapons = append(weapons, &wi)
	}

	return weapons, nil
}

// RWeaponInfoByGameID retrieves player weapon info records by their game ID
// Returns back the list of distinct weapons, map[player_game_stat_id][weapon_cd]*WeaponItem, and any error.
func (ds *PGDatastore) RWeaponInfoByGameID(gameID int) ([]*WeaponInfo, error) {
	sql := `select ws.player_weapon_stats_id, ws.player_id, gs.nick, ws.game_id, 
	ws.player_game_stat_id, ws.weapon_cd, ws.actual, ws.max, ws.hit, ws.fired, ws.frags
	from player_weapon_stats ws join player_game_stats gs on ws.player_game_stat_id = gs.player_game_stat_id
	where ws.game_id = $1
    order by gs.scoreboardpos;`

	rows, err := ds.db.Query(sql, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	wis, err := scanWeaponInfos(rows)
	if err != nil {
		return nil, err
	}

	return wis, nil
}
