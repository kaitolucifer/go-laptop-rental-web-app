package main

import (
	"testing"

	"github.com/go-chi/chi"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/config"
)

func TestRoutes(t *testing.T) {
	var app config.AppConfig

	mux := routes(&app)
	switch v := mux.(type) {
	case *chi.Mux:
		// do nothing
	default:
		t.Errorf("return type is not *chi.Mux: %s", v)
	}
}
