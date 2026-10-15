package helper

import (
	"context"

	db "web-app-template-core/internal/gen/sqlc"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/herodot"
)

// WithTx runs fn inside a transaction, committing only when fn succeeds.
func WithTx[T any](
	ctx context.Context,
	pool *pgxpool.Pool,
	queries *db.Queries,
	fn func(q *db.Queries) (T, *herodot.DefaultError),
) (T, *herodot.DefaultError) {
	var zero T

	tx, err := pool.Begin(ctx)
	if err != nil {
		return zero, herodot.ErrInternalServerError.WithReasonf("failed to begin transaction: %v", err)
	}
	// Rollback after a successful commit is a no-op.
	defer func() { _ = tx.Rollback(ctx) }()

	res, errH := fn(queries.WithTx(tx))
	if errH != nil {
		return zero, errH
	}
	if err = tx.Commit(ctx); err != nil {
		return zero, herodot.ErrInternalServerError.WithReasonf("failed to commit transaction: %v", err)
	}
	return res, nil
}
