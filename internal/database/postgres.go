package database

import (
	"context"
	"errors"
	"time"

	"github.com/kaitolucifer/go-laptop-rental-site/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func (p *postgres) AllUsers() bool {
	return true
}

// InsertReservation inserts a reservation into the database
func (p *postgres) InsertReservation(res *models.Reservation) (int, error) {
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
func (p *postgres) InsertLaptopRestriction(lr *models.LaptopRestriction) error {
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
		      WHERE laptop_id = $1 and $2 <= end_date and $3 >= start_date`

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
				  WHERE $1 <= lr.end_date AND $2 >= lr.start_date)`
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

// GetUserByID returns a user by id
func (p *postgres) GetUserByID(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT id, first_name, last_name, email, password, access_level, created_at, updated_at
			  FROM users WHERE id = $1`
	row := p.DB.QueryRowContext(ctx, query, id)

	var u models.User
	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
		&u.AccessLevel,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return u, err
	}

	return u, nil
}

// UpdateUser updates a user in the database
func (p *postgres) UpdateUser(u *models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `UPDATE users SET first_name = $1, last_name = $2, email = $3, access_level = $4, updated_at = $5`

	_, err := p.DB.ExecContext(ctx, query,
		u.FirstName,
		u.LastName,
		u.Email,
		u.AccessLevel,
		time.Now(),
	)
	if err != nil {
		return err
	}

	return nil
}

// Authenticate authenticates a user
func (p *postgres) Authenticate(email, password string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string
	row := p.DB.QueryRowContext(ctx, "SELECT id, password FROM users WHERE email = $1", email)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return id, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", errors.New("incorrect password")
	} else if err != nil {
		return 0, "", err
	}

	return id, hashedPassword, nil
}

// AllReservations returns a slice of all reservations
func (p *postgres) AllReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `SELECT r.id, r.first_name, r.last_name, r.email, r.phone,
			  r.start_date, r.end_date, r.laptop_id, r.created_at, r.updated_at, r.processed,
			  lp.id, lp.laptop_name
			  FROM reservations r
		   	  LEFT JOIN laptops lp ON (r.laptop_id = lp.id)
			  ORDER BY r.start_date asc, r.end_date asc`
	rows, err := p.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close()

	for rows.Next() {
		var r models.Reservation
		err := rows.Scan(
			&r.ID,
			&r.FirstName,
			&r.LastName,
			&r.Email,
			&r.Phone,
			&r.StartDate,
			&r.EndDate,
			&r.LaptopID,
			&r.CreatedAt,
			&r.UpdatedAt,
			&r.Processed,
			&r.Laptop.ID,
			&r.Laptop.LaptopName,
		)

		if err != nil {
			return reservations, err
		}
		reservations = append(reservations, r)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil
}

// AllNewReservations returns a slice of all new reservations
func (p *postgres) AllNewReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `SELECT r.id, r.first_name, r.last_name, r.email, r.phone,
			  r.start_date, r.end_date, r.laptop_id, r.created_at, r.updated_at, r.processed,
			  lp.id, lp.laptop_name
			  FROM reservations r
		   	  LEFT JOIN laptops lp ON (r.laptop_id = lp.id)
			  WHERE processed = 0
			  ORDER BY r.start_date asc, r.end_date asc`
	rows, err := p.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close()

	for rows.Next() {
		var r models.Reservation
		err := rows.Scan(
			&r.ID,
			&r.FirstName,
			&r.LastName,
			&r.Email,
			&r.Phone,
			&r.StartDate,
			&r.EndDate,
			&r.LaptopID,
			&r.CreatedAt,
			&r.UpdatedAt,
			&r.Processed,
			&r.Laptop.ID,
			&r.Laptop.LaptopName,
		)

		if err != nil {
			return reservations, err
		}
		reservations = append(reservations, r)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil
}

// GetReservatioByID returns one reservation by id
func (p *postgres) GetReservatioByID(id int) (models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var res models.Reservation

	query := `SELECT r.id, r.first_name, r.last_name, r.email, r.phone,
			  r.start_date, r.end_date, r.laptop_id, r.created_at, r.updated_at, r.processed,
			  lp.id, lp.laptop_name
			  FROM reservations r
			  LEFT JOIN laptops lp ON (r.laptop_id = lp.id)
			  WHERE r.id = $1`
	row := p.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&res.ID,
		&res.FirstName,
		&res.LastName,
		&res.Email,
		&res.Phone,
		&res.StartDate,
		&res.EndDate,
		&res.LaptopID,
		&res.CreatedAt,
		&res.UpdatedAt,
		&res.Processed,
		&res.Laptop.ID,
		&res.Laptop.LaptopName,
	)

	if err != nil {
		return res, err
	}

	return res, nil
}

// UpdateReservation updates a reservation in the database
func (p *postgres) UpdateReservation(res *models.Reservation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `UPDATE reservations SET first_name = $1, last_name = $2, email = $3, phone = $4, updated_at = $5
			  WHERE id = $6`

	_, err := p.DB.ExecContext(ctx, query,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		time.Now(),
		res.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

// DeleteReservation deletes one reservation by id
func (p *postgres) DeleteReservation(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `DELETE FROM reservations where id = $1`
	_, err := p.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

// UpdateReservationProcessed updates a reservation processed status by id
func (p *postgres) UpdateReservationProcessed(id, processed int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `UPDATE reservations SET processed = $2
			  WHERE id = $1`

	_, err := p.DB.ExecContext(ctx, query, id, processed)
	if err != nil {
		return err
	}

	return nil
}

// AllLaptops returns all laptops
func (p *postgres) AllLaptops() ([]models.Laptop, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var laptops []models.Laptop

	query := `SELECT id, laptop_name, created_at, updated_at
			  FROM laptops order by laptop_name`
	rows, err := p.DB.QueryContext(ctx, query)
	if err != nil {
		return laptops, err
	}
	defer rows.Close()

	for rows.Next() {
		var laptop models.Laptop
		err := rows.Scan(
			&laptop.ID,
			&laptop.LaptopName,
			&laptop.CreatedAt,
			&laptop.UpdatedAt,
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

// GetLaptopRestrictionsByDate returns restrictions for a laptop by date range
func (p *postgres) GetLaptopRestrictionsByDate(laptopID int, start, end time.Time) ([]models.LaptopRestriction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var restrictions []models.LaptopRestriction

	query := `SELECT id, COALESCE(reservation_id, 0) reservation_id, restriction_id, laptop_id, start_date, end_date
			  FROM laptop_restrictions 
			  WHERE $1 <= end_date AND $2 >= start_date AND laptop_id = $3`

	rows, err := p.DB.QueryContext(ctx, query, start, end, laptopID)
	if err != nil {
		return restrictions, nil
	}
	defer rows.Close()

	for rows.Next() {
		var l models.LaptopRestriction
		err := rows.Scan(
			&l.ID,
			&l.ReservationID,
			&l.RestrictionID,
			&l.LaptopID,
			&l.StartDate,
			&l.EndDate,
		)
		if err != nil {
			return restrictions, err
		}
		restrictions = append(restrictions, l)
	}

	if err = rows.Err(); err != nil {
		return restrictions, nil
	}

	return restrictions, nil
}

// InsertOneDayBlockByLaptopID inserts a one day block restriction by laptop id
func (p *postgres) InsertOneDayBlockByLaptopID(id int, startDate time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `INSERT INTO laptop_restrictions (start_date, end_date, laptop_id, restriction_id, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := p.DB.ExecContext(ctx, query, startDate, startDate, id, 2, time.Now(), time.Now())
	if err != nil {
		return err
	}
	
	return nil
}

// DeleteBlockByLaptopID deletes a laptop restriction by id
func (p *postgres) DeleteBlockByID(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `DELETE FROM laptop_restrictions WHERE id = $1`

	_, err := p.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	
	return nil
}
