package handlers

import (
	"context"
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
var getTests = []struct {
	name               string
	path               string
	expectedStatusCode int
}{
	{"home", "/", http.StatusOK},
	{"about", "/about", http.StatusOK},
	{"alienware", "/alienware", http.StatusOK},
	{"macbook", "/macbook", http.StatusOK},
	{"search-availability", "/search-availability", http.StatusOK},
	{"contact", "/contact", http.StatusOK},
	{"non-existent", "/not/exist", http.StatusNotFound},
	{"login", "/user/login", http.StatusOK},
	{"logout", "/user/logout", http.StatusOK},
	{"dashboard", "/admin/dashboard", http.StatusOK},
	{"new reservations", "/admin/reservations-new", http.StatusOK},
	{"all reservations", "/admin/reservations-all", http.StatusOK},
	{"show reservation", "/admin/reservations/new/1/show", http.StatusOK},
}

var loginTests = []struct {
	name               string
	email              string
	expectedStatusCode int
	expectedHTML       string
	expectedLocation   string
}{
	{
		"valid-credentials",
		"test@test.com",
		http.StatusSeeOther,
		"",
		"/",
	},
	{
		"invalid-credentials",
		"failed@test.com",
		http.StatusSeeOther,
		"",
		"/user/login",
	},
	{
		"invalid-data",
		"test",
		http.StatusOK,
		`action="/user/login"`,
		"",
	},
}

var adminDeleteReservationTests = []struct {
	name                 string
	queryParams          string
	expectedResponseCode int
	expectedLocation     string
}{
	{
		name:                 "delete-reservation",
		queryParams:          "",
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "",
	},
	{
		name:                 "delete-reservation-back-to-cal",
		queryParams:          "?y=2021&m=12",
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "",
	},
}

var adminProcessReservationTests = []struct {
	name                 string
	queryParams          string
	expectedResponseCode int
	expectedLocation     string
}{
	{
		name:                 "process-reservation",
		queryParams:          "",
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "",
	},
	{
		name:                 "process-reservation-back-to-cal",
		queryParams:          "?y=2021&m=12",
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "",
	},
}

var postAdminReservationCalendarTests = []struct {
	name                 string
	postedData           url.Values
	expectedResponseCode int
	expectedLocation     string
	expectedHTML         string
	blocks               int
	reservations         int
}{
	{
		name: "cal",
		postedData: url.Values{
			"year":  {time.Now().Format("2006")},
			"month": {time.Now().Format("01")},
			fmt.Sprintf("add_block_1_%s", time.Now().AddDate(0, 0, 2).Format("2006-01-2")): {"1"},
		},
		expectedResponseCode: http.StatusSeeOther,
	},
	{
		name:                 "cal-blocks",
		postedData:           url.Values{},
		expectedResponseCode: http.StatusSeeOther,
		blocks:               1,
	},
	{
		name:                 "cal-res",
		postedData:           url.Values{},
		expectedResponseCode: http.StatusSeeOther,
		reservations:         1,
	},
}

var PostAdminShowReservationTests = []struct {
	name                 string
	url                  string
	postedData           url.Values
	expectedResponseCode int
	expectedLocation     string
	expectedHTML         string
}{
	{
		name: "valid-data-from-new",
		url:  "/admin/reservations/new/1/show",
		postedData: url.Values{
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"555-555-5555"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "/admin/reservations-new",
		expectedHTML:         "",
	},
	{
		name: "valid-data-from-all",
		url:  "/admin/reservations/all/1/show",
		postedData: url.Values{
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"555-555-5555"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "/admin/reservations-all",
		expectedHTML:         "",
	},
	{
		name: "valid-data-from-calendar",
		url:  "/admin/reservations/calendar/1/show",
		postedData: url.Values{
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"555-555-5555"},
			"year":       {"2022"},
			"month":      {"01"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "/admin/reservations-calendar?y=2022&m=01",
		expectedHTML:         "",
	},
}

func TestMain(m *testing.M) {
	gob.Register(models.Reservation{})
	gob.Register(models.Restriction{})
	gob.Register(models.User{})
	gob.Register(models.Laptop{})
	gob.Register(models.LaptopRestriction{})
	gob.Register(map[string]int{})

	app.InProduction = false

	app.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	app.Session = scs.New()
	app.Session.Lifetime = 24 * time.Hour
	app.Session.Cookie.Persist = true
	app.Session.Cookie.SameSite = http.SameSiteLaxMode
	app.Session.Cookie.Secure = app.InProduction

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan
	defer close(mailChan)

	listenForMail()

	tc, err := render.CreateTemplateCache("./../../templates")
	if err != nil {
		log.Fatal(fmt.Sprintf("cannot create template cache: %s", err))
	}
	app.TemplateCache = tc
	app.UseCache = true // if set to false, render.RenderTemplate will use wrong path for render.CreateTemplateCache

	repo := NewMockRepo(&app)
	NewHandlers(repo)
	render.NewRenderer(&app)

	os.Exit(m.Run())
}

func listenForMail() {
	go func() {
		for {
			<-app.MailChan
		}
	}()
}

func getRoutes() http.Handler {
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
	mux.Get("/make-reservation", Repo.MakeReservation)
	mux.Get("/reservation-summary", Repo.ReservationSummary)

	mux.Get("/user/login", Repo.Login)
	mux.Get("/user/logout", Repo.Logout)

	mux.Route("/admin", func(mux chi.Router) {
		mux.Get("/dashboard", Repo.AdminDashbord)

		mux.Get("/reservations-new", Repo.AdminNewReservations)
		mux.Get("/reservations-all", Repo.AdminAllReservations)
		mux.Get("/reservations/{type}/{id}/show", Repo.AdminShowReservation)
		mux.Get("/process-reservation/{type}/{id}/do", Repo.AdminProcessReservation)
		mux.Get("/delete-reservation/{type}/{id}/do", Repo.AdminDeleteReservation)
		mux.Get("/reservations-calendar", Repo.AdminReservationsCalendar)
		mux.Post("/reservations-calendar", Repo.PostAdminReservationsCalendar)
	})

	// static files
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))
	return mux
}

func SessionLoad(next http.Handler) http.Handler {
	return app.Session.LoadAndSave(next)
}

func getCtx(req *http.Request) context.Context {
	ctx, err := app.Session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}
	return ctx
}
