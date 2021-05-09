package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/kaitolucifer/go-laptop-rental-site/internal/driver"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/models"
)

func TestNewRepo(t *testing.T) {
	var db driver.DB
	testRepo := NewRepo(&app, &db)

	if reflect.TypeOf(testRepo).String() != "*handlers.Repository" {
		t.Errorf("Did not get correct type from NewRepo: got %s, wanted *Repository", reflect.TypeOf(testRepo).String())
	}
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	for _, test := range getTests {
		var resp *http.Response
		var err error
		resp, err = ts.Client().Get(ts.URL + test.path)
		if err != nil {
			t.Log(err)
			t.Fatal(err)
		}

		if resp.StatusCode != test.expectedStatusCode {
			t.Errorf("for %s, expected status code %d but got %d", test.name, test.expectedStatusCode, resp.StatusCode)
		}
	}
}

func TestRepository_MakeReservation(t *testing.T) {
	reservation := models.Reservation{
		LaptopID: 1,
		Laptop: models.Laptop{
			ID:         1,
			LaptopName: "Alienware M15 R2",
		},
		StartDate: time.Now().Add(48 * time.Hour),
		EndDate:   time.Now().Add(72 * time.Hour),
	}

	req, _ := http.NewRequest("GET", "/make-reservation", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	app.Session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.MakeReservation)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("MakeReservation handler returned wrong response code: got: %d, expected: %d", rr.Code, http.StatusOK)
	}

	// test case: reservation is not in session
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("MakeReservation handler returned wrong response code: got: %d, expected: %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// test case: non-existent laptop id
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()
	reservation.LaptopID = 100
	app.Session.Put(ctx, "reservation", reservation)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("MakeReservation handler returned wrong response code: got: %d, expected: %d", rr.Code, http.StatusTemporaryRedirect)
	}
}

func TestRepository_PostMakeReservation(t *testing.T) {
	reservation := models.Reservation{
		LaptopID: 1,
		Laptop: models.Laptop{
			ID:         1,
			LaptopName: "Alienware M15 R2",
		},
		StartDate: time.Now().Add(48 * time.Hour),
		EndDate:   time.Now().Add(72 * time.Hour),
	}

	reqBody := "start_date=" + reservation.StartDate.Format("2006-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date="+reservation.EndDate.Format("2006-01-02"))
	reqBody = fmt.Sprintf("%s&%s", reqBody, "first_name=John")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Smith")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=john@smith.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=123456789")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "laptop_id=1")

	req, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx := getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	app.Session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.PostMakeReservation)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostMakeReservation handler returned wrong response code: got: %d, expected: %d", rr.Code, http.StatusSeeOther)
	}

	// test case: reservation is not in session
	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostMakeReservation handler returned wrong response code: got: %d, expected: %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// test case: no request body
	req, _ = http.NewRequest("POST", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()
	app.Session.Put(ctx, "reservation", reservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostMakeReservation handler returned wrong response code: got: %d, expected: %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// test case: invalid start date
	reqBody = "start_date=invalid"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date="+reservation.EndDate.Format("2006-01-02"))
	reqBody = fmt.Sprintf("%s&%s", reqBody, "first_name=John")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Smith")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=john@smith.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=123456789")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "laptop_id=1")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()
	app.Session.Put(ctx, "reservation", reservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostMakeReservation handler returned wrong response code: got: %d, expected: %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// test case: invalid end date
	reqBody = "start_date=" + reservation.StartDate.Format("2006-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date=invalid")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "first_name=John")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Smith")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=john@smith.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=123456789")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "laptop_id=1")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()
	app.Session.Put(ctx, "reservation", reservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostMakeReservation handler returned wrong response code: got: %d, expected: %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// test case: invalid laptop id
	reqBody = "start_date=" + reservation.StartDate.Format("2006-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date="+reservation.EndDate.Format("2006-01-02"))
	reqBody = fmt.Sprintf("%s&%s", reqBody, "first_name=John")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Smith")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=john@smith.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=123456789")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "laptop_id=invalid")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()
	app.Session.Put(ctx, "reservation", reservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostMakeReservation handler returned wrong response code: got: %d, expected: %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// test case: invalid form data
	reqBody = "start_date=" + reservation.StartDate.Format("2006-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date="+reservation.EndDate.Format("2006-01-02"))
	reqBody = fmt.Sprintf("%s&%s", reqBody, "first_name=J")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Smith")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=john@smith.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=123456789")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "laptop_id=1")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()
	app.Session.Put(ctx, "reservation", reservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostMakeReservation handler returned wrong response code: got: %d, expected: %d", rr.Code, http.StatusSeeOther)
	}

	// test case: failure insert reservation into the database
	reqBody = "start_date=" + reservation.StartDate.Format("2006-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date="+reservation.EndDate.Format("2006-01-02"))
	reqBody = fmt.Sprintf("%s&%s", reqBody, "first_name=Test")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Smith")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=john@smith.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=123456789")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "laptop_id=1")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()
	app.Session.Put(ctx, "reservation", reservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostMakeReservation handler returned wrong response code: got: %d, expected: %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// test case: failure insert restriction into the database
	reqBody = "start_date=" + reservation.StartDate.Format("2006-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date="+reservation.EndDate.Format("2006-01-02"))
	reqBody = fmt.Sprintf("%s&%s", reqBody, "first_name=John")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Smith")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=john@smith.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=123456789")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "laptop_id=1000")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()
	app.Session.Put(ctx, "reservation", reservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostMakeReservation handler returned wrong response code: got: %d, expected: %d", rr.Code, http.StatusTemporaryRedirect)
	}
}

