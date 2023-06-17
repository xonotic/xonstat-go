package models

import (
	"database/sql"
	"fmt"
	"time"
)

// CPlayerGameStat inserts a PlayerGameStat record into the database.
func (ds *PGDatastore) CPlayerGameStat(tx *sql.Tx, pgs PlayerGameStat) (int64, error) {
	// The player_game_stats table is partitioned, so the "returning" clause will not work like
	// the other tables. We must grab the sequence value explicitly.
	var pgsID int64

	seqVal, err := ds.nextSeqVal("player_game_stats_player_game_stat_id_seq")
	if err != nil {
		return pgsID, err
	}
	pgsID = seqVal

	aliveTimeLiteral := durationToMSStr(pgs.AliveTime)
	timeLiteral := durationToMSStr(pgs.Time)
	fastestLiteral := durationToMSStr(pgs.Fastest)

	rawsql := `insert into player_game_stats (player_id, game_id, nick, stripped_nick, team, rank, 
		alivetime, kills, deaths, suicides, score, time, captures, pickups, drops, returns, 
		collects, destroys, pushes, carrier_frags, elo_delta, fastest, avg_latency, scoreboardpos, 
		laps, revivals, lives, create_dt, player_game_stat_id) 
		values ($1, $2, $3, $4, $5, $6, %s, $7, $8, $9, $10, %s, $11, $12, $13, $14, $15, $16, 
		$17, $18, $19, %s, $20, $21, $22, $23, $24, now() at time zone 'UTC', $25) 
		returning player_game_stat_id`

	sql := fmt.Sprintf(rawsql, aliveTimeLiteral, timeLiteral, fastestLiteral)

	_, err = tx.Exec(sql, pgs.PlayerID, pgs.GameID, pgs.Nick, pgs.StrippedNick, pgs.Team,
		pgs.Rank, pgs.Kills, pgs.Deaths, pgs.Suicides, pgs.Score, pgs.Captures, pgs.Pickups,
		pgs.Drops, pgs.Returns, pgs.Collects, pgs.Destroys, pgs.Pushes, pgs.CarrierFrags,
		pgs.EloDelta, pgs.AvgLatency, pgs.ScoreboardPos, pgs.Laps, pgs.Revivals, pgs.Lives,
		pgsID)

	if err != nil {
		return pgsID, err
	}

	return pgsID, nil
}

// scanPlayerGameStats is a helper function to parse full player game stat records out of a resultset.
func scanPlayerGameStats(rows *sql.Rows) ([]*PlayerGameStat, error) {
	var pgs []*PlayerGameStat
	for rows.Next() {
		var s PlayerGameStat
		var alivetimeMS, timeMS, fastestMS int

		err := rows.Scan(&s.PlayerID, &s.GameID, &s.Nick, &s.StrippedNick, &s.Team, &s.Rank,
			&alivetimeMS, &s.Kills, &s.Deaths, &s.Suicides, &s.Score, &timeMS, &s.Captures,
			&s.Pickups, &s.Drops, &s.Returns, &s.Collects, &s.Destroys, &s.Pushes,
			&s.CarrierFrags, &s.EloDelta, &fastestMS, &s.AvgLatency, &s.ScoreboardPos,
			&s.Laps, &s.Revivals, &s.Lives, &s.CreateDt, &s.PlayerGameStatID)

		alivetime := time.Duration(alivetimeMS) * time.Millisecond
		s.AliveTime = &alivetime

		ttime := time.Duration(timeMS) * time.Millisecond
		s.Time = &ttime

		fastest := time.Duration(fastestMS) * time.Millisecond
		s.Fastest = &fastest

		if err != nil {
			return nil, err
		}

		pgs = append(pgs, &s)
	}

	return pgs, nil
}

// RPlayerGameStatsByGameID retrieves player game stat records by their game ID
func (ds *PGDatastore) RPlayerGameStatsByGameID(gameID int, gameTypeCd string) ([]*PlayerGameStat, error) {
	sql := `select player_id, game_id, nick, stripped_nick, team, rank, 
		cast(coalesce(extract(epoch from alivetime), 0)*1000 as integer), 
		kills, deaths, suicides, score, 
		cast(coalesce(extract(epoch from time), 0)*1000 as integer), 
		captures, pickups, drops, returns, collects, 
		destroys, pushes, carrier_frags, elo_delta, 
		cast(coalesce(extract(epoch from fastest), 0)*1000 as integer), 
		avg_latency, scoreboardpos, laps, revivals, lives, create_dt, player_game_stat_id
		from player_game_stats
		where game_id = $1`

	// CTS needs to be sorted by fastest lap. Scoreboardpos is not reliable.
	if gameTypeCd == "cts" {
		sql += " order by fastest"
	} else {
		sql += " order by scoreboardpos"
	}

	rows, err := ds.db.Query(sql, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pgs, err := scanPlayerGameStats(rows)
	if err != nil {
		return nil, err
	}

	return pgs, nil
}
