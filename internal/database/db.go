package database

import (
	"database/sql"

	"github.com/kaitolucifer/go-laptop-rental-site/internal/config"
)

type postgres struct {
	App *config.AppConfig
	DB  *sql.DB
}

type mockPostgres struct {
	App *config.AppConfig
	DB *sql.DB
}

func NewPostgres(conn *sql.DB, a *config.AppConfig) Database {
	return &postgres{
		App: a,
		DB:  conn,
	}
}

func NewMockPostgres(a *config.AppConfig) Database {
	return &mockPostgres{
		App: a,
	}
}
