package database

import (
	"time"

	"github.com/kaitolucifer/go-laptop-rental-site/internal/models"
)

type Database interface {
	AllUsers() bool

	InsertReservation(res *models.Reservation) (int, error)
	InsertLaptopRestriction(lr *models.LaptopRestrictions) error
	SearchAvailabilityByDatesByLaptopID(start, end time.Time, laptopID int) (bool, error)
	SearchAvailabilityForAllLaptops(start, end time.Time) ([]models.Laptop, error)
	GetLaptopByID(id int) (models.Laptop, error)
	GetUserByID(id int) (models.User, error)
	UpdateUser(u *models.User) error
	Authenticate(email, password string) (int, string, error)
}
