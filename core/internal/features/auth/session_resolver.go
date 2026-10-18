package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	db "web-app-template-core/internal/db/sqlc"
)

// SessionResolver looks up the user behind a session token. It satisfies
// helper.SessionResolver; internal/app passes it to the auth interceptor.
type SessionResolver struct {
	queries *db.Queries
	secret  string
}

// NewSessionResolver returns a SessionResolver that hashes tokens with secret.
func NewSessionResolver(queries *db.Queries, secret string) *SessionResolver {
	return &SessionResolver{queries: queries, secret: secret}
}

// UserIDByToken returns the owner of the unexpired session identified by token.
func (r *SessionResolver) UserIDByToken(ctx context.Context, token string) (uuid.UUID, error) {
	sess, err := r.queries.SessionGetByTokenHash(ctx, hashSessionToken(r.secret, token))
	if err != nil {
		return uuid.Nil, fmt.Errorf("get session: %w", err)
	}

	return sess.UserID, nil
}
