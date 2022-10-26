package controller

import (
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"golang.org/x/net/context"
)

type Controller struct {
	db  *sqlx.DB
	rdb *redis.Client
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

	redisAddr := fmt.Sprintf("%s:%s", getEnv("REDIS_HOST", "localhost"), getEnv("REDIS_PORT", "6379"))

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	rdbCtx, rdbCancel := context.WithTimeout(context.Background(), time.Second*3)
	defer rdbCancel()
	if err := rdb.Ping(rdbCtx).Err(); err != nil {
		return &Controller{}, err
	}

	return &Controller{db, rdb}, nil

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