func TestRepository_PostSearchAvailability(t *testing.T) {
	reqBody := "start_date=" + time.Now().Add(48*time.Hour).Format("2006-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date="+time.Now().Add(72*time.Hour).Format("2006-01-02"))
	req, _ := http.NewRequest("POST", "/search-availability", strings.NewReader(reqBody))
	ctx := getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Repo.PostSearchAvailability)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("PostSearchAvailability handler returned wrong response code: got: %d, expected: %d", rr.Code, http.StatusOK)
	}

	// test case: failure search availibility from database
	reqBody = "start_date=" + time.Now().Add(72*time.Hour).Format("2006-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date="+time.Now().Add(96*time.Hour).Format("2006-01-02"))
	req, _ = http.NewRequest("POST", "/search-availability", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostSearchAvailability handler returned wrong response code: got: %d, expected: %d", rr.Code, http.StatusSeeOther)
	}

	// test case: no availability
	reqBody = "start_date=" + time.Now().Add(96*time.Hour).Format("2006-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date="+time.Now().Add(120*time.Hour).Format("2006-01-02"))
	req, _ = http.NewRequest("POST", "/search-availability", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostSearchAvailability handler returned wrong response code: got: %d, expected: %d", rr.Code, http.StatusSeeOther)
	}

	// test case: invalid request body
	req, _ = http.NewRequest("POST", "/search-availability", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.PostSearchAvailability)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostSearchAvailability handler returned wrong response code: got: %d, expected: %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// test case: invalid start date
	reqBody = "start_date=invalid"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date="+time.Now().Add(48*time.Hour).Format("2006-01-02"))
	req, _ = http.NewRequest("POST", "/search-availability", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostSearchAvailability handler returned wrong response code: got: %d, expected: %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// test case: invalid end date
	reqBody = "start_date=" + time.Now().Add(48*time.Hour).Format("2006-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date=invalid")
	req, _ = http.NewRequest("POST", "/search-availability", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostSearchAvailability handler returned wrong response code: got: %d, expected: %d", rr.Code, http.StatusTemporaryRedirect)
	}
}

