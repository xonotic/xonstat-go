package models

import (
	"database/sql"
)

// CHashkey inserts a hashkey record into the database.
func (ds *PGDatastore) CHashkey(tx *sql.Tx, hashkey string, playerID int64) error {
	sql := `insert into hashkeys (hashkey, player_id) values ($1, $2)`

	_, err := tx.Exec(sql, hashkey, playerID)
	if err != nil {
		return err
	}

	return nil
}
