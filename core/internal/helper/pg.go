package helper

import (
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ory/herodot"
)

var (
	ErrAlreadyExists = errors.New("already exists")
	ErrConflict      = errors.New("conflict")
	ErrInvalid       = errors.New("invalid data")
)

func ClassifyPgError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	switch pgErr.Code {
	case pgerrcode.UniqueViolation:
		return fmt.Errorf("%w: %w", ErrAlreadyExists, err)
	case pgerrcode.ForeignKeyViolation:
		return fmt.Errorf("%w: %w", ErrConflict, err)
	case pgerrcode.NotNullViolation, pgerrcode.CheckViolation:
		return fmt.Errorf("%w: %w", ErrInvalid, err)
	}
	return err
}

func MapDBErrorHTTP(err error, resourceName string) *herodot.DefaultError {
	err = ClassifyPgError(err)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return herodot.ErrNotFound.WithReasonf("%s not found", resourceName).WithDebug(err.Error())
	case errors.Is(err, ErrAlreadyExists):
		return herodot.ErrConflict.WithReasonf("%s already exists", resourceName).WithDebug(err.Error())
	case errors.Is(err, ErrConflict):
		return herodot.ErrConflict.WithReasonf("%s conflict", resourceName).WithDebug(err.Error())
	case errors.Is(err, ErrInvalid):
		return herodot.ErrBadRequest.WithReasonf("%s invalid data", resourceName).WithDebug(err.Error())
	default:
		return herodot.ErrInternalServerError.WithReasonf("internal server error").WithDebug(err.Error())
	}
}