func TestRepository_SearchAvailabilityModal(t *testing.T) {
	reqBody := "start_date=" + time.Now().Add(48*time.Hour).Format("2006-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date="+time.Now().Add(72*time.Hour).Format("2006-01-02"))
	reqBody = fmt.Sprintf("%s&%s", reqBody, "laptop_id=1")
	req, _ := http.NewRequest("POST", "/search-availability-modal", strings.NewReader(reqBody))
	ctx := getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Repo.SearchAvailabilityModal)
	handler.ServeHTTP(rr, req)

	var jr jsonResponse
	err := json.Unmarshal(rr.Body.Bytes(), &jr)
	if err != nil {
		t.Error("failed to parse json")
	}
	if !jr.OK {
		t.Errorf("SearchAvailabilityModal handler returned wrong JSON message: got: %v, expected: %v", jr.OK, !jr.OK)
	}

	// test case: invalid start date
	reqBody = "start_date=invalid"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date="+time.Now().Add(72*time.Hour).Format("2006-01-02"))
	reqBody = fmt.Sprintf("%s&%s", reqBody, "laptop_id=1")
	req, _ = http.NewRequest("POST", "/search-availability-modal", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	err = json.Unmarshal(rr.Body.Bytes(), &jr)
	if err != nil {
		fmt.Println(rr.Body.String())
		t.Error("failed to parse json")
	}
	if jr.OK {
		t.Errorf("SearchAvailabilityModal handler returned wrong JSON message: got: %v, expected: %v", jr.OK, !jr.OK)
	}

	// test case: invalid end date
	reqBody = "start_date=" + time.Now().Add(48*time.Hour).Format("2006-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date=invalid")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "laptop_id=1")
	req, _ = http.NewRequest("POST", "/search-availability-modal", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	err = json.Unmarshal(rr.Body.Bytes(), &jr)
	if err != nil {
		t.Error("failed to parse json")
	}
	if jr.OK {
		t.Errorf("SearchAvailabilityModal handler returned wrong JSON message: got: %v, expected: %v", jr.OK, !jr.OK)
	}

	// test case: invalid laptop id
	reqBody = "start_date=" + time.Now().Add(48*time.Hour).Format("2006-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date="+time.Now().Add(72*time.Hour).Format("2006-01-02"))
	reqBody = fmt.Sprintf("%s&%s", reqBody, "laptop_id=invalid")
	req, _ = http.NewRequest("POST", "/search-availability-modal", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	err = json.Unmarshal(rr.Body.Bytes(), &jr)
	if err != nil {
		t.Error("failed to parse json")
	}
	if jr.OK {
		t.Errorf("SearchAvailabilityModal handler returned wrong JSON message: got: %v, expected: %v", jr.OK, !jr.OK)
	}

	// test case: laptop not available
	reqBody = "start_date=" + time.Now().Add(48*time.Hour).Format("2006-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date="+time.Now().Add(72*time.Hour).Format("2006-01-02"))
	reqBody = fmt.Sprintf("%s&%s", reqBody, "laptop_id=1000")
	req, _ = http.NewRequest("POST", "/search-availability-modal", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	err = json.Unmarshal(rr.Body.Bytes(), &jr)
	if err != nil {
		t.Error("failed to parse json")
	}
	if jr.OK {
		t.Errorf("SearchAvailabilityModal handler returned wrong JSON message: got: %v, expected: %v", jr.OK, !jr.OK)
	}

	// test case: failure search availability from database
	reqBody = "start_date=" + time.Now().Add(48*time.Hour).Format("2006-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date="+time.Now().Add(72*time.Hour).Format("2006-01-02"))
	reqBody = fmt.Sprintf("%s&%s", reqBody, "laptop_id=5000")
	req, _ = http.NewRequest("POST", "/search-availability-modal", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	err = json.Unmarshal(rr.Body.Bytes(), &jr)
	if err != nil {
		t.Error("failed to parse json")
	}
	if jr.OK {
		t.Errorf("SearchAvailabilityModal handler returned wrong JSON message: got: %v, expected: %v", jr.OK, !jr.OK)
	}

	// test case: no request body
	req, _ = http.NewRequest("POST", "/search-availability-modal", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	err = json.Unmarshal(rr.Body.Bytes(), &jr)
	if err != nil {
		t.Error("failed to parse json")
	}
	if jr.OK {
		t.Errorf("SearchAvailabilityModal handler returned wrong JSON message: got: %v, expected: %v", jr.OK, !jr.OK)
	}
}

func TestRepository_ReservationSummary(t *testing.T) {
	reservation := models.Reservation{
		LaptopID: 1,
		Laptop: models.Laptop{
			ID:         1,
			LaptopName: "Alienware M15 R2",
		},
	}

	req, _ := http.NewRequest("GET", "/reservation-summary", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	app.Session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.ReservationSummary)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ReservationSummary handler returned wrong response code: got: %d, expected: %d", rr.Code, http.StatusOK)
	}

	// test case: reservation not in session
	req, _ = http.NewRequest("GET", "/reservation-summary", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.ReservationSummary)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("ReservationSummary handler returned wrong response code: got: %d, expected: %d", rr.Code, http.StatusOK)
	}
}

func TestRepository_ChooseLaptop(t *testing.T) {
	reservation := models.Reservation{
		LaptopID: 1,
		Laptop: models.Laptop{
			ID:         1,
			LaptopName: "Alienware M15 R2",
		},
	}

	req, _ := http.NewRequest("GET", "/choose-laptop/1", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)
	// set the RequestURI on the request to get url parameters
	req.RequestURI = "/choose-laptop/1"

	rr := httptest.NewRecorder()
	app.Session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.ChooseLaptop)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("ChooseLaptop handler returned wrong response code: got: %d, expected: %d", rr.Code, http.StatusSeeOther)
	}

	// test case: reservation not in session
	req, _ = http.NewRequest("GET", "/choose-laptop/1", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.RequestURI = "/choose-laptop/1"

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.ChooseLaptop)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("ChooseLaptop handler returned wrong response code: got: %d, expected: %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// test case: missing url parameter, or malformed parameter
	req, _ = http.NewRequest("GET", "/choose-laptop/fish", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.RequestURI = "/choose-laptop/fish"

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.ChooseLaptop)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("ChooseLaptop handler returned wrong response code: got: %d, expected: %d", rr.Code, http.StatusTemporaryRedirect)
	}
}

