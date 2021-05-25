package database

import (
	"time"

	"github.com/kaitolucifer/go-laptop-rental-site/internal/models"
)

type DBRepository interface {
	AllUsers() bool

	InsertReservation(res *models.Reservation) (int, error)
	InsertLaptopRestriction(lr *models.LaptopRestriction) error
	SearchAvailabilityByDatesByLaptopID(start, end time.Time, laptopID int) (bool, error)
	SearchAvailabilityForAllLaptops(start, end time.Time) ([]models.Laptop, error)
	GetLaptopByID(id int) (models.Laptop, error)
	GetUserByID(id int) (models.User, error)
	UpdateUser(u *models.User) error
	Authenticate(email, password string) (int, string, error)
	AllReservations() ([]models.Reservation, error)
	AllNewReservations() ([]models.Reservation, error)
	GetReservatioByID(id int) (models.Reservation, error)
	UpdateReservation(res *models.Reservation) error
	DeleteReservation(id int) error
	UpdateReservationProcessed(id, processed int) error
	AllLaptops() ([]models.Laptop, error)
	GetLaptopRestrictionsByDate(laptopID int, start, end time.Time) ([]models.LaptopRestriction, error)
	InsertOneDayBlockByLaptopID(id int, startDate time.Time) error
	DeleteBlockByID(id int) error
}
