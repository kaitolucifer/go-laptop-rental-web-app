package render

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"text/template"

	"github.com/justinas/nosurf"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/config"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/helpers"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/models"
)

var functions = template.FuncMap{}

var app *config.AppConfig

const PathTemplates = "./templates"

// NewRenderer sets the config for the template package
func NewRenderer(a *config.AppConfig) {
	app = a
}

// AddDefaultData add default data to templates
func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.CSRFToken = nosurf.Token(r)
	if app.Session.Exists(r.Context(), "user_id") {
		td.IsAuthenticated = 1
	}
	return td
}

// Tamplate renders templates using html/template
func Template(w http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) error {
	var tc map[string]*template.Template
	// get the template cache from the app config
	if app.UseCache {
		tc = app.TemplateCache
	} else {
		tc, _ = CreateTemplateCache(PathTemplates)
	}

	t, ok := tc[tmpl]
	if !ok {
		return errors.New("can't get template from cache")
	}

	buf := new(bytes.Buffer)

	_ = t.Execute(buf, AddDefaultData(td, r))

	_, err := buf.WriteTo(w)
	if err != nil {
		helpers.ServerError(w, err)
		return err
	}

	return nil
}

//CreateTemplateCache creates a template cache as a map
func CreateTemplateCache(pathTemplates string) (map[string]*template.Template, error) {
	tmplCache := make(map[string]*template.Template)

	layouts, err := filepath.Glob(fmt.Sprintf("%s/*.layout.html", pathTemplates))
	if err != nil {
		return tmplCache, err
	}

	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.html", pathTemplates))
	if err != nil {
		return tmplCache, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return tmplCache, err
		}

		if len(layouts) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.html", pathTemplates))
			if err != nil {
				return tmplCache, err
			}
		}

		tmplCache[name] = ts
	}

	return tmplCache, nil
}
