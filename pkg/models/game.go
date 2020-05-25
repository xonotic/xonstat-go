package models

import (
	"github.com/lib/pq"
	"database/sql"
	"fmt"
)

// CGame inserts a Game record into the database.
func (ds *PGDatastore) CGame(tx *sql.Tx, game Game) (int64, error) {
	// TODO: validate start_dt is set properly
	// TODO: validate winner is set properly in Submission object before getting here
	// TODO: validate players is set properly to an array type
	var gameID int64

	// The games table is partitioned, so the "returning" clause will not work like
	// the other tables. We must grab the sequence value explicitly.
	seqVal, err := ds.nextSeqVal("games_game_id_seq")
	if err != nil {
		return gameID, err
	}
	gameID = seqVal

	durationLiteral := durationToMSStr(game.Duration)

	sql := `insert into games (game_id, game_type_cd, server_id, map_id, winner, match_id, mod,
		start_dt, duration, players)
		values ($1, $2, $3, $4, $5, $6, $7, $8, %s, $9)`

	_, err = tx.Exec(fmt.Sprintf(sql, durationLiteral), seqVal, game.GameTypeCd, game.ServerID,
		game.MapID, game.Winner, game.MatchID, game.Mod, game.StartDt, pq.Array(game.Players))

	if err != nil {
		return gameID, err
	}

	return gameID, nil
}

// scanGames is a helper function to parse full Game records out of a resultset.
func scanGames(rows *sql.Rows) ([]*Game, error) {
	var games []*Game
	for rows.Next() {
		var g Game
		err := rows.Scan(&g.GameID, &g.GameTypeCd, &g.ServerID, &g.MapID, &g.Winner,
			&g.MatchID, &g.Mod, &g.StartDt)

		if err != nil {
			return nil, err
		}

		games = append(games, &g)
	}

	return games, nil
}

// RGamesByMatchID retrives game records by their MatchID value.
func (ds *PGDatastore) RGamesByMatchID(matchID string) ([]*Game, error) {
	// TODO: retrieve the duration as a string, convert it back into a time.Duration
	sql := `select game_id, game_type_cd, server_id, map_id, winner, match_id, mod, 
	start_dt
	from games
	where match_id = $1
	order by create_dt`

	rows, err := ds.db.Query(sql, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGames(rows)
}
