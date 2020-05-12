package models

import (
	"database/sql"

	_ "github.com/lib/pq"
)

// IdleConnections is the maximum number of idle connections we should maintain for the database.
const IdleConnections = 5

// Datastore is the interface representing all database operations.
// This is useful for mocking or implementing different backends.
type Datastore interface {
	RServersByHashkey(hashkey string) ([]*Server, error)
	RServersByName(name string) ([]*Server, error)
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
