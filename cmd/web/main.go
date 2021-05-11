package main

import (
	"encoding/gob"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/config"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/driver"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/handlers"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/helpers"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/models"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/render"
)

// portNumber is the server port number to use
const portNumber = ":8080"

// app contains all app config
var app config.AppConfig

// main is the main application function
func main() {
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Conn.Close()

	defer close(app.MailChan)
	log.Println("Starting mail listener")
	listenForMail()

	log.Printf("Starting application on port %s\n", portNumber)

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run() (*driver.DB, error) {
	// need to register the data to put into the session
	gob.Register(models.Reservation{})
	gob.Register(models.Restriction{})
	gob.Register(models.User{})
	gob.Register(models.Laptop{})
	gob.Register(models.LaptopRestrictions{})

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

	// change this to true when in production
	app.InProduction = false

	app.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	app.Session = scs.New()
	app.Session.Lifetime = 24 * time.Hour
	app.Session.Cookie.Persist = true
	app.Session.Cookie.SameSite = http.SameSiteLaxMode
	app.Session.Cookie.Secure = app.InProduction

	// connect to database
	app.InfoLog.Println("Connecting to database...")
	db, err := driver.ConnectSQL("host=localhost port=5432 dbname=laptop_rental_app user=postgres password=261519")
	if err != nil {
		log.Printf("Cannot connect to database: %s\n", err)
		return nil, err
	}
	log.Println("Connected to database")

	tc, err := render.CreateTemplateCache(render.PathTemplates)
	if err != nil {
		log.Printf("Cannot create template cache: %s\n", err)
		return db, err
	}
	app.TemplateCache = tc
	app.UseCache = false

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	return db, nil
}
