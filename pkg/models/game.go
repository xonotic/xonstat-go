package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// CGame inserts a Game record into the database.
func (ds *PGDatastore) CGame(tx *sql.Tx, game Game) (int64, error) {
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
		var durationSecs int

		err := rows.Scan(&g.GameID, &g.GameTypeCd, &g.ServerID, &g.MapID, &g.Winner,
			&g.MatchID, &g.Mod, &g.StartDt, &durationSecs)

		d := time.Duration(durationSecs) * time.Second
		g.Duration = &d

		if err != nil {
			return nil, err
		}

		games = append(games, &g)
	}

	return games, nil
}

// RGamesByMatchID retrives game records by their MatchID value.
func (ds *PGDatastore) RGamesByMatchID(matchID string) ([]*Game, error) {
	sql := `select game_id, game_type_cd, server_id, map_id, winner, match_id, mod, start_dt,
	cast(extract(epoch from duration) as integer) as duration
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

// RGameByID retrives a single game record by its ID value.
func (ds *PGDatastore) RGameByID(gameID int) (*Game, error) {
	sql := `select game_id, game_type_cd, server_id, map_id, winner, match_id, mod, start_dt,
	cast(extract(epoch from duration) as integer) as duration
	from games
	where game_id = $1`

	rows, err := ds.db.Query(sql, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	games, err := scanGames(rows)
	if err != nil || len(games) != 1 {
		return nil, err
	}

	return games[0], nil
}
