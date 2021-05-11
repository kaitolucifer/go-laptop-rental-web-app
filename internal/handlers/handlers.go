package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/kaitolucifer/go-laptop-rental-site/internal/config"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/database"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/driver"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/forms"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/models"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/render"
)

// Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
	DB  database.Database
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  database.NewPostgres(db.Conn, a),
	}
}

// NewMockRepo creates a new repository
func NewMockRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
		DB:  database.NewMockPostgres(a),
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Home renders home page
func (repo *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "home.page.html", &models.TemplateData{})
}

// About renders about page
func (repo *Repository) About(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "about.page.html", &models.TemplateData{})
}

// Contact renders the contact page
func (repo *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "contact.page.html", &models.TemplateData{})
}

// MakeReservation renders the make a reservation page and displays form
func (repo *Repository) MakeReservation(w http.ResponseWriter, r *http.Request) {
	res, ok := repo.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		repo.App.Session.Put(r.Context(), "error", "can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	laptop, err := repo.DB.GetLaptopByID(res.LaptopID)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't find laptop by ID")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res.Laptop.LaptopName = laptop.LaptopName

	repo.App.Session.Put(r.Context(), "reservation", res)

	stringMap := make(map[string]string)
	stringMap["start_date"] = res.StartDate.Format("2006-01-02")
	stringMap["end_date"] = res.EndDate.Format("2006-01-02")

	data := make(map[string]interface{})
	data["reservation"] = res
	render.Template(w, r, "make-reservation.page.html", &models.TemplateData{
		Form:      forms.New(nil),
		Data:      data,
		StringMap: stringMap,
	})
}

// PostMakeReservation handles the posting of a reservation form
func (repo *Repository) PostMakeReservation(w http.ResponseWriter, r *http.Request) {
	reservation, ok := repo.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		repo.App.Session.Put(r.Context(), "error", "can't get reservation from session")
		r.Method = "GET"
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	err := r.ParseForm()
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't parse form")
		r.Method = "GET"
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	form := forms.New(r.PostForm)

	form.Required("first_name", "last_name", "email", "start_date", "end_date")
	form.IsAboveMinLength("first_name", 3)
	form.IsEmail("email")
	form.ValidateDate("start_date")
	form.ValidateDate("end_date")
	form.EndDateGreaterThanStartDate("start_date", "end_date")

	startDate, err := form.GetTimeObj("start_date")
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't get reservation from session")
		r.Method = "GET"
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	endDate, err := form.GetTimeObj("end_date")
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't parse start date")
		r.Method = "GET"
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	laptopID, err := strconv.Atoi(r.Form.Get("laptop_id"))
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "invalid data")
		r.Method = "GET"
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Email = r.Form.Get("email")
	reservation.Phone = r.Form.Get("phone")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		http.Error(w, "form invalid", http.StatusSeeOther)
		r.Method = "GET"
		render.Template(w, r, "make-reservation.page.html", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	newReservationID, err := repo.DB.InsertReservation(&reservation)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't insert reservation into the database")
		r.Method = "GET"
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	restriction := models.LaptopRestrictions{
		StartDate:     startDate,
		EndDate:       endDate,
		LaptopID:      laptopID,
		ReservationID: newReservationID,
		RestrictionID: 1,
	}
	err = repo.DB.InsertLaptopRestriction(&restriction)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't insert laptop restriction into the database")
		r.Method = "GET"
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// send notification mail to user
	htmlMessage := fmt.Sprintf(`
	<strong>Reservation Confirmation</strong><br>
	Dear %s:, <br>
	This is a confirmation of your reservation from %s to %s.
	`, reservation.FirstName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))
	mail := models.MailData{
		To:       reservation.Email,
		From:     "kaito@laptop-rental.com",
		Subject:  "Reservation Confirmation",
		Content:  htmlMessage,
		Template: "basic.email.html",
	}
	repo.App.MailChan <- mail

	// send notification mail to website Administrator
	htmlMessage = fmt.Sprintf(`
	<strong>Reservation Confirmation</strong><br>
	A reservation has been made for %s from %s to %s.
	`, reservation.Laptop.LaptopName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))
	mail = models.MailData{
		To:       "kaito@laptop-rental.com",
		From:     "kaito@laptop-rental.com",
		Subject:  "Reservation Confirmation",
		Content:  htmlMessage,
		Template: "basic.email.html",
	}
	repo.App.MailChan <- mail

	repo.App.Session.Put(r.Context(), "reservation", reservation)
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

// Alienware renders the Alienware laptop page
func (repo *Repository) Alienware(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "alienware.page.html", &models.TemplateData{})
}

// Macbook renders the Macbook laptop page
func (repo *Repository) Macbook(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "macbook.page.html", &models.TemplateData{})
}

// SearchAvailability renders the search availalibity page
func (repo *Repository) SearchAvailability(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "search-availability.page.html", &models.TemplateData{})
}

