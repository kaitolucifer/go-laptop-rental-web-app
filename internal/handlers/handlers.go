package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kaitolucifer/go-laptop-rental-site/internal/config"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/database"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/driver"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/forms"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/helpers"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/models"
	"github.com/kaitolucifer/go-laptop-rental-site/internal/render"
)

// Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
	DB  database.DBRepository
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
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	laptop, err := repo.DB.GetLaptopByID(res.LaptopID)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't find laptop by ID")
		http.Redirect(w, r, "/", http.StatusSeeOther)
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
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	err := r.ParseForm()
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't parse form")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	form := forms.New(r.PostForm)

	form.Required("first_name", "last_name", "email", "start_date", "end_date")
	form.IsAboveMinLength("first_name", 3)
	form.IsEmail("email")
	form.ValidateDate("start_date")
	form.ValidateDate("end_date")

	startDate, err := form.GetTimeObj("start_date")
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	endDate, err := form.GetTimeObj("end_date")
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't parse start date")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	laptopID, err := strconv.Atoi(r.Form.Get("laptop_id"))
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "invalid data")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Email = r.Form.Get("email")
	reservation.Phone = r.Form.Get("phone")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		stringMap := make(map[string]string)
		stringMap["start_date"] = r.Form.Get("start_date")
		stringMap["end_date"] = r.Form.Get("end_date")
		w.WriteHeader(http.StatusSeeOther)
		render.Template(w, r, "make-reservation.page.html", &models.TemplateData{
			Form:      form,
			Data:      data,
			StringMap: stringMap,
		})
		return
	}

	newReservationID, err := repo.DB.InsertReservation(&reservation)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't insert reservation into the database")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	restriction := models.LaptopRestriction{
		StartDate:     startDate,
		EndDate:       endDate,
		LaptopID:      laptopID,
		ReservationID: newReservationID,
		RestrictionID: 1,
	}
	err = repo.DB.InsertLaptopRestriction(&restriction)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't insert laptop restriction into the database")
		http.Redirect(w, r, "/", http.StatusSeeOther)
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
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	form := forms.New(r.PostForm)

	startDate, err := form.GetTimeObj("start_date")
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't parse start date")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	endDate, err := form.GetTimeObj("end_date")
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't parse end date")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	laptops, err := repo.DB.SearchAvailabilityForAllLaptops(startDate, endDate)
	if err != nil {
		fmt.Println(err)
		repo.App.Session.Put(r.Context(), "error", "can't search availability")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if len(laptops) == 0 {
		repo.App.Session.Put(r.Context(), "error", "no availability")
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
		http.Redirect(w, r, "/", http.StatusSeeOther)
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
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	res, ok := repo.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		repo.App.Session.Put(r.Context(), "error", "can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusSeeOther)
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
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	form := forms.New(r.Form)
	startDate, err := form.GetTimeObj("s")
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't parse start date")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	endDate, err := form.GetTimeObj("e")
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't parse end date")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var res models.Reservation
	res.LaptopID = LaptopID
	res.StartDate = startDate
	res.EndDate = endDate

	laptop, err := repo.DB.GetLaptopByID(res.LaptopID)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "error connecting to the database")
		http.Redirect(w, r, "/", http.StatusSeeOther)
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
		http.Redirect(w, r, "/", http.StatusSeeOther)
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
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	repo.App.Session.Put(r.Context(), "user_id", id)
	repo.App.Session.Put(r.Context(), "flash", "Logged in successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout logs a user out
func (repo *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = repo.App.Session.Destroy(r.Context())
	_ = repo.App.Session.RenewToken(r.Context())
	repo.App.Session.Put(r.Context(), "flash", "Logged out successfully")

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

// AdminDashbord shows admin dashboard page
func (repo *Repository) AdminDashbord(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-dashboard.page.html", &models.TemplateData{})
}

// AdminNewReservations shows all new reservations
func (repo *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := repo.DB.AllNewReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	data := make(map[string]interface{})
	data["reservations"] = reservations
	render.Template(w, r, "admin-new-reservations.page.html", &models.TemplateData{
		Data: data,
	})
}

// AdminAllReservations shows all reservations
func (repo *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := repo.DB.AllReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	data := make(map[string]interface{})
	data["reservations"] = reservations
	render.Template(w, r, "admin-all-reservations.page.html", &models.TemplateData{
		Data: data,
	})
}

// AdminShowReservation shows the reservation page
func (repo *Repository) AdminShowReservation(w http.ResponseWriter, r *http.Request) {
	splited := strings.Split(r.RequestURI, "/")

	id, err := strconv.Atoi(splited[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	tp := splited[3]

	stringMap := make(map[string]string)
	stringMap["type"] = tp

	stringMap["year"] = r.URL.Query().Get("y")
	stringMap["month"] = r.URL.Query().Get("m")

	res, err := repo.DB.GetReservatioByID(id)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't find reservation")
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations/%s/%d", tp, id), http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(w, r, "admin-show-reservation.page.html", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		Form:      forms.New(nil),
	})
}

// PostAdminShowReservation
func (repo *Repository) PostAdminShowReservation(w http.ResponseWriter, r *http.Request) {
	splited := strings.Split(r.RequestURI, "/")

	id, err := strconv.Atoi(splited[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	tp := splited[3]

	err = r.ParseForm()
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't parse form")
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations/%s/%d", tp, id), http.StatusSeeOther)
		return
	}

	res, err := repo.DB.GetReservatioByID(id)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't find reservation")
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations/%s/%d", tp, id), http.StatusSeeOther)
		return
	}

	res.FirstName = r.Form.Get("first_name")
	res.LastName = r.Form.Get("last_name")
	res.Email = r.Form.Get("email")
	res.Phone = r.Form.Get("phone")

	err = repo.DB.UpdateReservation(&res)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't update database")
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations/%s/%d", tp, id), http.StatusSeeOther)
		return
	}

	month := r.Form.Get("month")
	year := r.Form.Get("year")

	repo.App.Session.Put(r.Context(), "flash", "Saved successfully")
	if year == "" || month == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", tp), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s?y=%s&m=%s", tp, year, month), http.StatusSeeOther)
	}
}

