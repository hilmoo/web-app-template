package auth

import (
	"context"
	"net"
	"net/http"
	modelsv1 "proto/models/v1"

	"github.com/google/uuid"
	msession "github.com/hilmoo/gomox/transport/middleware/session"
	"github.com/matthewhartstonge/argon2"
	"github.com/ory/herodot"
	"google.golang.org/protobuf/types/known/timestamppb"

	db "web-app-template-core/internal/gen/sqlc"
	"web-app-template-core/internal/helper"
)

const (
	sessionCookieAge = 7 * 24 * 60 * 60 // seconds; matches msession's session duration
)

func hashPasswordArgon(plain string) (string, error) {
	argon := argon2.DefaultConfig()
	hash, err := argon.HashEncoded([]byte(plain))
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func peerIP(addr string) *string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	if host == "" {
		return nil
	}
	return &host
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func setSessionCookie(h http.Header, token string, isProd bool) {
	h.Add("Set-Cookie", (&http.Cookie{ //nolint:gosec // we want to set the cookie
		Name:     helper.SessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   sessionCookieAge,
		HttpOnly: true,
		Secure:   isProd,
		SameSite: http.SameSiteLaxMode,
	}).String())
}

func clearSessionCookie(h http.Header, isProd bool) {
	h.Add("Set-Cookie", (&http.Cookie{ //nolint:gosec // we want to clear the cookie
		Name:     helper.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isProd,
		SameSite: http.SameSiteLaxMode,
	}).String())
}

func toMeUser(u db.User) *modelsv1.AuthMeUser {
	return &modelsv1.AuthMeUser{
		Id:        u.ID.String(),
		Email:     u.Email,
		FullName:  u.FullName,
		CreatedAt: timestamppb.New(u.CreatedAt),
	}
}

func createUserSession(
	ctx context.Context,
	queries *db.Queries,
	secret string,
	userID uuid.UUID,
	ipAddress *string,
	userAgent *string,
) (string, *herodot.DefaultError) {
	return msession.CreateSession(ctx, msession.CreateSessionParams{
		Secret:    secret,
		UserID:    userID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		CreateSession: func(ctx context.Context, params msession.NewSessionParams) *herodot.DefaultError {
			// params.Token is already the hashed token (see msession.CreateSession).
			err := queries.SessionCreateNew(ctx, db.SessionCreateNewParams{
				UserID:    params.UserID,
				TokenHash: params.Token,
				ExpiresAt: params.ExpiresAt,
				IpAddress: params.IPAddress,
				UserAgent: params.UserAgent,
			})
			if err != nil {
				return herodot.ErrInternalServerError.WithReason("failed to create session").WithDebug(err.Error())
			}
			return nil
		},
	})
}
