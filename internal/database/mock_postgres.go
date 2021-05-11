package database

import (
	"errors"
	"time"

	"github.com/kaitolucifer/go-laptop-rental-site/internal/models"
)

func (p *mockPostgres) AllUsers() bool {
	return true
}

// InsertReservation inserts a reservation into the database
func (p *mockPostgres) InsertReservation(res *models.Reservation) (int, error) {
	// if the first name is Test, then failed
	if res.FirstName == "Test" {
		return 0, errors.New("error")
	}
	return 1, nil
}

// InsertLaptopRestriction inserts a laptop restriction into the database
func (p *mockPostgres) InsertLaptopRestriction(lr *models.LaptopRestrictions) error {
	if lr.LaptopID == 1000 {
		return errors.New("error")
	}
	return nil
}

// SearchAvailabilityByDatesLaptopID returns true if availability exists for laptop id, and false if no availability exists
func (p *mockPostgres) SearchAvailabilityByDatesByLaptopID(start, end time.Time, laptopID int) (bool, error) {
	if laptopID == 1 {
		return true, nil
	} else if laptopID == 1000 {
		return false, nil
	}
	return false, errors.New("error")
}

// SearchAvailabilityForAllLaptops returns a slice of available laptops if any, for given date range
func (p *mockPostgres) SearchAvailabilityForAllLaptops(start, end time.Time) ([]models.Laptop, error) {
	var laptops []models.Laptop
	year, month, day := time.Now().Add(48 * time.Hour).Date()
	year2, month2, day2 := time.Now().Add(72 * time.Hour).Date()
	if start.Year() == year && start.Month() == month && start.Day() == day {
		laptops = append(laptops, models.Laptop{
			ID:         0,
			LaptopName: "",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		})
		return laptops, nil
	} else if start.Year() == year2 && start.Month() == month2 && start.Day() == day2 {
		return laptops, nil
	}
	return laptops, errors.New("error")
}

// GetLaptopByID gets a laptop by id
func (p *mockPostgres) GetLaptopByID(id int) (models.Laptop, error) {
	var laptop models.Laptop

	if id > 2 {
		return laptop, errors.New("error")
	}

	return laptop, nil
}

// GetUserByID returns a user by id
func (p *mockPostgres) GetUserByID(id int) (models.User, error) {
	var u models.User
	return u, nil
}

func (p *mockPostgres) UpdateUser(u *models.User) error {
	return nil
}

func (p *mockPostgres) Authenticate(email, password string) (int, string, error) {
	return 0, "", nil
}
