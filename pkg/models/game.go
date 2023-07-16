package models

import (
	"bytes"
	"database/sql"
	"fmt"
	"time"
)

// BlankEndingGameID is a blank ending game ID.
const BlankEndingGameID = -1

// BlankStartingGameID is a blank ending game ID.
const BlankStartingGameID = -1

// BlankLimit is a blank limit value.
const BlankLimit = -1

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
		game.MapID, game.Winner, game.MatchID, game.Mod, game.StartDt, game.Players)

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

		err := rows.Scan(&g.GameID, &g.GameTypeCd, &g.GameTypeDescr, &g.ServerID, &g.MapID,
			&g.Winner, &g.MatchID, &g.Mod, &g.CreateDt, &durationSecs)

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
func (ds *PGDatastore) RGamesByMatchID(serverID int, matchID string) ([]*Game, error) {
	sql := `select game_id, c.game_type_cd, c.descr, server_id, map_id, winner, match_id, mod, create_dt,
	coalesce(cast(extract(epoch from duration) as integer), 0) as duration
	from games g join cd_game_type c on g.game_type_cd = c.game_type_cd
	where match_id = $1
	and server_id = $2
	order by create_dt`

	rows, err := ds.db.Query(sql, matchID, serverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGames(rows)
}

// RGameByID retrives a single game record by its ID value.
func (ds *PGDatastore) RGameByID(gameID int) (*Game, error) {
	sql := `select game_id, c.game_type_cd, c.descr, server_id, map_id, winner, match_id, mod, create_dt,
	coalesce(cast(extract(epoch from duration) as integer), 0) as duration
	from games g join cd_game_type c on g.game_type_cd = c.game_type_cd
	where game_id = $1`

	rows, err := ds.db.Query(sql, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	games, err := scanGames(rows)
	if err != nil {
		return nil, err
	}

	if len(games) != 1 {
		return nil, fmt.Errorf("unable to find a matching game")
	}

	return games[0], nil
}

// RGamesByRange retrives game records within a specified game ID range.
func (ds *PGDatastore) RGamesByRange(start, end, limit int) ([]*Game, error) {
	// Build up the SQL that will eventually be executed.
	var sqlBuf bytes.Buffer

	// Keep track of the bind placeholders and their parameters
	placeholder := 1
	params := make([]interface{}, 0)

	sqlBuf.WriteString(fmt.Sprintf(`select game_id, c.game_type_cd, c.descr, server_id, map_id, winner, match_id, mod, create_dt,
	coalesce(cast(extract(epoch from duration) as integer), 0) as duration
	from games g join cd_game_type c on g.game_type_cd = c.game_type_cd
	where game_id >= $%d `, placeholder))

	placeholder++
	params = append(params, start)

	if end != BlankEndingGameID {
		sqlBuf.WriteString(fmt.Sprintf("and game_id <= $%d ", placeholder))
		placeholder++
		params = append(params, end)
	}

	sqlBuf.WriteString("order by game_id asc ")

	if limit != BlankLimit {
		sqlBuf.WriteString(fmt.Sprintf("limit $%d ", placeholder))
		placeholder++
		params = append(params, limit)
	}

	rows, err := ds.db.Query(sqlBuf.String(), params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGames(rows)
}
