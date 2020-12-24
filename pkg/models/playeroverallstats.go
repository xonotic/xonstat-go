package models

import (
	"time"
)

// RPlayerOverallStats retrieves summary information about a player's stats.
func (ds *PGDatastore) RPlayerOverallStats(playerID int) ([]*PlayerOverallStats, error) {
	sql := `
	SELECT g.game_type_cd, 
		   gt.descr game_type_descr, 
		   SUM(pgs.kills)         total_kills, 
		   SUM(pgs.deaths)        total_deaths, 
		   MAX(pgs.create_dt)     last_played, 
	       cast(extract(epoch from sum(pgs.alivetime)) AS INTEGER) AS total_playing_time,
		   SUM(pgs.pickups)       total_pickups, 
		   SUM(pgs.captures)      total_captures, 
		   SUM(pgs.carrier_frags) total_carrier_frags 
	FROM  games g, 
		  cd_game_type gt, 
		  player_game_stats pgs 
	WHERE g.game_id = pgs.game_id 
	AND   g.game_type_cd = gt.game_type_cd 
	AND   g.players @> ARRAY[$1] 
	AND   pgs.player_id = $2 
	GROUP BY g.game_type_cd, game_type_descr 
	UNION 
	SELECT 'overall'              game_type_cd, 
		   'Overall'              game_type_descr, 
		   Sum(pgs.kills)         total_kills, 
		   Sum(pgs.deaths)        total_deaths, 
		   Max(pgs.create_dt)     last_played, 
		   cast(extract(epoch from sum(pgs.alivetime)) AS INTEGER) AS total_playing_time,
		   Sum(pgs.pickups)       total_pickups, 
		   Sum(pgs.captures)      total_captures, 
		   Sum(pgs.carrier_frags) total_carrier_frags 
	FROM   player_game_stats pgs 
	WHERE  pgs.player_id = $3 `

	rows, err := ds.db.Query(sql, playerID, playerID, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*PlayerOverallStats
	for rows.Next() {
		var s PlayerOverallStats
		alivetime := 0

		err := rows.Scan(&s.GameTypeCd, &s.GameTypeDescr, &s.Kills, &s.Deaths, &s.LastPlayed, &alivetime, &s.Pickups, &s.Captures, &s.CarrierFrags)

		s.TimePlayed = time.Duration(alivetime) * time.Second

		if err != nil {
			return nil, err
		}

		stats = append(stats, &s)
	}

	return stats, nil
}
