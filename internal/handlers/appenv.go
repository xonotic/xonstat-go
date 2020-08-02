package handlers

import (
	"gitlab.com/xonotic/xonstat/pkg/models"
	"html/template"
	"io"
	"log"
	"path/filepath"
)

// AppEnv houses the runtime environment for the application. Database connections, cache, etc.
// All web handlers are methods off of the application environment.
type AppEnv struct {
	db            models.Datastore
	requestLogger io.WriteCloser
	templates     map[string]*template.Template
}

func loadTemplates(templateDir string) map[string]*template.Template {
	templates := make(map[string]*template.Template)

	// Layouts are the base files upon which lots of other ones are built.
	layouts, err := filepath.Glob(filepath.Join(templateDir, "*layout*"))
	if err != nil {
		log.Print(err)
	}

	// Partials are little bits of reusable components.
	partials, err := filepath.Glob(filepath.Join(templateDir, "*partial*"))
	if err != nil {
		log.Print(err)
	}

	// Pages are the ones that utilize the two types above, redefining pieces
	// to form a whole, standalone page.
	pages, err := filepath.Glob(filepath.Join(templateDir, "*page*"))
	if err != nil {
		log.Print(err)
	}

	baseFiles := append(layouts, partials...)

	for _, page := range pages {
		fileName := filepath.Base(page)
		filenames := append([]string{page}, baseFiles...)
		templates[fileName], err = template.ParseFiles(filenames...)
		if err != nil {
			log.Println(err)
		}
	}

	return templates
}

// NewAppEnv creates a new AppEnv
func NewAppEnv(db models.Datastore, rl io.WriteCloser) *AppEnv {
	ae := AppEnv{
		db:            db,
		requestLogger: rl,
		templates:     loadTemplates("web/template"),
	}

	return &ae
}
