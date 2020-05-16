package models

import (
	"database/sql"

	_ "github.com/lib/pq" // PostgreSQL-specific datastore implementation
)

// IdleConnections is the maximum number of idle connections we should maintain for the database.
const IdleConnections = 5

// Datastore is the interface representing all database operations.
// This is useful for mocking or implementing different backends.
type Datastore interface {
	Begin() (*sql.Tx, error)

	// Server-oriented methods
	CServer(tx *sql.Tx, server Server) (int64, error)
	RServersByHashkey(hashkey string) ([]*Server, error)
	RServersByName(name string) ([]*Server, error)
	UServer(tx *sql.Tx, server Server) error

	// Map-oriented methods
	CMap(tx *sql.Tx, m Map) (int64, error)
	RMapsByName(name string) ([]*Map, error)
}

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
