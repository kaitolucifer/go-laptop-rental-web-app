package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/joho/godotenv"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/config"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/driver"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/handlers"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/helpers"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/models"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/render"
)

// portNumber is the server port number to use
const portNumber = ":8080"

// read flags
var (
	inProduction                                      = flag.Bool("production", true, "Application is in production")
	useCache                                          = flag.Bool("cache", true, "Use template cache")
	dbHost, dbName, dbUser, dbPassword, dbPort, dbSSL string
)

// app contains all app config
var app config.AppConfig

// main is the main application function
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	dbHost = os.Getenv("dbhost")
	dbName = os.Getenv("dbname")
	dbUser = os.Getenv("dbuser")
	dbPassword = os.Getenv("dbpassword")
	dbPort = os.Getenv("dbport")
	dbSSL = os.Getenv("dbssl") // (disbale, prefer, require)

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
	gob.Register(models.LaptopRestriction{})
	gob.Register(map[string]int{})

	flag.Parse()

	if dbName == "" || dbUser == "" {
		log.Fatal("Missing required flags")
		os.Exit(1)
	}

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

	app.InProduction = *inProduction

	app.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	app.Session = scs.New()
	app.Session.Lifetime = 24 * time.Hour
	app.Session.Cookie.Persist = true
	app.Session.Cookie.SameSite = http.SameSiteLaxMode
	app.Session.Cookie.Secure = app.InProduction

	// connect to database
	app.InfoLog.Println("Connecting to database...")
	connectionString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", dbHost, dbPort, dbName, dbUser, dbPassword, dbSSL)
	db, err := driver.ConnectSQL(connectionString)
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
	app.UseCache = *useCache

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	return db, nil
}
