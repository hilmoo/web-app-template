package app

import (
	"context"
	"fmt"

	"web-app-template-core/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func initDB(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create DB pool: %w", err)
	}

	if pingErr := pool.Ping(ctx); pingErr != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", pingErr)
	}

	return pool, nil
}