// AdminReservationsCalendar shows the reservations calendar
func (repo *Repository) AdminReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	if r.URL.Query().Get("y") != "" {
		year, err := strconv.Atoi(r.URL.Query().Get("y"))
		if err != nil {
			repo.App.Session.Put(r.Context(), "error", "can't get year")
			http.Redirect(w, r, "/admin/reservations-calendar", http.StatusSeeOther)
			return
		}

		month, err := strconv.Atoi(r.URL.Query().Get("m"))
		if err != nil {
			repo.App.Session.Put(r.Context(), "error", "can't get month")
			http.Redirect(w, r, "/admin/reservations-calendar", http.StatusSeeOther)
			return
		}

		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	}

	data := make(map[string]interface{})
	data["now"] = now

	next := now.AddDate(0, 1, 0)
	last := now.AddDate(0, -1, 0)

	stringMap := make(map[string]string)
	stringMap["next_month"] = next.Format("01")
	stringMap["next_month_year"] = next.Format("2006")
	stringMap["last_month"] = last.Format("01")
	stringMap["last_month_year"] = last.Format("2006")

	stringMap["this_month"] = now.Format("01")
	stringMap["this_month_year"] = now.Format("2006")

	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstDayOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastDayOfMonth := firstDayOfMonth.AddDate(0, 1, -1)

	intMap := make(map[string]int)
	intMap["days_in_month"] = lastDayOfMonth.Day()

	laptops, err := repo.DB.AllLaptops()
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't get all laptops from database")
		http.Redirect(w, r, "/admin/reservations-calendar", http.StatusSeeOther)
		return
	}

	data["laptops"] = laptops

	for _, lp := range laptops {
		reservationMap := make(map[string]int)
		blockMap := make(map[string]int)

		for d := firstDayOfMonth; !d.After(lastDayOfMonth); d = d.AddDate(0, 0, 1) {
			reservationMap[d.Format("2006-01-2")] = 0
			blockMap[d.Format("2006-01-2")] = 0
		}

		laptopRestrictions, err := repo.DB.GetLaptopRestrictionsByDate(lp.ID, firstDayOfMonth, lastDayOfMonth)
		if err != nil {
			repo.App.Session.Put(r.Context(), "error", "can't get laptop restrictions from database")
			http.Redirect(w, r, "/admin/reservations-calendar", http.StatusSeeOther)
			return
		}

		for _, lr := range laptopRestrictions {
			if lr.ReservationID > 0 {
				for d := lr.StartDate; !d.After(lr.EndDate); d = d.AddDate(0, 0, 1) {
					reservationMap[d.Format("2006-01-2")] = lr.ReservationID
				}
			} else {
				for d := lr.StartDate; !d.After(lr.EndDate); d = d.AddDate(0, 0, 1) {
					blockMap[d.Format("2006-01-2")] = lr.ID
				}
			}
		}

		data[fmt.Sprintf("reservation_map_%d", lp.ID)] = reservationMap
		data[fmt.Sprintf("block_map_%d", lp.ID)] = blockMap

		repo.App.Session.Put(r.Context(), fmt.Sprintf("block_map_%d", lp.ID), blockMap)
	}

	render.Template(w, r, "admin-reservations-calendar.page.html", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		IntMap:    intMap,
	})
}

