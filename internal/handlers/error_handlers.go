package handlers

import (
	"net/http"

	"github.com/kaitolucifer/go-laptop-rental-site/internal/models"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/render"
)

func (repo *Repository) NotFound(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["url"] = r.URL.Path
	render.Template(w, r, "404.page.html", &models.TemplateData{
		StringMap: stringMap,
	})
}
