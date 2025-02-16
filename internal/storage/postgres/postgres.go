package postgres

import (
	"avito-tech-winter-2025/internal/config"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type Storage struct {
	DB *sql.DB
}

const (
	maxRetryAttemptsForTransaction = 7
	retryDelay                     = 10 * time.Millisecond
	dbPingTimeout                  = 5 * time.Second
)

func New(cfg *config.Postgres) (*Storage, error) {
	pool, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbPingTimeout)
	defer cancel()

	if err := pool.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return &Storage{DB: pool}, nil
}

func isRetryableError(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		// 40001 - serialization_failure
		// 40P01 - deadlock_detected
		return pqErr.Code == "40001" || pqErr.Code == "40P01"
	}
	return false
}
