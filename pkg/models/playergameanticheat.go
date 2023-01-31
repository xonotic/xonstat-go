package models

import (
	"database/sql"
)

// CPlayerGameAnticheat inserts a PlayerGameAnticheat record into the database.
func (ds *PGDatastore) CPlayerGameAnticheat(tx *sql.Tx, ac PlayerGameAnticheat) error {
	sql := `insert into player_game_anticheats(player_id, game_id, key, value) 
		values($1, $2, $3, $4);`

	_, err := tx.Exec(sql, ac.PlayerID, ac.GameID, ac.Key, ac.Value)

	if err != nil {
		return err
	}

	return nil
}