func TestRepository_RentLaptop(t *testing.T) {
	reservation := models.Reservation{
		LaptopID: 1,
		Laptop: models.Laptop{
			ID:         1,
			LaptopName: "Alienware M15 R2",
		},
	}

	req, _ := http.NewRequest("GET", "/rent-laptop?s=2050-01-01&e=2050-01-02&id=1", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	app.Session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.RentLaptop)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("RentLaptop handler returned wrong response code: got: %d, expected: %d", rr.Code, http.StatusSeeOther)
	}

	// test case: database failed
	req, _ = http.NewRequest("GET", "/rent-laptop?s=2040-01-01&e=2040-01-02&id=4", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.RentLaptop)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("RentLaptop handler returned wrong response code: got: %d, expected %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// test case: invalid laptop id
	req, _ = http.NewRequest("GET", "/rent-laptop?s=2040-01-01&e=2040-01-02&id=invalid", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.RentLaptop)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("RentLaptop handler returned wrong response code: got: %d, expected %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// test case: invalid start date
	req, _ = http.NewRequest("GET", "/rent-laptop?s=invalid&e=2040-01-02&id=1", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.RentLaptop)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("RentLaptop handler returned wrong response code: got: %d, expected %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// test case: invalid end date
	req, _ = http.NewRequest("GET", "/rent-laptop?s=2040-01-01&e=invalid&id=1", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.RentLaptop)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("RentLaptop handler returned wrong response code: got: %d, expected %d", rr.Code, http.StatusTemporaryRedirect)
	}
}
