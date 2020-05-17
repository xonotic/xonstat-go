package models

import (
	"database/sql"

	_ "github.com/lib/pq" // PostgreSQL-specific datastore implementation
)

// IdleConnections is the maximum number of idle connections we should maintain for the database.
const IdleConnections = 5

// PGDatastore is an implementation of the Datastore interface for a Postgresql database.
type PGDatastore struct {
	db *sql.DB
}

// NewPGDatastore creates a new concrete implementation of a Datastore.
func NewPGDatastore(dsn string) (*PGDatastore, error) {
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
	db.SetMaxIdleConns(IdleConnections)

	return &PGDatastore{db}, nil
}

// Begin starts a transaction.
func (ds *PGDatastore) Begin() (*sql.Tx, error) {
	return ds.db.Begin()
}
