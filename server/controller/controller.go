package controller

import (
	"fmt"
	"os"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
)

type Controller struct {
	db *sqlx.DB
}

func NewController() (*Controller, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("POSTGRES_HOST", "localhost"),
		getEnv("POSTGRES_PORT", "5432"),
		getEnv("POSTGRES_USER", "postgres"),
		getEnv("POSTGRES_PASSWORD", "postgres"),
		getEnv("POSTGRES_DBNAME", "postgres"),
	)

	db := sqlx.MustOpen("pgx", psqlInfo)
	if err := db.Ping(); err != nil {
		return &Controller{db}, err
	}
	return &Controller{db}, nil
}

func (crtl *Controller) Close() error {
	return crtl.db.Close()
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
