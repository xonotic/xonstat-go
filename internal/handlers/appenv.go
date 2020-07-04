package handlers

import (
	"html/template"
	"io"
	"gitlab.com/xonotic/xonstat/pkg/models"
)

// AppEnv houses the runtime environment for the application. Database connections, cache, etc.
// All web handlers are methods off of the application environment.
type AppEnv struct {
	db models.Datastore
	requestLogger io.WriteCloser
	templates *template.Template 
}

// NewAppEnv creates a new AppEnv
func NewAppEnv(db models.Datastore, rl io.WriteCloser, templateDirGlob string) *AppEnv {
	ae := AppEnv{
		db: db, 
		requestLogger: rl,
	}
	ae.templates = template.Must(template.ParseGlob(templateDirGlob))
	return &ae
}