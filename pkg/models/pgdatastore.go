package models

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL-specific datastore implementation
)

// IdleConnections is the maximum number of idle connections we should maintain for the database.
const IdleConnections = 2

// OpenConnections is the maximum number of open connections we should maintain for the database.
const OpenConnections = 5

// PGDatastore is an implementation of the Datastore interface for a Postgresql database.
type PGDatastore struct {
	db *sql.DB
}

// NewPGDatastore creates a new concrete implementation of a Datastore.
func NewPGDatastore(dsn string) (*PGDatastore, error) {
	// establish a database connection
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	// connection pooling
	db.SetMaxIdleConns(IdleConnections)
	db.SetMaxOpenConns(OpenConnections)

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
