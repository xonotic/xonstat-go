package handlers

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	texttemplate "text/template"
	"time"

	"github.com/alehano/reverse"
	"gitlab.com/xonotic/xonstat/pkg/models"
)

// AppEnv houses the runtime environment for the application. Database connections, cache, etc.
// All web handlers are methods off of the application environment.
type AppEnv struct {
	db            models.Datastore
	cacheEnabled  bool
	cache         models.Cache
	requestLogger io.WriteCloser
	templates     map[string]*template.Template
	textTemplates map[string]*texttemplate.Template
}

func loadTemplates(templateDir string) map[string]*template.Template {
	// The core data structure is a map of filename -> template. We only
	// compile templates for pages which are named with the "page.html" suffix.
	templates := make(map[string]*template.Template)

	// These functions will be available to all templates.
	funcMap := template.FuncMap{
		"urlFor":      reverse.Rev,
		"inc":         func(i int) int { return i + 1 },
		"intToString": func(i int) string { return fmt.Sprintf("%d", i) },
		"unixts": func(t time.Time) string { return fmt.Sprintf("%d", t.Unix()) },
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

// loadTextTemplates compiles the plain-text templates (e.g. protocol responses
// consumed by game servers) that must NOT be HTML-escaped. They live alongside
// the html templates and are distinguished by the .txt suffix.
func loadTextTemplates(templateDir string) map[string]*texttemplate.Template {
	textTemplates := make(map[string]*texttemplate.Template)

	allFiles, err := filepath.Glob(filepath.Join(templateDir, "*.txt"))
	if err != nil {
		log.Print(err)
		return textTemplates
	}

	for _, file := range allFiles {
		t, err := texttemplate.ParseFiles(file)
		if err != nil {
			log.Printf("Error loading text template %s: %s", file, err)
			continue
		}

		textTemplates[filepath.Base(file)] = t
	}

	return textTemplates
}

// NewAppEnv creates a new AppEnv
func NewAppEnv(db models.Datastore, cacheEnabled bool, cache models.Cache, rl io.WriteCloser) *AppEnv {
	cwd, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	templatePath := path.Join(cwd, "web/template")

	ae := AppEnv{
		db:            db,
		cacheEnabled:  cacheEnabled,
		cache:         cache,
		requestLogger: rl,
		templates:     loadTemplates(templatePath),
		textTemplates: loadTextTemplates(templatePath),
	}

	return &ae
}