// PostAdminReservationsCalendar handles post of reservation calendar
func (repo *Repository) PostAdminReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't parse form")
		http.Redirect(w, r, "/admin/reservations-calendar", http.StatusSeeOther)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("y", "m")
	if !form.Valid() {
		repo.App.Session.Put(r.Context(), "error", "can't get year or month")
		http.Redirect(w, r, "/admin/reservations-calendar", http.StatusSeeOther)
		return
	}

	year, err := strconv.Atoi(form.Get("y"))
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't get year")
		http.Redirect(w, r, "/admin/reservations-calendar", http.StatusSeeOther)
		return
	}

	month, err := strconv.Atoi(form.Get("m"))
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't get month")
		http.Redirect(w, r, "/admin/reservations-calendar", http.StatusSeeOther)
		return
	}

	laptops, err := repo.DB.AllLaptops()
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't get all laptops from database")
		http.Redirect(w, r, "/admin/reservations-calendar", http.StatusSeeOther)
		return
	}

	for _, lp := range laptops {
		// get block_map from_session, if form data has remove_block but block_map
		curMap := repo.App.Session.Get(r.Context(), fmt.Sprintf("block_map_%d", lp.ID)).(map[string]int)
		for date, laptopRestrictionID := range curMap {
			if laptopRestrictionID > 0 && !form.Has(fmt.Sprintf("remove_block_%d_%s", lp.ID, date)) {
				err := repo.DB.DeleteBlockByID(laptopRestrictionID)
				if err != nil {
					repo.App.Session.Put(r.Context(), "error", "can't delete block from datebase")
					http.Redirect(w, r, "/admin/reservations-calendar", http.StatusSeeOther)
					return
				}
			}
		}
	}

	for inputName := range r.PostForm {
		if strings.HasPrefix(inputName, "add_block") {
			splited := strings.Split(inputName, "_")

			laptopID, err := strconv.Atoi(splited[2])
			if err != nil {
				helpers.ServerError(w, err)
			}

			startDate, err := time.Parse("2006-01-2", splited[3])
			if err != nil {
				helpers.ServerError(w, err)
			}

			err = repo.DB.InsertOneDayBlockByLaptopID(laptopID, startDate)
			if err != nil {
				repo.App.Session.Put(r.Context(), "error", "can't insert block into datebase")
				http.Redirect(w, r, "/admin/reservations-calendar", http.StatusSeeOther)
				return
			}
		}
	}

	repo.App.Session.Put(r.Context(), "flash", "changes saved")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%d&m=%d", year, month), http.StatusSeeOther)
}

// AdminProcessReservation marks a reservation as processed
func (repo *Repository) AdminProcessReservation(w http.ResponseWriter, r *http.Request) {
	splited := strings.Split(r.RequestURI, "/")

	tp := splited[3]

	id, err := strconv.Atoi(splited[4])
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't get id")
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", tp), http.StatusSeeOther)
		return
	}

	err = repo.DB.UpdateReservationProcessed(id, 1)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't update database")
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", tp), http.StatusSeeOther)
		return
	}

	month := r.URL.Query().Get("m")
	year := r.URL.Query().Get("y")

	repo.App.Session.Put(r.Context(), "flash", "Reservation marked as processed")

	if year == "" || month == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", tp), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s?y=%s&m=%s", tp, year, month), http.StatusSeeOther)
	}

}

// AdminDeleteReservation deletes a reservation as processed
func (repo *Repository) AdminDeleteReservation(w http.ResponseWriter, r *http.Request) {
	splited := strings.Split(r.RequestURI, "/")

	tp := splited[3]

	id, err := strconv.Atoi(splited[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	err = repo.DB.DeleteReservation(id)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't delete from database")
		http.Redirect(w, r, r.RequestURI, http.StatusSeeOther)
		return
	}

	month := r.URL.Query().Get("m")
	year := r.URL.Query().Get("y")

	repo.App.Session.Put(r.Context(), "flash", "Reservation deleted")

	if year == "" || month == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", tp), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s?y=%s&m=%s", tp, year, month), http.StatusSeeOther)
	}
}
