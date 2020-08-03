package handlers

import (
	"html/template"
	"io"
	"log"
	"path/filepath"
	"strings"

	"github.com/alehano/reverse"
	"gitlab.com/xonotic/xonstat/pkg/models"
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

	baseTemplate := template.New("base")
	baseTemplate.Funcs(template.FuncMap{"urlFor": reverse.Rev})

	// Separate pages from partials and layouts
	allFiles, err := filepath.Glob(filepath.Join(templateDir, "*html"))
	if err != nil {
		log.Print(err)
	}

	var baseFilenames, pageFilenames []string
	for _, t := range allFiles {
		if strings.Contains(t, "page") {
			pageFilenames = append(pageFilenames, t)
		} else {
			baseFilenames = append(baseFilenames, t)
		}
	}

	for _, pageFilename := range pageFilenames {
		basePageFilename := filepath.Base(pageFilename)
		filenames := append([]string{pageFilename}, baseFilenames...)

		// If we don't clone the base template, the calls to ParseFiles are accumulated and will
		// result in template compilation errors (e.g. leaderboard has a non-blank hero_unit that
		// isn't present in other pages)
		t, _ := baseTemplate.Clone()
		templates[basePageFilename], err = t.ParseFiles(filenames...)
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
