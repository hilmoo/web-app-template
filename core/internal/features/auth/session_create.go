package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"time"

	db "web-app-template-core/internal/gen/sqlc"

	"github.com/google/uuid"
	"github.com/hilmoo/gomox/sqlx"
	"github.com/ory/herodot"
)

const (
	// sessionDuration is how long a newly created session remains valid.
	sessionDuration = 7 * 24 * time.Hour
	// randomTokenBytes is the number of random bytes used to generate a session token.
	randomTokenBytes = 32
)

func hashSessionToken(secret, token string) string {
	hashToken := hmac.New(sha256.New, []byte(secret))
	hashToken.Write([]byte(token))
	return base64.URLEncoding.EncodeToString(hashToken.Sum(nil))
}

type createSessionParams struct {
	queries   *db.Queries
	secret    string
	userID    uuid.UUID
	iPAddress *string
	userAgent *string
}

func createSession(ctx context.Context, args createSessionParams) (string, *herodot.DefaultError) {
	b := make([]byte, randomTokenBytes)
	_, err := rand.Read(b)
	if err != nil {
		return "", herodot.ErrInternalServerError.WithReason("failed to generate session token").WithDebug(err.Error())
	}

	token := base64.URLEncoding.EncodeToString(b)

	err = args.queries.SessionCreateNew(ctx, db.SessionCreateNewParams{
		UserID:    args.userID,
		TokenHash: hashSessionToken(args.secret, token),
		ExpiresAt: time.Now().Add(sessionDuration),
		IpAddress: args.iPAddress,
		UserAgent: args.userAgent,
	})
	return token, sqlx.MapPgxErrorHTTP(err, "session")
}
