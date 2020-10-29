package models

import (
	"database/sql"
	"fmt"
	"time"
)

// CPlayerGameNonParticipant inserts a PlayerGameNonParticipant record into the database.
func (ds *PGDatastore) CPlayerGameNonParticipant(tx *sql.Tx, pgnp PlayerGameNonParticipant) (int64, error) {
	// The player_game_nonparticipants table is partitioned, so the "returning" clause will not work like
	// the other tables. We must grab the sequence value explicitly.
	var pgnpID int64

	seqVal, err := ds.nextSeqVal("player_game_nonparticipants_player_game_nonparticipants_id_seq")
	if err != nil {
		return pgnpID, err
	}
	pgnpID = seqVal

	aliveTimeLiteral := durationToMSStr(pgnp.AliveTime)

	rawsql := `insert into player_game_nonparticipants (player_game_nonparticipants_id, player_id, game_id, 
		status, nick, stripped_nick, alivetime, score, create_dt) 
		values ($1, $2, $3, $4, $5, $6, %s, $7, now() at time zone 'UTC') 
		returning player_game_nonparticipants_id`

	sql := fmt.Sprintf(rawsql, aliveTimeLiteral)

	_, err = tx.Exec(sql, pgnpID, pgnp.PlayerID, pgnp.GameID, pgnp.Status, pgnp.Nick.String, 
		pgnp.StrippedNick.String, pgnp.Score.Int32)

	if err != nil {
		return pgnpID, err
	}

	return pgnpID, nil
}

// scanPlayerGameNonParticipants is a helper function to parse full player game nonparticipant records out of a resultset.
func scanPlayerGameNonParticipants(rows *sql.Rows) ([]*PlayerGameNonParticipant, error) {
	var pgnps []*PlayerGameNonParticipant
	for rows.Next() {
		var s PlayerGameNonParticipant
		var alivetimeMS int

		err := rows.Scan(&s.PlayerGameNonParticipantID, &s.PlayerID, &s.GameID, &s.Status, &s.Nick, 
			&s.StrippedNick, &alivetimeMS, &s.Score, &s.CreateDt)

		alivetime := time.Duration(alivetimeMS) * time.Millisecond
		s.AliveTime = &alivetime

		if err != nil {
			return nil, err
		}

		pgnps = append(pgnps, &s)
	}

	return pgnps, nil
}

// RPlayerGameNonParticipantsByGameID retrieves player game nonparticipant records by their game ID
func (ds *PGDatastore) RPlayerGameNonParticipantsByGameID(gameID int) ([]*PlayerGameNonParticipant, error) {
	sql := `select player_game_nonparticipants_id, player_id, game_id, status, nick, stripped_nick,
	cast(coalesce(extract(epoch from alivetime), 0)*1000 as integer), score, create_dt
	from player_game_nonparticipants
	where game_id = $1
	order by player_game_nonparticipants_id`

	rows, err := ds.db.Query(sql, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pgs, err := scanPlayerGameNonParticipants(rows)
	if err != nil {
		return nil, err
	}

	return pgs, nil
}
