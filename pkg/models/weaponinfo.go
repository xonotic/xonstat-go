package models

import (
	"database/sql"
	"fmt"
	"strings"
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
	and gs.game_id = $2
    order by gs.scoreboardpos;`

	rows, err := ds.db.Query(sql, gameID, gameID)
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

// RPlayerWeaponStatsByGameList retrieves PlayerWeaponsStat rows for a list of game IDs using an IN list.
func (ds *PGDatastore) RPlayerWeaponStatsByGameList(playerID int, gameIDs []int) ([]*PlayerWeaponStat, error) {
	// Convert the game IDs to strings... 
	var strGameIDs []string
	for _, gameID := range gameIDs {
		strGameIDs = append(strGameIDs, fmt.Sprintf("%d", gameID))
	}

	// ... and build IN list from those strings
	gameIDINstr := strings.Join(strGameIDs, ",")

	sql := fmt.Sprintf(`select ws.player_weapon_stats_id, ws.player_id, ws.game_id, 
	ws.player_game_stat_id, ws.weapon_cd, ws.actual, ws.max, ws.hit, ws.fired, ws.frags
	from player_weapon_stats ws 
	where ws.player_id = $1
	and ws.game_id in (%s) 
	order by ws.game_id, ws.weapon_cd;`, gameIDINstr)
	
	rows, err := ds.db.Query(sql, playerID)
	if err != nil {
		return nil, err
	}

	var weaponStats []*PlayerWeaponStat
	for rows.Next() {
		var ws PlayerWeaponStat

		err := rows.Scan(&ws.PlayerWeaponStatID, &ws.PlayerID, &ws.GameID, &ws.PlayerGameStatID, 
			&ws.WeaponCd, &ws.Actual, &ws.Max, &ws.Hit, &ws.Fired, &ws.Frags)

		if err != nil {
			return nil, err
		}

		weaponStats = append(weaponStats, &ws)
	}

	return weaponStats, nil
}