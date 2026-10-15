package auth

import (
	"context"
	"errors"

	servicesv1 "proto/services/v1"

	db "web-app-template-core/internal/gen/sqlc"

	"github.com/hilmoo/gomox/sqlx"
	msession "github.com/hilmoo/gomox/transport/middleware/session"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/matthewhartstonge/argon2"
	"github.com/ory/herodot"
)

// pgUniqueViolation is the SQLSTATE for unique_violation.
const pgUniqueViolation = "23505"

type passwordLoginServiceParams struct {
	queries   *db.Queries
	secret    string
	ipAddress *string
	userAgent *string
	in        *servicesv1.AuthLoginRequest
}

type passwordLoginServiceResult struct {
	user  db.User
	token string
}

func passwordLoginService(
	ctx context.Context,
	args passwordLoginServiceParams,
) (*passwordLoginServiceResult, *herodot.DefaultError) {
	user, err := args.queries.UserGetByEmail(ctx, args.in.GetEmail())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, herodot.ErrUnauthorized.WithReason("invalid email or password").WithDebug(err.Error())
		}
		return nil, sqlx.MapDBErrorHTTP(err, "user")
	}

	hashPassword := user.PasswordHash
	if hashPassword == nil {
		return nil, herodot.ErrUnauthorized.WithReason("invalid email or password")
	}

	ok, err := argon2.VerifyEncoded([]byte(args.in.GetPassword()), []byte(*hashPassword))
	if err != nil {
		return nil, herodot.ErrInternalServerError.WithReason("failed to verify password").WithDebug(err.Error())
	}
	if !ok {
		return nil, herodot.ErrUnauthorized.WithReason("invalid email or password")
	}

	token, errH := createUserSession(
		ctx,
		args.queries,
		args.secret,
		user.ID,
		args.ipAddress,
		args.userAgent,
	)
	if errH != nil {
		return nil, errH
	}

	return &passwordLoginServiceResult{
		user:  user,
		token: token,
	}, nil
}

type passwordRegisterServiceParams struct {
	queries   *db.Queries
	in        *servicesv1.AuthRegisterRequest
	secret    string
	ipAddress *string
	userAgent *string
}

type passwordRegisterServiceResult struct {
	user  db.User
	token string
}

func passwordRegisterService(
	ctx context.Context,
	args passwordRegisterServiceParams,
) (*passwordRegisterServiceResult, *herodot.DefaultError) {
	hash, err := hashPasswordArgon(args.in.GetPassword())
	if err != nil {
		return nil, herodot.ErrInternalServerError.WithReason("failed to hash password").WithDebug(err.Error())
	}

	user, err := args.queries.UserCreatePassword(ctx, db.UserCreatePasswordParams{
		Email:        args.in.GetEmail(),
		FullName:     args.in.GetFullName(),
		PasswordHash: &hash,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return nil, herodot.ErrConflict.WithReason("email already registered").WithDebug(err.Error())
		}
		return nil, sqlx.MapDBErrorHTTP(err, "user")
	}

	token, errH := createUserSession(
		ctx,
		args.queries,
		args.secret,
		user.ID,
		args.ipAddress,
		args.userAgent,
	)
	if errH != nil {
		return nil, errH
	}

	return &passwordRegisterServiceResult{
		user:  user,
		token: token,
	}, nil
}

type passwordLogoutServiceParams struct {
	queries *db.Queries
	secret  string
	token   string
}

func passwordLogoutService(
	ctx context.Context,
	args passwordLogoutServiceParams,
) *herodot.DefaultError {
	err := args.queries.SessionDeleteByTokenHash(ctx, msession.HashSessionToken(args.secret, args.token))
	if err != nil {
		return herodot.ErrInternalServerError.WithReason("failed to delete session").WithDebug(err.Error())
	}
	return nil
}
