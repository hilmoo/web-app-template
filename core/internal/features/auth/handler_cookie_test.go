package auth //nolint:testpackage // exercises unexported helpers

import (
	"net/http"
	"strings"
	"testing"

	"web-app-template-core/internal/helper"
)

func TestPeerIP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		addr string
		want string // empty means nil
	}{
		{"203.0.113.7:51234", "203.0.113.7"},
		{"[2001:db8::1]:443", "2001:db8::1"},
		{"203.0.113.7", "203.0.113.7"},
		{"", ""},
	}
	for _, tt := range tests {
		got := peerIP(tt.addr)
		switch {
		case tt.want == "" && got != nil:
			t.Errorf("peerIP(%q) = %q, want nil", tt.addr, *got)
		case tt.want != "" && (got == nil || *got != tt.want):
			t.Errorf("peerIP(%q) = %v, want %q", tt.addr, got, tt.want)
		}
	}
}

func TestHashPasswordArgon(t *testing.T) {
	t.Parallel()

	hash, err := hashPasswordArgon("correct horse battery staple")
	if err != nil {
		t.Fatalf("hashPasswordArgon() error = %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2") {
		t.Errorf("hash = %q, want an argon2 encoded string", hash)
	}
}

func TestSetSessionCookie(t *testing.T) {
	t.Parallel()

	for _, isProd := range []bool{false, true} {
		h := http.Header{}
		setSessionCookie(h, "token", isProd)

		cookies := (&http.Response{Header: h}).Cookies()
		if len(cookies) != 1 {
			t.Fatalf("got %d cookies, want 1", len(cookies))
		}
		c := cookies[0]
		if c.Name != helper.SessionCookieName || c.Value != "token" {
			t.Errorf("cookie = %s=%s, want %s=token", c.Name, c.Value, helper.SessionCookieName)
		}
		if !c.HttpOnly || c.Secure != isProd {
			t.Errorf("HttpOnly=%v Secure=%v, want true/%v", c.HttpOnly, c.Secure, isProd)
		}
	}
}
