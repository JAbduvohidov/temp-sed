package db

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"os"
)

var Pool *pgxpool.Pool

func Connect() (err error) {
	Pool, err = pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	return err
}
