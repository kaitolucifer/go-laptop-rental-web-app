package database

import (
	"time"

	"github.com/kaitolucifer/go-laptop-rental-site/internal/models"
)

type Database interface {
	AllUsers() bool

	InsertReservation(res models.Reservation) (int, error)
	InsertLaptopRestriction(lr models.LaptopRestrictions) error
	SearchAvailabilityByDatesByLaptopID(start, end time.Time, laptopID int) (bool, error)
	SearchAvailabilityForAllLaptops(start, end time.Time) ([]models.Laptop, error)
	GetLaptopByID(id int) (models.Laptop, error)
}
