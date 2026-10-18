package helper

import (
	"errors"
	"fmt"

	"github.com/hilmoo/gomox/sqlx"
	"github.com/jackc/pgx/v5"
)

// MapDBError turns a database error into a domain error for a repository call named op
// (e.g. "auth: find user"). It returns nil for a nil err. [pgx.ErrNoRows] becomes notFound
// and a unique violation becomes conflict, both matchable with [errors.Is]; pass nil for a
// sentinel the call cannot produce. Anything else is wrapped with op unchanged. The original
// error stays in the chain for logs.
func MapDBError(err error, op string, notFound, conflict error) error {
	switch {
	case err == nil:
		return nil
	case notFound != nil && errors.Is(err, pgx.ErrNoRows):
		return fmt.Errorf("%s: %w: %w", op, notFound, err)
	case conflict != nil && errors.Is(sqlx.ClassifyPgError(err), sqlx.ErrAlreadyExists):
		return fmt.Errorf("%s: %w: %w", op, conflict, err)
	default:
		return fmt.Errorf("%s: %w", op, err)
	}
}
