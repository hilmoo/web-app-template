package helper

import (
	"log/slog"

	"web-app-template-core/internal/config"
	db "web-app-template-core/internal/db/sqlc"

	"github.com/hilmoo/gomox/transport/validation"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/herodot"
)

// HandlerParams is the infra container built by internal/app. Features never receive it;
// internal/app builds each feature's own Deps from it.
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
