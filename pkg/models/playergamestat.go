package models

import (
	"database/sql"
)

// CPlayerGameStat inserts a PlayerGameStat record into the database.
func (ds *PGDatastore) CPlayerGameStat(tx *sql.Tx, pgs PlayerGameStat) (int64, error) {
	// TODO: Convert AliveTime, Time, and Fastest to PG intervals.
	sql := `insert into player_game_stats (player_id, game_id, nick, stripped_nick, team, rank, 
		alivetime, kills, deaths, suicides, score, time, captures, pickups, drops, returns, 
		collects, destroys, pushes, carrier_frags, elo_delta, fastest, avg_latency, scoreboardpos, 
		laps, revivals, lives) 
		values ($1, $2, $3, $4, $5, $6, null, $7, $8, $10, $11, null, $13, $14, $15, $16, $17, $18, 
		$19, $20, $21, null, $23, $24, $25, $26, $27) returning player_game_stat_id`

	row := tx.QueryRow(sql, pgs.PlayerID, pgs.GameID, pgs.Nick, pgs.StrippedNick,pgs.Team, 
		pgs.Rank, pgs.AliveTime, pgs.Kills, pgs.Deaths, pgs.Suicides, pgs.Score, pgs.Time, 
		pgs.Captures, pgs.Pickups, pgs.Drops, pgs.Returns, pgs.Collects, pgs.Destroys, pgs.Pushes, 
		pgs.CarrierFrags, pgs.EloDelta, pgs.Fastest, pgs.AvgLatency, pgs.ScoreboardPos, pgs.Laps, 
		pgs.Revivals, pgs.Lives)

	var pgsID int64
	err := row.Scan(&pgsID)
	if err != nil {
		return pgsID, err
	}

	return pgsID, nil
}
