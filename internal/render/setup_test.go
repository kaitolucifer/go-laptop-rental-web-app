package render

import (
	"encoding/gob"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/config"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/models"
)

var testApp config.AppConfig

type testWriter struct{}

func (tw *testWriter) Header() http.Header {
	return http.Header{}
}

func (tw *testWriter) WriteHeader(i int) {

}

func (tw *testWriter) Write(b []byte) (int, error) {
	length := len(b)
	return length, nil
}

func getSession() (*http.Request, error) {
	r, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		return nil, err
	}

	ctx := r.Context()
	ctx, _ = testApp.Session.Load(ctx, r.Header.Get("X-Session"))
	r = r.WithContext(ctx)
	return r, nil
}

func TestMain(m *testing.M) {
	// need to register the data to put into the session
	gob.Register(models.Reservation{})

	// change this to true when in production
	testApp.InProduction = false

	testApp.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	testApp.ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	testApp.Session = scs.New()
	testApp.Session.Lifetime = 24 * time.Hour
	testApp.Session.Cookie.Persist = true
	testApp.Session.Cookie.SameSite = http.SameSiteLaxMode
	testApp.Session.Cookie.Secure = false

	app = &testApp

	os.Exit(m.Run())
}
