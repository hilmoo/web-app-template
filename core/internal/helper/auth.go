package helper

import (
	"context"
	"crypto/subtle"
	"maps"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	errort "github.com/hilmoo/gomox/transport/error"
	"github.com/ory/herodot"
)

const (
	// SessionCookieName is the cookie that carries the user session token.
	SessionCookieName = "session"
	// ServiceUserIDHeader lets a service caller act on behalf of a user. It is only
	// honoured together with valid service credentials.
	ServiceUserIDHeader = "X-User-Id"
)

// Access is the credential level a procedure requires.
type Access int

const (
	// AccessAuthenticated requires either a user session cookie or the service secret.
	AccessAuthenticated Access = iota
	// AccessPublic requires no credentials.
	AccessPublic
	// AccessService requires the service secret; user sessions are rejected.
	AccessService
)

// AccessPolicy overrides the default (AccessAuthenticated) per procedure. Procedures that
// are not listed require authentication, so new services are protected by default.
type AccessPolicy map[string]Access

// Merge returns a new policy holding the entries of p and all of others.
func (p AccessPolicy) Merge(others ...AccessPolicy) AccessPolicy {
	out := make(AccessPolicy, len(p))
	maps.Copy(out, p)
	for _, o := range others {
		maps.Copy(out, o)
	}
	return out
}

type principalCtxKey struct{}

// Principal is the authenticated caller attached to the request context by the interceptor.
type Principal struct {
	// Service is true when the caller presented the service secret.
	Service bool
	// UserID is the user session's owner, or for service callers the user named in
	// ServiceUserIDHeader. It is uuid.Nil when there is no user.
	UserID uuid.UUID
}

// PrincipalFromContext returns the caller authenticated by the interceptor, if any.
func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalCtxKey{}).(Principal)
	return p, ok
}

// UserIDFromContext returns the user the request acts as: the session's user, or for
// service callers the user passed in ServiceUserIDHeader. It fails when there is none.
func UserIDFromContext(ctx context.Context) (uuid.UUID, *herodot.DefaultError) {
	p, ok := PrincipalFromContext(ctx)
	if !ok || p.UserID == uuid.Nil {
		return uuid.Nil, herodot.ErrUnauthorized.WithReason("user session required")
	}
	return p.UserID, nil
}

// SessionResolver maps a raw session token to the id of its user. It is what the
// interceptor needs from whichever feature owns sessions; internal/app wires it in.
// Any error means the token is not a valid, unexpired session.
type SessionResolver interface {
	UserIDByToken(ctx context.Context, token string) (uuid.UUID, error)
}

type authInterceptor struct {
	sessions      SessionResolver
	serviceSecret string
	policy        AccessPolicy
}

// NewAuthInterceptor returns a Connect interceptor that authenticates every call. Users
// authenticate with the session cookie; services with `Authorization: Bearer <SERVICE_SECRET>`
// and may add `X-User-Id: <uuid>` to act as a user.
// Per-procedure requirements come from policy; unlisted procedures need authentication.
func NewAuthInterceptor(arg HandlerParams, sessions SessionResolver, policy ...AccessPolicy) connect.Interceptor {
	p := make(AccessPolicy)
	for _, v := range policy {
		maps.Copy(p, v)
	}
	return &authInterceptor{
		sessions:      sessions,
		serviceSecret: arg.Config.ServiceSecret,
		policy:        p,
	}
}

func (i *authInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		ctx, err := i.authenticate(ctx, req.Spec().Procedure, req.Header())
		if err != nil {
			return nil, err
		}
		return next(ctx, req)
	}
}

func (i *authInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (i *authInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		ctx, err := i.authenticate(ctx, conn.Spec().Procedure, conn.RequestHeader())
		if err != nil {
			return err
		}
		return next(ctx, conn)
	}
}

func (i *authInterceptor) authenticate(
	ctx context.Context,
	procedure string,
	header http.Header,
) (context.Context, error) {
	policy := i.policy[procedure]

	principal, errH := i.resolve(ctx, header)
	if errH != nil {
		// Credentials on public procedures are ignored, so a stale cookie can't block login.
		if policy == AccessPublic {
			return ctx, nil
		}
		return ctx, errort.ConnectError(ctx, errH)
	}

	switch policy {
	case AccessPublic:
	case AccessService:
		if !principal.Service {
			return ctx, errort.ConnectError(ctx, herodot.ErrForbidden.WithReason("service credentials required"))
		}
	case AccessAuthenticated:
		if !principal.Service && principal.UserID == uuid.Nil {
			return ctx, errort.ConnectError(ctx, herodot.ErrUnauthorized.WithReason("authentication required"))
		}
	}

	return context.WithValue(ctx, principalCtxKey{}, principal), nil
}

// resolve identifies the caller. No credentials yields an empty Principal and no error.
func (i *authInterceptor) resolve(ctx context.Context, header http.Header) (Principal, *herodot.DefaultError) {
	if bearer, ok := bearerToken(header); ok {
		if i.serviceSecret == "" || subtle.ConstantTimeCompare([]byte(bearer), []byte(i.serviceSecret)) != 1 {
			return Principal{}, herodot.ErrUnauthorized.WithReason("invalid service credentials")
		}
		principal := Principal{Service: true, UserID: uuid.Nil}
		if v := header.Get(ServiceUserIDHeader); v != "" {
			userID, err := uuid.Parse(v)
			if err != nil {
				return Principal{}, herodot.ErrBadRequest.WithReasonf("invalid %s header", ServiceUserIDHeader)
			}
			principal.UserID = userID
		}
		return principal, nil
	}

	token, errH := SessionTokenFromHeader(header)
	if errH != nil {
		return Principal{}, nil //nolint:nilerr // no credentials; the access policy decides
	}
	userID, err := i.sessions.UserIDByToken(ctx, token)
	if err != nil {
		return Principal{}, herodot.ErrUnauthorized.WithReason("authentication required").WithDebug(err.Error())
	}
	return Principal{UserID: userID, Service: false}, nil
}

func bearerToken(h http.Header) (string, bool) {
	const prefix = "Bearer "
	v := h.Get("Authorization")
	if len(v) <= len(prefix) || !strings.EqualFold(v[:len(prefix)], prefix) {
		return "", false
	}
	return strings.TrimSpace(v[len(prefix):]), true
}

// SessionTokenFromHeader returns the raw session token from the request's Cookie header.
func SessionTokenFromHeader(h http.Header) (string, *herodot.DefaultError) {
	req := http.Request{Header: h}
	cookie, err := req.Cookie(SessionCookieName)
	if err != nil || cookie.Value == "" {
		return "", herodot.ErrUnauthorized.WithReason("authentication required").WithDebug("missing session cookie")
	}
	return cookie.Value, nil
}
