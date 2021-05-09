package handlers

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/config"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/models"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/render"
)

var app config.AppConfig

// tests is the test case for handlers
var tests = []struct {
	name               string
	path               string
	method             string
	data               url.Values
	expectedStatusCode int
}{
	{"home", "/", "GET", url.Values{}, http.StatusOK},
	{"about", "/about", "GET", url.Values{}, http.StatusOK},
	{"alienware", "/alienware", "GET", url.Values{}, http.StatusOK},
	{"macbook", "/macbook", "GET", url.Values{}, http.StatusOK},
	{"search-availability", "/search-availability", "GET", url.Values{}, http.StatusOK},
	{"contact", "/contact", "GET", url.Values{}, http.StatusOK},
	{"make-reservation", "/make-reservation", "GET", url.Values{}, http.StatusOK},
	{"post-search-availability", "/search-availability", "POST", url.Values{
		"start": {"2020-01-01"},
		"end":   {"2020-01-02"},
	}, http.StatusOK},
	{"post-search-availability-modal", "/search-availability-modal", "POST", url.Values{
		"start": {"2020-01-01"},
		"end":   {"2020-01-02"},
	}, http.StatusOK},
	{"post-make-reservation", "/make-reservation", "POST", url.Values{
		"first_name": {"The"},
		"last_name":  {"Test"},
		"email":      {"test@test.com"},
		"phone":      {"555-5555-5555"},
	}, http.StatusOK},
}

func getRoutes() http.Handler {
	gob.Register(models.Reservation{})

	app.InProduction = false

	app.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	app.Session = scs.New()
	app.Session.Lifetime = 24 * time.Hour
	app.Session.Cookie.Persist = true
	app.Session.Cookie.SameSite = http.SameSiteLaxMode
	app.Session.Cookie.Secure = app.InProduction

	tc, err := render.CreateTemplateCache("./../../templates")
	if err != nil {
		log.Fatal(fmt.Sprintf("cannot create template cache: %s", err))
	}
	app.TemplateCache = tc
	app.UseCache = true // if set to false, render.RenderTemplate will use wrong path for render.CreateTemplateCache

	repo := NewRepo(&app)
	NewHandlers(repo)
	render.NewRenderer(&app)
	mux := chi.NewRouter()

	// middleware
	mux.Use(middleware.Recoverer)
	mux.Use(SessionLoad)

	// endpoint
	mux.Get("/", Repo.Home)
	mux.Get("/about", Repo.About)
	mux.Get("/contact", Repo.Contact)
	mux.Get("/alienware", Repo.Alienware)
	mux.Get("/macbook", Repo.Macbook)
	mux.Get("/search-availability", Repo.SearchAvailability)
	mux.Post("/search-availability", Repo.PostSearchAvailability)
	mux.Post("/search-availability-modal", Repo.SearchAvailabilityModal)
	mux.Get("/make-reservation", Repo.MakeReservation)
	mux.Post("/make-reservation", Repo.PostMakeReservation)
	mux.Get("/reservation-summary", Repo.ReservationSummary)

	// static files
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))
	return mux
}

func SessionLoad(next http.Handler) http.Handler {
	return app.Session.LoadAndSave(next)
}

func TestMain(m *testing.M) {
	fmt.Println("Start testing package: handlers")
	os.Exit(m.Run())
}
