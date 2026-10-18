package helper

import (
	"context"
	"log/slog"

	"web-app-template-core/internal/config"
	db "web-app-template-core/internal/db/sqlc"

	errort "github.com/hilmoo/gomox/transport/error"
	"github.com/hilmoo/gomox/transport/validation"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/herodot"
)

// HandlerParams are the dependencies shared by every feature handler.
type HandlerParams struct {
	Validate *validation.Vld
	Pool     *pgxpool.Pool
	Queries  *db.Queries
	Config   config.Config
	Logger   *slog.Logger
}

func StrToUUIDHTTP(s string) (uuid.UUID, *herodot.DefaultError) {
	uuid, err := uuid.Parse(s)
	if err != nil {
		return uuid, herodot.ErrBadRequest.WithReasonf("invalid id: %s", s)
	}
	return uuid, nil
}

// WithUserID resolves the calling user and converts fn's error into a connect error.
func WithUserID[T any](ctx context.Context, fn func(userID uuid.UUID) (T, *herodot.DefaultError)) (T, error) {
	userID, errH := UserIDFromContext(ctx)
	if errH == nil {
		var out T
		if out, errH = fn(userID); errH == nil {
			return out, nil
		}
	}
	var zero T
	return zero, errort.ConnectError(ctx, errH)
}
