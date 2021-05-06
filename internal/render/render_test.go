package render

import (
	"testing"

	"github.com/kaitolucifer/go-laptop-rental-site/internal/models"
)

func TestAddDefaultData(t *testing.T) {
	var td models.TemplateData

	r, err := getSession()
	if err != nil {
		t.Error(err)
	}
	session.Put(r.Context(), "flash", "123")
	result := AddDefaultData(&td, r)
	if result.Flash != "123" {
		t.Error("flash value not found in session")
	}
}

func TestRenderTemplate(t *testing.T) {
	PathTemplates := "./../../templates"
	tc, err := CreateTemplateCache(PathTemplates)
	if err != nil {
		t.Error(err)
	}

	app.TemplateCache = tc

	r, err := getSession()
	if err != nil {
		t.Error(err)
	}

	var tw testWriter
	app.UseCache = false
	err = RenderTemplate(&tw, r, "home.page.html", &models.TemplateData{})
	if err == nil {
		t.Error(err)
	}

	app.UseCache = true
	err = RenderTemplate(&tw, r, "home.page.html", &models.TemplateData{})
	if err != nil {
		t.Error(err)
	}

	err = RenderTemplate(&tw, r, "non-existent.page.html", &models.TemplateData{})
	if err == nil {
		t.Error("rendered template that does not exist")
	}
}

func TestNewTemplates(t *testing.T) {
	NewTemplates(app)
}

func TestCreateTemplateCache(t *testing.T) {
	_, err := CreateTemplateCache("./../../templates")
	if err != nil {
		t.Error(err)
	}
}
