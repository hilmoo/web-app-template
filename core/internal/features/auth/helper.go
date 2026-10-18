package auth

import (
	"net"
	"net/http"
	modelsv1 "proto/models/v1"

	"github.com/matthewhartstonge/argon2"
	"google.golang.org/protobuf/types/known/timestamppb"

	db "web-app-template-core/internal/db/sqlc"
	"web-app-template-core/internal/helper"
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
		MaxAge:   int((sessionDuration).Seconds()),
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
