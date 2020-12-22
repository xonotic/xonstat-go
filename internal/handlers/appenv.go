package handlers

import (
	"fmt"
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
	// The core data structure is a map of filename -> template. We only
	// compile templates for pages which are named with the "page.html" suffix.
	templates := make(map[string]*template.Template)

	// These functions will be available to all templates.
	funcMap := template.FuncMap{
		"urlFor": reverse.Rev,
		"inc":    func(i int) int { return i + 1 },
		"intToString":    func(i int) string { return fmt.Sprintf("%d", i) },
	}

	// First, grab the filenames of all of the templates in the directory.
	allFiles, err := filepath.Glob(filepath.Join(templateDir, "*html"))
	if err != nil {
		log.Print(err)
	}

	// Then we separate out the pages from the rest to avoid re-defining blocks multiple times
	// which leads to unexpected renders. 
	var baseFilenames, pageFilenames []string
	for _, t := range allFiles {
		if strings.Contains(t, "page") {
			pageFilenames = append(pageFilenames, t)
		} else {
			baseFilenames = append(baseFilenames, t)
		}
	}

	// Build the map such that templates[page filename] yields the compiled template.
	for _, pageFilename := range pageFilenames {
		basePageFilename := filepath.Base(pageFilename)
		filenames := append([]string{pageFilename}, baseFilenames...)

		// This is needed as a preprocessing step, else we get template compilation errors
		// for functions that are used by not defined yet (like urlFor).
		sharedTemplate := template.New(basePageFilename)
		sharedTemplate.Funcs(funcMap)

		t, err := sharedTemplate.ParseFiles(filenames...)
		if err != nil {
			log.Println(err)
		}

		templates[basePageFilename] = t
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