// PostSearchAvailability handles request for availability
func (repo *Repository) PostSearchAvailability(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't parse form")
		r.Method = "GET"
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	form := forms.New(r.PostForm)

	startDate, err := form.GetTimeObj("start_date")
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't parse start date")
		r.Method = "GET"
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	endDate, err := form.GetTimeObj("end_date")
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't parse end date")
		r.Method = "GET"
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	laptops, err := repo.DB.SearchAvailabilityForAllLaptops(startDate, endDate)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't search availability")
		r.Method = "GET"
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if len(laptops) == 0 {
		repo.App.Session.Put(r.Context(), "error", "no availability")
		r.Method = "GET"
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["laptops"] = laptops

	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	repo.App.Session.Put(r.Context(), "reservation", res)

	render.Template(w, r, "choose-laptop.page.html", &models.TemplateData{
		Data: data,
	})
}

// jsonResponse defines the schema of JSON repsonse sent by AvailabilityModal handler
type jsonResponse struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	LaptopID  string `json:"laptop_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// SearchAvailabilityModal handles request for availability on modal window and send JSON response
func (repo *Repository) SearchAvailabilityModal(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "Internal server error",
		}
		out, _ := json.MarshalIndent(resp, "", "     ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	form := forms.New(r.PostForm)

	form.Required("start_date", "end_date")
	form.ValidateDate("start_date")
	form.ValidateDate("end_date")
	form.EndDateGreaterThanStartDate("start_date", "end_date")

	startDate, err := form.GetTimeObj("start_date")
	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "Invalid Start Date",
		}
		out, _ := json.MarshalIndent(resp, "", "     ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}
	endDate, err := form.GetTimeObj("end_date")
	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "Invalid End Date",
		}
		out, _ := json.MarshalIndent(resp, "", "     ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	laptopID, err := strconv.Atoi(r.Form.Get("laptop_id"))
	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "Invalid Laptop ID",
		}
		out, _ := json.MarshalIndent(resp, "", "     ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	available, err := repo.DB.SearchAvailabilityByDatesByLaptopID(startDate, endDate, laptopID)
	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "Error connecting to the database",
		}
		out, _ := json.MarshalIndent(resp, "", "     ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	msg := "Available!"
	if !available {
		msg = "Not Available!"
	}

	resp := jsonResponse{
		OK:        available,
		Message:   msg,
		StartDate: r.Form.Get("start_date"),
		EndDate:   r.Form.Get("end_date"),
		LaptopID:  r.Form.Get("laptop_id"),
	}

	// the validity of the json response is certain at this point
	out, _ := json.MarshalIndent(resp, "", "     ")

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

// ReservationSummary displays the reservation summary page
func (repo *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := repo.App.Session.Get(r.Context(), "reservation").(models.Reservation) // type assertion
	if !ok {
		repo.App.Session.Put(r.Context(), "error", "can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	stringMap := make(map[string]string)
	stringMap["start_date"] = reservation.StartDate.Format("2006-01-02")
	stringMap["end_date"] = reservation.EndDate.Format("2006-01-02")

	repo.App.Session.Remove(r.Context(), "reservation")
	data := make(map[string]interface{})
	data["reservation"] = reservation
	render.Template(w, r, "reservation-summary.page.html", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})
}

// ChooseLaptop displays list of available laptops
func (repo *Repository) ChooseLaptop(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")
	laptopID, err := strconv.Atoi(exploded[2])
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "invalid Laptop ID")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	res, ok := repo.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		repo.App.Session.Put(r.Context(), "error", "can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res.LaptopID = laptopID
	repo.App.Session.Put(r.Context(), "reservation", res)
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

// RentLaptop takes URL parameters, builds a sessional variable, and takes user to make reservation page
func (repo *Repository) RentLaptop(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	LaptopID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "invalid Laptop ID")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	form := forms.New(r.Form)
	startDate, err := form.GetTimeObj("s")
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't parse start date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	endDate, err := form.GetTimeObj("e")
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't parse end date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	var res models.Reservation
	res.LaptopID = LaptopID
	res.StartDate = startDate
	res.EndDate = endDate

	laptop, err := repo.DB.GetLaptopByID(res.LaptopID)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "error connecting to the database")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res.Laptop.LaptopName = laptop.LaptopName

	repo.App.Session.Put(r.Context(), "reservation", res)
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

// Login shows the login page
func (repo *Repository) Login(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "login.page.html", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// PostLogin handles logging the user in 
func (repo *Repository) PostLogin(w http.ResponseWriter, r *http.Request) {
	_ = repo.App.Session.RenewToken(r.Context()) // to prevent session fixation attack

	err := r.ParseForm()
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't parse form")
		r.Method = "GET"
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")
	if !form.Valid() {
		render.Template(w, r, "login.page.html", &models.TemplateData{
			Form: form,
		})
		return 
	}
	email := r.Form.Get("email")
	password := r.Form.Get("password")

	id, _, err := repo.DB.Authenticate(email, password)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "invalid login credentials")
		r.Method = "GET"
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	repo.App.Session.Put(r.Context(), "user_id", id)
	repo.App.Session.Put(r.Context(), "flash", "Logged in successfully")
	r.Method = "GET"
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout logs a user out
func (repo *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = repo.App.Session.Destroy(r.Context())
	_ = repo.App.Session.RenewToken(r.Context())
	repo.App.Session.Put(r.Context(), "flash", "Logged out successfully")

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (repo *Repository) AdminDashbord(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-dashboard.page.html", &models.TemplateData{})
}
