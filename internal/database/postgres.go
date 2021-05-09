package database

import (
	"context"
	"time"

	"github.com/kaitolucifer/go-laptop-rental-site/internal/models"
)

func (p *postgres) AllUsers() bool {
	return true
}

// InsertReservation inserts a reservation into the database
func (p *postgres) InsertReservation(res models.Reservation) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var newID int

	query := `INSERT INTO reservations (first_name, last_name, email, phone,
			  start_date, end_date, laptop_id, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			  RETURNING id`

	err := p.DB.QueryRowContext(ctx, query,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.LaptopID,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

// InsertLaptopRestriction inserts a laptop restriction into the database
func (p *postgres) InsertLaptopRestriction(lr models.LaptopRestrictions) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `INSERT INTO laptop_restrictions (start_date, end_date, laptop_id, reservation_id,
			  created_at, updated_at, restriction_id)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := p.DB.ExecContext(ctx, query,
		lr.StartDate,
		lr.EndDate,
		lr.LaptopID,
		lr.ReservationID,
		time.Now(),
		time.Now(),
		lr.RestrictionID)
	if err != nil {
		return err
	}

	return nil
}

// SearchAvailabilityByDatesLaptopID returns true if availability exists for laptop id, and false if no availability exists
func (p *postgres) SearchAvailabilityByDatesByLaptopID(start, end time.Time, laptopID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT Count(id) FROM laptop_restrictions
		      WHERE laptop_id = $1 and $2 < end_date and $3 > start_date`

	var numRows int
	row := p.DB.QueryRowContext(ctx, query, laptopID, start, end)
	err := row.Scan(&numRows)
	if err != nil {
		return false, err
	}

	if numRows == 0 {
		return true, nil
	}

	return false, nil
}

// SearchAvailabilityForAllLaptops returns a slice of available laptops if any, for given date range
func (p *postgres) SearchAvailabilityForAllLaptops(start, end time.Time) ([]models.Laptop, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var laptops []models.Laptop

	query := `SELECT l.id, l.laptop_name
			  FROM laptops l
			  WHERE l.id not in (
				  SELECT lr.laptop_id FROM laptop_restrictions lr
				  WHERE $1 < lr.end_date AND $2 > lr.start_date)`
	rows, err := p.DB.QueryContext(ctx, query, start, end)
	if err != nil {
		return laptops, err
	}

	for rows.Next() {
		var laptop models.Laptop
		err = rows.Scan(
			&laptop.ID,
			&laptop.LaptopName,
		)
		if err != nil {
			return laptops, err
		}

		laptops = append(laptops, laptop)
	}

	if err = rows.Err(); err != nil {
		return laptops, err
	}

	return laptops, nil
}

// GetLaptopByID gets a laptop by id
func (p *postgres) GetLaptopByID(id int) (models.Laptop, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var laptop models.Laptop

	query := `SELECT id, laptop_name, created_at, updated_at FROM laptops
			 WHERE id = $1`
	row := p.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&laptop.ID,
		&laptop.LaptopName,
		&laptop.CreatedAt,
		&laptop.UpdatedAt,
	)

	if err != nil {
		return laptop, err
	}

	return laptop, nil
}
