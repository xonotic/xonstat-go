package handlers

import (
	"io"
	"gitlab.com/xonotic/xonstat/pkg/models"
)

// AppEnv houses the runtime environment for the application. Database connections, cache, etc.
// All web handlers are methods off of the application environment.
type AppEnv struct {
	db models.Datastore
	requestLogger io.WriteCloser
}

// NewAppEnv creates a new AppEnv
func NewAppEnv(db models.Datastore, rl io.WriteCloser) *AppEnv {
	return &AppEnv{db, rl}
}
