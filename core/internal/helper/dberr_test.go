package helper_test

import (
	"errors"
	"testing"

	"web-app-template-core/internal/helper"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	errNotFound = errors.New("not found")
	errConflict = errors.New("conflict")
	errOther    = errors.New("boom")
)

func TestMapDBError(t *testing.T) {
	t.Parallel()

	unique := &pgconn.PgError{Code: "23505"} //nolint:exhaustruct_v5 // only the code matters

	tests := []struct {
		name     string
		err      error
		notFound error
		conflict error
		want     error
	}{
		{"no rows", pgx.ErrNoRows, errNotFound, errConflict, errNotFound},
		{"unique violation", unique, errNotFound, errConflict, errConflict},
		{"other", errOther, errNotFound, errConflict, errOther},
		{"no rows without sentinel", pgx.ErrNoRows, nil, errConflict, pgx.ErrNoRows},
		{"unique without sentinel", unique, errNotFound, nil, unique},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := helper.MapDBError(tt.err, "op", tt.notFound, tt.conflict)
			if !errors.Is(got, tt.want) {
				t.Fatalf("got %v, want errors.Is %v", got, tt.want)
			}
		})
	}
}

func TestMapDBErrorNil(t *testing.T) {
	t.Parallel()

	if got := helper.MapDBError(nil, "op", errNotFound, errConflict); got != nil {
		t.Fatalf("got %v, want nil", got)
	}
}
