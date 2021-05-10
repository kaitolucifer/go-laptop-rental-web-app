package models

import (
	"time"
)

// User is the user model
type User struct {
	ID          int
	FirstName   string
	LastName    string
	Email       string
	Password    string
	AccessLevel int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Laptop is the laptop model
type Laptop struct {
	ID         int
	LaptopName string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Restriction is the restriction model
type Restriction struct {
	ID              int
	RestrictionName string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Reservation it the reservation model
type Reservation struct {
	ID        int
	FirstName string
	LastName  string
	Email     string
	Phone     string
	StartDate time.Time
	EndDate   time.Time
	LaptopID  int
	CreatedAt time.Time
	UpdatedAt time.Time
	Laptop    Laptop
}

// LaptopRestriction is the laptop restriction model
type LaptopRestrictions struct {
	ID            int
	StartDate     time.Time
	EndDate       time.Time
	LaptopID      int
	ReservationID int
	RestrictionID int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Laptop        Laptop
	Reservation   Reservation
	Restriction   Restriction
}

// MailData holds an email message
type MailData struct {
	To      string
	From    string
	Subject string
	Content string
	Template string
}
