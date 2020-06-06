package models

import (
	"database/sql"
	"encoding/json"
)

// CPlayerGameFragMatrix inserts a PlayerGameFragMatrix record into the database.
func (ds *PGDatastore) CPlayerGameFragMatrix(tx *sql.Tx, fm PlayerGameFragMatrix) error {
	matrix, err := json.Marshal(fm.Matrix)
	if err != nil {
		return err
	}

	sql := `insert into player_game_frag_matrix 
	(game_id, player_game_stat_id, player_id, player_index, matrix) 
	values ($1, $2, $3, $4, $5)`

	_, err = tx.Exec(sql, fm.GameID, fm.PlayerGameStatID, fm.PlayerID, fm.PlayerIndex, matrix)
	if err != nil {
		return err
	}

	return nil
}
