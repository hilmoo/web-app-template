package helper

import (
	"context"
	"fmt"

	db "web-app-template-core/internal/db/sqlc"

	"github.com/jackc/pgx/v5/pgxpool"
)

// WithTx runs fn inside a transaction, committing only when fn succeeds.
func WithTx[T any](
	ctx context.Context,
	pool *pgxpool.Pool,
	queries *db.Queries,
	fn func(q *db.Queries) (T, error),
) (T, error) {
	var zero T

	tx, err := pool.Begin(ctx)
	if err != nil {
		return zero, fmt.Errorf("begin transaction: %w", err)
	}
	// Rollback after a successful commit is a no-op.
	defer func() { _ = tx.Rollback(ctx) }()

	res, err := fn(queries.WithTx(tx))
	if err != nil {
		return zero, err
	}
	if err = tx.Commit(ctx); err != nil {
		return zero, fmt.Errorf("commit transaction: %w", err)
	}
	return res, nil
}
