package storage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPoolWithPing(connectUrl string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connectUrl)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse config from connect url: %v\n", err)
	}

	dbpool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("Unable to create connection pool: %v\n", err)
	}

	err = dbpool.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("Unable to ping: %v\n", err)
	}
	return dbpool, nil
}
