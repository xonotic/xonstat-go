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

// RPlayerGameFragMatrixByGameID retrieves frag matrix records by the corresponding game ID.
func (ds *PGDatastore) RPlayerGameFragMatrixByGameID(gameID int) ([]*PlayerGameFragMatrix, error) {
	sql := `select game_id, player_game_stat_id, player_id, player_index, matrix
	from player_game_frag_matrix
	where game_id = $1`

	var results []*PlayerGameFragMatrix

	rows, err := ds.db.Query(sql, gameID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var pgfm PlayerGameFragMatrix
		var rawMatrix []byte

		err := rows.Scan(&pgfm.GameID, &pgfm.PlayerGameStatID, &pgfm.PlayerID, &pgfm.PlayerIndex, &rawMatrix)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(rawMatrix, &pgfm.Matrix)
		if err != nil {
			return nil, err
		}

		results = append(results, &pgfm)
	}

	return results, nil
}
