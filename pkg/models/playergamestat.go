package models

import (
	"database/sql"
)

// CPlayerGameStat inserts a PlayerGameStat record into the database.
func (ds *PGDatastore) CPlayerGameStat(tx *sql.Tx, pgs PlayerGameStat) (int64, error) {
	// TODO: Convert AliveTime, Time, and Fastest to PG intervals.

	// The player_game_stats table is partitioned, so the "returning" clause will not work like
	// the other tables. We must grab the sequence value explicitly.
	var pgsID int64

	seqVal, err := ds.nextSeqVal("player_game_stats_player_game_stat_id_seq")
	if err != nil {
		return pgsID, err
	}
	pgsID = seqVal

	sql := `insert into player_game_stats (player_id, game_id, nick, stripped_nick, team, rank, 
		alivetime, kills, deaths, suicides, score, time, captures, pickups, drops, returns, 
		collects, destroys, pushes, carrier_frags, elo_delta, fastest, avg_latency, scoreboardpos, 
		laps, revivals, lives, create_dt, player_game_stat_id) 
		values ($1, $2, $3, $4, $5, $6, null, $7, $8, $9, $10, null, $11, $12, $13, $14, $15, $16, 
		$17, $18, $19, null, $20, $21, $22, $23, $24, now() at time zone 'UTC', $25) 
		returning player_game_stat_id`

	_, err = tx.Exec(sql, pgs.PlayerID, pgs.GameID, pgs.Nick, pgs.StrippedNick,pgs.Team, 
		pgs.Rank, pgs.Kills, pgs.Deaths, pgs.Suicides, pgs.Score, pgs.Captures, pgs.Pickups, 
		pgs.Drops, pgs.Returns, pgs.Collects, pgs.Destroys, pgs.Pushes, pgs.CarrierFrags, 
		pgs.EloDelta, pgs.AvgLatency, pgs.ScoreboardPos, pgs.Laps, pgs.Revivals, pgs.Lives,
	    pgsID)

	if err != nil {
		return pgsID, err
	}

	return pgsID, nil
}
