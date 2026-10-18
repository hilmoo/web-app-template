package auth

import (
	"context"
	"errors"

	errort "github.com/hilmoo/gomox/transport/error"
	"github.com/ory/herodot"
)

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrEmailAlreadyRegistered = errors.New("email already registered")
	ErrSessionNotFound        = errors.New("session not found")
)

func toConnectError(ctx context.Context, err error) error {
	return errort.ConnectError(ctx, toAppError(err))
}

func toAppError(err error) *herodot.DefaultError {
	switch {
	case errors.Is(err, ErrUserNotFound):
		return herodot.ErrNotFound.WithID("USER_NOT_FOUND").
			WithReason("user not found").WithWrap(err).WithDebug(err.Error())
	case errors.Is(err, ErrInvalidCredentials):
		return herodot.ErrUnauthorized.WithID("AUTH_INVALID_CREDENTIALS").
			WithReason("invalid email or password").WithWrap(err).WithDebug(err.Error())
	case errors.Is(err, ErrEmailAlreadyRegistered):
		return herodot.ErrConflict.WithID("AUTH_EMAIL_ALREADY_REGISTERED").
			WithReason("email already registered").WithWrap(err).WithDebug(err.Error())
	case errors.Is(err, ErrSessionNotFound):
		return herodot.ErrUnauthorized.WithID("AUTH_SESSION_NOT_FOUND").
			WithReason("authentication required").WithWrap(err).WithDebug(err.Error())
	}
	return herodot.ErrInternalServerError.WithReason("internal error").WithWrap(err).WithDebug(err.Error())
}
