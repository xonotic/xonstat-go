package models

import (
	"database/sql"
)

// CGame inserts a Game record into the database.
func (ds *PGDatastore) CGame(tx *sql.Tx, game Game) (int64, error) {
	// TODO: validate start_dt is set properly
	// TODO: validate winner is set properly in Submission object before getting here
	// TODO: validate players is set properly to an array type
	sql := `insert into games (game_type_cd, server_id, map_id, duration, winner, match_id, mod,
		category, players, start_dt)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) returning game_id`

	row := tx.QueryRow(sql, game.GameTypeCd, game.ServerID, game.MapID, game.Duration, game.Winner,
		game.MatchID, game.Mod, game.Category, game.Players, game.StartDt)

	var gameID int64
	err := row.Scan(&gameID)
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
		err := rows.Scan(&g.GameID, &g.GameTypeCd, &g.ServerID, &g.MapID, &g.Duration, &g.Winner,
			&g.MatchID, &g.Mod, &g.Category, &g.Players, &g.StartDt)

		if err != nil {
			return nil, err
		}

		games = append(games, &g)
	}

	return games, nil
}

// RGamesByMatchID retrives game records by their MatchID value.
func (ds *PGDatastore) RGamesByMatchID(matchID string) ([]*Game, error) {
	sql := `select game_id, game_type_cd, server_id, map_id, duration, winner, match_id, mod, 
	category, players, start_dt
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
