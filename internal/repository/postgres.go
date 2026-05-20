package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Pool *pgxpool.Pool
}

func NewPostgresConnection(ctx context.Context, connString string) (*Database, error) {
	pool, err := pgxpool.New(ctx, connString)

	if err != nil {
		return nil, fmt.Errorf("Cannot create connection pool: %v", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("DB does not respond: %v", err)
	}

	fmt.Println("Successful connection to DB")
	return &Database{Pool: pool}, nil
}
