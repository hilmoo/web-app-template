package auth

import (
	"context"
	servicesv1 "proto/services/v1"

	db "web-app-template-core/internal/db/sqlc"

	"github.com/google/uuid"
	"github.com/hilmoo/gomox/sqlx"
	"github.com/ory/herodot"
)

func meGetUser(
	ctx context.Context,
	queries *db.Queries,
	userID uuid.UUID,
) (db.User, *herodot.DefaultError) {
	user, err := queries.UserGetById(ctx, userID)
	if err != nil {
		return db.User{}, sqlx.MapPgxErrorHTTP(err, "user")
	}

	return user, nil
}

type meUpdateUserServiceParams struct {
	in      *servicesv1.AuthUpdateMeRequest
	queries *db.Queries
	userID  uuid.UUID
}

func meUpdateUserService(
	ctx context.Context,
	args meUpdateUserServiceParams,
) (db.User, *herodot.DefaultError) {
	user, err := args.queries.UserUpdate(ctx, db.UserUpdateParams{
		ID:           args.userID,
		Email:        args.in.Email,
		FullName:     args.in.FullName,
		PasswordHash: nil,
	})
	if err != nil {
		return db.User{}, sqlx.MapPgxErrorHTTP(err, "user")
	}

	return user, nil
}
