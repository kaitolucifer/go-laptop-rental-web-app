package driver

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/jackc/pgx/v4"
)

// DB holds the database connection pool
type DB struct {
	SQL *sql.DB
}

var dbConn = &DB{}

const maxOpenDbConn = 10
const maxIdleDbConn = 5
const maxDbLifetime = 5 * time.Minute

// ConnectSQL creates database pool for Postgres
func ConnectSQL(dsn string) (*DB, error) {
	db, err := NewDatabase(dsn)
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(maxOpenDbConn)
	db.SetMaxIdleConns(maxIdleDbConn)
	db.SetConnMaxLifetime(maxDbLifetime)
	dbConn.SQL = db

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return dbConn, nil
}

// NewDatabase creates a new database for the application
func NewDatabase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
