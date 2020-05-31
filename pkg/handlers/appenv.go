package handlers

import (
	"gitlab.com/xonotic/xonstat/pkg/models"
)

// AppEnv houses the runtime environment for the application. Database connections, cache, etc.
// All web handlers are methods off of the application environment.
type AppEnv struct {
	db models.Datastore
}

// NewAppEnv creates a new AppEnv
func NewAppEnv(db models.Datastore) *AppEnv {
	return &AppEnv{db}
}
