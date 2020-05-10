package models

import (
	"database/sql"

	_ "github.com/lib/pq"
)

// IdleConnections is the maximum number of idle connections we should maintain for the database.
const IdleConnections = 5

// Connect creates a PostgreSQL connection
func Connect(connStr string) (*sql.DB, error) {
	// establish a database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	// connection pooling
	db.SetMaxIdleConns(IdleConnections)

	return db, nil
}
