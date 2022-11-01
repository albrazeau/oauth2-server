package controller

import (
	"fmt"
	"os"
	"time"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"golang.org/x/net/context"
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

	dbCtx, dbCancel := context.WithTimeout(context.Background(), time.Second*3)
	defer dbCancel()
	if err := db.PingContext(dbCtx); err != nil {
		return &Controller{}, err
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
