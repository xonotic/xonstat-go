package models

import (
	"database/sql"
)

// CTeamGameStat inserts a TeamGameStat record into the database.
func (ds *PGDatastore) CTeamGameStat(tx *sql.Tx, tgs TeamGameStat) (int64, error) {
	// The team_game_stats table is partitioned, so the "returning" clause will not work like
	// the other tables. We must grab the sequence value explicitly.
	var tgsID int64

	seqVal, err := ds.nextSeqVal("team_game_stats_team_game_stat_id_seq")
	if err != nil {
		return tgsID, err
	}
	tgsID = seqVal

	sql := `insert into team_game_stats (team_game_stat_id, game_id, team, score, rounds, caps) 
	values ($1, $2, $3, $4, $5, $6)`

	_, err = tx.Exec(sql, tgsID, tgs.GameID, tgs.Team, tgs.Score, tgs.Rounds, tgs.Caps)

	if err != nil {
		return tgsID, err
	}

	return tgsID, nil
}
