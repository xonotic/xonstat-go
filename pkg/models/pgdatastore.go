package models

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // PostgreSQL-specific datastore implementation
)

// PGDatastore is an implementation of the Datastore interface for a Postgresql database.
type PGDatastore struct {
	db *sql.DB
}

// NewPGDatastore creates a new concrete implementation of a Datastore.
func NewPGDatastore(dsn string, maxidle int, maxopen int) (*PGDatastore, error) {
	// establish a database connection
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	// connection pooling
	db.SetMaxIdleConns(maxidle)
	db.SetMaxOpenConns(maxopen)

	return &PGDatastore{db}, nil
}

// Begin starts a transaction.
func (ds *PGDatastore) Begin() (*sql.Tx, error) {
	return ds.db.Begin()
}

// nextSeqVal grabs the next value of the given sequence name.
func (ds *PGDatastore) nextSeqVal(sequenceName string) (int64, error) {
	sql := fmt.Sprintf("select nextval('%s')", sequenceName)
	row := ds.db.QueryRow(sql)

	var seqVal int64
	err := row.Scan(&seqVal)
	if err != nil {
		return seqVal, err
	}

	return seqVal, nil
}
