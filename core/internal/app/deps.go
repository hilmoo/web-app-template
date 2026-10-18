package app

import (
	"context"
	"fmt"
	"log/slog"

	"web-app-template-core/internal/config"
	db "web-app-template-core/internal/db/sqlc"
	"web-app-template-core/internal/helper"

	"github.com/hilmoo/gomox/transport/validation"
)

// setupDependencies initializes all infrastructure clients and returns the
// handler dependencies together with a cleanup function.
func setupDependencies(
	ctx context.Context,
	cfg config.Config,
	logger *slog.Logger,
) (*helper.HandlerParams, func(), error) {
	dbPool, err := initDB(ctx, cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	deps := &helper.HandlerParams{
		Validate: validation.InitValidation(),
		Pool:     dbPool,
		Queries:  db.New(dbPool),
		Config:   cfg,
		Logger:   logger,
	}

	cleanup := func() {
		deps.Logger.InfoContext(ctx, "Cleaning up resources...")
		dbPool.Close()
	}

	return deps, cleanup, nil
}
