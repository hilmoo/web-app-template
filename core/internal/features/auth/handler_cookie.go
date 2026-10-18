package auth

import (
	"net"
	"net/http"

	"web-app-template-core/internal/helper"
)

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
